package auth

import (
	"context"
	"net/http"
	"strings"

	apperr "github.com/codersirojiddin/reddit-leads/internal/http"
	"github.com/google/uuid"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

// Middleware returns an http middleware that requires a valid "Authorization:
// Bearer <token>" header and injects the authenticated user ID into the
// request context.
func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				apperr.WriteError(w, apperr.Unauthorized("missing or malformed authorization header"))
				return
			}

			claims, err := svc.ParseToken(parts[1])
			if err != nil {
				apperr.WriteError(w, apperr.Unauthorized("invalid or expired token"))
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the authenticated user's ID set by Middleware.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}
