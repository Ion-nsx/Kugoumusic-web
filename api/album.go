package api

import (
	"fmt"
)

// 专辑详情
// POST https://gateway.kugou.com/kmr/v2/albums（x-router: openapi.kugou.com, kg-tid: 255）
// body: data=[{album_id}], fields 指定返回字段
func GetAlbumDetail(albumID string, creds *Credentials) *APIResponse {
	dataMap := map[string]interface{}{
		"data": []map[string]interface{}{
			{"album_id": albumID},
		},
		"is_buy": 0,
		"fields": "album_id,album_name,publish_date,sizable_cover,intro,language,is_publish,heat,type,quality,authors,exclusive,author_name,trans_param",
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/kmr/v2/albums",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "openapi.kugou.com", "kg-tid": "255"},
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 专辑歌曲列表
// POST https://gateway.kugou.com/v1/album_audio/lite（x-router: openapi.kugou.com, kg-tid: 255）
// body: album_id, page, pagesize
func GetAlbumSongs(albumID string, page, pagesize int, creds *Credentials) *APIResponse {
	dataMap := map[string]interface{}{
		"album_id": albumID,
		"is_buy":   "",
		"page":     page,
		"pagesize": pagesize,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v1/album_audio/lite",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "openapi.kugou.com", "kg-tid": "255"},
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 解析专辑详情响应（kmr/v2/albums）
// 响应结构: data[0] 为专辑信息
func ParseAlbumDetail(resp *APIResponse) *Album {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	album := &Album{}

	if data, ok := resp.Data["data"]; ok {
		listArr, ok := data.([]interface{})
		if !ok || len(listArr) == 0 {
			return nil
		}
		dataMap, ok := listArr[0].(map[string]interface{})
		if !ok {
			return nil
		}

		if id, ok := dataMap["album_id"]; ok {
			album.ID = fmt.Sprintf("%v", id)
		} else if id, ok := dataMap["id"]; ok {
			album.ID = fmt.Sprintf("%v", id)
		}
		if name, ok := dataMap["album_name"]; ok {
			album.Name = fmt.Sprintf("%v", name)
		} else if name, ok := dataMap["name"]; ok {
			album.Name = fmt.Sprintf("%v", name)
		}
		if singer, ok := dataMap["author_name"]; ok {
			album.Singer = fmt.Sprintf("%v", singer)
		} else if singer, ok := dataMap["singername"]; ok {
			album.Singer = fmt.Sprintf("%v", singer)
		}
		if cover, ok := dataMap["sizable_cover"]; ok {
			album.Cover = fmt.Sprintf("%v", cover)
		} else if cover, ok := dataMap["cover"]; ok {
			album.Cover = fmt.Sprintf("%v", cover)
		} else if cover, ok := dataMap["imgurl"]; ok {
			album.Cover = fmt.Sprintf("%v", cover)
		}
		if t, ok := dataMap["publish_date"]; ok {
			album.Time = fmt.Sprintf("%v", t)
		} else if t, ok := dataMap["time"]; ok {
			album.Time = fmt.Sprintf("%v", t)
		}
		if lang, ok := dataMap["language"]; ok {
			album.Language = fmt.Sprintf("%v", lang)
		}
	}

	return album
}

// 解析专辑歌曲列表响应（v1/album_audio/lite）
// 响应结构: data.songs
func ParseAlbumSongs(resp *APIResponse) []Song {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	var songs []Song

	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
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
				song := parseOpenAPISong(itemMap)
				if song != nil {
					songs = append(songs, *song)
				}
			}
		}
	}

	return songs
}

// 从 openapi 接口的歌曲项解析，兼容两种结构：
//  1. 顶层字段（artist_audios）：hash/song_name/author_name/album_id/album_name/audio_info
//  2. 嵌套结构（album_songs）：base.audio_name + audio_info.hash + authors[0].author_name + album_info.album_name
func parseOpenAPISong(item map[string]interface{}) *Song {
	song := &Song{}

	// hash：尝试 audio_info.hash，否则顶层 hash
	if audioInfo, ok := item["audio_info"].(map[string]interface{}); ok {
		if h, ok := audioInfo["hash"]; ok {
			song.ID = playlistIDString(h)
		}
		if ext, ok := audioInfo["extname"]; ok {
			song.Extname = fmt.Sprintf("%v", ext)
		}
		if br, ok := audioInfo["bitrate"]; ok {
			song.BitRate = int(toFloat64(br))
		}
		if dur, ok := audioInfo["duration"]; ok {
			song.Duration = int64(toFloat64(dur)) / 1000
		}
	}
	if song.ID == "" {
		if hash, ok := item["hash"]; ok {
			song.ID = playlistIDString(hash)
		} else if hash, ok := item["audio_hash"]; ok {
			song.ID = playlistIDString(hash)
		}
	}
	// 名称
	if name, ok := item["song_name"]; ok {
		song.Name = fmt.Sprintf("%v", name)
	} else if name, ok := item["audio_name"]; ok {
		song.Name = fmt.Sprintf("%v", name)
	}
	// 歌手
	if sn, ok := item["author_name"]; ok {
		song.Singer = fmt.Sprintf("%v", sn)
	} else if sn, ok := item["singer_name"]; ok {
		song.Singer = fmt.Sprintf("%v", sn)
	}
	if sid, ok := item["author_id"]; ok {
		song.SingerID = playlistIDString(sid)
	} else if sid, ok := item["singer_id"]; ok {
		song.SingerID = playlistIDString(sid)
	}
	// 专辑
	if aid, ok := item["album_id"]; ok {
		song.AlbumID = playlistIDString(aid)
	}
	if an, ok := item["album_name"]; ok {
		song.Album = fmt.Sprintf("%v", an)
	}
	// 封面：album_sizable_cover / sizable_cover / image
	if cover, ok := item["album_sizable_cover"]; ok {
		song.Img = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["image"]; ok {
		song.Img = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["sizable_cover"]; ok {
		song.Img = fmt.Sprintf("%v", cover)
	} else if cover, ok := item["cover"]; ok {
		song.Img = fmt.Sprintf("%v", cover)
	}
	// 时长：顶层 timelength（毫秒）
	if song.Duration == 0 {
		if dur, ok := item["timelength"]; ok {
			song.Duration = int64(toFloat64(dur)) / 1000
		}
	}
	// 比特率/文件大小（顶层可能有）
	if song.BitRate == 0 {
		if br, ok := item["bitrate"]; ok {
			song.BitRate = int(toFloat64(br))
		}
	}
	if sz, ok := item["filesize"]; ok {
		song.Size = int64(toFloat64(sz))
	}
	// mixsongid（收藏/添加歌单用）
	if mix, ok := item["mixsongid"]; ok {
		song.MixSongID = playlistIDString(mix)
	} else if mix, ok := item["mix_song_id"]; ok {
		song.MixSongID = playlistIDString(mix)
	} else if aid, ok := item["audio_id"]; ok {
		song.MixSongID = playlistIDString(aid)
	}

	// 专辑音频专用嵌套结构
	if base, ok := item["base"].(map[string]interface{}); ok {
		if song.Name == "" {
			if an, ok := base["audio_name"]; ok {
				song.Name = fmt.Sprintf("%v", an)
			}
		}
		if song.AlbumID == "" {
			if aid, ok := base["album_id"]; ok {
				song.AlbumID = playlistIDString(aid)
			}
		}
		if song.ID == "" {
			if aid, ok := base["audio_id"]; ok {
				song.ID = playlistIDString(aid)
			}
		}
	}
	if albumInfo, ok := item["album_info"].(map[string]interface{}); ok {
		if song.Album == "" {
			if an, ok := albumInfo["album_name"]; ok {
				song.Album = fmt.Sprintf("%v", an)
			}
		}
	}
	if authors, ok := item["authors"].([]interface{}); ok && song.Singer == "" {
		for _, a := range authors {
			if am, ok := a.(map[string]interface{}); ok {
				if an, ok := am["author_name"]; ok {
					song.Singer = fmt.Sprintf("%v", an)
					break
				}
			}
		}
	}

	if song.ID == "" {
		return nil
	}
	return song
}
