package api

import (
	"encoding/base64"
	"fmt"
)

// ========== 歌词接口（lyrics.kugou.com） ==========
// 流程：
//  1. GET https://lyrics.kugou.com/v1/search  按 hash/keyword 查歌词 id + accesskey
//  2. GET https://lyrics.kugou.com/download   用 id/accesskey 下载歌词（fmt=lrc 返回 base64 文本，fmt=krc 返回 krc18 加密）

// 歌词搜索结果
type LyricSearchCandidate struct {
	ID        string `json:"id"`
	AccessKey string `json:"accesskey"`
	Singer    string `json:"singer"`
	Song      string `json:"song"`
	Duration  int64  `json:"duration"`
	ContentType int  `json:"contenttype"`
	Krctype   int    `json:"krctype"`
}

// 查询歌词信息（v1/search）
// hash 与 keyword 至少提供一个；albumAudioID 可选（错误值会导致 404，谨慎使用）
func GetLyricSearch(hash, keyword, albumAudioID string, creds *Credentials) *APIResponse {
	params := map[string]string{
		"duration": "0",
		"lrctxt":   "1",
		"man":      "no",
	}
	if hash != "" {
		params["hash"] = hash
	}
	if keyword != "" {
		params["keyword"] = keyword
	}
	if albumAudioID != "" {
		params["album_audio_id"] = albumAudioID
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://lyrics.kugou.com",
		URL:         "/v1/search",
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 解析歌词搜索结果，返回候选歌词（优先下载数最高的）
func ParseLyricSearch(resp *APIResponse) *LyricSearchCandidate {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	candidates, ok := resp.Data["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return nil
	}

	// 取第一个候选
	cand, ok := candidates[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &LyricSearchCandidate{}
	if id, ok := cand["id"]; ok {
		result.ID = fmt.Sprintf("%v", id)
	}
	if ak, ok := cand["accesskey"]; ok {
		result.AccessKey = fmt.Sprintf("%v", ak)
	}
	if singer, ok := cand["singer"]; ok {
		result.Singer = fmt.Sprintf("%v", singer)
	}
	if song, ok := cand["song"]; ok {
		result.Song = fmt.Sprintf("%v", song)
	}
	if dur, ok := cand["duration"]; ok {
		result.Duration = int64(toFloat64(dur))
	}
	if ct, ok := cand["contenttype"]; ok {
		result.ContentType = int(toFloat64(ct))
	}

	if result.ID == "" || result.AccessKey == "" {
		return nil
	}
	return result
}

// 下载歌词内容（/download）
// fmt: "lrc"（base64 纯文本）或 "krc"（krc18 加密格式）
func GetLyricContent(id, accessKey, format string, creds *Credentials) *APIResponse {
	params := map[string]string{
		"ver":       "1",
		"client":    "android",
		"id":        id,
		"accesskey": accessKey,
		"fmt":       format,
		"charset":   "utf8",
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://lyrics.kugou.com",
		URL:         "/download",
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 解析歌词下载响应，返回解码后的歌词文本
// 返回 (文本, 歌词格式)
func ParseLyricContent(resp *APIResponse) (string, string) {
	if resp.Error != nil || resp.Data == nil {
		return "", ""
	}

	var content string
	if c, ok := resp.Data["content"]; ok {
		content = fmt.Sprintf("%v", c)
	}
	if content == "" {
		return "", ""
	}

	// 歌词格式：fmt 字段决定解码方式
	fmt_, _ := resp.Data["fmt"].(string)
	contenttype := int(toFloat64(resp.Data["contenttype"]))

	raw, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", ""
	}

	// lrc：直接是 base64 文本；krc：contenttype=0 时是 krc18 加密格式
	if fmt_ == "lrc" || contenttype != 0 {
		return string(raw), "lrc"
	}

	text, err := DecodeKRC(raw)
	if err != nil {
		return "", ""
	}
	return text, "krc"
}

// ========== 兼容旧接口 ==========

// 获取歌词（兼容旧入口）
// 先按 hash 搜索歌词拿 id/accesskey，再下载；优先 lrc，缺失则 krc
func GetLyric(hash, albumID, timelength string, creds *Credentials) *APIResponse {
	return getLyricCommon(hash, "", albumID, creds)
}

// 按关键词搜索并下载歌词（用于仅 hash 搜索失败时的回退）
func GetLyricByKeyword(hash, keyword, albumID, timelength string, creds *Credentials) *APIResponse {
	return getLyricCommon(hash, keyword, albumID, creds)
}

// 歌词获取公共逻辑：搜索 → 下载（lrc/krc 依次尝试）
func getLyricCommon(hash, keyword, albumID string, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}

	// 注意：albumID 是专辑 ID，不能作为 album_audio_id 传给歌词搜索（会导致 404）
	searchResp := GetLyricSearch(hash, keyword, "", creds)
	if searchResp.Error != nil {
		return searchResp
	}
	cand := ParseLyricSearch(searchResp)
	if cand == nil {
		return &APIResponse{Error: fmt.Errorf("未找到歌词")}
	}

	for _, format := range []string{"lrc", "krc"} {
		dlResp := GetLyricContent(cand.ID, cand.AccessKey, format, creds)
		if dlResp.Error != nil {
			return dlResp
		}
		text, _ := ParseLyricContent(dlResp)
		if text != "" {
			dlResp.Data["content"] = text
			dlResp.Data["format"] = format
			return dlResp
		}
	}

	return &APIResponse{Error: fmt.Errorf("歌词下载失败")}
}

// 解析歌词响应（兼容旧入口）
func ParseLyric(resp *APIResponse) *LyricResult {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	result := &LyricResult{}
	if content, ok := resp.Data["content"]; ok {
		result.Content = fmt.Sprintf("%v", content)
	}
	if t, ok := resp.Data["title"]; ok {
		result.Title = fmt.Sprintf("%v", t)
	}
	if a, ok := resp.Data["author"]; ok {
		result.Author = fmt.Sprintf("%v", a)
	}
	if al, ok := resp.Data["album"]; ok {
		result.Album = fmt.Sprintf("%v", al)
	}
	if ts, ok := resp.Data["timestamp"]; ok {
		result.Time = fmt.Sprintf("%v", ts)
	}

	if result.Content == "" {
		return nil
	}
	return result
}

// ========== 歌词文本辅助 ==========

// 提取纯歌词文本（去除 KRC 时间标签等，保留原样兼容旧调用）
func ExtractLyricText(krcContent string) string {
	var lines []string
	var cur string
	for _, c := range krcContent {
		if c == '\n' {
			if cur != "" {
				lines = append(lines, cur)
			}
			cur = ""
		} else {
			cur += string(c)
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}

	var out []string
	for _, line := range lines {
		// 去除行首的 [tag] 时间标签，保留歌词文本
		text := stripKRCHeadTags(line)
		if text != "" {
			out = append(out, text)
		}
	}
	return joinLines(out)
}

func stripKRCHeadTags(line string) string {
	for {
		if len(line) < 2 || line[0] != '[' {
			break
		}
		idx := indexByte(line, ']')
		if idx < 0 {
			break
		}
		line = line[idx+1:]
	}
	return line
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func joinLines(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n"
		}
		out += l
	}
	return out
}

