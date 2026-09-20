package github

import (
	"fmt"
	"net/http"
	"net/url"
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
