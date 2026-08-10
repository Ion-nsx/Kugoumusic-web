package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ============================================================
// 评论相关（对齐 KuGouMusicApi /mcomment 接口）
// ============================================================

const (
	commentSongCode    = "fc4be23b4e972707f36b8a828a93ba8a" // 歌曲
	commentPlaylistCode = "ca53b96fe5a1d9c22d71c8f522ef7c4f" // 歌单
	commentAlbumCode   = "94f1792ced1df89aa68a7939eaf2efca" // 专辑
)

// 歌曲评论列表
// POST https://gateway.kugou.com/mcomment/v1/cmtlist
func GetSongComments(mixsongid string, page, pagesize int, creds *Credentials) *APIResponse {
	if pagesize < 1 {
		pagesize = 30
	}
	if page < 1 {
		page = 1
	}
	params := map[string]string{
		"mixsongid":          mixsongid,
		"need_show_image":    "1",
		"p":                  strconv.Itoa(page),
		"pagesize":           strconv.Itoa(pagesize),
		"show_classify":      "1",
		"show_hotword_list":  "1",
		"extdata":            "0",
		"code":               commentSongCode,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/mcomment/v1/cmtlist",
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// 歌单评论列表
// POST https://gateway.kugou.com/m.comment.service/v1/cmtlist
func GetPlaylistComments(id string, page, pagesize int, creds *Credentials) *APIResponse {
	if pagesize < 1 {
		pagesize = 30
	}
	if page < 1 {
		page = 1
	}
	if creds == nil {
		creds = &Credentials{}
	}
	params := map[string]string{
		"childrenid":          id,
		"need_show_image":     "1",
		"p":                   strconv.Itoa(page),
		"pagesize":            strconv.Itoa(pagesize),
		"show_classify":       "1",
		"show_hotword_list":   "1",
		"code":                commentPlaylistCode,
		"content_type":        "0",
		"tag":                 "5",
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/m.comment.service/v1/cmtlist",
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// 专辑评论列表
func GetAlbumComments(id string, page, pagesize int, creds *Credentials) *APIResponse {
	if pagesize < 1 {
		pagesize = 30
	}
	if page < 1 {
		page = 1
	}
	params := map[string]string{
		"childrenid":         id,
		"need_show_image":    "1",
		"p":                  strconv.Itoa(page),
		"pagesize":           strconv.Itoa(pagesize),
		"show_classify":      "1",
		"show_hotword_list":  "1",
		"code":               commentAlbumCode,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/m.comment.service/v1/cmtlist",
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// 评论数
// GET gateway.kugou.com/index.php?r=comments/getcommentsnum（web 签名）
func GetCommentCount(hash, specialID string, creds *Credentials) *APIResponse {
	params := map[string]string{
		"r":    "comments/getcommentsnum",
		"code": commentSongCode,
	}
	if hash != "" {
		params["hash"] = hash
	}
	if specialID != "" {
		params["childrenid"] = specialID
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/index.php",
		Params:      params,
		EncryptType: "web",
		Headers:     map[string]string{"x-router": "sum.comment.service.kugou.com"},
		MID:         creds.MID,
		DFID:        creds.DFID,
	})
}

// 楼层评论（热评回复）
// 歌曲: /mcomment/v1/hot_replylist；歌单/专辑: /m.comment.service/v1/hot_replylist
func GetCommentFloor(resourceType, code, mixsongid, specialID, tid string, page, pagesize int, creds *Credentials) *APIResponse {
	if pagesize < 1 {
		pagesize = 30
	}
	if page < 1 {
		page = 1
	}
	rt := strings.ToLower(resourceType)
	if code == "" {
		switch rt {
		case "playlist":
			code = commentPlaylistCode
		case "album":
			code = commentAlbumCode
		default:
			code = commentSongCode
		}
	}
	useService := rt == "playlist" || rt == "album" || code == commentPlaylistCode || code == commentAlbumCode

	params := map[string]string{
		"childrenid":        specialID,
		"need_show_image":   "1",
		"p":                 strconv.Itoa(page),
		"pagesize":          strconv.Itoa(pagesize),
		"show_classify":     "1",
		"show_hotword_list": "1",
		"code":              code,
		"tid":               tid,
	}
	if mixsongid != "" {
		params["mixsongid"] = mixsongid
	}

	path := "/mcomment/v1/hot_replylist"
	if useService {
		path = "/m.comment.service/v1/hot_replylist"
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         path,
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// 评论热词
// POST https://gateway.kugou.com/mcomment/v1/get_hot_word
func GetCommentHotWords(mixsongid string, page, pagesize int, creds *Credentials) *APIResponse {
	if pagesize < 1 {
		pagesize = 30
	}
	if page < 1 {
		page = 1
	}
	params := map[string]string{
		"mixsongid":       mixsongid,
		"need_show_image": "1",
		"p":               strconv.Itoa(page),
		"pagesize":        strconv.Itoa(pagesize),
		"hot_word":        "",
		"extdata":         "0",
		"code":            commentSongCode,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/mcomment/v1/get_hot_word",
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// 发送评论
// POST https://gateway.kugou.com/mcomment/v1/add_cmt（歌曲/歌单/专辑）
func SendComment(resourceType, code, mixsongid, childrenid, content string, creds *Credentials) *APIResponse {
	content = strings.TrimSpace(content)
	if content == "" {
		return &APIResponse{Error: fmt.Errorf("评论内容不能为空")}
	}
	rt := strings.ToLower(resourceType)
	if code == "" {
		switch rt {
		case "playlist":
			code = commentPlaylistCode
		case "album":
			code = commentAlbumCode
		default:
			code = commentSongCode
		}
	}
	useService := rt == "playlist" || rt == "album"

	params := map[string]string{
		"need_show_image": "1",
		"content":         content,
		"code":            code,
	}
	if mixsongid != "" {
		params["mixsongid"] = mixsongid
	}
	if childrenid != "" {
		params["childrenid"] = childrenid
	}

	path := "/mcomment/v1/add_cmt"
	if useService {
		path = "/m.comment.service/v1/add_cmt"
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         path,
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// ============================================================
// 歌单管理（对齐 KuGouMusicApi cloudlist.service 接口）
// ============================================================

// 创建歌单
// POST gateway.kugou.com/cloudlist.service/v5/add_list（x-router: cloudlist.service.kugou.com）
func CreatePlaylist(name string, isPri bool, creds *Credentials) *APIResponse {
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}
	clienttime := time.Now().Unix()

	pri := 0
	if isPri {
		pri = 1
	}

	dataMap := map[string]interface{}{
		"userid":            userid,
		"token":             creds.Token,
		"total_ver":         0,
		"name":              name,
		"type":              0,
		"source":            1,
		"is_pri":            pri,
		"list_create_userid": userid,
		"list_create_listid": "",
		"list_create_gid":    "",
		"from_shupinmv":     0,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/cloudlist.service/v5/add_list",
		Data:        dataMap,
		Params: map[string]string{
			"last_time": strconv.FormatInt(clienttime, 10),
			"last_area": "gztx",
			"userid":    userid,
			"token":     creds.Token,
		},
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "cloudlist.service.kugou.com"},
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      userid,
		Token:       creds.Token,
	})
}

// 删除歌单 / 取消收藏
// POST gateway.kugou.com/v2/delete_list（AES 加密 body + RSA 加密 key）
// 返回 aesKey 用于解密加密响应
func DeletePlaylist(listid string, creds *Credentials) (*APIResponse, string, error) {
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}
	clienttime := time.Now().Unix()

	rawListID, err := strconv.ParseInt(listid, 10, 64)
	if err != nil {
		// 兼容 collection_3_xxx 格式，尝试取末尾数字段
		parts := strings.Split(listid, "_")
		if len(parts) > 0 {
			rawListID, err = strconv.ParseInt(parts[len(parts)-1], 10, 64)
		}
		if err != nil {
			return nil, "", fmt.Errorf("无效的歌单 id: %s", listid)
		}
	}

	dataMap := map[string]interface{}{
		"listid":    rawListID,
		"total_ver": 0,
		"type":      1,
	}

	aes, err := PlaylistAESEncrypt(dataMap)
	if err != nil {
		return nil, "", err
	}

	p, err := RSAEncrypt2(map[string]interface{}{"aes": aes.Key, "uid": userid, "token": creds.Token}, UseLite)
	if err != nil {
		return nil, "", err
	}

	params := map[string]string{
		"clienttime": strconv.FormatInt(clienttime, 10),
		"key":        signParamsKey(strconv.FormatInt(clienttime, 10)),
		"last_area":  "gztx",
		"clientver":  strconv.Itoa(ClientVer),
		"appid":      strconv.Itoa(AppID),
		"last_time":  strconv.FormatInt(clienttime, 10),
		"p":          strings.ToUpper(p),
	}

	resp := SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v2/delete_list",
		Params:      params,
		RawBody:     aes.Str,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "cloudlist.service.kugou.com"},
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      userid,
		Token:       creds.Token,
	})
	return resp, aes.Key, nil
}

// NormalizeListID 把歌单 id 统一为数字 listid（兼容 collection_3_xxx 格式，取末尾数字段）。
// add_song / delete_songs 上游只认数字 listid。
func NormalizeListID(listid string) string {
	parts := strings.Split(listid, "_")
	if len(parts) > 1 {
		if _, err := strconv.ParseInt(parts[len(parts)-1], 10, 64); err == nil {
			return parts[len(parts)-1]
		}
	}
	return listid
}

// 从 global_collection_id 提取数字 listid（add_song 不接受 collection_3_xxx 格式）
// 格式: collection_3_{userid}_{listid}_{sub}
// 如 "collection_3_725779799_2_0" → listid 是第4段 "2"
func extractListID(id string) string {
	if strings.HasPrefix(id, "collection_") {
		parts := strings.Split(id, "_")
		if len(parts) >= 5 {
			// collection_3_userid_listid_0 → listid = parts[3]
			return parts[3]
		}
	}
	return id
}

// 歌单添加歌曲
// POST gateway.kugou.com/cloudlist.service/v6/add_song
func AddPlaylistTracks(listid string, tracks []Song, creds *Credentials) *APIResponse {
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}
	clienttime := time.Now().Unix()

	resource := make([]map[string]interface{}, 0, len(tracks))
	for _, t := range tracks {
		albumID, _ := strconv.ParseInt(t.AlbumID, 10, 64)
		resource = append(resource, map[string]interface{}{
			"number": 1,
			"name":   t.Name,
			"hash":   t.ID,
			"size":   0,
			"sort":   0,
			"timelen": 0,
			"bitrate": 0,
			"album_id": albumID,
			// mixsongid 传0让云端用hash匹配，避免audio_id误匹配到别的歌
			"mixsongid": 0,
		})
	}

	dataMap := map[string]interface{}{
		"userid":      userid,
		"token":       creds.Token,
		"listid":      listid,
		"list_ver":    0,
		"type":        0,
		"slow_upload": 1,
		"scene":       "false;null",
		"data":        resource,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/cloudlist.service/v6/add_song",
		Data:        dataMap,
		Params: map[string]string{
			"last_time": strconv.FormatInt(clienttime, 10),
			"last_area": "gztx",
			"userid":    userid,
			"token":     creds.Token,
		},
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      userid,
		Token:       creds.Token,
	})
}

// 歌单删除歌曲
// POST gateway.kugou.com/v4/delete_songs（fileid 取自歌单内歌曲 file_id）
func DeletePlaylistTracks(listid string, fileIDs []string, creds *Credentials) *APIResponse {
	listid = NormalizeListID(listid)
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}

	resource := make([]map[string]interface{}, 0, len(fileIDs))
	for _, f := range fileIDs {
		if fid, err := strconv.ParseInt(f, 10, 64); err == nil {
			resource = append(resource, map[string]interface{}{"fileid": fid})
		}
	}

	dataMap := map[string]interface{}{
		"listid":   listid,
		"userid":   userid,
		"data":     resource,
		"type":     0,
		"token":    creds.Token,
		"list_ver": 0,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v4/delete_songs",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "cloudlist.service.kugou.com"},
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      userid,
		Token:       creds.Token,
	})
}

// ============================================================
// VIP 详情（对齐 KuGouMusicApi user_vip_detail）
// ============================================================

// 联合 VIP 详情
// GET https://kugouvip.kugou.com/v1/get_union_vip?busi_type=concept
func GetUserVIPDetail(creds *Credentials) *APIResponse {
	return SendRequest(&RequestOptions{
		Method:  "GET",
		BaseURL: "https://kugouvip.kugou.com",
		URL:     "/v1/get_union_vip",
		Params: map[string]string{
			"busi_type":         "concept",
			"opt_product_types": "dvip,qvip",
			"product_type":      "svip",
		},
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// ParseVIPDetail 解析 VIP 详情响应，提取 is_vip
func ParseVIPDetail(resp *APIResponse) map[string]interface{} {
	if resp.Error != nil || resp.Data == nil {
		return map[string]interface{}{"is_vip": false}
	}
	result := map[string]interface{}{}
	if data, ok := resp.Data["data"]; ok {
		if dm, ok := data.(map[string]interface{}); ok {
			if v, ok := dm["is_vip"]; ok { result["is_vip"] = v }
			if v, ok := dm["vip"]; ok && result["is_vip"] == nil { result["is_vip"] = v }
			if v, ok := dm["vip_type"]; ok { result["vip_type"] = fmt.Sprintf("%v", v) }
		}
	}
	if v, ok := resp.Data["is_vip"]; ok { result["is_vip"] = v }
	if v, ok := resp.Data["vip"]; ok && result["is_vip"] == nil { result["is_vip"] = v }
	// 确保 is_vip 字段存在，默认 false
	if _, ok := result["is_vip"]; !ok {
		result["is_vip"] = false
	}
	return result
}

// ============================================================
// 歌手关注 / 歌手列表
// ============================================================

// 关注 / 取消关注歌手
// POST gateway.kugou.com/followservice/v3/{follow_singer|unfollow_singer}
func FollowSinger(singerID string, follow bool, creds *Credentials) *APIResponse {
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}
	clienttime := time.Now().Unix()

	key, encryptedStr, err := CryptoAESEncrypt(map[string]interface{}{
		"singerid": singerID,
		"token":    creds.Token,
	})
	if err != nil {
		return &APIResponse{Error: err}
	}

	p, err := RSAEncrypt2(map[string]interface{}{"clienttime": clienttime, "key": key}, UseLite)
	if err != nil {
		return &APIResponse{Error: err}
	}

	path := "/followservice/v3/follow_singer"
	if !follow {
		path = "/followservice/v3/unfollow_singer"
	}

	dataMap := map[string]interface{}{
		"plat":     0,
		"userid":   userid,
		"singerid": singerID,
		"source":   7,
		"p":        strings.ToUpper(p),
		"params":   encryptedStr,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         path,
		Data:        dataMap,
		Params:      map[string]string{"clienttime": strconv.FormatInt(clienttime, 10)},
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      userid,
		Token:       creds.Token,
	})
}

// 获取用户关注歌手列表
// POST gateway.kugou.com/v4/follow_list（x-router: relationuser.kugou.com）
func GetUserFollowList(creds *Credentials) *APIResponse {
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}
	dateTime := time.Now().Unix()

	pk, err := RawRSAEncryptJSON(map[string]interface{}{
		"token":      creds.Token,
		"clienttime": dateTime,
	}, UseLite)
	if err != nil {
		return &APIResponse{Error: err}
	}

	dataMap := map[string]interface{}{
		"merge":          2,
		"need_iden_type": 1,
		"ext_params":     "k_pic,jumptype,singerid,score",
		"userid":         userid,
		"type":           0,
		"id_type":        0,
		"p":              strings.ToUpper(pk),
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v4/follow_list",
		Data:        dataMap,
		Params:      map[string]string{"plat": "1"},
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "relationuser.kugou.com"},
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      userid,
		Token:       creds.Token,
	})
}

// 歌手列表
// GET https://gateway.kugou.com/ocean/v6/singer/list
func GetSingerList(typ, sextype, musician, hotsize int, creds *Credentials) *APIResponse {
	if hotsize < 1 {
		hotsize = 60
	}
	params := map[string]string{
		"musician": strconv.Itoa(musician),
		"sextype":  strconv.Itoa(sextype),
		"showtype": "2",
		"type":     strconv.Itoa(typ),
		"hotsize":  strconv.Itoa(hotsize),
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/ocean/v6/singer/list",
		Params:      params,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// 解析 delete_list 的 AES 加密响应（arraybuffer）
func ParseDeletePlaylistResp(resp *APIResponse, key string) (map[string]interface{}, error) {
	if resp == nil {
		return nil, fmt.Errorf("空响应")
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	plain, err := PlaylistAESDecrypt(string(resp.Body), key)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := json.Unmarshal(plain, &out); err != nil {
		return nil, err
	}
	return out, nil
}
