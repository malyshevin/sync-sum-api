package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

func RequestLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			if logger != nil {
				logger.Info("api request",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Duration("duration", time.Since(start)),
				)
			}
		})
	}
}
