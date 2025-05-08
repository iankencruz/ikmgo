package main

import (
	"ikm/utils"
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

func (app *Application) SentryMiddleware(next http.Handler) http.Handler {
	return sentryHandler.Handle(next)
}

func (app *Application) SecureHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce := utils.GenerateNonce()
		csp := "default-src 'self'; script-src 'self' 'nonce-" + nonce + "'; style-src 'self'; img-src 'self' data: https://sgp1.vultrobjects.com/;"

		w.Header().Set("Content-Security-Policy", csp)
		ctx := utils.WithNonce(r.Context(), nonce)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
