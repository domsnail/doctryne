package utils

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RetryCodes maps response status code to max number of retries for it
type RetryCodes map[int]int

var DefaultRetryCodes = RetryCodes{
	http.StatusForbidden:          1,
	http.StatusTooManyRequests:    1,
	http.StatusGatewayTimeout:     3,
	http.StatusServiceUnavailable: 3,
	522:                           3, // Connection Timed Out (Cloudflare)
	524:                           3, // Timeout Occurred (Cloudflare)
}

// Retrier repeats http requests that return one of configured status codes
type Retrier struct {
	codes RetryCodes
	delay time.Duration
}

// NewRetrier returns nil when no codes provided, nil retrier sends requests without retries
func NewRetrier(codes map[int]int, delay time.Duration) *Retrier {
	if len(codes) == 0 {
		return nil
	}

	if delay <= 0 {
		panic("retry delay must greater than zero")
	}

	r := &Retrier{
		codes: make(RetryCodes, len(codes)),
		delay: delay,
	}

	for k, v := range codes {
		if v <= 0 {
			panic(fmt.Sprintf("invalid retry count for status code '%d': %d", k, v))
		}

		r.codes[k] = v
	}

	return r
}

// Do sends request and repeats it while response status code is listed in retry codes,
// each status code is retried up to its configured number of times,
// Retry-After response header overrides default delay when present
func (r *Retrier) Do(client *http.Client, req *http.Request) (*http.Response, error) {
	if r == nil {
		return client.Do(req)
	}

	var attempts = make(map[int]int)

	for {
		resp, err := client.Do(req.Clone(req.Context()))
		if err != nil {
			return nil, err
		}

		limit, ok := r.codes[resp.StatusCode]
		if !ok || attempts[resp.StatusCode] >= limit {
			return resp, nil
		}

		delay := r.delay
		if d, ok := ParseRetryAfter(resp.Header.Get("Retry-After"), time.Now()); ok {
			delay = d
		}

		attempts[resp.StatusCode]++
		slog.WarnContext(req.Context(), "http request failed, retrying...",
			slog.String("host", req.URL.Host),
			slog.String("path", req.URL.Path),
			slog.Int("status_code", resp.StatusCode),
			slog.Int("attempt", attempts[resp.StatusCode]),
			slog.Int("max_attempts", limit),
			slog.Duration("delay", delay),
		)

		// drain body so underlying connection can be reused
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		timer := time.NewTimer(delay)
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
		}
	}
}

// ParseRetryAfter parses Retry-After header value as delay in seconds or http date
// info: https://www.rfc-editor.org/rfc/rfc9110#field.retry-after
func ParseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.ParseUint(value, 10, 32); err == nil {
		return time.Duration(seconds) * time.Second, true
	}

	at, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}

	// date in the past means retry immediately
	return max(at.Sub(now), 0), true
}
