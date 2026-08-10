package api

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func uploadUA() string {
	return fmt.Sprintf("Android15-1070-%d-201-0-wifi", ClientVer)
}

func randomKGTHash() string {
	n := rand.Intn(0xFFFFFFF)
	return fmt.Sprintf("%07x", n)
}

func bssVerifyCode(bucket string) string {
	return MD5(fmt.Sprintf("%d%s8ae10344e9738dcb", AppID, bucket))
}

// 签名 bss 参数（lite Android 签名，无 body，与客户端上传行为一致）
func signBssParams(params map[string]string) map[string]string {
	params["signature"] = signatureAndroidParamsWithSalt(params, nil, LiteAndroidSalt)
	return params
}

// bss 原始 HTTP 请求（步骤 1-4），不经过 SendRequest 的默认参数注入
func doBssRequest(method, urlStr string, params map[string]string, body []byte, extraHeaders map[string]string) (*APIResponse, error) {
	qs := encodeParams(params)
	fullURL := urlStr + "?" + qs

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return &APIResponse{Error: err}, err
	}

	req.Header.Set("User-Agent", uploadUA())
	req.Header.Set("KG-RC", "1")
	req.Header.Set("KG-Rec", "1")
	req.Header.Set("KG-THash", randomKGTHash())

	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	dialer := &net.Dialer{Timeout: 15 * time.Second}
	client := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, "tcp4", addr)
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return &APIResponse{Error: err}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIResponse{Error: err}, err
	}

	result := &APIResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}

	var dataObj map[string]interface{}
	if err := json.Unmarshal(respBody, &dataObj); err == nil {
		result.Data = dataObj
	}

	return result, nil
}

// ===== 曲库匹配 =====
// POST http://kmr.service.kugou.com/v2/album_audio/audio + x-router
// 根据文件 hash 匹配酷狗曲库歌曲信息
func GetCloudMatch(hash string, albumAudioID int64) *APIResponse {
	if hash == "" {
		return &APIResponse{Error: fmt.Errorf("hash 参数必填")}
	}

	clienttime := time.Now().Unix()

	dataItem := map[string]interface{}{
		"hash": strings.ToLower(hash),
	}
	if albumAudioID > 0 {
		dataItem["album_audio_id"] = strconv.FormatInt(albumAudioID, 10)
	}
	data := []map[string]interface{}{dataItem}

	dataMap := map[string]interface{}{
		"appid":                      AppID,
		"clienttime":                 clienttime,
		"clientver":                  ClientVer,
		"data":                       data,
		"dfid":                       GetDFID(),
		"key":                        signParamsKey(strconv.FormatInt(clienttime, 10)),
		"mid":                        GetMID(),
		"show_privilege":             0,
		"show_author_alias":          0,
		"show_rel_album_audio_info":  0,
		"show_remarks":               0,
	}

	return SendRequest(&RequestOptions{
		Method:       "POST",
		BaseURL:      "http://kmr.service.kugou.com",
		URL:          "/v2/album_audio/audio",
		Data:         dataMap,
		EncryptType:  "android",
		ClearDefault: true,
		NoSign:       true,
		Headers: map[string]string{
			"x-router":     "kmr.service.kugou.com",
			"Content-Type": "application/json",
		},
	})
}

// 从匹配结果中提取有用信息
type CloudMatchInfo struct {
	AlbumAudioID int64  `json:"album_audio_id"`
	AudioID      int64  `json:"audio_id"`
	HashStd      string `json:"hash_std"`
	AuthorName   string `json:"author_name"`
	AudioName    string `json:"audio_name"`
	SizableCover string `json:"sizable_cover"` // album_info.sizable_cover
}

// 批量曲库匹配，返回 hash→匹配结果 的映射
func GetCloudMatchBatch(hashes []string) (map[string]*CloudMatchInfo, error) {
	if len(hashes) == 0 {
		return nil, fmt.Errorf("hash 列表为空")
	}

	// 去重
	seen := map[string]bool{}
	uniqueHashes := make([]string, 0, len(hashes))
	for _, h := range hashes {
		h = strings.ToLower(h)
		if h != "" && !seen[h] {
			seen[h] = true
			uniqueHashes = append(uniqueHashes, h)
		}
	}
	if len(uniqueHashes) == 0 {
		return nil, fmt.Errorf("无有效 hash")
	}

	clienttime := time.Now().Unix()

	data := make([]map[string]interface{}, len(uniqueHashes))
	for i, h := range uniqueHashes {
		data[i] = map[string]interface{}{
			"hash": h,
		}
	}

	dataMap := map[string]interface{}{
		"appid":                     AppID,
		"clienttime":                clienttime,
		"clientver":                 ClientVer,
		"data":                      data,
		"dfid":                      GetDFID(),
		"key":                       signParamsKey(strconv.FormatInt(clienttime, 10)),
		"mid":                       GetMID(),
		"show_privilege":            0,
		"show_author_alias":         0,
		"show_rel_album_audio_info": 0,
		"show_remarks":              0,
	}

	resp := SendRequest(&RequestOptions{
		Method:       "POST",
		BaseURL:      "http://kmr.service.kugou.com",
		URL:          "/v2/album_audio/audio",
		Data:         dataMap,
		EncryptType:  "android",
		ClearDefault: true,
		NoSign:       true,
		Headers: map[string]string{
			"x-router":     "kmr.service.kugou.com",
			"Content-Type": "application/json",
		},
	})
	if resp.Error != nil {
		return nil, resp.Error
	}

	status, _ := resp.Data["status"].(float64)
	if status != 1 {
		return nil, fmt.Errorf("匹配失败")
	}

	dataArr, ok := resp.Data["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("匹配结果为空")
	}

	result := map[string]*CloudMatchInfo{}
	for i, item := range dataArr {
		if i >= len(uniqueHashes) {
			break
		}
		hash := uniqueHashes[i]

		// 每个元素可能是候选数组
		var first map[string]interface{}
		if arr, ok := item.([]interface{}); ok && len(arr) > 0 {
			first, _ = arr[0].(map[string]interface{})
		} else {
			first, _ = item.(map[string]interface{})
		}
		if first == nil {
			continue
		}

		info := &CloudMatchInfo{}
		if v, ok := first["album_audio_id"]; ok {
			info.AlbumAudioID, _ = toInt64(v)
		}
		if audioInfo, ok := first["audio_info"].(map[string]interface{}); ok {
			if v, ok := audioInfo["audio_id"]; ok {
				info.AudioID, _ = toInt64(v)
			}
			if v, ok := audioInfo["hash"]; ok {
				info.HashStd, _ = v.(string)
			}
		}
		if info.AudioID == 0 {
			if v, ok := first["audio_id"]; ok {
				info.AudioID, _ = toInt64(v)
			}
		}
		if info.HashStd == "" {
			if v, ok := first["hash"]; ok {
				info.HashStd, _ = v.(string)
			}
		}
		if v, ok := first["author_name"]; ok {
			info.AuthorName, _ = v.(string)
		}
		if v, ok := first["ori_audio_name"]; ok {
			info.AudioName, _ = v.(string)
		}
		if info.AudioName == "" {
			if v, ok := first["audio_name"]; ok {
				info.AudioName, _ = v.(string)
			}
		}
		if v, ok := first["songname"]; ok {
			if info.AudioName == "" {
				info.AudioName, _ = v.(string)
			}
		}
		// 提取封面
		if albumInfo, ok := first["album_info"].(map[string]interface{}); ok {
			if v, ok := albumInfo["sizable_cover"]; ok {
				info.SizableCover, _ = v.(string)
			}
		}

		if info.AlbumAudioID > 0 || info.AudioID > 0 || info.HashStd != "" {
			result[hash] = info
		}
	}

	return result, nil
}

func extractMatchInfo(resp *APIResponse) *CloudMatchInfo {
	if resp == nil || resp.Data == nil {
		return nil
	}
	status, _ := resp.Data["status"].(float64)
	if status != 1 {
		return nil
	}

	// 尝试从 data 数组第一个元素中提取
	dataArr, ok := resp.Data["data"].([]interface{})
	if !ok || len(dataArr) == 0 {
		return nil
	}

	// data 的每个元素可能是数组，取第一个候选项
	var first map[string]interface{}
	if arr, ok := dataArr[0].([]interface{}); ok && len(arr) > 0 {
		first, _ = arr[0].(map[string]interface{})
	} else {
		first, _ = dataArr[0].(map[string]interface{})
	}
	if first == nil {
		return nil
	}

	info := &CloudMatchInfo{}
	if v, ok := first["album_audio_id"]; ok {
		info.AlbumAudioID, _ = toInt64(v)
	}
	if v, ok := first["audio_id"]; ok {
		info.AudioID, _ = toInt64(v)
	}
	if audioInfo, ok := first["audio_info"].(map[string]interface{}); ok {
		if v, ok := audioInfo["audio_id"]; ok {
			i, _ := toInt64(v)
			if i > 0 {
				info.AudioID = i
			}
		}
		if v, ok := audioInfo["hash"]; ok {
			info.HashStd, _ = v.(string)
		}
	}
	if info.HashStd == "" {
		if v, ok := first["hash"]; ok {
			info.HashStd, _ = v.(string)
		}
	}
	if v, ok := first["author_name"]; ok {
		info.AuthorName, _ = v.(string)
	}
	if v, ok := first["ori_audio_name"]; ok {
		info.AudioName, _ = v.(string)
	}
	if info.AudioName == "" {
		if v, ok := first["audio_name"]; ok {
			info.AudioName, _ = v.(string)
		}
	}
	if info.AudioName == "" {
		if v, ok := first["songname"]; ok {
			info.AudioName, _ = v.(string)
		}
	}
	// 提取封面
	if albumInfo, ok := first["album_info"].(map[string]interface{}); ok {
		if v, ok := albumInfo["sizable_cover"]; ok {
			info.SizableCover, _ = v.(string)
		}
	}

	if info.AlbumAudioID == 0 && info.AudioID == 0 && info.HashStd == "" {
		return nil
	}
	return info
}

func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	case string:
		i, err := strconv.ParseInt(n, 10, 64)
		return i, err == nil
	case int64:
		return n, true
	case int:
		return int64(n), true
	}
	return 0, false
}

// ===== 云盘上传（完整5步流程） =====
func UploadCloudFile(fileData []byte, filename, extendname string, autoMatch bool, userid, token string) *APIResponse {
	if len(fileData) == 0 {
		return &APIResponse{Error: fmt.Errorf("文件数据为空")}
	}
	if userid == "" || token == "" {
		return &APIResponse{Error: fmt.Errorf("请先登录")}
	}

	bucket := "musicclound"

	// 文件 MD5 作为 BSS 上传标识（酷狗 BSS 系统要求，不能用原始文件名）
	h := md5.Sum(fileData)
	fileMD5 := hex.EncodeToString(h[:])

	// 保留原始文件名用于匹配失败时的显示
	originalBaseName := filename
	if originalBaseName == "" {
		originalBaseName = fileMD5
	}
	filename = fileMD5

	if extendname == "" {
		extendname = "mp3"
	}
	extendname = strings.TrimPrefix(extendname, ".")

	version := strconv.Itoa(ClientVer)
	dfid := GetDFID()
	mid := GetMID()
	uuid := GetGUID()

	// 可选：曲库匹配
	var matchInfo *CloudMatchInfo
	if autoMatch {
		matchResp := GetCloudMatch(filename, 0)
		if matchResp.Error == nil {
			matchInfo = extractMatchInfo(matchResp)
		}
	}

	hashStd := filename
	audioID := int64(0)
	albumAudioID := int64(0)
	authorName := ""
	trackName := originalBaseName

	if matchInfo != nil {
		if matchInfo.HashStd != "" {
			hashStd = strings.ToLower(matchInfo.HashStd)
		}
		audioID = matchInfo.AudioID
		albumAudioID = matchInfo.AlbumAudioID
		if matchInfo.AuthorName != "" {
			authorName = matchInfo.AuthorName
		}
		if matchInfo.AudioName != "" {
			trackName = matchInfo.AudioName
		}
	}

	// 构建云盘文件名
	name := trackName
	if authorName != "" {
		name = authorName + " - " + trackName
	}
	name = name + "." + extendname

	// ========== 步骤1 获取上传授权 ==========
	log.Printf("[cloud_upload:step1] 获取上传授权, filename=%s, userid=%s", filename, userid)
	loginType := 0
	if token != "" && userid != "0" {
		loginType = 1
	}
	authParams := signBssParams(map[string]string{
		"bucket":       bucket,
		"filename":     filename,
		"method":       "POST",
		"loginType":    strconv.Itoa(loginType),
		"buVerifyCode": bssVerifyCode(bucket),
		"extranet":     "1",
		"userid":       userid,
		"token":        token,
		"version":      version,
		"dfid":         dfid,
		"mid":          mid,
		"uuid":         uuid,
		"appid":        strconv.Itoa(AppID),
		"clientver":    version,
		"clienttime":   strconv.FormatInt(time.Now().Unix(), 10),
	})

	authResp, err := doBssRequest("GET", "https://gateway.kugou.com/bsstrackercdngz/v1/upload/auth", authParams, nil, nil)
	if err != nil {
		log.Printf("[cloud_upload:step1] 获取上传授权失败: %v", err)
		return &APIResponse{Error: fmt.Errorf("获取上传授权失败: %v", err)}
	}
	authData, ok := authResp.Data["data"].(map[string]interface{})
	if !ok {
		log.Printf("[cloud_upload:step1] 授权响应异常, body=%s", string(authResp.Body))
		return &APIResponse{Error: fmt.Errorf("授权响应异常: %s", string(authResp.Body))}
	}
	authorization, _ := authData["authorization"].(string)
	if authorization == "" {
		log.Printf("[cloud_upload:step1] 未获取到 authorization, body=%s", string(authResp.Body))
		return &APIResponse{Error: fmt.Errorf("未获取到 authorization: %s", string(authResp.Body))}
	}
	log.Printf("[cloud_upload:step1] 获取授权成功")

	// ========== 步骤2 初始化分片上传 ==========
	log.Printf("[cloud_upload:step2] 初始化分片上传")
	initParams := signBssParams(map[string]string{
		"bucket":        bucket,
		"filename":      filename,
		"ssl":           "1",
		"extendname":    extendname,
		"version":       version,
		"userid":        userid,
		"token":         token,
		"authorization": authorization,
		"dfid":          dfid,
		"mid":           mid,
		"uuid":          uuid,
		"appid":         strconv.Itoa(AppID),
		"clientver":     version,
		"clienttime":    strconv.FormatInt(time.Now().Unix(), 10),
	})

	initResp, err := doBssRequest("POST", "http://bssulbig.kugou.com/v2/multipart/initiate/music", initParams, nil,
		map[string]string{"Authorization": authorization})
	if err != nil {
		log.Printf("[cloud_upload:step2] 初始化上传失败: %v", err)
		return &APIResponse{Error: fmt.Errorf("初始化上传失败: %v", err)}
	}

	initData, ok := initResp.Data["data"].(map[string]interface{})
	if !ok {
		log.Printf("[cloud_upload:step2] 初始化响应异常, body=%s", string(initResp.Body))
		return &APIResponse{Error: fmt.Errorf("初始化响应异常: %s", string(initResp.Body))}
	}

	externalHost, _ := initData["external_host"].(string)
	uploadID, _ := initData["upload_id"].(string)
	bssFileHash := filename
	if v, ok := initData["x-bss-filename"].(string); ok && v != "" {
		bssFileHash = v
	}
	log.Printf("[cloud_upload:step2] 初始化成功 upload_id=%s host=%s 秒传=%v", uploadID, externalHost, uploadID == "")

	// 秒传分支：upload_id 为空说明文件已在服务器
	if uploadID != "" {
		if externalHost == "" {
			return &APIResponse{Error: fmt.Errorf("未获取到 external_host: %s", string(initResp.Body))}
		}

		// ========== 步骤3 上传分片（1MB 一片） ==========
		partSize := 1024 * 1024
		partCount := (len(fileData) + partSize - 1) / partSize
		if partCount < 1 {
			partCount = 1
		}

		for i := 0; i < partCount; i++ {
			start := i * partSize
			end := start + partSize
			if end > len(fileData) {
				end = len(fileData)
			}
			chunk := fileData[start:end]

			uploadParams := signBssParams(map[string]string{
				"bucket":        bucket,
				"authorization": authorization,
				"filename":      filename,
				"partnumber":    strconv.Itoa(i + 1),
				"upload_id":     uploadID,
				"body_empty":    "1",
				"version":       version,
				"userid":        userid,
				"token":         token,
				"dfid":          dfid,
				"mid":           mid,
				"uuid":          uuid,
				"appid":         strconv.Itoa(AppID),
				"clientver":     version,
				"clienttime":    strconv.FormatInt(time.Now().Unix(), 10),
			})

			uploadResp, err := doBssRequest("POST", "http://"+externalHost+"/v3/multipart/upload", uploadParams, chunk,
				map[string]string{
					"Authorization": authorization,
					"Content-Type":  "application/octet-stream",
				})
			if err != nil {
				return &APIResponse{Error: fmt.Errorf("分片 %d/%d 上传失败: %v", i+1, partCount, err)}
			}
			status, _ := uploadResp.Data["status"].(float64)
			if status != 1 {
				return &APIResponse{Error: fmt.Errorf("分片 %d/%d 上传失败: %s", i+1, partCount, string(uploadResp.Body))}
			}
		}

		// ========== 步骤4 完成上传 ==========
		completeParams := signBssParams(map[string]string{
			"bucket":        bucket,
			"authorization": authorization,
			"filename":      filename,
			"partnumber":    strconv.Itoa(partCount),
			"upload_id":     uploadID,
			"md5":           filename,
			"version":       version,
			"userid":        userid,
			"token":         token,
			"if_id3":        "1",
			"dfid":          dfid,
			"mid":           mid,
			"uuid":          uuid,
			"appid":         strconv.Itoa(AppID),
			"clientver":     version,
			"clienttime":    strconv.FormatInt(time.Now().Unix(), 10),
		})

		completeResp, err := doBssRequest("POST", "http://"+externalHost+"/v3/multipart/complete", completeParams, nil,
			map[string]string{"Authorization": authorization})
		if err != nil {
			return &APIResponse{Error: fmt.Errorf("完成上传失败: %v", err)}
		}
		completeStatus, _ := completeResp.Data["status"].(float64)
		if completeStatus != 1 {
			return &APIResponse{Error: fmt.Errorf("完成上传失败: %s", string(completeResp.Body))}
		}
		if cd, ok := completeResp.Data["data"].(map[string]interface{}); ok {
			if v, ok := cd["x-bss-filename"].(string); ok && v != "" {
				bssFileHash = v
			}
		}
	}

	// ========== 步骤5 添加文件到云盘 ==========
	log.Printf("[cloud_upload:step5] 添加文件到云盘, name=%s, bssHash=%s", name, bssFileHash)
	aesResult, err := PlaylistAESEncrypt(map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"name":           name,
				"ext":            extendname,
				"author_name":    authorName,
				"hash":           bssFileHash,
				"hash_std":       hashStd,
				"audio_id":       audioID,
				"bitrate":        4,
				"album_audio_id": albumAudioID,
				"size":           len(fileData),
				"timelen":        0,
			},
		},
		"list_ver": 0,
	})
	if err != nil {
		return &APIResponse{Error: err}
	}

	p, err := RSAEncrypt2(map[string]interface{}{
		"aes":   aesResult.Key,
		"uid":   userid,
		"token": token,
	}, UseLite)
	if err != nil {
		return &APIResponse{Error: err}
	}

	bodyBytes, err := base64.StdEncoding.DecodeString(aesResult.Str)
	if err != nil {
		return &APIResponse{Error: err}
	}

	clienttime := time.Now().Unix()
	addParams := map[string]string{
		"clienttime": strconv.FormatInt(clienttime, 10),
		"mid":        mid,
		"key":        signParamsKey(strconv.FormatInt(clienttime, 10)),
		"clientver":  version,
		"appid":      strconv.Itoa(AppID),
		"p":          strings.ToUpper(p),
	}

	addResp := SendRequest(&RequestOptions{
		Method:       "POST",
		BaseURL:      "https://mcloudservice.kugou.com",
		URL:          "/v1/add_files",
		Params:       addParams,
		RawBodyBytes: bodyBytes,
		EncryptType:  "android",
		ClearDefault: true,
		NoSign:       true,
	})

	// 尝试解密响应
	if addResp.Error == nil && len(addResp.Body) > 0 {
		plainBytes, decErr := PlaylistAESDecrypt(base64.StdEncoding.EncodeToString(addResp.Body), aesResult.Key)
		if decErr == nil {
			log.Printf("[cloud_upload:step5] AES 解密成功")
			var decrypted map[string]interface{}
			if json.Unmarshal(plainBytes, &decrypted) == nil {
				addResp.Data = decrypted
				// 附加上传流程信息
				addResp.Data["upload_info"] = map[string]interface{}{
					"authorization":  authorization,
					"external_host":  externalHost,
					"upload_id":      uploadID,
					"hash":           bssFileHash,
					"local_hash":     filename,
					"hash_std":       hashStd,
					"audio_id":       audioID,
					"album_audio_id": albumAudioID,
					"matched":        matchInfo != nil,
					"filesize":       len(fileData),
				}
			}
		} else {
			log.Printf("[cloud_upload:step5] AES 解密失败 (尝试明文解析): %v", decErr)
			// 尝试明文 JSON
			var plain map[string]interface{}
			if json.Unmarshal(addResp.Body, &plain) == nil {
				addResp.Data = plain
				log.Printf("[cloud_upload:step5] 明文 JSON 解析成功")
			}
		}
	} else if addResp.Error != nil {
		log.Printf("[cloud_upload:step5] 请求失败: %v", addResp.Error)
	}

	return addResp
}
