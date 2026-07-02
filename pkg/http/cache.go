package http

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry

	ttl time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	if ttl <= 0 {
		panic("http request cache ttl must be greater than zero")
	}

	return &Cache{
		mu:      sync.RWMutex{},
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
}

func (c *Cache) Get(u *url.URL) (*CacheEntry, bool) {
	now := time.Now()
	key := cacheKey(u)

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if now.After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()

		return nil, false
	}

	return entry, true
}

func (c *Cache) Set(key *url.URL, entry *CacheEntry) {
	c.mu.Lock()
	c.entries[cacheKey(key)] = entry
	c.mu.Unlock()
}

type CacheEntry struct {
	statusCode int
	header     http.Header

	body []byte

	expiresAt time.Time
}

func (c CacheEntry) response(req *http.Request, fromCache bool) *http.Response {
	h := cloneHeader(c.header)
	if fromCache {
		h.Set("X-From-Cache", "true")
	}

	return &http.Response{
		StatusCode:    c.statusCode,
		Status:        fmt.Sprintf("%d %s", c.statusCode, http.StatusText(c.statusCode)),
		Header:        h,
		Body:          io.NopCloser(bytes.NewReader(c.body)),
		ContentLength: int64(len(c.body)),
		Request:       req,
	}
}

func cacheKey(u *url.URL) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(u.Host+u.RequestURI())))
}
