// cmd/api/routes.go
package main

import (
	"ikm/internal/handlers"
	appMiddleware "ikm/internal/middleware" // Alias custom middleware to appMiddleware
	"ikm/internal/session"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware" // Alias Chi middleware
)

func mainRouter(h *handlers.Handlers, sessionManager *session.Manager) http.Handler {
	router := chi.NewRouter()

	// General middleware for all routes
	router.Use(chiMiddleware.RequestID)
	router.Use(chiMiddleware.RealIP)
	router.Use(chiMiddleware.Logger)
	router.Use(chiMiddleware.Recoverer)

	// Add a file server to serve static assets
	workDir, _ := os.Getwd()
	staticDir := http.Dir(filepath.Join(workDir, "ui", "static"))
	fileServer := http.FileServer(staticDir)
	router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Public routes
	router.Group(func(r chi.Router) {
		r.Get("/", h.HomePageHandler)
		r.Get("/contact", h.ContactPageHandler)
		r.Get("/galleries", h.ListGalleriesHandler)

		r.Get("/register", h.RegisterUserHandler)
		r.Post("/register", h.RegisterUserHandler)
		r.Get("/login", h.LoginUserHandler)
		r.Post("/login", h.LoginUserHandler)
		r.Post("/logout", h.LogoutUserHandler)

	})

	// Mounting admin and authenticated routers
	router.Mount("/admin", adminRouter(h, sessionManager))
	router.Mount("/user", authenticatedRouter(h, sessionManager))

	return router
}

func authenticatedRouter(h *handlers.Handlers, sessionManager *session.Manager) http.Handler {
	r := chi.NewRouter()
	r.Use(appMiddleware.RequireAuthentication(sessionManager)) // Apply custom authentication middleware

	// Routes requiring user authentication
	// Add other authenticated routes here
	r.Get("/about", h.AboutPageHandler)

	return r
}

func adminRouter(h *handlers.Handlers, sessionManager *session.Manager) http.Handler {
	r := chi.NewRouter()
	r.Use(appMiddleware.RequireAuthentication(sessionManager)) // Apply custom authentication middleware
	r.Use(appMiddleware.RequireAdmin(sessionManager))          // Apply admin authorization middleware

	// Admin routes
	r.Get("/file-browser", h.AdminUploadFileHandler)
	// Add other admin-specific routes here

	return r
}
