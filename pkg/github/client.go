package github

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

const (
	defaultBaseURL        = "https://api.github.com"
	defaultAdvisoriesPath = "/advisories"
	defaultApiVersion     = "2026-03-10"

	defaultResultsPerPage = 100
)

type Client struct {
	baseURL string
	h       *http.Client

	token string
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		h:       http.DefaultClient,
		token:   "",
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := url.Parse(defaultBaseURL)
	if err != nil {
		panic(fmt.Errorf("invalid base url: %s", err))
	}

	if c.token == "" {
		slog.Warn("github bearer token is not set",
			slog.String("details", "please consider using github personal access token (PAT), see: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens"),
		)
	}

	return c
}

func (c *Client) GetAdvisories(ctx context.Context, opts *AdvisoriesQueryOptions, perPage, page int) ([]AdvisoryRecord, error) {
	if perPage > 100 {
		return nil, fmt.Errorf("per_page exceeds 100 records")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+defaultAdvisoriesPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	q := req.URL.Query()
	if opts.ModifiedAfter != nil && opts.ModifiedBefore != nil {
		q.Add("modified", fmt.Sprintf("%s..%s", opts.ModifiedAfter.Format("2006-01-02"), opts.ModifiedBefore.Format("2006-01-02")))
	} else if opts.ModifiedAfter != nil {
		q.Add("modified", fmt.Sprintf(">=%s", opts.ModifiedAfter.Format("2006-01-02")))
	} else if opts.ModifiedBefore != nil {
		q.Add("modified", fmt.Sprintf("<=%s", opts.ModifiedBefore.Format("2006-01-02")))
	}

	if opts.CveID != "" {
		q.Add("cve_id", opts.CveID)
	}

	if opts.GhsaID != "" {
		q.Add("ghsa_id", opts.GhsaID)
	}

	if opts.Direction != "" {
		q.Add("direction", opts.Direction)
	}

	if opts.Sort != "" {
		q.Add("sort", opts.Sort)
	}

	q.Add("per_page", strconv.Itoa(perPage))
	q.Add("page", strconv.Itoa(page))

	req.URL.RawQuery = q.Encode()
	c.setHeaders(req)

	resp, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var apiError ApiError
		_ = json.NewDecoder(resp.Body).Decode(&apiError)

		slog.WarnContext(ctx, "failed to fetch github advisories",
			slog.Int("status_code", resp.StatusCode),
			slog.Any("details", apiError),
		)

		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read gzipped response body: %w", err)
		}

		defer reader.Close()
	default:
		reader = resp.Body
	}

	var records []AdvisoryRecord
	err = json.NewDecoder(reader).Decode(&records)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return records, nil
}
