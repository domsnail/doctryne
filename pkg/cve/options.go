package cve

import (
	"net/http"
	"net/url"
	"time"
)

type Option func(*Client)

func WithAccessToken(token string) Option {
	return func(s *Client) {
		s.token = token
	}
}

func WithBaseURL(url *url.URL) Option {
	return func(s *Client) {
		s.baseURL = url.String()
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(s *Client) {
		s.h = client
	}
}

func WithHistoryRange(period time.Duration) Option {
	return func(s *Client) {
		s.historyRange = period
	}
}

func WithPerPage(limit uint16) Option {
	return func(s *Client) {
		s.perPage = limit
	}
}
