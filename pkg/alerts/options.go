package alerts

import (
	"log/slog"
	"net/http"
)

type Option func(service *GitlabAlertHandler)

func WithToken(token string) Option {
	return func(s *GitlabAlertHandler) {
		s.token = token
	}
}

func WithEndpoint(url string) Option {
	return func(s *GitlabAlertHandler) {
		s.endpoint = url
	}
}

func WithHost(host string) Option {
	return func(s *GitlabAlertHandler) {
		s.host = host
	}
}

func WithEnvironment(env string) Option {
	return func(s *GitlabAlertHandler) {
		s.environment = env
	}
}

func WithMinLevel(level slog.Level) Option {
	return func(s *GitlabAlertHandler) {
		s.minLevel = level
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(s *GitlabAlertHandler) {
		s.c = client
	}
}
