package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 每日推荐（新接口）
// POST https://gateway.kugou.com/everyday_song_recommend（x-router: everydayrec.service.kugou.com）
// 响应 data 含 song_list（30首每日推荐歌曲）
func GetEverydayRecommend(creds *Credentials) *APIResponse {
	params := map[string]string{"platform": "ios"}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/everyday_song_recommend",
		Params:      params,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "everydayrec.service.kugou.com"},
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// 个性推荐
// GET http://mobilecdn.kugou.com/api/v3/recommend
// 参数: userid, token
func GetRecommend(creds *Credentials) *APIResponse {
	params := creds.InjectParams(map[string]string{
		"plat":    "0",
		"version": "1",
	})

	return SendRequest(&RequestOptions{
		Method: "GET",
		URL:    "http://mobilecdn.kugou.com/api/v3/recommend",
		Params: params,
	})
}

// 推荐歌单
// POST https://gateway.kugou.com/v2/special_recommend
// body 含 signParamsKey；响应 data.special_list 为推荐歌单列表
func GetRecommendPlaylist(creds *Credentials) *APIResponse {
	dateTime := time.Now().Unix()
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}

	dataMap := map[string]interface{}{
		"appid":        AppID,
		"mid":          mid,
		"clientver":    ClientVer,
		"platform":     "android",
		"clienttime":   dateTime,
		"userid":       userid,
		"module_id":    1,
		"page":         1,
		"pagesize":     20,
		"key":          signParamsKey(strconv.FormatInt(dateTime, 10)),
		"special_recommend": map[string]interface{}{
			"withtag": 1, "withsong": 1, "sort": 1, "ugc": 1, "is_selected": 0,
			"withrecommend": 1, "area_code": 1, "categoryid": 0,
		},
		"req_multi": 1, "retrun_min": 5, "return_special_falg": 1,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v2/special_recommend",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "specialrec.service.kugou.com"},
		MID:         mid,
		DFID:        creds.DFID,
		UserID:      userid,
	})
}

// 解析推荐歌单响应（special_recommend）
func ParseRecommendPlaylist(resp *APIResponse) []Playlist {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	var playlists []Playlist

	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}

		if list, ok := dataMap["special_list"]; ok {
			listArr, ok := list.([]interface{})
			if !ok {
				return nil
			}

			for _, item := range listArr {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				pl := parsePlaylist(itemMap)
				if pl != nil {
					playlists = append(playlists, *pl)
				}
			}
		}
	}

	return playlists
}

// 排行榜（歌单分类）
// POST https://gateway.kugou.com/v2/special_recommend（categoryid 区分分类）
// 参数: page, pagesize, category_id, sort
func GetTopPlaylist(page, pagesize int, categoryID, sort string, creds *Credentials) *APIResponse {
	dateTime := time.Now().Unix()
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}
	category := 0
	if categoryID != "" {
		if v, err := strconv.Atoi(categoryID); err == nil {
			category = v
		}
	}
	sortVal := 1
	if sort == "hot" {
		sortVal = 0
	}

	dataMap := map[string]interface{}{
		"appid":        AppID,
		"mid":          mid,
		"clientver":    ClientVer,
		"platform":     "android",
		"clienttime":   dateTime,
		"userid":       userid,
		"module_id":    1,
		"page":         page,
		"pagesize":     pagesize,
		"key":          signParamsKey(strconv.FormatInt(dateTime, 10)),
		"special_recommend": map[string]interface{}{
			"withtag": 1, "withsong": 1, "sort": sortVal, "ugc": 1, "is_selected": 0,
			"withrecommend": 1, "area_code": 1, "categoryid": category,
		},
		"req_multi": 1, "retrun_min": 5, "return_special_falg": 1,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v2/special_recommend",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "specialrec.service.kugou.com"},
		MID:         mid,
		DFID:        creds.DFID,
		UserID:      userid,
	})
}

// 歌单详情
// POST https://gateway.kugou.com/v3/get_list_info
// 参数: data=[{global_collection_id}], userid, token
// 注意: id 需为 global_collection_id 格式（如 collection_3_xxx）
func GetPlaylistDetail(playlistID string, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}

	dataMap := map[string]interface{}{
		"data":   []map[string]interface{}{{"global_collection_id": playlistID}},
		"userid": userid,
		"token":  creds.Token,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v3/get_list_info",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "pubsongs.kugou.com"},
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      userid,
		Token:       creds.Token,
	})
}

// 歌单歌曲列表
// GET https://gateway.kugou.com/pubsongs/v2/get_other_list_file_nofilt
// 参数: global_collection_id, begin_idx, pagesize, ...（响应 data.info 为歌曲列表）
func GetPlaylistSongs(playlistID string, page, pagesize int, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	beginIdx := (page - 1) * pagesize
	if beginIdx < 0 {
		beginIdx = 0
	}

	params := map[string]string{
		"area_code":            "1",
		"begin_idx":            strconv.Itoa(beginIdx),
		"plat":                 "1",
		"type":                 "1",
		"mode":                 "1",
		"personal_switch":      "1",
		"pagesize":             strconv.Itoa(pagesize),
		"global_collection_id": playlistID,
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/pubsongs/v2/get_other_list_file_nofilt",
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 歌单分类
// 评榜分类通过 special_recommend 的 categoryid 区分，直接返回空（由前端展示固定分类）
func GetPlaylistCategories(creds *Credentials) *APIResponse {
	return &APIResponse{Data: map[string]interface{}{"categories": []interface{}{}}}
}

// 歌单 ID 转字符串：JSON 中数值可能以 float64 形式出现，需避免科学计数法
func playlistIDString(v interface{}) string {
	switch t := v.(type) {
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case json.Number:
		return t.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 解析排行榜/歌单列表响应
func ParseTopPlaylist(resp *APIResponse) []Playlist {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	var playlists []Playlist

	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}

		if info, ok := dataMap["info"]; ok {
			infoList, ok := info.([]interface{})
			if !ok {
				return nil
			}

			for _, item := range infoList {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				pl := parsePlaylist(itemMap)
				if pl != nil {
					playlists = append(playlists, *pl)
				}
			}
		}
	}

	return playlists
}

// 解析歌单详情响应（get_list_info）
// 响应结构: data[0] 为歌单信息（含 global_collection_id/listid/name/pic/count）
func ParsePlaylistDetail(resp *APIResponse) *Playlist {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	if data, ok := resp.Data["data"]; ok {
		listArr, ok := data.([]interface{})
		if !ok {
			return nil
		}
		if len(listArr) == 0 {
			return nil
		}
		itemMap, ok := listArr[0].(map[string]interface{})
		if !ok {
			return nil
		}

		pl := parsePlaylist(itemMap)

		// 歌曲列表（若有）
		if songs, ok := resp.Data["songs"]; ok {
			songList, ok := songs.([]interface{})
			if ok {
				for _, item := range songList {
					itemMap, ok := item.(map[string]interface{})
					if !ok {
						continue
					}
					song := parseSongFromTrack(itemMap)
					if song != nil {
						pl.Songs = append(pl.Songs, *song)
					}
				}
			}
		}

		return pl
	}

	return nil
}

// 解析歌单歌曲列表响应（get_other_list_file_nofilt）
// 响应结构: data.songs 为歌曲数组
func ParsePlaylistSongs(resp *APIResponse) []Song {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	var songs []Song

	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}

		// 歌单信息（供详情页使用）
		if listInfo, ok := dataMap["list_info"]; ok {
			if infoMap, ok := listInfo.(map[string]interface{}); ok {
				pl := parsePlaylist(infoMap)
				if pl != nil {
					_ = pl
				}
			}
		}

		if info, ok := dataMap["songs"]; ok {
			songList, ok := info.([]interface{})
			if !ok {
				return nil
			}

			for _, item := range songList {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				song := parseSongFromTrack(itemMap)
				if song != nil {
					songs = append(songs, *song)
				}
			}
		}
	}

	return songs
}

// 从歌单歌曲项解析歌曲
func parseSongFromTrack(item map[string]interface{}) *Song {
	song := &Song{}
	if hash, ok := item["hash"]; ok {
		song.ID = playlistIDString(hash)
	}
	if mix, ok := item["mixsongid"]; ok {
		song.MixSongID = playlistIDString(mix)
	} else if mix, ok := item["mix_song_id"]; ok {
		song.MixSongID = playlistIDString(mix)
	}
	if f, ok := item["fileid"]; ok {
		song.FileID = playlistIDString(f)
	} else if f, ok := item["file_id"]; ok {
		song.FileID = playlistIDString(f)
	} else if f, ok := item["audio_id"]; ok {
		song.FileID = playlistIDString(f)
	}
	if name, ok := item["name"]; ok {
		song.Name = fmt.Sprintf("%v", name)
	}
	if sn, ok := item["songname"]; ok {
		song.Name = fmt.Sprintf("%v", sn)
	}
	if sn, ok := item["filename"]; ok {
		song.Name = fmt.Sprintf("%v", sn)
	}
	// 歌手信息可能在 singerinfo / singers / authors 数组
	if singers, ok := item["singerinfo"].([]interface{}); ok && len(singers) > 0 {
		if s0, ok := singers[0].(map[string]interface{}); ok {
			if n, ok := s0["name"]; ok {
				song.Singer = fmt.Sprintf("%v", n)
			}
			if id, ok := s0["id"]; ok {
				song.SingerID = playlistIDString(id)
			}
		}
	}
	if song.Singer == "" {
		if sn, ok := item["singername"]; ok {
			song.Singer = fmt.Sprintf("%v", sn)
		}
	}
	if song.Singer == "" {
		if authors, ok := item["authors"].([]interface{}); ok && len(authors) > 0 {
			if a0, ok := authors[0].(map[string]interface{}); ok {
				if n, ok := a0["author_name"]; ok {
					song.Singer = fmt.Sprintf("%v", n)
				}
				if id, ok := a0["author_id"]; ok {
					song.SingerID = playlistIDString(id)
				}
			}
		}
	}
	if song.Singer == "" {
		if sn, ok := item["author_name"]; ok {
			song.Singer = fmt.Sprintf("%v", sn)
		}
	}

	// 曲名去歌手前缀：歌单接口 name 为 "歌手 - 歌名" 格式，去掉前半部分
	if song.Singer != "" && song.Name != "" {
		prefix := song.Singer + " - "
		if strings.HasPrefix(song.Name, prefix) {
			song.Name = song.Name[len(prefix):]
		} else if idx := strings.Index(song.Name, " - "); idx > 0 {
			// 多歌手情况：name="A、B - 歌名"，singer=首个歌手A
			fullSinger := song.Name[:idx]
			if strings.Contains(fullSinger, song.Singer) {
				song.Name = song.Name[idx+3:]
			}
		}
	}
	if aid, ok := item["album_id"]; ok {
		song.AlbumID = playlistIDString(aid)
	}
	if an, ok := item["album_name"]; ok {
		song.Album = fmt.Sprintf("%v", an)
	}
	if alb, ok := item["albuminfo"].(map[string]interface{}); ok {
		if n, ok := alb["name"]; ok {
			song.Album = fmt.Sprintf("%v", n)
		}
		if id, ok := alb["id"]; ok {
			song.AlbumID = playlistIDString(id)
		}
	}
	if dur, ok := item["timelen"]; ok {
		song.Duration = int64(toFloat64(dur)) / 1000
	} else if dur, ok := item["timelength"]; ok {
		song.Duration = int64(toFloat64(dur)) / 1000
	} else if dur, ok := item["duration"]; ok {
		song.Duration = int64(toFloat64(dur))
	}
	if ext, ok := item["extname"]; ok {
		song.Extname = fmt.Sprintf("%v", ext)
	}
	if cover, ok := item["cover"]; ok {
		song.Img = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["album_sizable_cover"]; ok {
		song.Img = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["image"]; ok {
		song.Img = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["sizable_cover"]; ok {
		song.Img = fmt.Sprintf("%v", cover)
	}
	if br, ok := item["bitrate"]; ok {
		song.BitRate = int(toFloat64(br))
	}

	if song.ID == "" {
		return nil
	}
	return song
}

func parsePlaylist(item map[string]interface{}) *Playlist {
	pl := &Playlist{}

	if id, ok := item["id"]; ok {
		pl.ID = playlistIDString(id)
	} else if id, ok := item["specialid"]; ok {
		pl.ID = playlistIDString(id)
	} else if id, ok := item["listid"]; ok {
		pl.ID = playlistIDString(id)
	}

	// 保留数字 listid（add_song 需要）
	if lid, ok := item["listid"]; ok {
		pl.ListID = playlistIDString(lid)
	}

	// 优先使用 global_collection_id（v3/get_list_info 接口需要此格式的 id）
	hasGCID := false
	if gcid, ok := item["global_collection_id"]; ok {
		if s := fmt.Sprintf("%v", gcid); s != "" && s != "<nil>" {
			pl.ID = s
			hasGCID = true
		}
	}

	// 评论用 childrenid：用户歌单（collection_3_）需要用原始创建者的 global_collection_id
	// 因为评论关联在创建者视角下，而非当前用户（收藏/转存）副本
	if hasGCID && strings.HasPrefix(pl.ID, "collection_3_") {
		creatorUID := playlistIDString(item["list_create_userid"])
		creatorLID := playlistIDString(item["list_create_listid"])
		if creatorUID != "" && creatorUID != "0" && creatorLID != "" && creatorLID != "0" {
			pl.CommentID = "collection_3_" + creatorUID + "_" + creatorLID + "_0"
			if pl.CommentID == pl.ID {
				pl.CommentID = ""
			}
		}
	}

	if name, ok := item["name"]; ok {
		pl.Name = fmt.Sprintf("%v", name)
	} else if name, ok := item["specialname"]; ok {
		pl.Name = fmt.Sprintf("%v", name)
	}

	if cover, ok := item["img"]; ok {
		pl.Cover = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["cover"]; ok {
		pl.Cover = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["imgurl"]; ok {
		pl.Cover = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["flexible_cover"]; ok {
		pl.Cover = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["pic"]; ok {
		pl.Cover = fmt.Sprintf("%v", cover)
	}

	if count, ok := item["count"]; ok {
		pl.Count = int(toFloat64(count))
	} else if count, ok := item["songcount"]; ok {
		pl.Count = int(toFloat64(count))
	}

	if pc, ok := item["playcount"]; ok {
		pl.PlayCount = int64(toFloat64(pc))
	} else if pc, ok := item["play_count"]; ok {
		pl.PlayCount = int64(toFloat64(pc))
	}

	if author, ok := item["author"]; ok {
		pl.Author = fmt.Sprintf("%v", author)
	} else if uname, ok := item["username"]; ok {
		pl.Author = fmt.Sprintf("%v", uname)
	} else if nickname, ok := item["nickname"]; ok {
		pl.Author = fmt.Sprintf("%v", nickname)
	}

	if desc, ok := item["intro"]; ok {
		pl.Desc = fmt.Sprintf("%v", desc)
	} else if desc, ok := item["desc"]; ok {
		pl.Desc = fmt.Sprintf("%v", desc)
	}

	// 标签
	if tag, ok := item["tag"]; ok {
		tagStr := fmt.Sprintf("%v", tag)
		pl.Tag = strings.Split(tagStr, ",")
	}

	return pl
}