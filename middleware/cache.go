package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"
)

// 内存缓存实现（带 TTL）
type CacheStore struct {
	mu       sync.Mutex
	items    map[string]cacheItem
	maxItems int
}

type cacheItem struct {
	data      []byte
	status    int
	headers   http.Header
	expiresAt time.Time
}

var DefaultCache = NewCacheStore(2000)

// 新建缓存
func NewCacheStore(maxItems int) *CacheStore {
	return &CacheStore{
		items:    make(map[string]cacheItem),
		maxItems: maxItems,
	}
}

// 可缓存的只读接口前缀白名单。
// 之所以用白名单而不是「检测响应头 no-store」：CORS 中间件会给所有 API 响应统一加
// Cache-Control: no-store（为修扫码轮询被浏览器缓存的问题），若靠响应头判断会导致
// 缓存永远不命中（曾长期空转）。这里显式列出公开、极少变化的接口。
// 不含：用户相关（user/cloud/login/captcha）、歌单编辑与歌单详情/歌曲（刚加完歌不能看到旧数据）、
// 播放地址/音质（依赖登录与请求参数）、图片与音频代理。
var cacheablePrefixes = []string{
	"/api/search",
	"/api/rank",
	"/api/top",
	"/api/everyday",
	"/api/recommend",
	"/api/album",
	"/api/artist",
	"/api/lyric",
	"/api/fm",
	"/api/song/climax",
	"/api/playlist/tags",
	"/api/playlist/category",
	"/api/comment/song",
	"/api/comment/playlist",
	"/api/comment/album",
	"/api/comment/count",
	"/api/comment/floor",
	"/api/comment/hotword",
}

func isCacheablePath(p string) bool {
	for _, prefix := range cacheablePrefixes {
		if strings.HasPrefix(p, prefix) {
			return true
		}
	}
	return false
}

// 计算缓存 key（基于方法、路径、查询参数）
func cacheKey(r *http.Request) string {
	h := sha256.Sum256([]byte(r.Method + "|" + r.URL.String()))
	return hex.EncodeToString(h[:])
}

// 缓存响应中间件：对白名单内的成功 GET 响应缓存 2 分钟。
// 静态文件不做缓存（已 embed 在内存中）
func Cache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 仅 GET 且命中白名单才缓存；其余直接放行
		if r.Method != http.MethodGet || !isCacheablePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		key := cacheKey(r)
		if item, ok := DefaultCache.Get(key); ok {
			for k, v := range item.headers {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(item.status)
			w.Write(item.data)
			return
		}

		// 包装 ResponseWriter 以捕获响应
		cw := &cacheWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			headers:        make(http.Header),
			status:         http.StatusOK,
		}
		next.ServeHTTP(cw, r)

		// 成功的响应默认缓存 2 分钟
		if cw.status < 400 {
			ttl := 120 * time.Second // 默认 2 分钟
			DefaultCache.Set(key, cacheItem{
				data:      cw.body.Bytes(),
				status:    cw.status,
				headers:   cw.headers.Clone(),
				expiresAt: time.Now().Add(ttl),
			})
		}
	})
}

// 获取缓存
func (c *CacheStore) Get(key string) (cacheItem, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.items[key]
	if !ok {
		return cacheItem{}, false
	}
	if time.Now().After(item.expiresAt) {
		delete(c.items, key)
		return cacheItem{}, false
	}
	return item, true
}

// 写入缓存
func (c *CacheStore) Set(key string, item cacheItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.items) >= c.maxItems {
		// 简单淘汰：删除最早的
		oldestKey := ""
		var oldest time.Time
		for k, v := range c.items {
			if oldestKey == "" || v.expiresAt.Before(oldest) {
				oldestKey = k
				oldest = v.expiresAt
			}
		}
		if oldestKey != "" {
			delete(c.items, oldestKey)
		}
	}
	c.items[key] = item
}

// 缓存响应写入器
type cacheWriter struct {
	http.ResponseWriter
	body     *bytes.Buffer
	headers  http.Header
	status   int
	started  bool
}

func (w *cacheWriter) WriteHeader(status int) {
	if w.started {
		return
	}
	w.started = true
	w.status = status
	// 复制响应头
	for k, v := range w.ResponseWriter.Header() {
		w.headers[k] = v
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *cacheWriter) Write(b []byte) (int, error) {
	if !w.started {
		w.WriteHeader(http.StatusOK)
	}
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}