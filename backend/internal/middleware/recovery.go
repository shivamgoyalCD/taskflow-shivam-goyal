package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"taskflow-shivam-goyal/backend/internal/response"
)

func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error(
						"http_request_panic_recovered",
						"request_id", chimiddleware.GetReqID(r.Context()),
						"method", r.Method,
						"path", r.URL.Path,
						"panic", rec,
						"stack_trace", string(debug.Stack()),
					)

					if err := response.InternalServerError(w); err != nil {
						logger.Error("http_recovery_response_failed", "error", err)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
