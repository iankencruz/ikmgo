package handlers

import (
	"html/template"
	"net/http"
	"strconv"
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
