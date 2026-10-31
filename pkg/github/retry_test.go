package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// retry behaviour itself is covered in pkg/utils, this only checks client wiring
func TestClient_GetAdvisoriesRetry(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	t.Run("retries when configured", func(t *testing.T) {
		calls.Store(0)
		c := NewClient(WithBaseURL(u), WithBearerToken("test"),
			WithRetry(map[int]int{http.StatusServiceUnavailable: 1}, time.Millisecond),
		)

		_, _, err := c.GetAdvisories(context.Background(), AdvisoriesQueryOptions{}, 10, 1)
		require.NoError(t, err)
		require.EqualValues(t, 2, calls.Load())
	})

	t.Run("does not retry without retry option", func(t *testing.T) {
		calls.Store(0)
		c := NewClient(WithBaseURL(u), WithBearerToken("test"))

		_, _, err := c.GetAdvisories(context.Background(), AdvisoriesQueryOptions{}, 10, 1)
		require.ErrorContains(t, err, "unexpected status code: 503")
		require.EqualValues(t, 1, calls.Load())
	})
}
