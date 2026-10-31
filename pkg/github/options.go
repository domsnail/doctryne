package github

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/domsnail/doctryne/pkg/utils"
)

type Option func(*Client)

func WithBearerToken(token string) Option {
	return func(s *Client) {
		s.token = fmt.Sprintf("Bearer %s", token)
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

func WithRetry(retryCodes map[int]int, delay time.Duration) Option {
	return func(s *Client) {
		s.retry = utils.NewRetrier(retryCodes, delay)
	}
}
