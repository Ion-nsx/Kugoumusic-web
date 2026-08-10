package api

import (
	"fmt"
	"strconv"
	"time"
)

// 歌手详情
// POST https://gateway.kugou.com/kmr/v3/author（x-router: openapi.kugou.com, kg-tid: 36）
// body: author_id
func GetArtistDetail(singerID string, creds *Credentials) *APIResponse {
	dataMap := map[string]interface{}{
		"author_id": singerID,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/kmr/v3/author",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "openapi.kugou.com", "kg-tid": "36"},
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 歌手作品列表
// POST https://openapi.kugou.com/kmr/v1/audio_group/author（kg-tid: 220）
// body: author_id, page, pagesize, sort(1最热/2最新), area_code
func GetArtistAudios(singerID string, page, pagesize int, creds *Credentials) *APIResponse {
	dateTime := time.Now().Unix()
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}

	dataMap := map[string]interface{}{
		"appid":      AppID,
		"clientver":  ClientVer,
		"mid":        mid,
		"clienttime": dateTime,
		"key":        signParamsKey(strconv.FormatInt(dateTime, 10)),
		"author_id":  singerID,
		"pagesize":   pagesize,
		"page":       page,
		"sort":       2, // 2：最新
		"area_code":  "all",
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://openapi.kugou.com",
		URL:         "/kmr/v1/audio_group/author",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "openapi.kugou.com", "kg-tid": "220"},
		MID:         mid,
		DFID:        creds.DFID,
	})
}

// 歌手热门歌曲（用作品列表的最热排序）
func GetArtistHot(singerID string, creds *Credentials) *APIResponse {
	return GetArtistAudios(singerID, 1, 30, creds)
}

// 歌手专辑列表
// POST https://gateway.kugou.com/kmr/v1/author/albums（x-router: openapi.kugou.com, kg-tid: 36）
// body: author_id, pagesize, page, sort(3最热/1最新), category, area_code
func GetArtistAlbums(singerID string, page, pagesize int, creds *Credentials) *APIResponse {
	dataMap := map[string]interface{}{
		"author_id": singerID,
		"pagesize":  pagesize,
		"page":      page,
		"sort":      1, // 1：最新
		"category":  1,
		"area_code": "all",
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/kmr/v1/author/albums",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "openapi.kugou.com", "kg-tid": "36"},
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 解析歌手详情响应（kmr/v3/author）
// 响应结构: data 为歌手信息
func ParseArtistDetail(resp *APIResponse) *Artist {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	artist := &Artist{}

	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}

		if id, ok := dataMap["author_id"]; ok {
			artist.ID = fmt.Sprintf("%v", id)
		} else if id, ok := dataMap["id"]; ok {
			artist.ID = fmt.Sprintf("%v", id)
		} else if id, ok := dataMap["singer_id"]; ok {
			artist.ID = fmt.Sprintf("%v", id)
		}
		if name, ok := dataMap["author_name"]; ok {
			artist.Name = fmt.Sprintf("%v", name)
		} else if name, ok := dataMap["name"]; ok {
			artist.Name = fmt.Sprintf("%v", name)
		} else if name, ok := dataMap["singername"]; ok {
			artist.Name = fmt.Sprintf("%v", name)
		}
		if pic, ok := dataMap["sizable_avatar"]; ok {
			artist.Pic = fmt.Sprintf("%v", pic)
		} else if pic, ok := dataMap["avatar"]; ok {
			artist.Pic = fmt.Sprintf("%v", pic)
		} else if pic, ok := dataMap["imgurl"]; ok {
			artist.Pic = fmt.Sprintf("%v", pic)
		}
		if count, ok := dataMap["song_count"]; ok {
			artist.Count = int(toFloat64(count))
		} else if count, ok := dataMap["count"]; ok {
			artist.Count = int(toFloat64(count))
		}
		if gender, ok := dataMap["gender"]; ok {
			artist.Gender = fmt.Sprintf("%v", gender)
		}
		if country, ok := dataMap["country"]; ok {
			artist.Country = fmt.Sprintf("%v", country)
		}
		if desc, ok := dataMap["intro"]; ok {
			artist.Desc = fmt.Sprintf("%v", desc)
		} else if desc, ok := dataMap["desc"]; ok {
			artist.Desc = fmt.Sprintf("%v", desc)
		}
	}

	return artist
}

// 解析歌手作品列表响应（kmr/v1/audio_group/author）
// 响应结构: data 为歌曲数组
// 歌手作品接口不返回封面，通过批量查询专辑详情补充
func ParseArtistAudios(resp *APIResponse) []Song {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	var songs []Song
	albumIDs := make(map[string]bool)

	if data, ok := resp.Data["data"]; ok {
		songList, ok := data.([]interface{})
		if !ok {
			return nil
		}

		for _, item := range songList {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			song := parseOpenAPISong(itemMap)
			if song != nil {
				songs = append(songs, *song)
				// 收集非零 album_id，后续批量获取专辑封面
				if aid, ok := itemMap["album_id"]; ok {
					idStr := playlistIDString(aid)
					if idStr != "0" {
						albumIDs[idStr] = true
					}
				}
			}
		}
	}

	// 批量查询专辑详情获取封面
	if len(albumIDs) > 0 && len(songs) > 0 {
		albumCovers := fetchAlbumCovers(albumIDs, nil)
		for i := range songs {
			if songs[i].Img == "" && songs[i].AlbumID != "" {
				if cover, ok := albumCovers[songs[i].AlbumID]; ok {
					songs[i].Img = cover
				}
			}
		}
	}

	return songs
}

// 批量获取专辑封面
func fetchAlbumCovers(albumIDs map[string]bool, creds *Credentials) map[string]string {
	albumList := make([]map[string]interface{}, 0, len(albumIDs))
	for id := range albumIDs {
		albumList = append(albumList, map[string]interface{}{"album_id": id})
	}
	if len(albumList) == 0 {
		return nil
	}

	dataMap := map[string]interface{}{
		"data":   albumList,
		"is_buy": 0,
		"fields": "album_id,sizable_cover",
	}

	resp := SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/kmr/v2/albums",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "openapi.kugou.com", "kg-tid": "255"},
		MID:         GetMID(),
		DFID:        "-",
	})
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	covers := make(map[string]string)
	if data, ok := resp.Data["data"].([]interface{}); ok {
		for _, item := range data {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			albumID := ""
			if aid, ok := itemMap["album_id"]; ok {
				albumID = playlistIDString(aid)
			}
			if cover, ok := itemMap["sizable_cover"]; ok {
				covers[albumID] = fmt.Sprintf("%v", cover)
			}
		}
	}
	return covers
}

// 解析歌手专辑列表响应（kmr/v1/author/albums）
// 响应结构: data.list 或 data.albums
func ParseArtistAlbums(resp *APIResponse) []Album {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	var albums []Album

	if data, ok := resp.Data["data"]; ok {
		var listArr []interface{}

		switch v := data.(type) {
		case []interface{}:
			// data 直接是数组（author/albums 接口）
			listArr = v
		case map[string]interface{}:
			// data 是对象，取 list/albums/album_list
			if list, ok := v["list"]; ok {
				listArr, _ = list.([]interface{})
			} else if albumsArr, ok := v["albums"]; ok {
				listArr, _ = albumsArr.([]interface{})
			} else if list, ok := v["album_list"]; ok {
				listArr, _ = list.([]interface{})
			}
		}

		for _, item := range listArr {
			itemMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			album := &Album{}
			if id, ok := itemMap["album_id"]; ok {
				album.ID = playlistIDString(id)
			}
			if name, ok := itemMap["album_name"]; ok {
				album.Name = fmt.Sprintf("%v", name)
			}
			if cover, ok := itemMap["sizable_cover"]; ok {
				album.Cover = fmt.Sprintf("%v", cover)
			} else if cover, ok := itemMap["cover"]; ok {
				album.Cover = fmt.Sprintf("%v", cover)
			} else if cover, ok := itemMap["imgurl"]; ok {
				album.Cover = fmt.Sprintf("%v", cover)
			}
			if t, ok := itemMap["publish_date"]; ok {
				album.Time = fmt.Sprintf("%v", t)
			}
			if album.ID != "" {
				albums = append(albums, *album)
			}
		}
	}

	return albums
}
