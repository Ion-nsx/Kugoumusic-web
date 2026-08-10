package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// 账号密码登录（新接口）
// POST https://gateway.kugou.com/v9/login_by_pwd（x-router: login.user.kugou.com）
// 密码经 AES 加密（params），临时 key 经裸 RSA 加密（pk）
// 响应 status=1 时 data.secu_params 为 AES 密文，解密后得到 token/userid/vip_type 等
func Login(username, password string, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	dateNow := time.Now().UnixMilli()

	// AES 加密 {pwd, code, clienttime_ms}
	enc, err := AESEncryptJSON(map[string]interface{}{
		"pwd": password, "code": "", "clienttime_ms": dateNow,
	})
	if err != nil {
		return &APIResponse{Error: err}
	}

	// 裸 RSA 加密 {clienttime_ms, key}
	pk, err := RawRSAEncryptJSON(map[string]interface{}{
		"clienttime_ms": dateNow, "key": enc.Key,
	}, UseLite)
	if err != nil {
		return &APIResponse{Error: err}
	}

	dataMap := map[string]interface{}{
		"plat":           1,
		"support_multi":  1,
		"clienttime_ms":  dateNow,
		"t1":             loginT1,
		"t2":             loginT2,
		"t3":             "MCwwLDAsMCwwLDAsMCwwLDA=",
		"username":       username,
		"params":         enc.Str,
		"pk":             strings.ToUpper(pk),
	}

	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}
	dfid := creds.DFID
	if dfid == "" {
		dfid = GetDFID()
	}

	resp := SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://gateway.kugou.com",
		URL:         "/v9/login_by_pwd",
		Data:        dataMap,
		EncryptType: "android",
		Headers:     map[string]string{"x-router": "login.user.kugou.com"},
		MID:         mid,
		DFID:        dfid,
	})

	// 解密 secu_params，将 token/userid/vip_type 等合并到 data
	if resp.Error == nil && resp.Data != nil {
		if data, ok := resp.Data["data"].(map[string]interface{}); ok {
			if sp, ok := data["secu_params"]; ok {
				if dec, err := AESDecryptJSON(fmt.Sprintf("%v", sp), enc.Key); err == nil {
					for k, v := range dec {
						data[k] = v
					}
				}
			}
		}
	}

	return resp
}

// login_by_pwd 固定预加密串（参考 login.js）
const (
	loginT1 = "562a6f12a6e803453647d16a08f5f0c2ff7eee692cba2ab74cc4c8ab47fc467561a7c6b586ce7dc46a63613b246737c03a1dc8f8d162d8ce1d2c71893d19f1d4b797685a4c6d3d81341cbde65e488c4829a9b4d42ef2df470eb102979fa5adcdd9b4eecfea8b909ff7599abeb49867640f10c3c70fc444effca9d15db44a9a6c907731e2bb0f22cd9b3536380169995693e5f0e2424e3378097d3813186e3fe96bbe7023808a0981b4e2b6135a76faac"
	loginT2 = "31c4daf4cf480169ccea1cb7d4a209295865a9d2b788510301694db229b87807469ea0d41b4d4b9173c2151da7294aeebfc9738df154bbdf11a4e117bb5dff6a3af8ce5ce333e681c1f29a44038f27567d58992eb81283e080778ac77db1400fdf49b7cf7e26be2e5af4da7830cc3be4"
)

// 默认开发设备标识（DEV），随机 10 位大写字符串
var serverDev = RandomString(10)

// ========== 扫码登录（新接口 login-user.kugou.com） ==========
// 参考：MakcRe/lfhy module/login_qr_key.js、login_qr_check.js
// 生成二维码：GET https://login-user.kugou.com/v2/qrcode
// 响应: data.qrcode=key, data.qrcode_img=base64 图片
// 检查状态：GET https://login-user.kugou.com/v2/get_userinfo_qrcode
// 响应: data.status（0过期/1等待扫码/2待确认/4登录成功带token）

const (
	QRSrcAppID = "2919" // 酷狗官方 srcappid
	QRWebAppID = "1014" // 扫码接口请求 appid
	QRAppID    = "3116" // 扫码签发 token 的 appid：概念版(lite)，保证 token 与概念版接口/VIP 兼容
)

// 二维码图片缓存（qrGet 接口使用）
var (
	qrImgMu    sync.Mutex
	qrImgCache = map[string]string{} // key -> base64 图片
)

// 获取二维码 key + 图片
// GET https://login-user.kugou.com/v2/qrcode
func GetQRKey(creds *Credentials) *APIResponse {
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}

	resp := SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://login-user.kugou.com",
		URL:         "/v2/qrcode",
		Params: map[string]string{
			"appid":      QRWebAppID,
			"type":       "1",
			"plat":       "4",
			"qrcode_txt": "https://h5.kugou.com/apps/loginQRCode/html/index.html?appid=3116&",
			"srcappid":   QRSrcAppID,
		},
		EncryptType: "web",
		MID:         mid,
		DFID:        "-",
	})

	// 缓存二维码图片，供后续 qrGet 使用
	if resp.Error == nil && resp.Data != nil {
		if data, ok := resp.Data["data"].(map[string]interface{}); ok {
			if img, ok := data["qrcode_img"]; ok {
				key, _ := data["qrcode"].(string)
				if key != "" {
					qrImgMu.Lock()
					qrImgCache[key] = fmt.Sprintf("%v", img)
					qrImgMu.Unlock()
				}
			}
		}
	}

	return resp
}

// 获取二维码图片（从缓存返回 base64）
func GetQRImage(key string, creds *Credentials) *APIResponse {
	qrImgMu.Lock()
	img, ok := qrImgCache[key]
	qrImgMu.Unlock()

	if !ok {
		return &APIResponse{Data: map[string]interface{}{"status": 0, "error": "二维码已失效，请重新获取"}}
	}
	return &APIResponse{Data: map[string]interface{}{"status": 1, "data": img}}
}

type QRCheckResult struct {
	Status   int    `json:"status"`
	Code     string `json:"code"`
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	MID      string `json:"mid"`
	GUID     string `json:"guid"`
	DFID     string `json:"dfid"`
	Expire   int64  `json:"expire"`
	User     User   `json:"user"`
	VIPToken string `json:"vip_token"`
	VIPType  string `json:"vip_type"`
}

// 检查扫码状态
// GET https://login-user.kugou.com/v2/get_userinfo_qrcode
// 参数: plat=4, appid, srcappid, qrcode=key
func CheckQRStatus(key string, creds *Credentials) *APIResponse {
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}

	return SendRequest(&RequestOptions{
		Method:      "GET",
		BaseURL:     "https://login-user.kugou.com",
		URL:         "/v2/get_userinfo_qrcode",
		Params: map[string]string{
			"plat":     "4",
			"appid":    QRAppID,
			"srcappid": QRSrcAppID,
			"qrcode":   key,
		},
		EncryptType: "web",
		MID:         mid,
		DFID:        "-",
	})
}

// 解析 QR check 响应（v2/get_userinfo_qrcode）
// data.status: 0=二维码过期 1=等待扫码 2=待确认 4=授权登录成功（返回 token/userid）
func ParseQRCheck(resp *APIResponse) *QRCheckResult {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	result := &QRCheckResult{}

	if status, ok := resp.Data["status"]; ok {
		result.Status = int(toFloat64(status))
	}

	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return result
		}

		if status, ok := dataMap["status"]; ok {
			result.Status = int(toFloat64(status))
		}
		if code, ok := dataMap["code"]; ok {
			result.Code = fmt.Sprintf("%v", code)
		}
		if token, ok := dataMap["token"]; ok {
			result.Token = fmt.Sprintf("%v", token)
		}
		if uid, ok := dataMap["userid"]; ok {
			result.UserID = playlistIDString(uid)
		} else if uid, ok := dataMap["user_id"]; ok {
			result.UserID = playlistIDString(uid)
		}
		if mid, ok := dataMap["mid"]; ok {
			result.MID = fmt.Sprintf("%v", mid)
		}
		if guid, ok := dataMap["guid"]; ok {
			result.GUID = fmt.Sprintf("%v", guid)
		}
		if dfid, ok := dataMap["dfid"]; ok {
			result.DFID = fmt.Sprintf("%v", dfid)
		}
		if expire, ok := dataMap["expire"]; ok {
			result.Expire = int64(toFloat64(expire))
		}
		if vtk, ok := dataMap["vip_token"]; ok {
			result.VIPToken = fmt.Sprintf("%v", vtk)
		}
		if vt, ok := dataMap["vip_type"]; ok {
			result.VIPType = playlistIDString(vt)
		} else if vt, ok := dataMap["vipType"]; ok {
			result.VIPType = playlistIDString(vt)
		}
	}
	// 打印 QR check 返回的 VIP 字段用于调试
	if result.Status == 4 {
		fmt.Printf("[QR] vip_token=%q vip_type=%q\n", result.VIPToken, result.VIPType)
	}

	return result
}

// 手机验证码登录（新接口）
// POST https://loginserviceretry.kugou.com/v7/login_by_verifycode
// 概念版(lite)需带 t1/t2（设备信息 AES 加密）、key(signParamsKey)、pk(裸 RSA)
// 响应 status=1 时 data.secu_params 为 AES 密文，解密后得到 token/userid
// userid 非空时指定登录账号（手机号绑定多账号时需选号，否则上游返回 34175）
func LoginCellphone(mobile, code, areaCode, userid string, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	dateNow := time.Now().UnixMilli()

	// AES 加密 {mobile, code}
	enc, err := AESEncryptJSON(map[string]interface{}{
		"mobile": mobile, "code": code,
	})
	if err != nil {
		return &APIResponse{Error: err}
	}

	// 手机号掩码：前2位 + ***** + 末位
	mobileMask := ""
	if len(mobile) >= 3 {
		mobileMask = mobile[:2] + "*****" + mobile[len(mobile)-1:]
	} else {
		mobileMask = mobile
	}

	// 设备标识（概念版 t1/t2 使用）
	guid := creds.GUID
	if guid == "" {
		guid = GetGUID()
	}
	dev := creds.DEV
	if dev == "" {
		dev = serverDev
	}
	mac := creds.MAC
	if mac == "" {
		mac = "02:00:00:00:00:00"
	}
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}
	dfid := creds.DFID
	if dfid == "" {
		dfid = RandomString(24)
	}

	// 概念版固定 key/iv 的 AES 加密（参考 login_cellphone.js）
	t1Str, _ := AESEncrypt([]byte("|"+fmt.Sprintf("%d", dateNow)), liteT1Key, liteT1Iv)
	t2Str, _ := AESEncrypt([]byte(fmt.Sprintf("%s|0f607264fc6318a92b9e13c65db7cd3c|%s|%s|%d", guid, mac, dev, dateNow)), liteT2Key, liteT2Iv)

	// 裸 RSA 加密 {clienttime_ms, key}
	pk, err := RawRSAEncryptJSON(map[string]interface{}{
		"clienttime_ms": dateNow, "key": enc.Key,
	}, UseLite)
	if err != nil {
		return &APIResponse{Error: err}
	}

	dataMap := map[string]interface{}{
		"plat":          1,
		"support_multi": 1,
		"t1":            t1Str.Str,
		"t2":            t2Str.Str,
		"clienttime_ms": dateNow,
		"mobile":        mobileMask,
		"key":           signParamsKey(fmt.Sprintf("%d", dateNow)),
		"pk":            strings.ToUpper(pk),
		"params":        enc.Str,
		"dfid":          dfid,
		"dev":           dev,
		"gitversion":    "5f0b7c4",
	}
	if userid != "" {
		dataMap["userid"] = userid
	}

	resp := SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "https://loginserviceretry.kugou.com",
		URL:         "/v7/login_by_verifycode",
		Data:        dataMap,
		EncryptType: "android",
		Headers: map[string]string{
			"support-calm": "1",
			"User-Agent":   "Android16-1070-11440-130-0-LOGIN-wifi",
		},
		MID:    mid,
		DFID:   dfid,
		UserID: creds.UserID,
		Token:  creds.Token,
	})

	// 解密 secu_params，合并 token/userid 等
	if resp.Error == nil && resp.Data != nil {
		if data, ok := resp.Data["data"].(map[string]interface{}); ok {
			if sp, ok := data["secu_params"]; ok {
				if dec, err := AESDecryptJSON(fmt.Sprintf("%v", sp), enc.Key); err == nil {
					for k, v := range dec {
						data[k] = v
					}
				}
			}
		}
	}

	return resp
}

// 概念版(lite) t1/t2 AES 固定密钥（参考 login_cellphone.js）
const (
	liteT1Key = "5e4ef500e9597fe004bd09a46d8add98"
	liteT1Iv  = "04bd09a46d8add98"
	liteT2Key = "fd14b35e3f81af3817a20ae7adae7020"
	liteT2Iv  = "17a20ae7adae7020"
)

// 发送验证码（新接口）
// POST http://login.user.kugou.com/v7/send_mobile_code
func SendCaptcha(mobile, areaCode string, creds *Credentials) *APIResponse {
	if creds == nil {
		creds = &Credentials{}
	}
	mid := creds.MID
	if mid == "" {
		mid = GetMID()
	}

	dataMap := map[string]interface{}{
		"businessid": 5,
		"mobile":     mobile,
		"plat":       3,
	}

	return SendRequest(&RequestOptions{
		Method:      "POST",
		BaseURL:     "http://login.user.kugou.com",
		URL:         "/v7/send_mobile_code",
		Data:        dataMap,
		EncryptType: "android",
		MID:         mid,
		DFID:        "-",
	})
}

// 解析登录响应
// 返回登录结果（token, userid, dfid, mid, guid）
func ParseLogin(resp *APIResponse) *LoginResult {
	if resp.Error != nil || resp.Data == nil {
		return nil
	}

	result := &LoginResult{}

	// 检查登录状态
	if status, ok := resp.Data["status"]; ok {
		statusInt := int(toFloat64(status))
		if statusInt != 1 {
			return nil
		}
	}

	if data, ok := resp.Data["data"]; ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}

		if token, ok := dataMap["token"]; ok {
			result.Token = fmt.Sprintf("%v", token)
		}
		if uid, ok := dataMap["user_id"]; ok {
			result.UserID = playlistIDString(uid)
		} else if uid, ok := dataMap["userid"]; ok {
			result.UserID = playlistIDString(uid)
		}
		if dfid, ok := dataMap["dfid"]; ok {
			result.DFID = fmt.Sprintf("%v", dfid)
		}
		if mid, ok := dataMap["mid"]; ok {
			result.MID = fmt.Sprintf("%v", mid)
		}
		if guid, ok := dataMap["guid"]; ok {
			result.GUID = fmt.Sprintf("%v", guid)
		}
		if expire, ok := dataMap["expire"]; ok {
			result.Expire = int64(toFloat64(expire))
		}
		if vt, ok := dataMap["vip_type"]; ok {
			result.VIPType = playlistIDString(vt)
		} else if vt, ok := dataMap["vipType"]; ok {
			result.VIPType = playlistIDString(vt)
		}
		if vtk, ok := dataMap["vip_token"]; ok {
			result.VIPToken = fmt.Sprintf("%v", vtk)
		}

		// 解析用户信息
		if user, ok := dataMap["user"]; ok {
			userMap, ok := user.(map[string]interface{})
			if ok {
				result.User = parseUser(userMap)
			}
		} else {
			result.User = parseUser(dataMap)
		}
	}

	return result
}

func parseUser(data map[string]interface{}) User {
	user := User{}

	if id, ok := data["user_id"]; ok {
		user.ID = fmt.Sprintf("%v", id)
	} else if id, ok := data["id"]; ok {
		user.ID = fmt.Sprintf("%v", id)
	}
	if name, ok := data["nickname"]; ok {
		user.Name = fmt.Sprintf("%v", name)
	} else if name, ok := data["username"]; ok {
		user.Name = fmt.Sprintf("%v", name)
	} else if name, ok := data["user_name"]; ok {
		user.Name = fmt.Sprintf("%v", name)
	}
	if avatar, ok := data["avatar"]; ok {
		user.Avatar = fmt.Sprintf("%v", avatar)
	} else if avatar, ok := data["user_pic"]; ok {
		user.Avatar = fmt.Sprintf("%v", avatar)
	} else if avatar, ok := data["pic"]; ok {
		user.Avatar = fmt.Sprintf("%v", avatar)
	} else if avatar, ok := data["k_pic"]; ok {
		user.Avatar = fmt.Sprintf("%v", avatar)
	}
	if sex, ok := data["sex"]; ok {
		user.Sex = fmt.Sprintf("%v", sex)
	}
	if level, ok := data["level"]; ok {
		user.Level = int(toFloat64(level))
	}
	if vip, ok := data["is_vip"]; ok {
		user.IsVIP = toBool(vip)
	} else if vip, ok := data["vip"]; ok {
		user.IsVIP = toBool(vip)
	}
	if vt, ok := data["vip_type"]; ok {
		user.VIPType = fmt.Sprintf("%v", vt)
	}
	if email, ok := data["email"]; ok {
		user.Email = fmt.Sprintf("%v", email)
	}

	return user
}

// 持久化存储登录凭证
type AuthStore struct {
	Token   string `json:"token"`
	UserID  string `json:"user_id"`
	DFID    string `json:"dfid"`
	MID     string `json:"mid"`
	GUID    string `json:"guid"`
	Expire  int64  `json:"expire"`
	User    User   `json:"user"`
}

// 序列化凭证到 JSON
func (a *AuthStore) ToJSON() string {
	data, _ := json.Marshal(a)
	return string(data)
}

// 将凭证转为 Credentials
func (a *AuthStore) ToCredentials() *Credentials {
	return &Credentials{
		Token:  a.Token,
		UserID: a.UserID,
		DFID:   a.DFID,
		MID:    a.MID,
		GUID:   a.GUID,
	}
}

// 转换为 Authorization header 格式
func (a *AuthStore) ToAuthHeader() string {
	return fmt.Sprintf("token=%s;userid=%s;dfid=%s;mid=%s;guid=%s",
		a.Token, a.UserID, a.DFID, a.MID, a.GUID)
}

// 从 JSON 反序列化凭证
func AuthStoreFromJSON(data string) *AuthStore {
	var store AuthStore
	if err := json.Unmarshal([]byte(data), &store); err != nil {
		return nil
	}
	return &store
}

// 管理端登录（用于管理后台）
// 返回 base64 编码的 token 字符串
func EncodeCredential(token, userid string) string {
	raw := fmt.Sprintf("%s:%s", token, userid)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func DecodeCredential(encoded string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid credential format")
	}
	return parts[0], parts[1], nil
}