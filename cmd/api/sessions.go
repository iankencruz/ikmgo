package main

import (
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

// Secure cookie instance for encoding & decoding session data
var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64), // Encryption key
	securecookie.GenerateRandomKey(32), // Signing key
)

// SetSession stores the user ID in a secure, encrypted cookie
func SetSession(userID int, w http.ResponseWriter) {
	encoded, err := cookieHandler.Encode("session", userID)
	if err == nil {
		cookie := &http.Cookie{
			Name:     "session",
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,                           // Prevents JavaScript access (XSS protection)
			Secure:   true,                           // Set to false for local development
			SameSite: http.SameSiteStrictMode,        // Protects against CSRF
			Expires:  time.Now().Add(24 * time.Hour), // 24-hour expiration
		}
		http.SetCookie(w, cookie)
	}
}

// GetSession retrieves the user ID from the session cookie
func GetSession(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return 0, err
	}

	var userID int
	err = cookieHandler.Decode("session", cookie.Value, &userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// ClearSession removes the session cookie (logs the user out)
func ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(-1 * time.Hour), // Expired in the past
	}
	http.SetCookie(w, cookie)
}
