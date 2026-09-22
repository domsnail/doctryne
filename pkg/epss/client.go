package epss

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL        = "https://api.first.org/epss/"
	defaultCsvDownloadURL = "https://epss.empiricalsecurity.com/epss_scores-current.csv.gz"

	defaultResultsPerPage = 100
)

type Client struct {
	baseURL     string
	downloadURL string

	h *http.Client
}

func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:     defaultBaseURL,
		downloadURL: defaultCsvDownloadURL,
		h:           http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := url.Parse(c.baseURL)
	if err != nil {
		panic(fmt.Errorf("invalid base url: %s", err))
	}

	_, err = url.Parse(c.downloadURL)
	if err != nil {
		panic(fmt.Errorf("invalid download  url: %s", err))
	}

	return c
}

func (c *Client) DownloadCurrentDatabase(ctx context.Context) (*EpssDatabase, error) {
	return c.downloadDatabase(ctx, nil)
}

func (c *Client) DownloadDatabaseOnDate(ctx context.Context, date time.Time) (*EpssDatabase, error) {
	return c.downloadDatabase(ctx, &date)
}

func (c *Client) downloadDatabase(ctx context.Context, date *time.Time) (*EpssDatabase, error) {
	var downloadURL string
	if date == nil {
		downloadURL = c.downloadURL
	} else {
		downloadURL = strings.ReplaceAll(c.downloadURL, "current", date.Format("2006-01-02"))
	}

	slog.DebugContext(ctx, "downloading epss database...",
		slog.String("download_url", downloadURL),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %w", err)
	}

	resp, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gzipped response: %w", err)
	}

	defer func() { _ = reader.Close() }()
	var db EpssDatabase

	var totalLines int
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	csvReader.Comma = ','

	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read epss file header: %w", err)
	}

	db.Version, db.Date, err = parseFirstLine(header)
	if err != nil {
		return nil, fmt.Errorf("failed to parse epss file header: %w", err)
	}

	_, err = csvReader.Read() // read header
	for {
		line, err := csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to parse csv line: %w", err)
		}

		totalLines++
		epss, err := strconv.ParseFloat(line[1], 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse epss value on line %d: %w", totalLines+2, err)
		}

		percentile, err := strconv.ParseFloat(line[2], 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse percentile value on line %d: %w", totalLines+2, err)
		}

		db.Records = append(db.Records, EpssRecord{
			CveID:      line[0],
			Epss:       float32(epss),
			Percentile: float32(percentile),
		})
	}

	slog.Debug("successfully downloaded epss records",
		slog.Time("date", db.Date),
		slog.String("version", db.Version),
		slog.Int("total_records", len(db.Records)),
	)

	return &db, nil
}

// parseFirstLine
// for example: #model_version:v2026.06.15,score_date:2026-09-21T12:03:23Z
func parseFirstLine(line []string) (version string, date time.Time, err error) {
	if len(line) != 2 {
		return "", time.Time{}, fmt.Errorf("invalid header line: %s", line)
	}

	v, ok := strings.CutPrefix(line[0], "#model_version:")
	if !ok {
		return "", time.Time{}, fmt.Errorf("missing model_version in header line: %s", line)
	}

	version = v

	v, ok = strings.CutPrefix(line[1], "score_date:")
	if !ok {
		return "", time.Time{}, fmt.Errorf("missing score_date in header line: %s", line)
	}

	date, err = time.Parse(time.RFC3339, v)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse score_date: %s", line)
	}

	return version, date, nil
}

func (c *Client) GetRemoteURL() url.URL {
	u, _ := url.Parse(c.baseURL)
	return *u
}
