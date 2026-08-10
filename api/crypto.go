package api

import (
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
)

// 标准版 RSA 公钥
const publicRSAKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDIAG7QOELSYoIJvTFJhMpe1s/g
bjDJX51HBNnEl5HXqTW6lQ7LC8jr9fWZTwusknp+sVGzwd40MwP6U5yDE27M/X1+
UR4tvOGOqp94TJtQ1EPnWGWXngpeIW5GxoQGao1rmYWAu6oi1z9XkChrsUdC6DJE
5E221wf/4WLFxwAtRQIDAQAB
-----END PUBLIC KEY-----`

// 概念版 RSA 公钥
const publicLiteRSAKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDECi0Np2UR87scwrvTr72L6oO0
1rBbbBPriSDFPxr3Z5syug0O24QyQO8bg27+0+4kBzTBTBOZ/WWU0WryL1JSXRTX
LgFVxtzIY41Pe7lPOgsfTCn5kZcvKhYKJesKnnJDNr5/abvTGf+rHG3YRwsCHcQ0
8/q6ifSioBszvb3QiwIDAQAB
-----END PUBLIC KEY-----`

// KRC 歌词 XOR 解密密钥（16字节）
var krcXORKey = []byte{64, 71, 97, 119, 94, 50, 116, 71, 81, 54, 49, 45, 206, 210, 110, 105}

// 随机字符串字符池
const randomChars = "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// MD5 哈希
func MD5(data string) string {
	h := md5.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

// MD5Bytes 对字节数组取 MD5
func MD5Bytes(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}

// SHA1 哈希
func SHA1(data string) string {
	h := sha1.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

// AES-128-CBC 加密，返回 hex 字符串
// 如果 key 和 iv 为空，自动生成随机 key
type AESResult struct {
	Str string // 密文 hex
	Key string // 密钥（自动生成时有效）
}

// 加密 JSON 对象（自动生成 key），对应参考 cryptoAesEncrypt(object)
func AESEncryptJSON(data interface{}) (*AESResult, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return AESEncrypt(b, "", "")
}

// 解密 JSON 对象，key 为自动生成的临时 key（自动推导 MD5）
func AESDecryptJSON(cipherHex, key string) (map[string]interface{}, error) {
	plain, err := AESDecrypt(cipherHex, key, "")
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(plain, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func AESEncrypt(plaintext []byte, key, iv string) (*AESResult, error) {
	var useKey, useIv string
	var tempKey string
	autoGen := false

	if key != "" && iv != "" {
		useKey = key
		useIv = iv
	} else {
		tempKey = RandomString(16)
		useKey = MD5(tempKey)[:32]
		useIv = useKey[len(useKey)-16:]
		autoGen = true
	}

	block, err := aes.NewCipher([]byte(useKey))
	if err != nil {
		return nil, err
	}

	plaintext = pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, []byte(useIv))
	mode.CryptBlocks(ciphertext, plaintext)

	result := &AESResult{Str: hex.EncodeToString(ciphertext)}
	if autoGen {
		result.Key = tempKey
	}
	return result, nil
}

// AES-128-CBC 解密，输入 hex 密文
func AESDecrypt(cipherHex, key, iv string) ([]byte, error) {
	var useKey, useIv string
	if iv != "" {
		useKey = key
		useIv = iv
	} else {
		useKey = MD5(key)[:32]
		useIv = useKey[len(useKey)-16:]
	}

	ciphertext, err := hex.DecodeString(cipherHex)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher([]byte(useKey))
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, []byte(useIv))
	mode.CryptBlocks(plaintext, ciphertext)

	return pkcs7Unpad(plaintext)
}

// 歌单 AES 加密（特殊 key 推导）
type PlaylistAESResult struct {
	Key string
	Str string
}

// 歌手关注等接口的 AES 加密（对齐 JS cryptoAesEncrypt）
// key = MD5(随机16位小写) 全量(32字节)作 AES-256 key，后16位作 iv，密文输出 hex
func CryptoAESEncrypt(data interface{}) (key, str string, err error) {
	var jsonStr string
	switch v := data.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return "", "", err
		}
		jsonStr = string(jsonBytes)
	}

	tempKey := strings.ToLower(RandomString(16))
	md5Hex := MD5(tempKey) // 32 位 hex 字符串
	keyBytes := []byte(md5Hex)
	iv := keyBytes[len(keyBytes)-16:]

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", "", err
	}

	plaintext := pkcs7Pad([]byte(jsonStr), aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return tempKey, hex.EncodeToString(ciphertext), nil
}

func PlaylistAESEncrypt(data interface{}) (*PlaylistAESResult, error) {	var jsonStr string
	switch v := data.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		jsonStr = string(jsonBytes)
	}

	key := strings.ToLower(RandomString(6))
	encryptKey := MD5(key)[:16]
	iv := MD5(key)[16:32]

	block, err := aes.NewCipher([]byte(encryptKey))
	if err != nil {
		return nil, err
	}

	plaintext := pkcs7Pad([]byte(jsonStr), aes.BlockSize)
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, plaintext)

	return &PlaylistAESResult{
		Key: key,
		Str: base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

func PlaylistAESDecrypt(str, key string) ([]byte, error) {
	encryptKey := MD5(key)[:16]
	iv := MD5(key)[16:32]

	ciphertext, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher([]byte(encryptKey))
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(plaintext, ciphertext)

	return pkcs7Unpad(plaintext)
}

// RSA 加密（PKCS1v15）
func RSAEncrypt(data []byte, lite bool) (string, error) {
	pemStr := publicRSAKey
	if lite {
		pemStr = publicLiteRSAKey
	}

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not RSA public key")
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, data)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encrypted), nil
}

// RSA 加密 v2（用于 register_dev 等接口，直接输出 hex）
func RSAEncrypt2(data interface{}, lite bool) (string, error) {
	var jsonStr string
	switch v := data.(type) {
	case string:
		jsonStr = v
	case []byte:
		jsonStr = string(v)
	default:
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		jsonStr = string(jsonBytes)
	}
	return RSAEncrypt([]byte(jsonStr), lite)
}

// 裸 RSA 加密（Raw RSA / textbook RSA，无 PKCS1 填充）
// 对应参考项目 crypto.js 的 cryptoRSAEncrypt：
//   明文右对齐填充到密钥长度（前导明文、尾随0，等价于明文作为大端整数）
//   c = m^e mod n，输出固定 keyLen*2 位大写 hex
// 用于登录接口（/v9/login_by_pwd、/v3/get_my_info 等）
func RawRSAEncryptJSON(data interface{}, lite bool) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return RawRSAEncrypt(jsonBytes, lite)
}

// 裸 RSA 加密字节数据
func RawRSAEncrypt(data []byte, lite bool) (string, error) {
	pemStr := publicRSAKey
	if lite {
		pemStr = publicLiteRSAKey
	}

	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not RSA public key")
	}

	keyLen := (rsaPub.N.BitLen() + 7) / 8
	if len(data) > keyLen {
		return "", fmt.Errorf("data length exceeds key size")
	}

	// 填充到密钥长度：明文放开头，尾随补 0（与参考 padded.set(buffer) 一致）
	m := make([]byte, keyLen)
	copy(m, data)

	c := new(big.Int).SetBytes(m)
	e := big.NewInt(int64(rsaPub.E))
	cipher := new(big.Int).Exp(c, e, rsaPub.N)

	return fmt.Sprintf("%0*X", keyLen*2, cipher), nil
}

// 生成随机字符串
func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		randBytes := make([]byte, 1)
		rand.Read(randBytes)
		b[i] = randomChars[int(randBytes[0])%len(randomChars)]
	}
	return string(b)
}

// 生成随机数字字符串
func RandomNumber(length int) string {
	chars := "1234567890"
	b := make([]byte, length)
	for i := range b {
		randBytes := make([]byte, 1)
		rand.Read(randBytes)
		b[i] = chars[int(randBytes[0])%len(chars)]
	}
	return string(b)
}

// 计算设备 MID：MD5(guid) 的 hex 转十进制大整数
func CalculateMID(str string) string {
	hash := md5.Sum([]byte(str))
	hexStr := hex.EncodeToString(hash[:])
	n := new(big.Int)
	n.SetString(hexStr, 16)
	return n.String()
}

// 解码 KRC 歌词
// 格式：文件头（4字节）+ 异或加密数据 + zlib 压缩
func DecodeKRC(data []byte) (string, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("data too short")
	}
	// 跳过前4字节文件头
	krcData := make([]byte, len(data)-4)
	copy(krcData, data[4:])

	// XOR 解密
	for i := 0; i < len(krcData); i++ {
		krcData[i] ^= krcXORKey[i%len(krcXORKey)]
	}

	// zlib 解压
	r, err := zlib.NewReader(strings.NewReader(string(krcData)))
	if err != nil {
		return "", err
	}
	defer r.Close()

	var buf strings.Builder
	b := make([]byte, 4096)
	for {
		n, err := r.Read(b)
		if n > 0 {
			buf.Write(b[:n])
		}
		if err != nil {
			break
		}
	}
	return buf.String(), nil
}

// 格式化 Cookie 字符串（移除 domain/path/expires/HttpOnly）
func ParseCookieString(cookie string) string {
	parts := strings.Split(cookie, ";")
	var filtered []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		lower := strings.ToLower(p)
		if strings.HasPrefix(lower, "domain=") ||
			strings.HasPrefix(lower, "path=") ||
			strings.HasPrefix(lower, "expires=") ||
			lower == "httponly" {
			continue
		}
		filtered = append(filtered, p)
	}
	return strings.Join(filtered, "; ")
}

// PKCS7 填充
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// PKCS7 去填充
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, fmt.Errorf("invalid padding byte")
		}
	}
	return data[:len(data)-padding], nil
}