package http

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/sync/singleflight"
)

// SmartTransport provides http.Transport with in-memory cache and basic request delay
type SmartTransport struct {
	base http.RoundTripper

	cache    *Cache
	throttle *ThrottleMap

	single singleflight.Group

	CacheTTL     time.Duration
	CacheMethods map[string]bool
}

type TransportOptions struct {
	BaseTransport http.RoundTripper

	CachedMethods []string
	CacheTTL      time.Duration

	ThrottleOptions *ThrottleOptions
}

func NewSmartTransport(opts TransportOptions) *SmartTransport {
	var t = SmartTransport{
		base:         opts.BaseTransport,
		throttle:     NewThrottleMap(opts.ThrottleOptions),
		cache:        NewCache(opts.CacheTTL),
		CacheTTL:     opts.CacheTTL,
		CacheMethods: make(map[string]bool),
	}

	if opts.BaseTransport == nil {
		t.base = http.DefaultTransport
	}

	for _, m := range opts.CachedMethods {
		t.CacheMethods[m] = true
	}

	return &t
}

func (t *SmartTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil && req.GetBody == nil {
		return t.base.RoundTrip(req)
	}

	if !t.CacheMethods[req.Method] {
		return t.sendWithWait(req)
	}

	if entry, ok := t.cache.Get(req.URL); ok {
		slog.Log(req.Context(), -8, "returning cached request",
			slog.String("method", req.Method),
			slog.String("url", req.URL.Redacted()),
		)

		return entry.response(req, true), nil
	}

	v, err, _ := t.single.Do(cacheKey(req.URL), func() (any, error) {
		if entry, ok := t.cache.Get(req.URL); ok {
			return entry, nil
		}

		resp, err := t.sendWithWait(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		entry := CacheEntry{
			statusCode: resp.StatusCode,
			header:     cloneHeader(resp.Header),
			body:       body,
			expiresAt:  time.Now().Add(t.CacheTTL),
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			t.cache.Set(req.URL, &entry)
		}

		return entry, nil
	})

	if err != nil {
		return nil, err
	}

	entry := v.(CacheEntry)
	return entry.response(req, false), nil
}

func (t *SmartTransport) sendWithWait(req *http.Request) (*http.Response, error) {
	host := req.URL.Hostname()
	if err := t.throttle.Wait(req.Context(), host); err != nil {
		return nil, err
	}

	clone, err := cloneRequest(req)
	if err != nil {
		return nil, err
	}

	return t.base.RoundTrip(clone)
}

func cloneHeader(h http.Header) http.Header {
	out := make(http.Header, len(h))
	for k, vv := range h {
		cp := make([]string, len(vv))
		copy(cp, vv)
		out[k] = cp
	}

	return out
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.Body != nil && req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}

		clone.Body = body
	}

	return clone, nil
}
