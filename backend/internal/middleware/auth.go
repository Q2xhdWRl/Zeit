package middleware

import (
	"context"
	"net/http"

	"github.com/newa/zeiterfassung/internal/model"
	"github.com/newa/zeiterfassung/internal/service"
)

const userContextKey contextKey = "user"

// Auth returns middleware that validates session cookies and injects the user into context.
func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("zeit_session")
			if err != nil || cookie.Value == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Unauthorized","message":"not authenticated"}`)) //nolint:errcheck
				return
			}

			user, err := authService.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Internal Server Error","message":"session validation failed"}`)) //nolint:errcheck
				return
			}
			if user == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"Unauthorized","message":"invalid or expired session"}`)) //nolint:errcheck
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext extracts the authenticated user from the request context.
func UserFromContext(ctx context.Context) *model.User {
	user, _ := ctx.Value(userContextKey).(*model.User)
	return user
}

// ContextWithUser injects a user into a context. Intended for testing.
func ContextWithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
