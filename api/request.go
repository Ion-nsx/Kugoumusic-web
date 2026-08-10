package api

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ========== 平台配置 ==========
// 标准版 (appid=1005) 与概念版 (lite, appid=3116) 共用同一套签名算法，仅盐值与 appid/clientver 不同。
const (
	StdAppIDInt     = 1005
	StdClientVerInt = 20489
	StdAndroidSalt  = "OIlwieks28dk2k092lksi2UIkp"
	StdSignKeySalt  = "57ae12eb6890223e355ccfcb74edf70d"

	LiteAppIDInt     = 3116
	LiteClientVerInt = 11440
	LiteAndroidSalt  = "LnT6xpN3khm36zse0QzvmgTZ3waWdRSA"
	LiteSignKeySalt  = "185672dd44712f60bb1736df5a377e82"

	WebSalt      = "NVPh5oo715z5DIWAeQlhMDsWXXQV4hwt" // Web 版签名盐值
	RegisterSalt = "1014"                              // 设备注册签名盐值

	UserAgent = "Android15-1070-11083-46-0-DiscoveryDRADProtocol-wifi"
)

// 当前平台（默认标准版；如需概念版可将 UseLite 设为 true）
var (
	UseLite      = false
	AppID        = StdAppIDInt
	ClientVer    = StdClientVerInt
	AndroidSalt  = StdAndroidSalt
	SignKeySalt  = StdSignKeySalt
	RSAKeyPEM    = publicRSAKey
)

func init() {
	if UseLite {
		AppID = LiteAppIDInt
		ClientVer = LiteClientVerInt
		AndroidSalt = LiteAndroidSalt
		SignKeySalt = LiteSignKeySalt
		RSAKeyPEM = publicLiteRSAKey
	}
}

// SetUseLite 切换平台为标准版/概念版（概念版 appid=3116，signKey 盐不同）
func SetUseLite(lite bool) {
	UseLite = lite
	if lite {
		AppID = LiteAppIDInt
		ClientVer = LiteClientVerInt
		AndroidSalt = LiteAndroidSalt
		SignKeySalt = LiteSignKeySalt
		RSAKeyPEM = publicLiteRSAKey
	} else {
		AppID = StdAppIDInt
		ClientVer = StdClientVerInt
		AndroidSalt = StdAndroidSalt
		SignKeySalt = StdSignKeySalt
		RSAKeyPEM = publicRSAKey
	}
}

// ========== 全局设备标识缓存 ==========
var (
	devMu   sync.Mutex
	devGUID string
	devMID  string
	devDFID string
)

// 设备 GUID 持久化文件：保证服务重启后设备标识不变（mid 与登录 token 绑定，变了会导致 VIP 播放 20018/不匹配）
const guidFile = "/root/X-music/.device-guid"

// 读取或创建持久化 GUID（重启后沿用同一设备标识）
func loadOrCreateGUID() string {
	if b, err := os.ReadFile(guidFile); err == nil {
		if s := strings.TrimSpace(string(b)); s != "" {
			return s
		}
	}
	guid := newUUID()
	_ = os.WriteFile(guidFile, []byte(guid), 0600)
	return guid
}

// 生成 UUID v4（设备 GUID）
func newUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// 获取全局设备 GUID（首次生成并持久化）
func GetGUID() string {
	devMu.Lock()
	defer devMu.Unlock()
	if devGUID == "" {
		devGUID = loadOrCreateGUID()
	}
	return devGUID
}

// 获取全局设备 MID（由 GUID 计算得到）
func GetMID() string {
	devMu.Lock()
	defer devMu.Unlock()
	if devGUID == "" {
		devGUID = loadOrCreateGUID()
	}
	if devMID == "" {
		devMID = CalculateMID(devGUID)
	}
	return devMID
}

// 获取全局设备 DFID（未注册时返回 '-'）
func GetDFID() string {
	devMu.Lock()
	defer devMu.Unlock()
	if devDFID == "" {
		return "-"
	}
	return devDFID
}

// 设置全局设备 DFID
func SetDFID(dfid string) {
	devMu.Lock()
	devDFID = dfid
	devMu.Unlock()
}

// ========== 请求选项 ==========
type RequestOptions struct {
	Method  string
	URL     string
	BaseURL string
	Params  map[string]string
	Data    map[string]interface{}
	RawBody      string
	RawBodyBytes []byte // 原始字节体（云盘等二进制 body），优先级高于 RawBody/Data
	Headers      map[string]string
	Cookies string

	// 设备/用户标识（发送前自动注入）
	DFID   string
	MID    string
	Token  string
	UserID string
	// 显式控制 mid/dfid 是否参与签名：为 true 时保留空值不自动 fallback（get_my_info 等接口需 mid 为空）
	MIDSet  bool
	DFIDSet bool
	// 强制使用标准版参数（appid=1005/clientver=20489/标准盐/标准 RSA key）参与签名
	// 登录后的用户类接口（get_my_info/get_all_list 等）不受概念版 lite 支持，必须用标准版签名
	StdPlatform bool

	// 签名控制
	EncryptType  string // "android" | "web" | "register" | ""（默认 android）
	EncryptKey   bool   // 是否生成 key 参数（signParamsKey）
	KeyHash      string // 非空时 EncryptKey 改用 signKey（song url 接口）：MD5(hash+SignKeySalt+appid+mid+userid)
	ClearDefault bool   // 是否清除默认参数（dfid/mid/uuid/appid/clientver/clienttime）
	NoSign       bool   // 是否跳过签名
	ResponseType string // 响应类型，如 "arraybuffer"

	// 兼容保留字段（不再参与签名逻辑）
	NoCrypt bool
	V2      bool
	V3      bool
	IsAPI   bool

	Timeout time.Duration
}

// API 响应
type APIResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Data       map[string]interface{}
	Error      error
}

// 发送请求（自动注入设备参数与签名）
func SendRequest(opts *RequestOptions) *APIResponse {
	// 本机无 IPv6，强制使用 IPv4，避免解析到 IPv6 地址后连接卡死
	dialer := &net.Dialer{Timeout: 15 * time.Second}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", addr)
		},
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
	if opts.Timeout > 0 {
		client.Timeout = opts.Timeout
	}
	if opts.Method == "" {
		opts.Method = "GET"
	}

	// 设备标识
	dfid := opts.DFID
	if dfid == "" && !opts.DFIDSet {
		dfid = GetDFID()
	}
	mid := opts.MID
	if mid == "" && !opts.MIDSet {
		mid = GetMID()
	}
	uuid := "-"
	token := opts.Token
	userid := opts.UserID
	clienttime := time.Now().Unix()

	// 平台参数：StdPlatform 强制标准版（登录后用户接口不支持 lite）
	appidInt := AppID
	clientVerInt := ClientVer
	salt := AndroidSalt
	if opts.StdPlatform {
		appidInt = StdAppIDInt
		clientVerInt = StdClientVerInt
		salt = StdAndroidSalt
	}

	// 构建默认参数
	defaultParams := map[string]string{
		"dfid":       dfid,
		"mid":        mid,
		"uuid":       uuid,
		"appid":      strconv.Itoa(appidInt),
		"clientver":  strconv.Itoa(clientVerInt),
		"clienttime": strconv.FormatInt(clienttime, 10),
	}
	if token != "" {
		defaultParams["token"] = token
	}
	if userid != "" {
		defaultParams["userid"] = userid
	}

	// 合并参数：自定义覆盖默认
	params := map[string]string{}
	if !opts.ClearDefault {
		for k, v := range defaultParams {
			params[k] = v
		}
	}
	for k, v := range opts.Params {
		params[k] = v
	}

	// 可选：生成 key 参数（signParamsKey 或 signKey）
	if opts.EncryptKey {
		if opts.KeyHash != "" {
			// signKey 需要感知平台：标准版请求必须用标准版的 appid+盐，否则上游校验 illegal key
			keyAppid := AppID
			keySalt := SignKeySalt
			if opts.StdPlatform {
				keyAppid = StdAppIDInt
				keySalt = StdSignKeySalt
			}
			keyUserID := userid
			if keyUserID == "" {
				keyUserID = "0"
			}
			params["key"] = MD5(opts.KeyHash + keySalt + strconv.Itoa(keyAppid) + mid + keyUserID)
		} else {
			params["key"] = signParamsKey(strconv.FormatInt(clienttime, 10))
		}
	}

	// 序列化请求体
	isRawBytes := len(opts.RawBodyBytes) > 0
	var data []byte
	if isRawBytes {
		data = opts.RawBodyBytes
	} else if opts.RawBody != "" {
		data = []byte(opts.RawBody)
	} else if opts.Data != nil {
		data, _ = json.Marshal(opts.Data)
	}

	// 生成签名
	if !opts.NoSign && params["signature"] == "" {
		et := opts.EncryptType
		if et == "" {
			et = "android"
		}
		switch et {
		case "register":
			params["signature"] = signatureRegisterParams(params)
		case "web":
			params["signature"] = signatureWebParams(params)
		default:
			params["signature"] = signatureAndroidParamsWithSalt(params, data, salt)
		}
	}

	// 构造 URL
	reqURL := opts.BaseURL + opts.URL
	if strings.Contains(reqURL, "?") {
		reqURL = reqURL + "&" + encodeParams(params)
	} else {
		reqURL = reqURL + "?" + encodeParams(params)
	}

	var bodyReader io.Reader
	if opts.Method == "POST" && len(data) > 0 {
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(opts.Method, reqURL, bodyReader)
	if err != nil {
		return &APIResponse{Error: err}
	}

	// 请求头：设备信息 + 内部标识
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("dfid", dfid)
	req.Header.Set("clienttime", strconv.FormatInt(clienttime, 10))
	req.Header.Set("mid", mid)
	req.Header.Set("kg-rc", "1")
	req.Header.Set("kg-thash", "5d816a0")
	req.Header.Set("kg-rec", "1")
	req.Header.Set("kg-rf", "B9EDA08A64250DEFFBCADDEE00F8F25F")

	if opts.Method == "POST" {
		if isRawBytes {
			req.Header.Set("Content-Type", "application/octet-stream")
		} else {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	if opts.Cookies != "" {
		req.Header.Set("Cookie", opts.Cookies)
	}

	for k, v := range opts.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return &APIResponse{Error: err}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIResponse{Error: err}
	}

	result := &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}

	// 尝试解析 JSON
	var dataObj map[string]interface{}
	if err := json.Unmarshal(body, &dataObj); err == nil {
		result.Data = dataObj
	}

	return result
}

// ========== 签名算法 ==========

// Android 版 signature 签名
// MD5(盐 + 排序key=value拼接 + data + 盐)，data 为请求体字节
func signatureAndroidParams(params map[string]string, data []byte) string {
	return signatureAndroidParamsWithSalt(params, data, AndroidSalt)
}

// 指定盐值的 Android 版签名（供 StdPlatform 等场景使用标准版盐）
func signatureAndroidParamsWithSalt(params map[string]string, data []byte, salt string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	paramsString := sb.String()

	if len(data) > 0 {
		h := md5.New()
		h.Write([]byte(salt))
		h.Write([]byte(paramsString))
		h.Write(data)
		h.Write([]byte(salt))
		return hex.EncodeToString(h.Sum(nil))
	}
	return MD5(salt + paramsString + salt)
}

// Web 版 signature 签名
// MD5(盐 + 排序key=value拼接 + 盐)
func signatureWebParams(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	return MD5(WebSalt + sb.String() + WebSalt)
}

// 设备注册 signature 签名
// MD5("1014" + 排序value拼接 + "1014")
func signatureRegisterParams(params map[string]string) string {
	values := make([]string, 0, len(params))
	for _, v := range params {
		values = append(values, v)
	}
	sort.Strings(values)
	return MD5(RegisterSalt + strings.Join(values, "") + RegisterSalt)
}

// 参数密钥签名 signParamsKey
// MD5(appid + 盐 + clientver + data)，使用当前平台配置
func signParamsKey(data string) string {
	return MD5(fmt.Sprintf("%d%s%d%s", AppID, AndroidSalt, ClientVer, data))
}

// 指定 appid/clientver 的 signParamsKey（云盘等接口需要显式传参）
func signParamsKeyWithApp(data string, appid, clientver int) string {
	// 根据 appid 判断平台盐值
	salt := StdAndroidSalt
	if appid == LiteAppIDInt {
		salt = LiteAndroidSalt
	}
	return MD5(fmt.Sprintf("%d%s%d%s", appid, salt, clientver, data))
}

// 云盘 URL 签名 signCloudKey
// MD5("musicclound" + hash + pid + signCloudKeySalt)
func signCloudKey(hash string, pid int) string {
	const signCloudKeySalt = "ebd1ac3134c880bda6a2194537843caa0162e2e7"
	return MD5(fmt.Sprintf("musicclound%s%d%s", hash, pid, signCloudKeySalt))
}

// 请求密钥签名 signKey
// MD5(hash + 盐 + appid + mid + userid)，userid 空时按 "0" 处理（与官方算法一致）
func signKey(hash, mid, userid string) string {
	if userid == "" {
		userid = "0"
	}
	return MD5(hash + SignKeySalt + strconv.Itoa(AppID) + mid + userid)
}

// 通用 sign 签名（signParams）
// MD5(排序key+value拼接 + data + 盐)
func signParams(params map[string]string, data string) string {
	const signParamsSalt = "R6snCXJgbCaj9WFRJKefTMIFp0ey6Gza"
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}
	return MD5(sb.String() + data + signParamsSalt)
}

// 编码 URL 参数
func encodeParams(params map[string]string) string {
	parts := make([]string, 0, len(params))
	for k, v := range params {
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
	}
	return strings.Join(parts, "&")
}

// 从 Authorization header 解析凭证
type Credentials struct {
	Token  string
	UserID string
	DFID   string
	MID    string
	GUID   string
	DEV    string
	MAC    string
}

func ParseCredentials(authHeader string) *Credentials {
	c := &Credentials{}
	if authHeader == "" {
		return c
	}

	parts := strings.Split(authHeader, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "token=") {
			c.Token = strings.TrimPrefix(part, "token=")
		} else if strings.HasPrefix(part, "userid=") {
			c.UserID = strings.TrimPrefix(part, "userid=")
		} else if strings.HasPrefix(part, "dfid=") {
			c.DFID = strings.TrimPrefix(part, "dfid=")
		} else if strings.HasPrefix(part, "mid=") || strings.HasPrefix(part, "KUGOU_API_MID=") {
			// 处理 KUGOU_API_MID= 前缀
			if strings.HasPrefix(part, "KUGOU_API_MID=") {
				c.MID = strings.TrimPrefix(part, "KUGOU_API_MID=")
			} else {
				c.MID = strings.TrimPrefix(part, "mid=")
			}
		} else if strings.HasPrefix(part, "guid=") || strings.HasPrefix(part, "KUGOU_API_GUID=") {
			if strings.HasPrefix(part, "KUGOU_API_GUID=") {
				c.GUID = strings.TrimPrefix(part, "KUGOU_API_GUID=")
			} else {
				c.GUID = strings.TrimPrefix(part, "guid=")
			}
		} else if strings.HasPrefix(part, "KUGOU_API_DEV=") {
			c.DEV = strings.TrimPrefix(part, "KUGOU_API_DEV=")
		} else if strings.HasPrefix(part, "KUGOU_API_MAC=") {
			c.MAC = strings.TrimPrefix(part, "KUGOU_API_MAC=")
		}
	}
	return c
}

// 将凭证注入到参数中
func (c *Credentials) InjectParams(params map[string]string) map[string]string {
	if params == nil {
		params = make(map[string]string)
	}
	params["token"] = c.Token
	params["userid"] = c.UserID
	params["dfid"] = c.DFID
	params["mid"] = c.MID
	params["guid"] = c.GUID
	return params
}

// 将凭证转换为 Cookie 字符串
func (c *Credentials) ToCookie() string {
	parts := []string{
		fmt.Sprintf("kg_mid=%s", c.MID),
		fmt.Sprintf("userid=%s", c.UserID),
		fmt.Sprintf("token=%s", c.Token),
	}
	if c.DFID != "" {
		parts = append(parts, fmt.Sprintf("dfid=%s", c.DFID))
	}
	return strings.Join(parts, "; ")
}

// 请求错误
type RequestError struct {
	Code    int
	Message string
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("request error: %d %s", e.Code, e.Message)
}

// JSON 响应包装
func JSONResponse(data interface{}, err error) map[string]interface{} {
	if err != nil {
		return map[string]interface{}{
			"status": 0,
			"error":  err.Error(),
		}
	}
	return map[string]interface{}{
		"status": 1,
		"data":   data,
	}
}
