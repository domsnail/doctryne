package cve

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL           = "https://services.nvd.nist.gov/rest/json"
	defaultCVEPath           = "/cves/2.0"
	defaultChangeHistoryPath = "/cvehistory/2.0"

	defaultResultsPerPage = 2000
	defaultHistoryRange   = 72 * time.Hour
)

type Client struct {
	baseURL string
	h       *http.Client

	token string

	perPage      uint16
	historyRange time.Duration
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:      defaultBaseURL,
		h:            http.DefaultClient,
		token:        "",
		perPage:      defaultResultsPerPage,
		historyRange: defaultHistoryRange,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.token == "" {
		slog.Warn("nist access token is not set",
			slog.String("details", "please consider using nist personal access token, see: https://nvd.nist.gov/developers/request-an-api-key"),
		)
	}

	return c
}

// GetRecords implements
// info: https://nvd.nist.gov/developers/vulnerabilities
func (c *Client) GetRecords(ctx context.Context, opts RecordsQueryOptions, offset, limit int) (*RecordsQueryResponse, error) {
	err := opts.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+defaultCVEPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	q := req.URL.Query()
	if opts.ModificationDateStart != nil {
		q.Add("lastModStartDate", opts.ModificationDateStart.Format(time.RFC3339))
	}

	if opts.ModificationDateStart != nil {
		q.Add("lastModEndDate", opts.ModificationDateStart.Format(time.RFC3339))
	}

	if len(opts.CveIDs) > 0 {
		q.Add("cveIds", strings.Join(opts.CveIDs, ","))
	}

	q.Add("resultsPerPage", strconv.Itoa(limit))
	q.Add("startIndex", strconv.Itoa(offset))

	req.URL.RawQuery = q.Encode()
	c.setHeaders(req)

	resp, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var records RecordsQueryResponse
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &records, nil
}
