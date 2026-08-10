package api

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// 搜索歌曲
// GET http://mobilecdn.kugou.com/api/v3/search/song
// 参数: keyword, page, pagesize, ...
func SearchSongs(keyword string, page, pagesize int, creds *Credentials) *APIResponse {
	params := creds.InjectParams(map[string]string{
		"keyword":         keyword,
		"page":            strconv.Itoa(page),
		"pagesize":        strconv.Itoa(pagesize),
		"showtype":        "1",
		"version":         "1",
		"correct":         "1",
		"privilege_filter": "0",
		"cmsv":            "1",
		"srcverid":        "0",
		"plat":            "0",
		"type":            "0",
		"tag":             "0",
		"flag":            "0",
		"ishighquality":   "0",
		"is_orphans":      "0",
		"search_type":     "0",
	})

	return SendRequest(&RequestOptions{
		Method: "GET",
		URL:    "http://mobilecdn.kugou.com/api/v3/search/song",
		Params: params,
	})
}

// 搜索建议（新接口）
// GET https://gateway.kugou.com/v2/getSearchTip（x-router: searchtip.kugou.com）
// 响应 data 为数组，每个元素含 RecordDatas[].HintInfo（歌手/MV/专辑分类建议）
func SearchSuggest(keyword string, creds *Credentials) *APIResponse {
	params := map[string]string{
		"keyword":         keyword,
		"AlbumTipCount":   "10",
		"CorrectTipCount": "10",
		"MVTipCount":      "10",
		"MusicTipCount":   "10",
		"radiotip":        "1",
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v2/getSearchTip",
		Params:      params,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "searchtip.kugou.com"},
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 综合搜索（聚合各分类型搜索结果）
// 说明: mobilecdn /search/complex 与 gateway /v6/search/complex、/v3/search/mixed 均已失效（error 152），
// 因此这里并行聚合 song/album/author/special/mv 五个分类型搜索，构造与旧接口一致的结构。
// song 走 mobilecdn（仍可用）；album/author/mv/special 走 gateway /v1/search/{type}。
func SearchComplex(keyword string, page, pagesize int, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	if page < 1 {
		page = 1
	}
	if pagesize < 1 {
		pagesize = 10
	}

	type segResult struct {
		typ   string
		total int
		lists []interface{}
	}

	results := make(chan segResult, 5)
	var wg sync.WaitGroup

	// song：mobilecdn 搜索
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := SearchSongs(keyword, page, pagesize, creds)
		parsed := ParseSearchSongs(r)
		var lists []interface{}
		for _, s := range parsed.Songs {
			lists = append(lists, songToMap(s))
		}
		results <- segResult{typ: "song", total: parsed.Total, lists: lists}
	}()

	// album / author / special / mv：gateway 分类型搜索
	for _, typ := range []string{"album", "author", "special", "mv"} {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			r := SearchByType(t, keyword, page, pagesize, creds)
			lists, total := ParseTypedSearch(r)
			results <- segResult{typ: t, total: total, lists: lists}
		}(typ)
	}

	wg.Wait()
	close(results)

	segments := []map[string]interface{}{
		{"type": "song"}, {"type": "mv"}, {"type": "special"},
		{"type": "album"}, {"type": "ksong"}, {"type": "program"},
		{"type": "author"}, {"type": "talent"},
	}
	byType := map[string]segResult{}
	for r := range results {
		byType[r.typ] = r
	}

	order := []string{"song", "mv", "special", "album", "ksong", "program", "author", "talent"}
	for i, t := range order {
		if sr, ok := byType[t]; ok {
			segments[i]["total"] = sr.total
			segments[i]["lists"] = sr.lists
		} else {
			segments[i]["total"] = 0
			segments[i]["lists"] = []interface{}{}
		}
	}

	return &APIResponse{
		StatusCode: 200,
		Data: map[string]interface{}{
			"status": 1,
			"data": map[string]interface{}{
				"keyword":   keyword,
				"indextotal": len(segments),
				"lists":     segments,
			},
		},
	}
}

// 分类型搜索
// song → mobilecdn /api/v3/search/song（仍可用）
// album/author/mv/special → gateway /v1/search/{type}（x-router: complexsearch.kugou.com）
func SearchByType(typ, keyword string, page, pagesize int, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	if typ == "" || typ == "song" {
		return SearchSongs(keyword, page, pagesize, creds)
	}

	params := map[string]string{
		"keyword":       keyword,
		"page":          strconv.Itoa(page),
		"pagesize":      strconv.Itoa(pagesize),
		"albumhide":     "0",
		"iscorrection":  "1",
		"nocollect":     "0",
		"platform":      "AndroidFilter",
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v1/search/" + typ,
		Params:      params,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "complexsearch.kugou.com"},
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 解析分类型搜索响应（gateway /v1/search/{type}）
// 结构: data.lists[]，字段因类型而异
func ParseTypedSearch(resp *APIResponse) ([]interface{}, int) {
	if resp.Error != nil || resp.Data == nil {
		return nil, 0
	}
	total := 0
	if t, ok := resp.Data["total"]; ok {
		total = int(toFloat64(t))
	}
	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil, total
		}
		if t, ok := dataMap["total"]; ok {
			total = int(toFloat64(t))
		}
		if lists, ok := dataMap["lists"]; ok {
			if listArr, ok := lists.([]interface{}); ok {
				return listArr, total
			}
		}
	}
	return nil, total
}

// 将 Song 转为 map（供综合搜索聚合使用）
func songToMap(s Song) map[string]interface{} {
	return map[string]interface{}{
		"id":        s.ID,
		"name":      s.Name,
		"singer":    s.Singer,
		"album":     s.Album,
		"album_id":  s.AlbumID,
		"duration":  s.Duration,
		"img":       s.Img,
		"hash":      s.ID,
		"mixsongid": s.MixSongID,
	}
}

// 热搜榜
// GET http://mobilecdn.kugou.com/api/v3/search/hot
func SearchHot(creds *Credentials) *APIResponse {
	params := creds.InjectParams(map[string]string{
		"plat": "0",
		"type": "0",
	})

	return SendRequest(&RequestOptions{
		Method: "GET",
		URL:    "http://mobilecdn.kugou.com/api/v3/search/hot",
		Params: params,
	})
}

// 解析搜索歌曲响应
func ParseSearchSongs(resp *APIResponse) *SearchResult {
	if resp.Error != nil || resp.Data == nil {
		return &SearchResult{}
	}

	result := &SearchResult{}

	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return result
		}

		if total, ok := dataMap["total"]; ok {
			result.Total = int(toFloat64(total))
		}

		if info, ok := dataMap["info"]; ok {
			infoList, ok := info.([]interface{})
			if !ok {
				return result
			}

			for _, item := range infoList {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				song := parseSongFromSearch(itemMap)
				if song != nil {
					result.Songs = append(result.Songs, *song)
				}
			}
		}
	}

	return result
}

// 解析搜索建议响应（v2/getSearchTip）
// 结构: data[].RecordDatas[].HintInfo，多个分类组（歌手/MV/专辑/歌曲）
func ParseSearchSuggest(resp *APIResponse) []string {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	var suggestions []string
	seen := map[string]bool{}

	if data, ok := resp.Data["data"]; ok {
		groupList, ok := data.([]interface{})
		if !ok {
			return nil
		}
		for _, group := range groupList {
			groupMap, ok := group.(map[string]interface{})
			if !ok {
				continue
			}
			records, ok := groupMap["RecordDatas"].([]interface{})
			if !ok {
				continue
			}
			for _, rec := range records {
				recMap, ok := rec.(map[string]interface{})
				if !ok {
					continue
				}
				if hint, ok := recMap["HintInfo"]; ok {
					s := fmt.Sprintf("%v", hint)
					if s != "" && !seen[s] {
						seen[s] = true
						suggestions = append(suggestions, s)
					}
				}
			}
		}
	}

	return suggestions
}

// 搜索默认词 — 搜索框未聚焦时展示的推荐搜索词
// POST https://gateway.kugou.com/searchnofocus/v1/search_no_focus_word
// 固定参数: plat=0, vip_type=65530, mode=normal, clientver=12329(query)
func GetSearchDefault(creds *Credentials) *APIResponse {
	userid := "0"
	vipType := "65530"
	if creds != nil {
		if creds.UserID != "" && creds.UserID != "0" {
			userid = creds.UserID
		}
	}

	dataMap := map[string]interface{}{
		"plat":     0,
		"userid":   toFloat64(userid),
		"tags":     "{}",
		"vip_type": toFloat64(vipType),
		"m_type":   0,
		"own_ads":  map[string]interface{}{},
		"ability":  "3",
		"sources":  []interface{}{},
		"bitmap":   2,
		"mode":     "normal",
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/searchnofocus/v1/search_no_focus_word",
		Data:        dataMap,
		Params:      map[string]string{"clientver": "12329"},
		EncryptType: "android",
		MID:         GetMID(),
	})
}

// 解析热搜响应
func ParseSearchHot(resp *APIResponse) []HotWord {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	var hotWords []HotWord
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
			for i, item := range infoList {
				itemMap, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				word := ""
				if w, ok := itemMap["keyword"]; ok {
					word = fmt.Sprintf("%v", w)
				}
				hot := int64(0)
				if h, ok := itemMap["hot"]; ok {
					hot = int64(toFloat64(h))
				}
				hotWords = append(hotWords, HotWord{
					Word: word,
					Rank: i + 1,
					Hot:  hot,
				})
			}
		}
	}

	return hotWords
}

// 从搜索条目解析歌曲
func parseSongFromSearch(item map[string]interface{}) *Song {
	song := &Song{}

	if hash, ok := item["hash"]; ok {
		song.ID = playlistIDString(hash)
	}
	if mix, ok := item["mixsongid"]; ok {
		song.MixSongID = playlistIDString(mix)
	} else if mix, ok := item["mix_song_id"]; ok {
		song.MixSongID = playlistIDString(mix)
	} else if aid, ok := item["audio_id"]; ok {
		song.MixSongID = playlistIDString(aid)
	}
	if songID, ok := item["songname_original"]; ok {
		song.Name = fmt.Sprintf("%v", songID)
	} else if sn, ok := item["songname"]; ok {
		song.Name = fmt.Sprintf("%v", sn)
	}
	if an, ok := item["album_name"]; ok {
		song.Album = fmt.Sprintf("%v", an)
	}
	if aid, ok := item["album_id"]; ok {
		song.AlbumID = playlistIDString(aid)
	}
	if sn, ok := item["singername"]; ok {
		// 可能包含多个歌手，用逗号分隔
		artists := strings.Split(fmt.Sprintf("%v", sn), "、")
		song.Singer = fmt.Sprintf("%v", sn)
		if len(artists) > 0 {
			song.Singer = artists[0]
		}
	}
	if si, ok := item["singer_id"]; ok {
		song.SingerID = playlistIDString(si)
	}
	if t, ok := item["duration"]; ok {
		song.Duration = int64(toFloat64(t))
	}
	if img, ok := item["img"]; ok {
		song.Img = fmt.Sprintf("%v", img)
	}
	// mobilecdn /v3/search/song 无 img 字段，封面在 trans_param.union_cover（带 {size} 占位符）
	if song.Img == "" {
		if tp, ok := item["trans_param"].(map[string]interface{}); ok {
			if uc, ok := tp["union_cover"]; ok {
				song.Img = fmt.Sprintf("%v", uc)
			}
		}
	}
	if img, ok := item["hq_img"]; ok {
		song.HQImg = fmt.Sprintf("%v", img)
	}
	if pt, ok := item["pay_type"]; ok {
		song.PayType = fmt.Sprintf("%v", pt)
	}
	if p, ok := item["pay"]; ok {
		song.Pay = int(toFloat64(p))
	}

	// 音质
	if et, ok := item["extname"]; ok {
		song.Extname = fmt.Sprintf("%v", et)
	}
	if bt, ok := item["bitrate"]; ok {
		song.BitRate = int(toFloat64(bt))
	}

	song.Playable = true

	return song
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// 非原版特征词（DJ/Remix/翻唱/伴奏等），大小写不敏感
var nonOriginalPatterns = []string{
	"dj", "remix", "rmx", "翻唱", "cover", "伴奏", "instrumental",
	"karaoke", "纯音乐", "慢速", "加速", "降调", "升调",
	"烟嗓", "烟酒嗓", "伤感版", "新版", "改版",
}

// IsNonOriginalVersion 判断是否为非原版（DJ/翻唱版等）
func IsNonOriginalVersion(name, singer string) bool {
	lower := strings.ToLower(name)
	for _, p := range nonOriginalPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	// 歌手名包含 DJ/翻唱也视为非原版
	lowerSinger := strings.ToLower(singer)
	if strings.Contains(lowerSinger, "dj") || strings.Contains(lowerSinger, "翻唱") {
		return true
	}
	return false
}

// SortOriginalFirst 智能排序：DJ/Remix版排最后，同歌名优先推送主唱版本
func SortOriginalFirst(songs []Song) []Song {
	if len(songs) <= 1 {
		return songs
	}

	// 统计每个歌名对应歌手的出现次数，找出主唱（出现最多的歌手）
	singerBySong := map[string]map[string]int{}
	for _, s := range songs {
		if singerBySong[s.Name] == nil {
			singerBySong[s.Name] = map[string]int{}
		}
		singerBySong[s.Name][s.Singer]++
	}
	primarySinger := map[string]string{}
	for name, singers := range singerBySong {
		best, maxN := "", 0
		for singer, n := range singers {
			if n > maxN {
				maxN, best = n, singer
			}
		}
		primarySinger[name] = best
	}

	type scoredItem struct {
		song  Song
		score int
	}
	items := make([]scoredItem, len(songs))
	for i, s := range songs {
		score := 100
		// DJ/Remix 严重扣分
		if IsNonOriginalVersion(s.Name, s.Singer) {
			score -= 90
		}
		// 主唱加分
		if primary, ok := primarySinger[s.Name]; ok && primary != "" && s.Singer == primary {
			score += 20
		}
		// 同歌名有多版本但非主唱：轻微扣分
		if singers, ok := singerBySong[s.Name]; ok && len(singers) > 1 {
			if s.Singer != primarySinger[s.Name] {
				score -= 5
			}
		}
		// 高质量加分
		if s.BitRate >= 320 {
			score += 10
		} else if s.BitRate > 128 {
			score += 5
		}
		if s.Size > 10*1024*1024 {
			score += 5
		}
		items[i] = scoredItem{song: s, score: score}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].score > items[j].score
	})

	result := make([]Song, len(songs))
	for i, it := range items {
		result[i] = it.song
	}
	return result
}

// SortSongMaps 对综合搜索中的歌曲 map 列表做原版优先排序
func SortSongMaps(songs []interface{}) []interface{} {
	if len(songs) <= 1 {
		return songs
	}

	// 从 map 中提取歌手/歌名信息
	type songMeta struct {
		idx    int
		name   string
		singer string
	}
	metas := make([]songMeta, len(songs))
	for i, s := range songs {
		m, ok := s.(map[string]interface{})
		if ok {
			name, _ := m["name"].(string)
			singer, _ := m["singer"].(string)
			metas[i] = songMeta{idx: i, name: name, singer: singer}
		} else {
			metas[i] = songMeta{idx: i}
		}
	}

	// 统计主唱
	singerBySong := map[string]map[string]int{}
	for _, m := range metas {
		if singerBySong[m.name] == nil {
			singerBySong[m.name] = map[string]int{}
		}
		singerBySong[m.name][m.singer]++
	}
	primarySinger := map[string]string{}
	for name, singers := range singerBySong {
		best, maxN := "", 0
		for singer, n := range singers {
			if n > maxN {
				maxN, best = n, singer
			}
		}
		primarySinger[name] = best
	}

	type scoredIdx struct {
		idx   int
		score int
	}
	scored := make([]scoredIdx, len(songs))
	for i, m := range metas {
		score := 100
		if IsNonOriginalVersion(m.name, m.singer) {
			score -= 90
		}
		if primary, ok := primarySinger[m.name]; ok && primary != "" && m.singer == primary {
			score += 20
		}
		if singers, ok := singerBySong[m.name]; ok && len(singers) > 1 {
			if m.singer != primarySinger[m.name] {
				score -= 5
			}
		}
		scored[i] = scoredIdx{idx: m.idx, score: score}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	result := make([]interface{}, len(songs))
	for i, sc := range scored {
		result[i] = songs[sc.idx]
	}
	return result
}