package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Application) routes() http.Handler {
	r := chi.NewRouter()

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Public Routes
	r.Get("/", app.Home)
	r.Get("/about", app.About)
	r.Get("/contact", app.Contact)
	r.Get("/gallery/{id}", app.GalleryView)

	// Authentication Routes
	r.Get("/login", app.Login)
	r.Post("/login", app.Login)
	r.Post("/logout", app.Logout)
	r.Get("/register", app.Register)
	r.Post("/register", app.Register)

	// Admin Routes (Protected)
	r.Route("/admin", func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		r.Get("/", app.AdminDashboard)
		r.Get("/gallery/create", app.CreateGalleryForm)
		r.Post("/gallery/create", app.CreateGallery)
		r.Get("/upload", app.UploadMediaForm)
		r.Post("/upload", app.UploadMedia)
		r.Delete("/media/{id}", app.DeleteMedia)
	})

	return r
}
