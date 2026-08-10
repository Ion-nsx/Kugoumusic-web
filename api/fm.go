package api

import (
	"strconv"
	"time"
)

// 猜你喜欢 — 获取个性化推荐歌曲
// POST https://gateway.kugou.com/v2/personal_recommend（x-router: persnfm.service.kugou.com）
// 参数: action/mode/song_pool_id/remain_songcnt/fakem + signParamsKey（毫秒级时间戳）
// 可选 hash/songid/playtime —— 用于播放后反馈（已听过/跳过等），下次推荐会避开已听歌曲
func GetPersonalFM(action, songPoolID string, remainSongCnt int, creds *Credentials) *APIResponse {
	clienttimeMs := time.Now().UnixMilli()

	dataMap := map[string]interface{}{
		"appid":                   AppID,
		"clienttime":              clienttimeMs,
		"mid":                     GetMID(),
		"action":                  action,
		"recommend_source_locked": 0,
		"song_pool_id":           toFloat64(songPoolID),
		"callerid":                0,
		"m_type":                  1,
		"platform":                "android",
		"area_code":               1,
		"remain_songcnt":          remainSongCnt,
		"clientver":               ClientVer,
		"is_overplay":             0,
		"mode":                    "normal",
		"fakem":                   "ca981cfc583a4c37f28d2d49000013c16a0a",
		"key":                     signParamsKey(strconv.FormatInt(clienttimeMs, 10)),
	}

	if creds != nil {
		if creds.UserID != "" && creds.UserID != "0" {
			dataMap["userid"] = creds.UserID
			dataMap["kguid"] = creds.UserID
		}
		if creds.Token != "" {
			dataMap["token"] = creds.Token
		}
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v2/personal_recommend",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "persnfm.service.kugou.com"},
		MID:         GetMID(),
	})
}
