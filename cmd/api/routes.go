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
		r.Get("/dashboard", app.AdminDashboard)
		// Galleries
		r.Get("/galleries", app.AdminGalleries)
		r.Get("/gallery/create", app.CreateGalleryForm)
		r.Post("/gallery/create", app.CreateGallery)
		//Users
		r.Get("/users", app.AdminUsers)
		r.Get("/users/edit/{id}", app.EditUserForm)
		r.Post("/users/edit/{id}", app.UpdateUser)
		r.Post("/users/delete/{id}", app.DeleteUser)
		// Media management
		r.Get("/media", app.AdminMedia)
		r.Get("/media/upload", app.UploadMediaForm)
		r.Post("/media/upload", app.UploadMedia)
		r.Post("/media/delete/{id}", app.DeleteMedia)
	})

	return r
}
