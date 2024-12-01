package handlers

import (
	"html/template"

	"net/http"
)

// CreateGalleryHandler handles gallery creation requests
func (h *Handlers) CreateGalleryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	coverImage := r.FormValue("cover_image")

	galleryID, err := h.Gallery.CreateGallery(title, description, coverImage)
	if err != nil {
		http.Error(w, "Failed to create gallery", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/gallery_success.html")
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, galleryID)
}
