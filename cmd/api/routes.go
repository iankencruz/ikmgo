// cmd/api/routes.go
package main

import (
	"ikm/internal/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes(h *handlers.Handlers) http.Handler {
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	fileServer := http.FileServer(http.Dir("./ui/static"))
	router.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Routes
	router.Get("/", h.HomePageHandler)
	// router.Get("/healthcheck", h.HealthCheck)

	router.Group(func(r chi.Router) {
		r.Get("/galleries", h.ListGalleriesHandler)
		// Add more gallery-related routes here
	})

	// router.Post("/media/upload", h.UploadMediaHandler)

	router.Get("/about", h.AboutPageHandler)

	return router
}
