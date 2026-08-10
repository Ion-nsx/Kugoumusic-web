package api

import (
	"strconv"
	"strings"
	"time"
)

// 用户听歌历史排行 — 按播放次数排名
// POST https://listenservice.kugou.com/v2/get_list
// 需登录: token + userid，RSA 加密 {clienttime, token}
func GetUserListen(userid, token string, listType int) *APIResponse {
	clienttimeSec := time.Now().Unix()

	p, err := RawRSAEncryptJSON(map[string]interface{}{
		"clienttime": clienttimeSec,
		"token":      token,
	}, true)
	if err != nil {
		return &APIResponse{Error: err}
	}

	dataMap := map[string]interface{}{
		"t_userid":  userid,
		"userid":    userid,
		"list_type": listType,
		"area_code": 1,
		"cover":     2,
		"p":         strings.ToUpper(p),
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://listenservice.kugou.com",
		URL:         "/v2/get_list",
		Data:        dataMap,
		Params:      map[string]string{"clienttime": strconv.FormatInt(clienttimeSec, 10), "plat": "0"},
		EncryptType: "android",
		MID:         GetMID(),
		DFID:        GetDFID(),
		UserID:      userid,
		Token:       token,
	})
}

// 听歌历史 — 获取用户播放历史排行
// POST https://gateway.kugou.com/playhistory/v1/get_songs
// 需登录: token + userid
// 可选 bp 参数用于分页
func GetUserHistory(userid, token, bp string) *APIResponse {
	dataMap := map[string]interface{}{
		"token":            token,
		"userid":           userid,
		"source_classify":  "app",
		"to_subdivide_sr":  1,
	}

	if bp != "" {
		dataMap["bp"] = bp
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/playhistory/v1/get_songs",
		Data:        dataMap,
		EncryptType: "android",
		MID:         GetMID(),
	})
}
