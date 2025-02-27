package middleware

import (
	"context"
	"ikm/internal/session" // Replace with your actual session package path
	"net/http"
	"strings"
)

func RequireAdmin(sessionManager *session.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := sessionManager.Get(r, "user-session")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			role, ok := session.Values["userRole"].(string)
			if !ok || strings.ToLower(role) != "admin" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Add userRole to the request context
			ctx := context.WithValue(r.Context(), "userRole", role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
