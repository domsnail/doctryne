package http

import (
	"log/slog"
	"net/http"
	"time"
)

func defaultSlogMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Log the completed request structured fields
			slog.InfoContext(r.Context(), "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Duration("latency", time.Since(start)),
				slog.String("ip", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)

			next.ServeHTTP(w, r)
		})
	}
}
