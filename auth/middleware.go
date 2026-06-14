package auth

import (
	"context"
	"net/http"
	"strings"
)

type authorIDKey struct{}

// AuthorIDKey is the context key used to store the authenticated author's ID.
var AuthorIDKey = authorIDKey{}

func Required(s *Service) func(http.Handler) http.Handler {
	if s == nil {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeAuthError(w, "authentication service unavailable")
			})
		}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				writeAuthError(w, "missing or malformed Authorization header")
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := s.ValidateToken(token)
			if err != nil {
				writeAuthError(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), AuthorIDKey, claims.AuthorID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthorIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(AuthorIDKey).(string)
	return id, ok
}

func writeAuthError(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": message})
}
