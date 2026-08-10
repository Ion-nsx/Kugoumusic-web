package api

import (
	"strconv"
	"time"
)

// 新歌速递（新接口）
// POST https://gateway.kugou.com/musicadservice/container/v1/newsong_publish
// 响应 data 为歌曲数组
func GetTopSong(page, pagesize int, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}

	dataMap := map[string]interface{}{
		"rank_id":  21608,
		"userid":   userid,
		"page":     page,
		"pagesize": pagesize,
		"tags":     []string{},
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/musicadservice/container/v1/newsong_publish",
		Data:        dataMap,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      userid,
	})
}

// 新碟上架（新接口）
// POST https://gateway.kugou.com/musicadservice/v1/mobile_newalbum_sp
// 响应 data 含 chn 等字段
func GetTopAlbum(page, pagesize int, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}

	dataMap := map[string]interface{}{
		"apiver":    20,
		"token":     creds.Token,
		"page":      page,
		"pagesize":  pagesize,
		"withpriv":  1,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/musicadservice/v1/mobile_newalbum_sp",
		Data:        dataMap,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 热门好歌精选（新接口）
// POST https://gateway.kugou.com/singlecardrec.service/v1/single_card_recommend
// card_id: 1=精选好歌随心听 2=经典怀旧金曲 3=热门好歌精选 4=小众宝藏佳作 6=vip专属推荐
// 响应 data.song_list[] 为歌曲数组
func GetTopCard(cardID string, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}
	if cardID == "" {
		cardID = "1"
	}

	dateMS := time.Now().UnixMilli()
	fakem := "60f7ebf1f812edbac3c63a7310001701760f"

	dataMap := map[string]interface{}{
		"appid":           AppID,
		"clientver":       ClientVer,
		"platform":        "android",
		"clienttime":      dateMS,
		"userid":          userid,
		"key":             signParamsKey(strconv.FormatInt(dateMS, 10)),
		"fakem":           fakem,
		"area_code":       1,
		"mid":             mid,
		"uuid":            "-",
		"client_playlist": []string{},
		"u_info":          "a0c35cd40af564444b5584c2754dedec",
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/singlecardrec.service/v1/single_card_recommend",
		Data:        dataMap,
		Params:      map[string]string{"card_id": cardID, "fakem": fakem, "area_code": "1", "platform": "ios"},
		EncryptType: "android",
		MID:         mid,
		DFID:        creds.DFID,
		UserID:      userid,
	})
}

// 歌单分类（新接口）
// POST https://gateway.kugou.com/pubsongs/v1/get_tags_by_type
// 响应 data 为分类数组
func GetPlaylistTags(creds *Credentials) *APIResponse {
	dataMap := map[string]interface{}{
		"tag_type": "collection",
		"tag_id":   0,
		"source":   3,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/pubsongs/v1/get_tags_by_type",
		Data:        dataMap,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 解析新歌速递响应（newsong_publish）：data 为歌曲数组
func ParseTopSongs(resp *APIResponse) []Song {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}
	var songs []Song
	if data, ok := resp.Data["data"]; ok {
		list, ok := data.([]interface{})
		if !ok {
			return nil
		}
		for _, item := range list {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			song := parseSongFromTrack(itemMap)
			if song == nil {
				continue
			}
			// 哈希字段：新歌接口为 hash 或 songname_original 等
			if song.ID == "" {
				if h, ok := itemMap["hash"]; ok {
					song.ID = playlistIDString(h)
				}
			}
			songs = append(songs, *song)
		}
	}
	return songs
}
