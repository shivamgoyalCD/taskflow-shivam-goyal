package middleware

import (
	"net/http"
	"strings"
)

var defaultAllowedOrigins = map[string]struct{}{
	"http://localhost:5173": {},
	"http://127.0.0.1:5173": {},
	"http://localhost:4173": {},
	"http://127.0.0.1:4173": {},
	"http://localhost:3000": {},
	"http://127.0.0.1:3000": {},
}

const (
	allowedMethods = "GET, POST, PATCH, DELETE, OPTIONS"
	allowedHeaders = "Accept, Authorization, Content-Type"
)

func CORS() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			headers := w.Header()
			headers.Add("Vary", "Origin")
			headers.Add("Vary", "Access-Control-Request-Method")
			headers.Add("Vary", "Access-Control-Request-Headers")

			if _, allowed := defaultAllowedOrigins[origin]; allowed {
				headers.Set("Access-Control-Allow-Origin", origin)
				headers.Set("Access-Control-Allow-Methods", allowedMethods)
				headers.Set("Access-Control-Allow-Headers", allowedHeaders)
				headers.Set("Access-Control-Max-Age", "600")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
