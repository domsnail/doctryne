package http

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/domsnail/doctryne/cfg"
	"github.com/stretchr/testify/require"
)

const (
	delay         = time.Second * 3
	refreshPeriod = time.Second * 30
	cacheTTL      = time.Minute * 30
)

func Test_CachedHTTPClient(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	config, err := cfg.NewConfigFromEnv()
	require.NoError(t, err)

	cfg.SetGlobalConfig(config)

	t.Run("test package.json inspection with errors", func(t *testing.T) {
		transport := NewSmartTransport(TransportOptions{
			CachedMethods: []string{http.MethodGet},
			CacheTTL:      cacheTTL,
			ThrottleOptions: &ThrottleOptions{
				MaxRequests:   5,
				RefreshPeriod: refreshPeriod,
				MinDelay:      delay,
			},
		})

		client := http.Client{
			Transport: transport,
		}

		resp, err := client.Get("https://api.weather.gov/stations?limit=1")
		require.NoError(t, err)

		body1, _ := io.ReadAll(resp.Body)
		slog.Debug("response_received",
			slog.Int("body_length", len(body1)),
			slog.Any("headers", resp.Header),
		)

		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, resp.Header, "Content-Type")
		require.NotContains(t, resp.Header, "X-From-Cache")

		require.Len(t, transport.cache.entries, 1)

		resp, err = client.Get("https://api.weather.gov/stations?limit=1")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, resp.Header, "Content-Type")
		require.Contains(t, resp.Header, "X-From-Cache", "must contain cached header on second request")

		body2, _ := io.ReadAll(resp.Body)
		slog.Debug("cached_response_received",
			slog.Int("body_length", len(body2)),
			slog.Any("headers", resp.Header),
		)

		require.Equal(t, body1, body2)

		require.Len(t, transport.cache.entries, 1)

		startedAt := time.Now()
		resp, err = client.Get("https://api.weather.gov/stations/001ID")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, resp.Header, "Content-Type")
		require.NotContains(t, resp.Header, "X-From-Cache", "must not contain cached header on third request")

		body3, _ := io.ReadAll(resp.Body)
		slog.Debug("response_received",
			slog.Int("body_length", len(body3)),
			slog.Any("headers", resp.Header),
			slog.Duration("time_taken", time.Since(startedAt)),
		)

		require.NotEqual(t, body2, body3)

		require.Len(t, transport.cache.entries, 2)

		resp, err = client.Get("https://api.weather.gov/stations/001BH")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, transport.cache.entries, 3)
		slog.Debug("response_received", slog.Duration("time_taken", time.Since(startedAt)))

		resp, err = client.Get("https://api.weather.gov/stations/001CE")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, transport.cache.entries, 4)
		slog.Debug("response_received", slog.Duration("time_taken", time.Since(startedAt)))

		resp, err = client.Get("https://api.weather.gov/stations/001HI")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, transport.cache.entries, 5)
		slog.Debug("response_received", slog.Duration("time_taken", time.Since(startedAt)))

		resp, err = client.Get("https://api.weather.gov/stations/001HI")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, transport.cache.entries, 5)
		slog.Debug("cached_response_received", slog.Duration("time_taken", time.Since(startedAt)))

		resp, err = client.Get("https://api.weather.gov/stations/001HE")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Len(t, transport.cache.entries, 6)
		slog.Debug("response_received", slog.Duration("time_taken", time.Since(startedAt)))
	})
}
