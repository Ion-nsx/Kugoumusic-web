package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// 图片代理
// 酷狗图片 URL 通常格式：https://img.kugou.com/.../{size}.jpg
// 本代理用于：
// 1. 避免前端直接访问外网图片
// 2. 统一图片大小格式
// 3. 添加缓存
// 4. 绕过可能的防盗链

// 图片尺寸常量
const (
	ImgSizeSmall  = "240"
	ImgSizeMedium = "480"
	ImgSizeLarge  = "640"
	ImgSizeOrigin = "0"
)

// 获取图片（通过代理）
// 输入: 原始图片 URL 或图片 ID
// 输出: 图片二进制数据 + content-type
func ProxyImage(imageURL, size string) ([]byte, string, error) {
	if imageURL == "" {
		return nil, "", fmt.Errorf("empty image url")
	}

	// 如果已经是完整 URL，直接代理
	if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
		return fetchImage(imageURL)
	}

	// 构造酷狗图片 URL
	// 通常格式: https://img.kugou.com/{type}/{id}/{size}.jpg
	// 或: https://imge.kugou.com/{type}/{id}/{size}.jpg
	if size == "" {
		size = ImgSizeMedium
	}

	// 根据图片 ID 判断类型
	imgURL := fmt.Sprintf("https://img.kugou.com/audio/%s/%s.jpg", size, imageURL)

	return fetchImage(imgURL)
}

// 获取歌手图片
func ProxyArtistImage(artistID, size string) ([]byte, string, error) {
	if size == "" {
		size = ImgSizeMedium
	}
	imgURL := fmt.Sprintf("https://img.kugou.com/singer/%s/%s.jpg", size, artistID)
	return fetchImage(imgURL)
}

// 获取专辑图片
func ProxyAlbumImage(albumID, size string) ([]byte, string, error) {
	if size == "" {
		size = ImgSizeMedium
	}
	imgURL := fmt.Sprintf("https://img.kugou.com/album/%s/%s.jpg", size, albumID)
	return fetchImage(imgURL)
}

// 获取歌单图片
func ProxyPlaylistImage(playlistID, size string) ([]byte, string, error) {
	if size == "" {
		size = ImgSizeMedium
	}
	imgURL := fmt.Sprintf("https://img.kugou.com/playlist/%s/%s.jpg", size, playlistID)
	return fetchImage(imgURL)
}

// 通用图片获取
func fetchImage(url string) ([]byte, string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Referer", "https://www.kugou.com/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	return body, contentType, nil
}

// 判断是否是图片链接
func IsImageURL(url string) bool {
	lower := strings.ToLower(url)
	return strings.HasSuffix(lower, ".jpg") ||
		strings.HasSuffix(lower, ".jpeg") ||
		strings.HasSuffix(lower, ".png") ||
		strings.HasSuffix(lower, ".webp") ||
		strings.HasSuffix(lower, ".gif") ||
		strings.Contains(lower, "img.kugou.com") ||
		strings.Contains(lower, "imge.kugou.com")
}