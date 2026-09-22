package epss

import (
	"log/slog"
	"net/http"
	"strings"
)

type Option func(*Client)

// WithDownloadURL overrides database download url
// provide only the current database url, for example: https://gitlab.domsnail.ru/public-data/epss/-/raw/main/epss_scores-current.csv.gz?ref_type=heads
func WithDownloadURL(url string) Option {
	return func(s *Client) {
		if !strings.Contains(url, "current") {
			slog.Warn("please provide current epss database url for correct fetching")
		}

		s.downloadURL = url
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(s *Client) {
		s.h = client
	}
}
