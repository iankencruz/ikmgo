package main

import (
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

// Secure cookie instance
var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

// SetSession stores the user ID in a secure, encrypted cookie
func SetSession(userID int, w http.ResponseWriter) {
	encoded, err := cookieHandler.Encode("session", userID)
	if err == nil {
		cookie := &http.Cookie{
			Name:     "session",
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(24 * time.Hour),
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

// ClearSession removes the session cookie
func ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(-1 * time.Hour),
	}
	http.SetCookie(w, cookie)
}
