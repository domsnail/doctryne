package stack_exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	http_smart_transport "github.com/domsnail/doctryne/pkg/http"
)

const (
	defaultApiURL = "https://api.stackexchange.com/2.3"

	defaultCacheTTL = 8 * time.Hour

	// ref: https://api.stackexchange.com/docs/throttle
	defaultRateLimit_Period      = time.Second * 60
	defaultRateLimit_MaxRequests = 1000
	defaultRateLimit_MinDelay    = 500 * time.Millisecond
)

// Client is StackExchange (StackOverflow etc.) API client
// ref: https://api.stackexchange.com/docs
// ref: https://stackapps.com/
type Client struct {
	h *http.Client

	api *url.URL
}

type Options struct {
	AccessToken string

	ApiURL string

	CacheTTL time.Duration
}

func NewClient(opts Options) (*Client, error) {
	slog.Debug("initializing stack exchange client...",
		slog.Bool("using_access_token", opts.AccessToken != ""),
	)

	var err error
	if opts.CacheTTL == 0 {
		slog.Warn("no cache ttl set for stack exchange, setting to default value",
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
	if opts.AccessToken != "" {
		client = http.Client{
			Transport: &bearerTransport{
				token: opts.AccessToken,
				base:  transport,
			},
			Timeout: http.DefaultClient.Timeout,
		}
	} else {
		slog.Warn("bearer token not set for stack exchange",
			slog.String("details", "please consider using access token"),
		)

		client = http.Client{
			Transport: transport,
			Timeout:   http.DefaultClient.Timeout,
		}
	}

	var (
		api *url.URL
	)

	if opts.ApiURL != "" {
		slog.Debug("overriding stack exchange api url", slog.String("url", opts.ApiURL))

		api, err = url.Parse(opts.ApiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse stack exchange api url: %w", err)
		}
	} else {
		slog.Debug("using default stack exchange api url...", slog.String("url", defaultApiURL))

		api, err = url.Parse(defaultApiURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse default api url: %w", err)
		}
	}

	slog.Info("initialized stack exchange client",
		slog.String("api_url", api.Redacted()),
		slog.Bool("using_access_token", opts.AccessToken != ""),
		slog.Duration("cache_ttl", opts.CacheTTL),
		slog.Group("rate_limiting",
			slog.Duration("period", defaultRateLimit_Period),
			slog.Int("max_requests", defaultRateLimit_MaxRequests),
			slog.Duration("min_delay", defaultRateLimit_MinDelay),
		),
	)

	return &Client{
		api: api,
		h:   &client,
	}, nil
}

func (c *Client) GetMe(ctx context.Context) (*User, json.RawMessage, error) {
	queryURL := fmt.Sprintf("%s/me?site=%s", c.api, StackSite_StackOverflow)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	slog.DebugContext(ctx, "querying stack exchange api",
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

		return nil, nil, fmt.Errorf("failed fetch stack exchange about me info: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.DebugContext(ctx, "received non-successful status code",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, nil, fmt.Errorf("failed to fetch stack exchange about me info: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.DebugContext(ctx, "failed to read stack exchange about me response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, nil, fmt.Errorf("failed to read stack exchange about me info: %w", err)
	}

	var me User
	if err = json.Unmarshal(body, &me); err != nil {
		slog.WarnContext(ctx, "failed to unmarshal stack exchange user response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, nil, fmt.Errorf("failed to unmarshal stack exchange about me response: %w", err)
	}

	slog.DebugContext(ctx, "successfully fetched stack exchange about me",
		slog.String("username", me.DisplayName),
	)

	return &me, body, nil
}

// GetUsersByUsername implements https://api.stackexchange.com/docs/users
func (c *Client) GetUsersByUsername(ctx context.Context, username string) ([]*User, json.RawMessage, error) {
	if len(username) == 0 {
		return nil, nil, fmt.Errorf("username is required")
	}

	queryURL := fmt.Sprintf("%s/users?inname=%s&site=%s&max=100", c.api, url.QueryEscape(username), StackSite_StackOverflow)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	slog.DebugContext(ctx, "querying stack exchange api",
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

		return nil, nil, fmt.Errorf("failed fetch stack exchange users info: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.DebugContext(ctx, "received non-successful status code",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, nil, fmt.Errorf("failed to fetch stack exchange users info: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.DebugContext(ctx, "failed to read stack exchange users response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, nil, fmt.Errorf("failed to read stack exchange users info: %w", err)
	}

	var list UserList
	if err = json.Unmarshal(body, &list); err != nil {
		slog.WarnContext(ctx, "failed to unmarshal stack exchange users response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, nil, fmt.Errorf("failed to unmarshal stack exchange user response: %w", err)
	}

	if len(list.Items) == 0 {
		slog.DebugContext(ctx, "users not found on stack exchange",
			slog.String("username", username),
			slog.Int("users_count", len(list.Items)),
		)

		return nil, nil, nil
	}

	slog.DebugContext(ctx, "successfully fetched stack exchange users",
		slog.String("username", username),
		slog.Int("users_count", len(list.Items)),
	)

	return list.Items, body, nil
}

// GetUserByID implements https://api.stackexchange.com/docs/users-by-ids
func (c *Client) GetUserByID(ctx context.Context, id int64) (*User, json.RawMessage, error) {
	if id == 0 {
		return nil, nil, fmt.Errorf("id is required")
	}

	queryURL := fmt.Sprintf("%s/users/%d?site=%s", c.api, id, StackSite_StackOverflow)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	slog.DebugContext(ctx, "querying stack exchange api",
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

		return nil, nil, fmt.Errorf("failed fetch stack exchange user info: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.DebugContext(ctx, "received non-successful status code",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
		)

		return nil, nil, fmt.Errorf("failed to fetch stack exchange user info: status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.DebugContext(ctx, "failed to read stack exchange user response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, nil, fmt.Errorf("failed to read stack exchange user info: %w", err)
	}

	var list UserList
	if err = json.Unmarshal(body, &list); err != nil {
		slog.WarnContext(ctx, "failed to unmarshal stack exchange user response",
			slog.String("method", http.MethodGet),
			slog.String("query_url", queryURL),
			slog.Int("status_code", resp.StatusCode),
			slog.String("error", err.Error()),
		)

		return nil, nil, fmt.Errorf("failed to unmarshal stack exchange user response: %w", err)
	}

	if len(list.Items) == 0 {
		slog.DebugContext(ctx, "user not found on stack exchange",
			slog.Int64("user_id", id),
			slog.Int("users_count", len(list.Items)),
		)

		return nil, nil, nil
	}

	slog.DebugContext(ctx, "successfully fetched stack exchange user",
		slog.Int64("user_id", id),
		slog.Int("users_count", len(list.Items)),
	)

	return list.Items[0], body, nil
}
