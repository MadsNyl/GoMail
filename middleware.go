package main

import (
	"GoMail/logger"
	"net/http"
	"os"
	"strings"
)

func APIKeyAuthMiddleware(next http.Handler) http.Handler {
	APIKey := os.Getenv("API_KEY")
	if APIKey == "" {
		logger.Error("API_KEY environment variable is not set", nil)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Error("Missing or invalid Authorization header", nil)
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		providedKey := strings.TrimPrefix(authHeader, "Bearer ")
		if providedKey != APIKey {
			logger.Error("Unauthorized access attempt", nil)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logger.Info("API key authentication successful", map[string]any{
			"ip": r.RemoteAddr,
			"endpoint":  r.URL.Path,
		})

		next.ServeHTTP(w, r)
	})
}