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
				http.Error(w, `{"error":"Unauthorized","message":"not authenticated"}`, http.StatusUnauthorized)
				return
			}

			user, err := authService.ValidateSession(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, `{"error":"Internal Server Error","message":"session validation failed"}`, http.StatusInternalServerError)
				return
			}
			if user == nil {
				http.Error(w, `{"error":"Unauthorized","message":"invalid or expired session"}`, http.StatusUnauthorized)
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
