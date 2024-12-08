package middleware

import (
	"context"
	"ikm/internal/models"
	"ikm/internal/session"
	"net/http"
)

type contextKey string

const UserKey contextKey = "user"

// WithUser populates the request context with the user's data.
func WithUser(sessionManager *session.Manager, userModel *models.UserModel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := sessionManager.Get(r, "user-session")
			userID, ok := session.Values["userID"].(int64)
			if !ok || userID == 0 {
				// User is not logged in, continue without setting user in context
				next.ServeHTTP(w, r)
				return
			}

			// Retrieve user from the database
			user, err := userModel.GetByID(userID)
			if err != nil {
				http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
				return
			}

			// Add user to the context
			ctx := context.WithValue(r.Context(), UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// internal/middleware/user_context.go
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserKey).(*models.User)
	return user, ok
}
