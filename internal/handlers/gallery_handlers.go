package handlers

import (
	"html/template"
	"ikm/internal/models"
	"ikm/internal/viewdata"

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

	galleryID, err := h.Gallery.CreateGallery(title, title, description, coverImage)
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

func (h *Handlers) ListGalleriesHandler(w http.ResponseWriter, r *http.Request) {
	galleries, err := h.Gallery.GetGalleries()
	if err != nil {
		http.Error(w, "Failed to retrieve galleries", http.StatusInternalServerError)
		return
	}

	data := viewdata.NewTemplateData()
	data.CurrentPath = r.URL.Path
	data.Data = struct {
		Galleries []*models.Gallery
	}{
		Galleries: galleries,
	}

	h.Render(w, r, "galleries", data)
}
