package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *Application) routes() http.Handler {
	r := chi.NewRouter()

	//Logger Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)

	// Serve static files
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Public Routes
	r.Get("/", app.Home)
	r.Get("/about", app.About)
	r.Get("/contact", app.Contact)
	r.Get("/galleries", app.Galleries)
	r.Get("/gallery/{id}", app.GalleryView)
	r.Get("/projects", app.PublicProjectsList)
	r.Get("/project/{id}", app.PublicProjectView)

	// Authentication Routes
	r.Get("/login", app.Login)
	r.Post("/login", app.Login)
	r.Post("/logout", app.Logout)
	r.Get("/register", app.Register)
	r.Post("/register", app.Register)
	// Contacts
	r.Get("/contact", app.Contact)
	r.Post("/contact", app.Contact)

	// Admin Routes (Protected)
	r.Route("/admin", func(r chi.Router) {
		// r.Use(app.AuthMiddleware)

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/dashboard", 301)
		})

		r.Get("/dashboard", app.AdminDashboard)

		// Galleries
		r.Get("/galleries", app.AdminGalleries)
		r.Get("/gallery/create", app.CreateGalleryForm)
		r.Post("/gallery/create", app.CreateGallery)
		r.Delete("/gallery/{id}", app.DeleteGallery)
		r.Post("/gallery/feature/{id}", app.SetFeaturedGallery) // Set Featured Galleryrou
		r.Get("/gallery/edit/{id}", app.EditGalleryForm)
		r.Post("/gallery/update/{id}", app.UpdateGallery)

		r.Post("/gallery/{galleryID}/cover", app.SetCoverImage)
		r.Post("/gallery/{id}/publish", app.SetGalleryVisibility)

		// Projects
		r.Get("/projects", app.AdminProjects)                   // list view
		r.Get("/project/create", app.CreateProjectForm)         // show form
		r.Post("/project/create", app.CreateProject)            // handle form submit
		r.Get("/project/edit/{id}", app.EditProjectForm)        // show edit form
		r.Post("/project/edit/{id}", app.UpdateProject)         // handle update
		r.Post("/project/{id}/cover", app.SetProjectCoverImage) // HTMX: update cover
		r.Post("/project/update-order", app.UpdateProjectMediaOrder)
		r.Post("/project/{id}/publish", app.SetProjectVisibility)
		r.Get("/project/{id}/info", app.ProjectInfoView)
		r.Get("/project/{id}/info/edit", app.ProjectInfoEdit)
		r.Post("/project/{id}/info", app.ProjectInfoUpdate)

		//Users
		r.Get("/users", app.AdminUsers)
		r.Get("/users/edit/{id}", app.EditUserForm)
		r.Post("/users/edit/{id}", app.UpdateUser)
		r.Delete("/users/{id}", app.DeleteUser)

		// Media management
		r.Get("/media", app.AdminMedia)
		r.Get("/media/upload-modal", app.UploadMediaModal)
		r.Get("/media/upload", app.UploadMediaForm)
		r.Post("/media/upload", app.UploadMedia)
		r.Delete("/media/{id}", app.DeleteMedia)
		r.Post("/media/attach", app.AttachMediaToItem)
		r.Post("/media/update-order-bulk", app.UpdateMediaOrderBulk)
		r.Put("/media/unlink", app.UnlinkMediaFromItem)

		// Contacts
		r.Get("/contacts", app.AdminContacts)

		// Settings
		r.Get("/settings", app.AdminSettings)
		r.Post("/settings", app.UpdateSettings)
		r.Get("/settings/select-about-image", app.GetAboutMeImageModal)
		r.Post("/settings/set-about-image", app.SetAboutMeImage)

		// Toast
		r.Get("/toast", app.Toast)

	})

	return r
}

func DebugRoutes(handler http.Handler) {
	r, ok := handler.(*chi.Mux) // ✅ Convert http.Handler to chi.Router
	if !ok {
		log.Println("❌ DebugRoutes: Handler is not a chi.Router")
		return
	}

	log.Println("📌 Registered Routes:")
	_ = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s", method, route)
		return nil
	})
}
