package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
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

// 计算缓存 key（基于方法、路径、查询参数）
func cacheKey(r *http.Request) string {
	h := sha256.Sum256([]byte(r.Method + "|" + r.URL.String()))
	return hex.EncodeToString(h[:])
}

// 缓存响应中间件
// 仅缓存 GET 请求，且响应需包含 Cache-Control: max-age
func Cache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
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

		// 检查是否设置了缓存时长
		maxAge := cw.headers.Get("Cache-Control")
		if maxAge != "" && cw.status < 400 {
			ttl := 60 * time.Second // 默认 60s
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