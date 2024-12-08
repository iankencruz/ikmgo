package middleware

import (
	"context"
	"fmt"
	"ikm/internal/session" // Replace with your actual session package path
	"net/http"
)

func RequireAuthentication(sessionManager *session.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := sessionManager.Get(r, "user-session")
			userID, ok := session.Values["userID"].(int64)
			if !ok || userID == 0 {
				fmt.Print("Authentication Required")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
