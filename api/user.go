package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 用户详情（新接口）
// POST https://gateway.kugou.com/v3/get_my_info（x-router: usercenter.kugou.com）
// 参数: p(RSA加密) + visit_time/usertype/userid；响应 data 含用户信息
func GetUserDetail(creds *Credentials) *APIResponse {
	dateTime := time.Now().Unix()
	userid := creds.UserID
	if userid == "" {
		userid = "0"
	}
	pk := ""

	// RSA 加密 token+clienttime
	// 注意：需保持与参考 JS 一致的 JSON 字段顺序（token 在前、clienttime 在后），
	// 否则 Go map 序列化会按字母序排成 {"clienttime":..,"token":..}，RSA 明文不同导致上游解不开
	// get_my_info 仅支持标准版平台，RSA 用标准版公钥
	// 概念版 token 需用概念版(lite) RSA 公钥 + lite 签名
	jsonStr := fmt.Sprintf(`{"token":%q,"clienttime":%d}`, creds.Token, dateTime)
	enc, err := RawRSAEncrypt([]byte(jsonStr), true)
	if err == nil {
		pk = strings.ToUpper(enc)
	}

	dataMap := map[string]interface{}{
		"visit_time": dateTime,
		"usertype":   1,
		"p":          pk,
		"userid":     userid,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v3/get_my_info",
		Data:        dataMap,
		Params:      map[string]string{"plat": "1"},
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "usercenter.kugou.com"},
		MID:         "undefined", // 对齐 MoeKoeMusic：mid 传字符串 "undefined"（空会 20006，真实 MID 会 20018）
		DFID:        "-",         // 对齐 MoeKoeMusic：dfid 用 "-"，真实 dfid 会导致 20006
		MIDSet:      true,        // 不自动 fallback
		UserID:      userid,
		Token:       creds.Token,
	})
}

// 用户歌单列表（新接口）
// POST https://gateway.kugou.com/v7/get_all_list（x-router: cloudlist.service.kugou.com）
// 参数: userid/token/total_ver/type(2=全部)/page/pagesize
func GetUserPlaylist(userid, token string, page, pagesize int, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	if userid == "" {
		userid = creds.UserID
	}
	if userid == "" {
		userid = "0"
	}
	if token == "" {
		token = creds.Token
	}

	dataMap := map[string]interface{}{
		"userid":    userid,
		"token":     token,
		"total_ver": 979,
		"type":      2,
		"page":      page,
		"pagesize":  pagesize,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v7/get_all_list",
		Data:        dataMap,
		Params:      map[string]string{"plat": "1", "userid": userid, "token": token},
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "cloudlist.service.kugou.com"},
		MID:         "undefined", // 对齐 MoeKoeMusic：mid 传字符串 "undefined"
		DFID:        "-",         // 对齐 MoeKoeMusic：dfid 用 "-"
		MIDSet:      true,        // 不自动 fallback
		UserID:      userid,
		Token:       token,
	})
}

// 用户收藏歌单
// GET http://mobilecdn.kugou.com/api/v3/user/favorite
// 参数: userid, token, page, pagesize
func GetUserFavorite(userid, token string, page, pagesize int, creds *Credentials) *APIResponse {
	params := creds.InjectParams(map[string]string{
		"userid":   userid,
		"token":    token,
		"page":     strconv.Itoa(page),
		"pagesize": strconv.Itoa(pagesize),
		"plat":     "0",
		"version":  "1",
	})

	return SendRequest(&RequestOptions{
		Method: "GET",
		URL:    "http://mobilecdn.kugou.com/api/v3/user/favorite",
		Params: params,
	})
}

// 每日领 VIP（新接口）
// POST https://gateway.kugou.com/youth/v1/ad/play_report
// 需登录；data 含 ad_id/play_start/play_end
func GetDailyVIP(creds *Credentials) *APIResponse {
	timeMS := time.Now().UnixMilli()

	dataMap := map[string]interface{}{
		"ad_id":      12307537187,
		"play_end":   timeMS,
		"play_start": timeMS - 30000,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/youth/v1/ad/play_report",
		Data:        dataMap,
		EncryptType: "android",
		MID:         creds.MID,
		DFID:        creds.DFID,
		UserID:      creds.UserID,
		Token:       creds.Token,
	})
}

// 解析用户详情响应
func ParseUserDetail(resp *APIResponse) *User {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}

		user := parseUser(dataMap)
		return &user
	}

	return nil
}

// 解析用户歌单列表响应
func ParseUserPlaylist(resp *APIResponse) []Playlist {
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

// 解析每日 VIP 响应
func ParseDailyVIP(resp *APIResponse) (bool, string) {
	if resp.Error != nil || resp.Data == nil {
		return false, ""
	}

	if status, ok := resp.Data["status"]; ok {
		statusInt := int(toFloat64(status))
		if statusInt == 1 {
			return true, ""
		}
	}

	errMsg := ""
	if err, ok := resp.Data["error"]; ok {
		errMsg = fmt.Sprintf("%v", err)
	} else if msg, ok := resp.Data["msg"]; ok {
		errMsg = fmt.Sprintf("%v", msg)
	}

	return false, errMsg
}