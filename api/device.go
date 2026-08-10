package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// 设备信息
type DeviceInfo struct {
	MID   string
	GUID  string
	DFID  string
	UUID  string
	AppID string
}

// 注册设备
// POST https://userservice.kugou.com/risk/v2/r_register_dev
// 设备信息用歌单 AES 加密（playlistAesEncrypt），AES key 用 RSA 加密（rsaEncrypt2）
func RegisterDevice() (*DeviceInfo, error) {
	guid := GetGUID()
	mid := GetMID()
	uuid := newUUID()

	// 设备信息数据
	dataMap := map[string]interface{}{
		"availableRamSize":  4983533568,
		"availableRomSize":  48114719,
		"availableSDSize":   48114717,
		"basebandVer":       "",
		"batteryLevel":      100,
		"batteryStatus":     3,
		"brand":             "Xiaomi",
		"buildSerial":       "unknown",
		"device":            "marble",
		"imei":              guid,
		"imsi":              "",
		"manufacturer":      "Xiaomi",
		"uuid":              uuid,
		"accelerometer":     false,
		"accelerometerValue": "",
		"gravity":           false,
		"gravityValue":      "",
		"gyroscope":         false,
		"gyroscopeValue":    "",
		"light":             false,
		"lightValue":        "",
		"magnetic":          false,
		"magneticValue":     "",
		"orientation":       false,
		"orientationValue":  "",
		"pressure":          false,
		"pressureValue":     "",
		"step_counter":      false,
		"step_counterValue": "",
		"temperature":       false,
		"temperatureValue":  "",
	}

	// 歌单 AES 加密设备信息
	aesResult, err := PlaylistAESEncrypt(dataMap)
	if err != nil {
		return nil, fmt.Errorf("playlist aes encrypt failed: %w", err)
	}

	// RSA 加密 AES key（rsaEncrypt2：RSAES-PKCS1-V1_5 输出 hex）
	rsaData := map[string]interface{}{
		"aes":   aesResult.Key,
		"uid":   0,
		"token": "",
	}
	rsaJSON, err := json.Marshal(rsaData)
	if err != nil {
		return nil, fmt.Errorf("marshal rsa data failed: %w", err)
	}
	p, err := RSAEncrypt(rsaJSON, UseLite)
	if err != nil {
		return nil, fmt.Errorf("rsa encrypt failed: %w", err)
	}

	// 请求参数：part=1, platid=1, p=RSA加密结果
	params := map[string]string{
		"part":   "1",
		"platid": "1",
		"p":      p,
	}

	resp := SendRequest(&RequestOptions{
		Method:       "POST",
		BaseURL:      "https://userservice.kugou.com",
		URL:          "/risk/v2/r_register_dev",
		Params:       params,
		RawBody:      aesResult.Str, // AES 加密后的 base64 字符串作为请求体
		EncryptType:  "android",
		MID:          mid,
		DFID:         "-",
		ResponseType: "arraybuffer", // 响应体是 AES 加密的二进制
	})

	if resp.Error != nil {
		return nil, resp.Error
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("register device: http %d", resp.StatusCode)
	}

	// 响应体是 AES 加密的，需用同样的 key 解密（PlaylistAESDecrypt 输入为 base64）
	bodyB64 := base64.StdEncoding.EncodeToString(resp.Body)
	decrypted, err := PlaylistAESDecrypt(bodyB64, aesResult.Key)
	if err != nil {
		return nil, fmt.Errorf("playlist aes decrypt failed: %w", err)
	}

	// 解析解密后的 JSON（data 偶发为数组/空数组，此时视为注册失败）
	var result struct {
		Status int `json:"status"`
		Data   struct {
			DFID string `json:"dfid"`
		} `json:"data"`
	}
	if err := json.Unmarshal(decrypted, &result); err != nil {
		// 兼容 data 为数组的情况（上游偶发返回 []），视为注册失败
		return nil, fmt.Errorf("register device: empty data")
	}

	if result.Status != 1 || result.Data.DFID == "" {
		return nil, fmt.Errorf("register device: status %d", result.Status)
	}

	// 缓存全局 DFID
	SetDFID(result.Data.DFID)

	return &DeviceInfo{
		MID:  mid,
		GUID: guid,
		DFID: result.Data.DFID,
		UUID: uuid,
	}, nil
}
