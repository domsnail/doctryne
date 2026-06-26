package http

import (
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

type CacheEntry struct {
	statusCode int
	header     http.Header

	body []byte

	expiresAt time.Time
}

func (c *Cache) Get(key *url.URL) (*CacheEntry, bool) {
	now := time.Now()
	uri := key.RequestURI()

	c.mu.RLock()
	entry, ok := c.entries[uri]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if now.After(entry.expiresAt) {
		c.mu.Lock()
		delete(c.entries, uri)
		c.mu.Unlock()

		return nil, false
	}

	return entry, true
}

func (c *Cache) Set(key *url.URL, entry *CacheEntry) {
	c.mu.Lock()
	c.entries[key.RequestURI()] = entry
	c.mu.Unlock()
}
