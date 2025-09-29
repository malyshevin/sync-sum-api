package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
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

func CORS(environment string, uiDomains []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Allow all origins in local environment
			if environment == "local" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				// For stage/prod, check against allowed domains
				if origin != "" && isAllowedOrigin(origin, uiDomains) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin string, uiDomains []string) bool {
	if len(uiDomains) == 0 {
		return false
	}

	for _, domain := range uiDomains {
		if origin == domain || strings.HasSuffix(origin, "."+domain) {
			return true
		}
	}
	return false
}
