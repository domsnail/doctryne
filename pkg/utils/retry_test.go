package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type sequenceResponse struct {
	code       int
	retryAfter string
}

// newSequenceServer responds with given responses in order, then with 200 OK
func newSequenceServer(t *testing.T, responses ...sequenceResponse) (*httptest.Server, *atomic.Int32) {
	var calls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(calls.Add(1)) - 1
		if i < len(responses) {
			if responses[i].retryAfter != "" {
				w.Header().Set("Retry-After", responses[i].retryAfter)
			}

			w.WriteHeader(responses[i].code)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(srv.Close)
	return srv, &calls
}

func codes(cs ...int) []sequenceResponse {
	var r = make([]sequenceResponse, len(cs))
	for i, c := range cs {
		r[i] = sequenceResponse{code: c}
	}

	return r
}

func doRequest(t *testing.T, ctx context.Context, r *Retrier, srv *httptest.Server) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := r.Do(srv.Client(), req)
	if resp != nil {
		t.Cleanup(func() { _ = resp.Body.Close() })
	}

	return resp, err
}

func TestRetrier_Do(t *testing.T) {
	const delay = time.Millisecond

	t.Run("succeeds after retrying listed status code", func(t *testing.T) {
		srv, calls := newSequenceServer(t, codes(http.StatusServiceUnavailable, http.StatusServiceUnavailable)...)
		r := NewRetrier(map[int]int{http.StatusServiceUnavailable: 2}, delay)

		resp, err := doRequest(t, context.Background(), r, srv)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.EqualValues(t, 3, calls.Load())
	})

	t.Run("returns last response when retries exhausted", func(t *testing.T) {
		srv, calls := newSequenceServer(t, codes(http.StatusTooManyRequests, http.StatusTooManyRequests, http.StatusTooManyRequests)...)
		r := NewRetrier(map[int]int{http.StatusTooManyRequests: 2}, delay)

		resp, err := doRequest(t, context.Background(), r, srv)
		require.NoError(t, err)
		require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		require.EqualValues(t, 3, calls.Load())
	})

	t.Run("does not retry unlisted status code", func(t *testing.T) {
		srv, calls := newSequenceServer(t, codes(http.StatusForbidden)...)
		r := NewRetrier(map[int]int{http.StatusServiceUnavailable: 3}, delay)

		resp, err := doRequest(t, context.Background(), r, srv)
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, resp.StatusCode)
		require.EqualValues(t, 1, calls.Load())
	})

	t.Run("nil retrier does not retry", func(t *testing.T) {
		srv, calls := newSequenceServer(t, codes(http.StatusServiceUnavailable)...)
		var r *Retrier

		resp, err := doRequest(t, context.Background(), r, srv)
		require.NoError(t, err)
		require.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		require.EqualValues(t, 1, calls.Load())
	})

	t.Run("counts retries per status code", func(t *testing.T) {
		srv, calls := newSequenceServer(t, codes(522, http.StatusGatewayTimeout, 522, http.StatusGatewayTimeout)...)
		r := NewRetrier(map[int]int{522: 2, http.StatusGatewayTimeout: 2}, delay)

		resp, err := doRequest(t, context.Background(), r, srv)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.EqualValues(t, 5, calls.Load())
	})

	t.Run("stops waiting when context cancelled", func(t *testing.T) {
		srv, calls := newSequenceServer(t, codes(http.StatusServiceUnavailable)...)
		r := NewRetrier(map[int]int{http.StatusServiceUnavailable: 1}, time.Hour)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := doRequest(t, ctx, r, srv)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.EqualValues(t, 1, calls.Load())
	})

	t.Run("respects retry-after header", func(t *testing.T) {
		srv, calls := newSequenceServer(t, sequenceResponse{code: http.StatusTooManyRequests, retryAfter: "0"})

		// default delay would exceed context timeout, header must override it
		r := NewRetrier(map[int]int{http.StatusTooManyRequests: 1}, time.Hour)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := doRequest(t, ctx, r, srv)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.EqualValues(t, 2, calls.Load())
	})
}

func TestNewRetrier(t *testing.T) {
	t.Run("default retry codes are accepted", func(t *testing.T) {
		require.NotPanics(t, func() {
			require.NotNil(t, NewRetrier(DefaultRetryCodes, time.Second))
		})
	})

	t.Run("empty retry codes return nil retrier", func(t *testing.T) {
		require.Nil(t, NewRetrier(nil, time.Second))
		require.Nil(t, NewRetrier(map[int]int{}, 0))
	})

	t.Run("invalid retry config panics", func(t *testing.T) {
		require.Panics(t, func() { NewRetrier(map[int]int{503: 1}, 0) })
		require.Panics(t, func() { NewRetrier(map[int]int{503: 0}, time.Second) })
	})
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)

	var cases = []struct {
		value string
		must  time.Duration
		ok    bool
	}{
		{value: "", ok: false},
		{value: "0", must: 0, ok: true},
		{value: "120", must: 2 * time.Minute, ok: true},
		{value: " 5 ", must: 5 * time.Second, ok: true},
		{value: "-5", ok: false},
		{value: "1.5", ok: false},
		{value: "soon", ok: false},
		{value: "Tue, 27 Oct 2026 12:00:30 GMT", must: 30 * time.Second, ok: true},
		{value: "Tue, 27 Oct 2026 11:00:00 GMT", must: 0, ok: true},
	}

	for _, c := range cases {
		d, ok := ParseRetryAfter(c.value, now)
		require.Equal(t, c.ok, ok, "value: %q", c.value)
		require.Equal(t, c.must, d, "value: %q", c.value)
	}
}
