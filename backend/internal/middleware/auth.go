package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"taskflow-shivam-goyal/backend/internal/auth"
	"taskflow-shivam-goyal/backend/internal/response"
)

type currentUserContextKey struct{}

type CurrentUser struct {
	ID    string
	Email string
}

func Authenticate(logger *slog.Logger, jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := bearerToken(r.Header.Get("Authorization"))
			if err != nil {
				writeUnauthorized(logger, w, "missing_or_malformed_authorization_header", err)
				return
			}

			claims, err := jwtManager.ParseToken(token)
			if err != nil {
				writeUnauthorized(logger, w, "invalid_jwt_token", err)
				return
			}

			currentUser := CurrentUser{
				ID:    claims.UserID,
				Email: claims.Email,
			}

			ctx := context.WithValue(r.Context(), currentUserContextKey{}, currentUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool) {
	currentUser, ok := ctx.Value(currentUserContextKey{}).(CurrentUser)
	return currentUser, ok
}

func CurrentUserIDFromContext(ctx context.Context) (string, bool) {
	currentUser, ok := CurrentUserFromContext(ctx)
	if !ok {
		return "", false
	}

	return currentUser.ID, true
}

func CurrentUserEmailFromContext(ctx context.Context) (string, bool) {
	currentUser, ok := CurrentUserFromContext(ctx)
	if !ok {
		return "", false
	}

	return currentUser.Email, true
}

func bearerToken(headerValue string) (string, error) {
	headerValue = strings.TrimSpace(headerValue)
	if headerValue == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(headerValue, " ")
	if len(parts) != 2 {
		return "", errors.New("authorization header must be in Bearer format")
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("authorization header must use Bearer scheme")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("bearer token is required")
	}

	return token, nil
}

func writeUnauthorized(logger *slog.Logger, w http.ResponseWriter, reason string, err error) {
	logger.Warn("http_auth_unauthorized", "reason", reason, "error", err)

	if writeErr := response.Error(w, http.StatusUnauthorized, "unauthorized"); writeErr != nil {
		logger.Error("http_auth_unauthorized_response_failed", "error", writeErr)
	}
}
