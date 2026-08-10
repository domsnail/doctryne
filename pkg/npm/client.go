package npm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	http_smart_transport "github.com/domsnail/doctryne/pkg/http"
)

const (
	defaultRegistryURL = "https://registry.npmjs.org"
	defaultApiURL      = "https://api.npmjs.org"

	defaultCacheTTL = 4 * time.Hour

	// npm does not explicitly set rate limits on its api, but this should be enough
	// ref: https://blog.npmjs.org/post/187698412060/acceptible-use.html
	defaultRateLimit_Period      = time.Second * 60
	defaultRateLimit_MaxRequests = 5000
	defaultRateLimit_MinDelay    = 500 * time.Millisecond
)

// Client is JavaScript NPM Registry API client
// ref: https://github.com/npm/registry/blob/main/docs/REGISTRY-API.md
type Client struct {
	h *http.Client

	registry *url.URL
	api      *url.URL
}

type Options struct {
	BearerToken string

	ApiURL      string
	RegistryURL string

	CacheTTL time.Duration
}

func NewClient(opts Options) (*Client, error) {
	slog.Debug("initializing npm client...",
		slog.Bool("using_bearer_token", opts.BearerToken != ""),
	)

	var err error
	if opts.CacheTTL == 0 {
		slog.Warn("no cache ttl set for npm, setting to default value",
			slog.Duration("default_cache_ttl", defaultCacheTTL),
		)

		opts.CacheTTL = defaultCacheTTL
	}

	var transport = http_smart_transport.NewSmartTransport(http_smart_transport.TransportOptions{
		BaseTransport: http.DefaultTransport,
		CachedMethods: []string{http.MethodGet},
		CacheTTL:      opts.CacheTTL,
		ThrottleOptions: &http_smart_transport.ThrottleOptions{
			RefreshPeriod: defaultRateLimit_Period,
			MaxRequests:   defaultRateLimit_MaxRequests,
			MinDelay:      defaultRateLimit_MinDelay,
		},
	})

	var client http.Client
	if opts.BearerToken != "" {
		client = http.Client{
			Transport: &bearerTransport{
				token: opts.BearerToken,
				base:  transport,
			},
			Timeout: http.DefaultClient.Timeout,
		}
	} else {
		slog.Warn("bearer token not set for npm",
			slog.String("details", "please consider using bearer token"),
		)

		client = http.Client{
			Transport: transport,
			Timeout:   http.DefaultClient.Timeout,
		}
	}

	var (
		registry *url.URL
		api      *url.URL
	)

	if opts.ApiURL != "" {
		slog.Debug("overriding npm api url", slog.String("url", opts.ApiURL))

		api, err = url.Parse(opts.ApiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse npm api url: %w", err)
		}
	} else {
		slog.Debug("using default npm api url...", slog.String("url", defaultApiURL))

		api, err = url.Parse(defaultApiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse default api registry url: %w", err)
		}
	}

	if opts.RegistryURL != "" {
		slog.Debug("overriding npm registry url", slog.String("url", opts.RegistryURL))

		registry, err = url.Parse(opts.RegistryURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse npm registry url: %w", err)
		}
	} else {
		slog.Debug("using default npm registry url...", slog.String("url", defaultRegistryURL))

		registry, err = url.Parse(defaultRegistryURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse default npm registry url: %w", err)
		}
	}

	slog.Info("initialized npm client",
		slog.String("registry_url", registry.Redacted()),
		slog.String("api_url", api.Redacted()),
		slog.Bool("using_bearer_token", opts.BearerToken != ""),
		slog.Duration("cache_ttl", opts.CacheTTL),
		slog.Group("rate_limiting",
			slog.Duration("period", defaultRateLimit_Period),
			slog.Int("max_requests", defaultRateLimit_MaxRequests),
			slog.Duration("min_delay", defaultRateLimit_MinDelay),
		),
	)

	return &Client{
		registry: registry,
		api:      api,
		h:        &client,
	}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	queryURL := fmt.Sprintf("%s/-/whoami", c.registry)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return fmt.Errorf("failed to prepare request: %w", err)
	}

	slog.DebugContext(ctx, "pinging npm registry",
		slog.String("method", http.MethodGet),
		slog.String("query_url", queryURL),
	)

	resp, err := c.h.Do(req)
	if err != nil {
		slog.DebugContext(ctx, "failed to perform request",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.String("error", err.Error()),
		)

		return fmt.Errorf("failed ping npm: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.DebugContext(ctx, "received non-successful status code",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
		)

		return fmt.Errorf("failed ping npm: %s (%d)", resp.Status, resp.StatusCode)
	}

	var me Me
	if err = json.NewDecoder(resp.Body).Decode(&me); err != nil {
		slog.WarnContext(ctx, "failed to unmarshal npm response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)
	} else {
		slog.InfoContext(ctx, "provided active npm bearer token",
			slog.String("username", me.Username),
		)
	}

	return nil
}

func (c *Client) GetPackage(ctx context.Context, name string) (*Package, json.RawMessage, error) {
	if len(name) == 0 {
		return nil, nil, fmt.Errorf("package name is required")
	}

	queryURL := fmt.Sprintf("%s/%s", c.registry, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	slog.DebugContext(ctx, "querying npm registry",
		slog.String("method", http.MethodGet),
		slog.String("query_url", queryURL),
	)

	resp, err := c.h.Do(req)
	if err != nil {
		slog.DebugContext(ctx, "failed to perform request",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.String("error", err.Error()),
		)

		return nil, nil, fmt.Errorf("failed fetch npm package info: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		slog.WarnContext(ctx, "package not found in npm registry",
			slog.String("package_name", name),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, nil, errors.New("package not found")
	}

	if resp.StatusCode != http.StatusOK {
		slog.DebugContext(ctx, "received non-successful status code",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, nil, fmt.Errorf("failed to fetch npm package info: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.DebugContext(ctx, "failed to read npm package response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, nil, fmt.Errorf("failed to read npm package info: %w", err)
	}

	var info Package
	if err = json.Unmarshal(body, &info); err != nil {
		slog.WarnContext(ctx, "failed to unmarshal npm response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, nil, fmt.Errorf("failed to unmarshal npm package response: %w", err)
	}

	slog.DebugContext(ctx, "successfully fetched npm package",
		slog.String("package_name", name),
		slog.Int("versions_count", len(info.Versions)),
	)

	return &info, body, nil
}

func (c *Client) GetPackageStats(ctx context.Context, name string, period time.Duration) (*Stats, error) {
	if len(name) == 0 || period == 0 {
		return nil, fmt.Errorf("package name and period are required")
	}

	end := time.Now()
	start := end.Add(-period)

	queryURL := fmt.Sprintf("%s/downloads/point/%s:%s/%s", c.api, start.Format("2006-01-02"), end.Format("2006-01-02"), name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	slog.DebugContext(ctx, "querying npm api",
		slog.String("method", http.MethodGet),
		slog.String("query_url", queryURL),
	)

	resp, err := c.h.Do(req)
	if err != nil {
		slog.DebugContext(ctx, "failed to perform request",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed fetch npm package stats: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		slog.WarnContext(ctx, "package not found in npm registry",
			slog.String("package_name", name),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, errors.New("package not found")
	}

	if resp.StatusCode != http.StatusOK {
		slog.DebugContext(ctx, "received non-successful status code",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, fmt.Errorf("failed fetch npm package stats: status code %d", resp.StatusCode)
	}

	var stats Stats
	if err = json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		slog.WarnContext(ctx, "failed to unmarshal npm response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed unmarshal npm stats response: %w", err)
	}

	slog.DebugContext(ctx, "successfully fetched npm package stats",
		slog.String("package_name", stats.Package),
		slog.Uint64("downloads", stats.Downloads),
	)

	return &stats, nil
}
