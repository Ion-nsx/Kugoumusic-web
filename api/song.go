package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 获取歌曲播放地址
// GET https://gateway.kugou.com/v5/url（x-router: trackercdn.kugou.com）
// 参数: hash + 播放上下文参数 + key(signKey) + signature(android)
// 免费歌曲返回 status:1 + url/backupUrl 数组；VIP/版权歌曲返回 status:2 无 url
func GetSongURL(hash, albumID string, bitrate int, quality string, creds *Credentials) *APIResponse {
	hash = strings.ToLower(strings.TrimSpace(hash))
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}
	dfid := creds.DFID
	if dfid == "" {
		dfid = GetDFID()
	}
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}

	// 音质：v5/url 接受 128/192/320/flac/high/super/multitrack 等。
	// 优先使用显式 quality 参数，否则按 bitrate 映射，默认 128。
	qualityStr := strings.TrimSpace(quality)
	switch qualityStr {
	case "128", "192", "320", "flac", "high":
	default:
		qualityStr = ""
	}
	if qualityStr == "" {
		qualityStr = "128"
		if bitrate >= 320 {
			qualityStr = "320"
		} else if bitrate >= 192 {
			qualityStr = "192"
		}
	}

	params := map[string]string{
		"album_id":       albumID,
		"area_code":      "1",
		"hash":           hash,
		"ssa_flag":       "is_fromtrack",
		"version":        strconv.Itoa(ClientVer),
		"page_id":        "967177915", // 概念版(lite) page_id
		"quality":        qualityStr,
		"album_audio_id": "0",
		"behavior":       "play",
		"pid":            "411", // 概念版(lite) pid
		"cmd":            "26",
		"pidversion":     "3001",
		"IsFreePart":     "0",
		"ppage_id":       "356753938,823673182,967485191", // 概念版(lite) ppage_id
		"cdnBackup":      "1",
		"kcard":          "0",
		"module":         "",
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v5/url",
		Params:      params,
		Headers:     map[string]string{"x-router": "trackercdn.kugou.com"},
		EncryptType: "android",
		EncryptKey:  true,
		KeyHash:     hash,
		MID:         mid,
		DFID:        dfid,
		UserID:      userid,
		Token:       creds.Token, // 关键：必须带 token，VIP 歌才能解锁（否则返回 status:2）
	})
}

// VIP 歌曲播放地址（对齐 MoeKoeMusic song_url_new）
// POST http://tracker.kugou.com/v6/priv_url
// v5/url 对 VIP/版权歌返回 status:2 无地址，VIP 播放必须走 priv_url。
// 概念版(lite) VIP 需概念版(lite) token（扫码登录是标准版 token，VIP 校验不通过）。
// tracker_param.key = MD5(hash + lite盐 + appid + mid + userid)
func GetPrivURL(hash string, creds *Credentials, vipToken, vipType string) *APIResponse {
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}
	dfid := creds.DFID
	if dfid == "" {
		dfid = GetDFID()
	}
	clienttimeMS := time.Now().UnixMilli()

	key := MD5(hash + LiteSignKeySalt + strconv.Itoa(LiteAppIDInt) + mid + userid)

	dataMap := map[string]interface{}{
		"area_code": "1",
		"behavior":  "play",
		"qualities": []string{"128", "320", "flac", "high", "multitrack", "viper_atmos", "viper_tape", "viper_clear", "super"},
		"resource": map[string]interface{}{
			"collect_list_id": "3",
			"collect_time":    clienttimeMS,
			"hash":            hash,
			"id":              0,
			"page_id":         1,
			"type":            "audio",
		},
		"token": creds.Token,
		"tracker_param": map[string]interface{}{
			"all_m": 1, "auth": "", "is_free_part": 0,
			"key": key, "module_id": 0, "need_climax": 0, "need_xcdn": 1,
			"open_time": "", "pid": "411", "pidversion": "3001",
			"priv_vip_type": "6", "viptoken": vipToken,
		},
		"userid": userid,
		"vip":    vipType,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "http://tracker.kugou.com",
		URL:         "/v6/priv_url",
		Data:        dataMap,
		EncryptType: "android",
		MID:         mid,
		DFID:        dfid,
		UserID:      userid,
		Token:       creds.Token,
	})
}

// 解析 priv_url 响应，返回可播放地址
// 结构: status:1, data[]（每个元素一种音质）: { hash, name, info: { bitrate, url: [..], .. }, quality }
// quality 为要选用的音质（128/320/flac/high 等），匹配 info.quality 或 bitrate
func ParsePrivURL(resp *APIResponse, quality string) *SongURL {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}
	if st, ok := resp.Data["status"]; ok {
		if int(toFloat64(st)) != 1 {
			return nil
		}
	}

	songURL := &SongURL{}
	seen := map[string]bool{}
	hasURL := false
	quality = strings.TrimSpace(quality)

	// 音质优先级：请求音质 > 128 > 其他
	rank := func(q string) int {
		if quality != "" && q == quality {
			return 0
		}
		if q == "128" {
			return 1
		}
		return 2
	}
	bestRank := 99

	if data, ok := resp.Data["data"]; ok {
		if arr, ok := data.([]interface{}); ok {
			for _, x := range arr {
				m, ok := x.(map[string]interface{})
				if !ok {
					continue
				}
				q := ""
				if qv, ok := m["quality"]; ok {
					q = fmt.Sprintf("%v", qv)
				}
				info, _ := m["info"].(map[string]interface{})
				if info == nil {
					continue
				}
				if q == "" {
					if br, ok := info["bitrate"]; ok {
						if int(toFloat64(br)) >= 320 {
							q = "320"
						} else {
							q = "128"
						}
					}
				}
				urls, _ := info["url"].([]interface{})
				if len(urls) == 0 {
					continue
				}
				r := rank(q)
				if r > bestRank {
					continue
				}
				if r < bestRank {
					songURL.URL = ""
					songURL.Backup = nil
					seen = map[string]bool{}
					bestRank = r
				}
				for _, u := range urls {
					s := strings.TrimSpace(fmt.Sprintf("%v", u))
					if s == "" || seen[s] {
						continue
					}
					seen[s] = true
					hasURL = true
					if songURL.URL == "" {
						songURL.URL = s
					} else {
						songURL.Backup = append(songURL.Backup, s)
					}
				}
			}
		}
	}

	if !hasURL {
		return nil
	}
	return songURL
}

// 获取音质权限（新接口）
// POST https://gateway.kugou.com/v2/get_res_privilege/lite（x-router: media.store.kugou.com）
// 参数: hash 列表；响应 data 为数组（每条含 hash/pay_type/quality/status 等）
func GetPrivilegeLite(songIDs []string, creds *Credentials) *APIResponse {
	resource := make([]map[string]interface{}, 0, len(songIDs))
	for _, s := range songIDs {
		resource = append(resource, map[string]interface{}{"type": "audio", "page_id": 0, "hash": s, "album_id": 0})
	}

	dataMap := map[string]interface{}{
		"appid":           AppID,
		"area_code":       1,
		"behavior":        "play",
		"clientver":       ClientVer,
		"need_hash_offset": 1,
		"relate":          1,
		"support_verify":  1,
		"resource":        resource,
		"qualities":       []string{"128", "320", "flac", "high", "viper_atmos", "viper_tape", "viper_clear", "super", "multitrack"},
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v2/get_res_privilege/lite",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "media.store.kugou.com", "Content-Type": "application/json"},
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// 获取歌曲详情（已迁移）
// 原 mobilecdn /api/v3/song/detail 已失效（Access Deny），改用 kmr.service.kugou.com/v1/audio/audio
// POST body: appid/clienttime(毫秒)/clientver/data[{hash,audio_id}]/dfid/key(signParamsKey)/mid
// 响应 data 为数组，每项含 hash/hash_128/320/flac/audio_name/timelength/bitrate/filesize 等
func GetKmrAudio(hashes []string, creds *Credentials) *APIResponse {
	dateTime := time.Now().UnixMilli()
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}
	dfid := creds.DFID
	if dfid == "" {
		dfid = "-"
	}

	dataList := make([]map[string]interface{}, 0, len(hashes))
	for _, h := range hashes {
		dataList = append(dataList, map[string]interface{}{
			"hash":     h,
			"audio_id": 0,
		})
	}

	dataMap := map[string]interface{}{
		"appid":      AppID,
		"clienttime": dateTime,
		"clientver":  ClientVer,
		"data":       dataList,
		"dfid":       dfid,
		"key":        signParamsKey(strconv.FormatInt(dateTime, 10)),
		"mid":        mid,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "http://kmr.service.kugou.com",
		URL:         "/v1/audio/audio",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "kmr.service.kugou.com", "Content-Type": "application/json"},
		MID:         mid,
		DFID:        dfid,
	})
}

// 获取歌曲详情（单曲）
func GetSongDetail(hash string, creds *Credentials) *APIResponse {
	return GetKmrAudio([]string{strings.ToLower(strings.TrimSpace(hash))}, creds)
}

// 批量获取歌曲详情
func GetSongsDetail(hashes []string, creds *Credentials) *APIResponse {
	for i, h := range hashes {
		hashes[i] = strings.ToLower(strings.TrimSpace(h))
	}
	return GetKmrAudio(hashes, creds)
}

// 解析歌曲详情响应（kmr audio），返回第一个歌曲详情 map
// 兼容原 mobilecdn song/detail 的 data.info 结构
func ParseSongDetail(resp *APIResponse) map[string]interface{} {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	data, ok := resp.Data["data"].([]interface{})
	if !ok || len(data) == 0 {
		return nil
	}

	first, ok := data[0].(map[string]interface{})
	if !ok {
		return nil
	}

	// 转换为信息结构
	info := make(map[string]interface{})
	for k, v := range first {
		info[k] = v
	}

	return info
}

// 解析歌曲 URL 响应（v5/url 顶层直接返回 url/backupUrl 数组）
func ParseSongURL(resp *APIResponse) *SongURL {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	// status != 1 表示版权/VIP 限制或无权限，不返回 URL
	if st, ok := resp.Data["status"]; ok {
		if int(toFloat64(st)) != 1 {
			return nil
		}
	}

	songURL := &SongURL{}
	seen := map[string]bool{}

	// url 数组
	if urls, ok := resp.Data["url"]; ok {
		for _, u := range stringSlice(urls) {
			if u != "" && !seen[u] {
				seen[u] = true
				if songURL.URL == "" {
					songURL.URL = u
				} else {
					songURL.Backup = append(songURL.Backup, u)
				}
			}
		}
	}

	// backupUrl 数组
	if bks, ok := resp.Data["backupUrl"]; ok {
		for _, u := range stringSlice(bks) {
			if u != "" && !seen[u] {
				seen[u] = true
				if songURL.URL == "" {
					songURL.URL = u
				} else {
					songURL.Backup = append(songURL.Backup, u)
				}
			}
		}
	}

	if songURL.URL == "" {
		return nil
	}

	if br, ok := resp.Data["bitRate"]; ok {
		songURL.BitRate = int(toFloat64(br))
	}
	if ext, ok := resp.Data["extName"]; ok {
		songURL.Ext = fmt.Sprintf("%v", ext)
	}
	if size, ok := resp.Data["fileSize"]; ok {
		songURL.Size = int64(toFloat64(size))
	}
	if tl, ok := resp.Data["timeLength"]; ok {
		songURL.Duration = int64(toFloat64(tl))
	}
	if h, ok := resp.Data["hash"]; ok {
		songURL.Hash = fmt.Sprintf("%v", h)
	}
	if fn, ok := resp.Data["fileName"]; ok {
		songURL.FileName = fmt.Sprintf("%v", fn)
	}

	// 根据响应的实际 bitRate + extName 推断真实音质，而非请求时传的 quality 参数
	songURL.Quality = inferQualityFromResponse(songURL.BitRate, songURL.Ext)

	return songURL
}

// 根据 bitRate 和扩展名推断实际音质
func inferQualityFromResponse(bitRate int, ext string) string {
	if bitRate > 2000000 {
		return "high"
	}
	if ext == "flac" {
		return "flac"
	}
	if bitRate >= 320000 {
		return "320"
	}
	if bitRate >= 192000 {
		return "192"
	}
	return "128"
}

// 音质等级排序：数字越大音质越高，用于降级链判断是否接受当前结果
func QualityRank(q string) int {
	switch q {
	case "high":
		return 6
	case "flac":
		return 5
	case "multitrack":
		return 4
	case "super":
		return 3
	case "320":
		return 2
	case "192":
		return 1
	case "128":
		return 0
	default:
		return 0
	}
}

// stringSlice 将接口值（字符串或字符串数组）归一化为字符串切片
func stringSlice(v interface{}) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, item := range t {
			out = append(out, fmt.Sprintf("%v", item))
		}
		return out
	case []string:
		return t
	default:
		return nil
	}
}

// 解析音质权限响应（v2/get_res_privilege/lite）
// 结构: data 为数组，每条含 hash/pay_type/quality/status/info
func ParsePrivilegeLite(resp *APIResponse) map[string]*Privilege {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	result := make(map[string]*Privilege)

	if data, ok := resp.Data["data"]; ok {
		listArr, ok := data.([]interface{})
		if !ok {
			return nil
		}

		for _, item := range listArr {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			p := &Privilege{}
			hash := ""

			if h, ok := itemMap["hash"]; ok {
				hash = fmt.Sprintf("%v", h)
			}
			if q, ok := itemMap["quality"]; ok {
				p.Format = fmt.Sprintf("%v", q)
			}
			if pt, ok := itemMap["pay_type"]; ok {
				p.PayType = fmt.Sprintf("%v", pt)
			}
			if st, ok := itemMap["status"]; ok {
				p.CanPlay = int(toFloat64(st)) == 0
			}

			// info 子对象（bitrate/filesize/duration/extname/image）
			if info, ok := itemMap["info"].(map[string]interface{}); ok {
				if br, ok := info["bitrate"]; ok {
					p.BitRate = int(toFloat64(br))
				}
				if ext, ok := info["extname"]; ok {
					p.Extname = fmt.Sprintf("%v", ext)
				}
				if dur, ok := info["duration"]; ok {
					p.Duration = int64(toFloat64(dur)) / 1000
				}
				if img, ok := info["image"]; ok {
					p.Image = fmt.Sprintf("%v", img)
				}
			}

			if hash != "" {
				result[hash] = p
			}
		}
	}

	return result
}

// 解析歌曲所有可用音质（从 privilege_lite 响应中提取）
// 响应 data 数组每条为一个音质级别，从 quality 字段提取
func ParsePrivilegeQualities(resp *APIResponse) []string {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}
	qualities := make(map[string]bool)
	if data, ok := resp.Data["data"]; ok {
		listArr, ok := data.([]interface{})
		if !ok {
			return nil
		}
		for _, item := range listArr {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if q, ok := itemMap["quality"]; ok {
				qStr := fmt.Sprintf("%v", q)
				if qStr != "" && qStr != "0" {
					qualities[qStr] = true
				}
			}
		}
	}
	result := make([]string, 0, len(qualities))
	for _, q := range []string{"high", "flac", "super", "multitrack", "320", "192", "128"} {
		if qualities[q] {
			result = append(result, q)
		}
	}
	return result
}

func toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val == 1
	case float64:
		return val == 1
	case string:
		return val == "1" || val == "true"
	}
	return false
}

// 歌曲高潮片段 — 获取音频高潮区间
// GET https://expendablekmrcdn.kugou.com/v1/audio_climax/audio
// 参数: data=JSON([{"hash":"xxx"},{"hash":"yyy"}])
func GetSongClimax(hashes []string) *APIResponse {
	type climaxItem struct {
		Hash string `json:"hash"`
	}
	items := make([]climaxItem, len(hashes))
	for i, h := range hashes {
		items[i] = climaxItem{Hash: h}
	}

	js, _ := json.Marshal(items)

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://expendablekmrcdn.kugou.com",
		URL:         "/v1/audio_climax/audio",
		Params:      map[string]string{"data": string(js)},
		EncryptType: "android",
	})
}