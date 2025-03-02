package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Home Page Handler
func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	galleries, _ := app.GalleryModel.GetAll()
	render(w, r, "index.html", map[string]interface{}{
		"Title":     "Home",
		"Galleries": galleries,
	})
}

// Register Handler (GET + POST)

func (app *Application) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		render(w, r, "register.html", nil)
		return
	}

	// Process registration form
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	// Validate input
	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Create user
	err := app.UserModel.Create(email, password)
	if err != nil {
		log.Printf("❌ Error creating user: %v", err) // ✅ Log the exact error
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Login Handler (GET + POST)
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		render(w, r, "login.html", nil)
		return
	}

	// Process login form submission
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := app.UserModel.Authenticate(email, password)
	if err != nil {
		http.Error(w, "Invalid login", http.StatusUnauthorized)
		return
	}

	// Store session using secure cookies
	SetSession(user.ID, w)

	// Redirect to admin dashboard
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// Logout Handler
func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// Admin Dashboard Handler
func (app *Application) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin/dashboard.html", nil)
}

// Form to Create a New Gallery
func (app *Application) CreateGalleryForm(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin/create_gallery.html", nil)
}

// Create a New Gallery (POST)
func (app *Application) CreateGallery(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	err := app.GalleryModel.Create(title)
	if err != nil {
		http.Error(w, "Error creating gallery", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusFound)
}

// Upload Media Form (GET)
func (app *Application) UploadMediaForm(w http.ResponseWriter, r *http.Request) {
	render(w, r, "admin/upload_media.html", nil)
}

// Upload Media File (POST)
func (app *Application) UploadMedia(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // Max 10MB file upload

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get gallery ID
	galleryID, err := strconv.Atoi(r.FormValue("gallery_id"))
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	// Save file name
	fileName := strconv.Itoa(galleryID) + "_" + handler.Filename
	err = app.MediaModel.Insert(fileName, galleryID)
	if err != nil {
		http.Error(w, "Error saving media", http.StatusInternalServerError)
		return
	}

	// HTMX response: Update media list dynamically
	w.Header().Set("HX-Trigger", "mediaAdded")
	render(w, r, "partials/media_item.html", map[string]interface{}{
		"FileName": fileName,
	})
}

// Delete Media File (DELETE)
func (app *Application) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	mediaID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	// Retrieve media file info but don't use the variable
	_, err = app.MediaModel.GetByID(mediaID)
	if err != nil {
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	// Delete media from database
	err = app.MediaModel.Delete(mediaID)
	if err != nil {
		http.Error(w, "Failed to delete media", http.StatusInternalServerError)
		return
	}

	// HTMX removes the media item dynamically
	w.WriteHeader(http.StatusOK)
}

// About Page Handler
func (app *Application) About(w http.ResponseWriter, r *http.Request) {
	render(w, r, "about.html", map[string]interface{}{
		"Title": "About Me",
	})
}

// Contact Page Handler
func (app *Application) Contact(w http.ResponseWriter, r *http.Request) {
	render(w, r, "contact.html", map[string]interface{}{
		"Title": "Contact Me",
	})
}

// View Gallery Page
func (app *Application) GalleryView(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	gallery, _ := app.GalleryModel.GetByID(id)
	media, _ := app.MediaModel.GetByGalleryID(id)

	render(w, r, "gallery.html", map[string]interface{}{
		"Title":   gallery.Title,
		"Gallery": gallery,
		"Media":   media,
	})
}
