package handlers

import (
	"context"
	"html/template"
	"ikm/internal/models"
	"ikm/internal/viewdata"

	"net/http"

	"github.com/jackc/pgx/v5"
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

	data := viewdata.NewTemplateData(r, h.session, h.User)
	data.CurrentPath = r.URL.Path
	data.Data = struct {
		Galleries []*models.Gallery
	}{
		Galleries: galleries,
	}

	h.Render(w, r, "galleriesPage", data)
}

func (h *Handlers) UpdateGalleryDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data for gallery update
	err := r.ParseForm()
	if err != nil {
		h.serverError(w, err)
		return
	}

	galleryID := r.FormValue("id")
	title := r.FormValue("title")
	description := r.FormValue("description")

	// Use pgx.NamedArgs to create a named argument map
	args := pgx.NamedArgs{
		"id":          galleryID,
		"title":       title,
		"description": description,
	}

	// Update the gallery details in the database
	query := `UPDATE galleries SET title=:title, description=:description WHERE id=:id`
	_, err = h.db.Exec(context.Background(), query, args)
	if err != nil {
		h.serverError(w, err)
		return
	}

	// Respond with the updated HTML for this part of the UI
	data := viewdata.TemplateData{
		Title:       title,
		CurrentPath: r.URL.Path,
	}

	// Render updated gallery details only
	h.Render(w, r, "gallery_details_partial", data)
}
