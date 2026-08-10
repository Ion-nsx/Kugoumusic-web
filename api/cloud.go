package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 云盘歌曲文件信息
type CloudFile struct {
	KVID           int64  `json:"kv_id"`
	AlbumAudioID   int64  `json:"album_audio_id"`
	AudioID        int64  `json:"audio_id"`
	Hash           string `json:"hash"`
	Name           string `json:"name"`
	AuthorName     string `json:"author_name"`
	Ext            string `json:"ext"`
	Size           int64  `json:"size"`
	TimeLen        int64  `json:"timelen"`
	HashStd        string `json:"hash_std"`
	Bitrate        int    `json:"bitrate"`
	SongName       string `json:"song_name"`
	SingerName     string `json:"singer_name"`
	AlbumName      string `json:"album_name"`
	AlbumID        string `json:"album_id"`
	Cover          string `json:"cover"`
	Flag           int    `json:"flag"`
}

// ===== 云盘列表 =====
// POST https://mcloudservice.kugou.com/v1/get_list
func GetCloudList(userid, token string, page, pageSize int) *APIResponse {
	if userid == "" || token == "" {
		return &APIResponse{Error: fmt.Errorf("请先登录")}
	}

	clienttime := time.Now().Unix()

	dataMap := map[string]interface{}{
		"page":     page,
		"pagesize": pageSize,
		"getkmr":   1,
	}

	aesResult, err := PlaylistAESEncrypt(dataMap)
	if err != nil {
		return &APIResponse{Error: err}
	}

	p, err := RSAEncrypt2(map[string]interface{}{
		"aes":   aesResult.Key,
		"uid":   userid,
		"token": token,
	}, UseLite)
	if err != nil {
		return &APIResponse{Error: err}
	}

	bodyBytes, err := base64.StdEncoding.DecodeString(aesResult.Str)
	if err != nil {
		return &APIResponse{Error: err}
	}

	params := map[string]string{
		"clienttime": strconv.FormatInt(clienttime, 10),
		"mid":        GetMID(),
		"key":        signParamsKey(strconv.FormatInt(clienttime, 10)),
		"clientver":  strconv.Itoa(ClientVer),
		"appid":      strconv.Itoa(AppID),
		"p":          strings.ToUpper(p),
	}

	resp := SendRequest(&RequestOptions{
		Method:       "POST",
		BaseURL:      "https://mcloudservice.kugou.com",
		URL:          "/v1/get_list",
		Params:       params,
		RawBodyBytes: bodyBytes,
		EncryptType:  "android",
		ClearDefault: true,
		NoSign:       true,
	})
	if resp.Error != nil {
		return resp
	}

	// 响应是 AES 加密的 binary，需要用原 key 解密
	plainBytes, err := PlaylistAESDecrypt(base64.StdEncoding.EncodeToString(resp.Body), aesResult.Key)
	if err != nil {
		return resp
	}

	var decrypted map[string]interface{}
	if err := json.Unmarshal(plainBytes, &decrypted); err != nil {
		return &APIResponse{Error: err}
	}
	resp.Data = decrypted

	// 对没有封面的文件做批量曲库匹配，补充封面数据
	if d, ok := decrypted["data"].(map[string]interface{}); ok {
		listRaw, _ := d["list"].([]interface{})
		if len(listRaw) > 0 {
			var unmatchedHashes []string
			needMatch := map[string]int{} // hash → list index
			for i, item := range listRaw {
				m, _ := item.(map[string]interface{})
				if m == nil {
					continue
				}
				// 已有 album_info 或 authors 的跳过
				if _, hasAlbumInfo := m["album_info"]; hasAlbumInfo {
					continue
				}
				if _, hasAuthors := m["authors"]; hasAuthors {
					continue
				}
				h, _ := m["hash"].(string)
				if h != "" {
					unmatchedHashes = append(unmatchedHashes, h)
					needMatch[h] = i
				}
			}

			if len(unmatchedHashes) > 0 {
				matches, err := GetCloudMatchBatch(unmatchedHashes)
				if err == nil {
					for hash, info := range matches {
						if info.SizableCover == "" {
							continue
						}
						idx, ok := needMatch[hash]
						if !ok || idx >= len(listRaw) {
							continue
						}
						item, _ := listRaw[idx].(map[string]interface{})
						if item == nil {
							continue
						}
						item["album_info"] = map[string]interface{}{
							"sizable_cover": info.SizableCover,
						}
					}
				}
			}
		}
	}

	return resp
}

// ===== 云盘删除 =====
// POST https://mcloudservice.kugou.com/v1/del_files
func DeleteCloudFile(userid, token string, kvIDs []int64, albumAudioIDs []int64) *APIResponse {
	if userid == "" || token == "" {
		return &APIResponse{Error: fmt.Errorf("请先登录")}
	}
	if len(kvIDs) == 0 {
		return &APIResponse{Error: fmt.Errorf("请提供文件ID")}
	}

	clienttime := time.Now().Unix()

	type delItem struct {
		KVID         int64 `json:"kv_id"`
		AlbumAudioID int64 `json:"album_audio_id"`
	}
	items := make([]delItem, len(kvIDs))
	for i, kvID := range kvIDs {
		aaID := int64(0)
		if i < len(albumAudioIDs) {
			aaID = albumAudioIDs[i]
		} else if len(albumAudioIDs) > 0 {
			aaID = albumAudioIDs[0]
		}
		items[i] = delItem{KVID: kvID, AlbumAudioID: aaID}
	}

	aesResult, err := PlaylistAESEncrypt(map[string]interface{}{
		"data": items,
	})
	if err != nil {
		return &APIResponse{Error: err}
	}

	p, err := RSAEncrypt2(map[string]interface{}{
		"aes":   aesResult.Key,
		"uid":   userid,
		"token": token,
	}, UseLite)
	if err != nil {
		return &APIResponse{Error: err}
	}

	bodyBytes, err := base64.StdEncoding.DecodeString(aesResult.Str)
	if err != nil {
		return &APIResponse{Error: err}
	}

	params := map[string]string{
		"clienttime": strconv.FormatInt(clienttime, 10),
		"mid":        GetMID(),
		"key":        signParamsKey(strconv.FormatInt(clienttime, 10)),
		"clientver":  strconv.Itoa(ClientVer),
		"appid":      strconv.Itoa(AppID),
		"p":          strings.ToUpper(p),
	}

	resp := SendRequest(&RequestOptions{
		Method:       "POST",
		BaseURL:      "https://mcloudservice.kugou.com",
		URL:          "/v1/del_files",
		Params:       params,
		RawBodyBytes: bodyBytes,
		EncryptType:  "android",
		ClearDefault: true,
		NoSign:       true,
	})
	if resp.Error != nil {
		return resp
	}

	// 响应可能也是 AES 加密的
	plainBytes, err := PlaylistAESDecrypt(base64.StdEncoding.EncodeToString(resp.Body), aesResult.Key)
	if err == nil {
		var decrypted map[string]interface{}
		if err := json.Unmarshal(plainBytes, &decrypted); err == nil {
			resp.Data = decrypted
		}
	}

	return resp
}

// ===== 云盘播放 URL =====
// GET gateway.kugou.com/bsstrackercdngz/v2/query_musicclound_url
func GetCloudURL(hash string, albumAudioID, audioID int64, name string) *APIResponse {
	if hash == "" {
		return &APIResponse{Error: fmt.Errorf("hash 参数必填")}
	}

	hash = strings.ToLower(hash)

	params := map[string]string{
		"hash":           hash,
		"ssa_flag":       "is_fromtrack",
		"version":        "20102",
		"ssl":            "0",
		"album_audio_id": strconv.FormatInt(albumAudioID, 10),
		"pid":            "20026",
		"audio_id":       strconv.FormatInt(audioID, 10),
		"kv_id":          "2",
		"key":            signCloudKey(hash, 20026),
		"bucket":         "musicclound",
		"name":           name,
		"with_res_tag":   "0",
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/bsstrackercdngz/v2/query_musicclound_url",
		Params:      params,
		EncryptType: "android",
	})
}
