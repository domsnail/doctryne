package npm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const defaultRegistryURL = "https://registry.npmjs.org"
const defaultApiURL = "https://api.npmjs.org"

// Client is JavaScript NPM Registry API client
// ref: https://github.com/npm/registry/blob/main/docs/REGISTRY-API.md
type Client struct {
	h *http.Client

	registry *url.URL
	api      *url.URL
}

type Options struct {
	Timeout  time.Duration
	ProxyURL *url.URL

	BearerToken string

	ApiURL      string
	RegistryURL string
}

func NewClient(opts Options) (*Client, error) {
	slog.Debug("initializing npm client...")
	var err error

	if opts.Timeout == 0 {
		return nil, errors.New("timeout is required")
	}

	var transport http.RoundTripper = &http.Transport{}
	if opts.ProxyURL != nil {
		slog.Debug("using proxy for npm client", slog.String("proxy_url", opts.ProxyURL.Redacted()))
		transport = &http.Transport{
			Proxy: http.ProxyURL(opts.ProxyURL),
		}
	}

	if opts.BearerToken == "" {
		slog.Debug("npm bearer token is not set")
	} else {
		slog.Debug("setting npm bearer token...")

		transport = &bearerTransport{
			token: opts.BearerToken,
			base:  transport,
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

	return &Client{
		registry: registry,
		api:      api,
		h: &http.Client{
			Transport: transport,
			Timeout:   opts.Timeout,
		},
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
		slog.DebugContext(ctx, "failed to unmarshal npm response",
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

func (c *Client) GetPackage(ctx context.Context, name string) (*Package, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("package name is required")
	}

	queryURL := fmt.Sprintf("%s/%s", c.registry, name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
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

		return nil, fmt.Errorf("failed fetch npm package info: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		slog.WarnContext(ctx, "package not found",
			slog.String("package_name", name),
			slog.String("package_version", version),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		slog.DebugContext(ctx, "received non-successful status code",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, fmt.Errorf("failed to fetch npm package info: status code %d", resp.StatusCode)
	}

	var info Package
	if err = json.NewDecoder(resp.Body).Decode(&info); err != nil {
		slog.DebugContext(ctx, "failed to unmarshal npm response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to unmarshal npm package response: %w", err)
	}

	slog.DebugContext(ctx, "successfully fetched npm package",
		slog.String("package_name", name),
		slog.Int("versions_count", len(info.Versions)),
	)

	return &info, nil
}

func (c *Client) GetPackageStats(ctx context.Context, name string, period PackageStatsPeriod) (*Stats, error) {
	if len(name) == 0 || len(period) == 0 {
		return nil, fmt.Errorf("package name and period are required")
	}

	queryURL := fmt.Sprintf("%s/downloads/point/%s/%s", c.api, period, name)

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
		slog.WarnContext(ctx, "package not found",
			slog.String("package_name", name),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, nil
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
		slog.DebugContext(ctx, "failed to unmarshal npm response",
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
