package api

// 排行榜列表（新接口）
// GET https://gateway.kugou.com/ocean/v6/rank/list
// 响应 data.info[] 含 id/rankname/img_9 等
func GetRankList(creds *Credentials) *APIResponse {
	params := map[string]string{
		"plat":     "2",
		"withsong": "1",
		"parentid": "0",
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/ocean/v6/rank/list",
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 排行榜歌曲（新接口）
// POST https://gateway.kugou.com/openapi/kmr/v2/rank/audio（kg-tid: 369）
// 响应 data.songlist[] 为歌曲数组
func GetRankAudio(rankID string, page, pagesize int, creds *Credentials) *APIResponse {
	dataMap := map[string]interface{}{
		"show_portrait_mv":       1,
		"show_type_total":        1,
		"filter_original_remarks": 1,
		"area_code":              1,
		"pagesize":               pagesize,
		"rank_cid":               0,
		"type":                   1,
		"page":                   page,
		"rank_id":                rankID,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/openapi/kmr/v2/rank/audio",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"kg-tid": "369"},
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 解析排行榜列表响应（ocean/v6/rank/list）
func ParseRankList(resp *APIResponse) []map[string]interface{} {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}
	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}
		if info, ok := dataMap["info"].([]interface{}); ok {
			out := make([]map[string]interface{}, 0, len(info))
			for _, item := range info {
				if m, ok := item.(map[string]interface{}); ok {
					out = append(out, m)
				}
			}
			return out
		}
	}
	return nil
}

// 解析排行榜歌曲响应（openapi/kmr/v2/rank/audio）
func ParseRankAudio(resp *APIResponse) []Song {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}
	var songs []Song
	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}
		if songlist, ok := dataMap["songlist"].([]interface{}); ok {
			for _, item := range songlist {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				song := parseSongFromTrack(itemMap)
				if song == nil {
					continue
				}
				// 补充哈希：排行榜歌曲 hash 在 deprecated.hash
				if song.ID == "" {
					if dep, ok := itemMap["deprecated"].(map[string]interface{}); ok {
						if h, ok := dep["hash"]; ok {
							song.ID = playlistIDString(h)
						}
					}
				}
				songs = append(songs, *song)
			}
		}
	}
	return songs
}
