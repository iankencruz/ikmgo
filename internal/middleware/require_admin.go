package middleware

import (
	"ikm/internal/session" // Replace with your actual session package path
	"net/http"
)

func RequireAdmin(sessionManager *session.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := sessionManager.Get(r, "user-session")
			userRole, ok := session.Values["userRole"].(string)
			if !ok || userRole != "admin" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
