package main

import (
	"net/http"

	sentryhttp "github.com/getsentry/sentry-go/http"
)

var sentryHandler = sentryhttp.New(sentryhttp.Options{
	Repanic: true,
})

func (app *Application) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetSession(r)
		if err != nil || userID == 0 {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SentryMiddleware(next http.Handler) http.Handler {
	return sentryHandler.Handle(next)
}
