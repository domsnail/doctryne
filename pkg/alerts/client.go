package alerts

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
	"uuid"
)

type GitlabAlertHandler struct {
	c *http.Client

	environment string
	endpoint    string
	token       string
	host        string

	minLevel slog.Level
}

func NewGitlabAlertClient(opts ...Option) *GitlabAlertHandler {
	c := &GitlabAlertHandler{
		c: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := url.Parse(c.endpoint)
	if err != nil {
		panic(fmt.Errorf("invalid endpoint: %s", err))
	}

	if c.token == "" {
		panic("token (authorization key) is required")
	}

	if c.environment == "" {
		panic("environment is required")
	}

	return c
}

func (h *GitlabAlertHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *GitlabAlertHandler) Handle(ctx context.Context, record slog.Record) error {
	severity, ok := slogSeverityMap[record.Level]
	if !ok {
		severity = SeverityUnknown
	}

	alert := Alert{
		Fingerprint: uuid.NewV7(),
		Environment: h.environment,
		Host:        h.host,
		Tool:        "slog",
		Title:       record.Message,
		Severity:    severity,
		StartedAt:   new(time.Now().Local()),
	}

	service := ctx.Value("service")
	if service != nil {
		alert.Service = new(service.(string))
	}

	if record.NumAttrs() > 0 {
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "error" {
				alert.Description = attr.Value.String()
				return false
			}

			return true
		})
	}

	payload, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", h.token))
	req.Header.Add("Content-Type", "application/json")
	resp, err := h.c.Do(req)
	if err != nil {
		return err
	}

	_ = resp.Body.Close()
	return nil
}

func (h *GitlabAlertHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	h2 := *h
	return &h2
}

func (h *GitlabAlertHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	h2 := *h
	return &h2
}
