package handlers

import (
	"context"
	"html/template"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// UploadMediaHandler handles media upload requests
func (h *Handlers) UploadMediaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	galleryIDStr := r.FormValue("gallery_id")
	galleryID, err := strconv.ParseInt(galleryIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	imagePath := r.FormValue("image_path")
	s3Key := r.FormValue("s3_key")

	mediaID, err := h.Media.UploadMedia(galleryID, imagePath, s3Key)
	if err != nil {
		http.Error(w, "Failed to upload media", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/upload_success.html")
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, mediaID)
}

func (h *Handlers) LinkImageToGalleryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		h.serverError(w, err)
		return
	}

	galleryID := r.FormValue("gallery_id")
	fileID := r.FormValue("file_id")

	// Use pgx.NamedArgs to create named arguments for the query
	args := pgx.NamedArgs{
		"gallery_id": galleryID,
		"file_id":    fileID,
	}

	// Insert relationship into the database
	query := `INSERT INTO gallery_images (gallery_id, file_id) VALUES (:gallery_id, :file_id)`
	_, err = h.db.Exec(context.Background(), query, args)
	if err != nil {
		h.serverError(w, err)
		return
	}

	// Respond with a success message or redirect as needed
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Image successfully linked to gallery"))
}
