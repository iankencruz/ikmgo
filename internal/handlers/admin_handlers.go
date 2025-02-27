package handlers

import (
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func (h *Handlers) AdminUploadFileHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is an admin (this is just an example, your actual check may vary)

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the multipart form to get the uploaded file
	err := r.ParseMultipartForm(10 << 20) // Limit upload size to 10 MB
	if err != nil {
		h.logger.Printf("Unable to parse form: %v", err)
		http.Error(w, "Unable to parse form", http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		h.logger.Printf("Unable to retrieve the file from form: %v", err)
		http.Error(w, "Unable to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Determine the content type of the file
	fileExtension := filepath.Ext(handler.Filename)
	contentType := mime.TypeByExtension(fileExtension)
	if contentType == "" {
		contentType = "application/octet-stream" // Default content type for unknown file types
	}

	s3bucket := os.Getenv("S3_BUCKET")

	// Upload file directly to S3 using the injected S3 Uploader
	_, err = h.S3Uploader.Upload(&s3manager.UploadInput{
		Bucket:      &s3bucket,
		Key:         &handler.Filename,
		Body:        file,
		ContentType: &contentType,
	})
	if err != nil {
		h.logger.Printf("Unable to upload file to S3: %v", err)
		http.Error(w, "Unable to upload file", http.StatusInternalServerError)
		return
	}

	h.logger.Printf("File uploaded successfully to S3: %s", handler.Filename)
	http.Redirect(w, r, "/admin/file-browser", http.StatusSeeOther)
}

func (h *Handlers) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Check if the user is an admin (this is just an example, your actual check may vary)
	w.Write([]byte("Admin Dashboard"))
	// h.Render(w, r, "admin-dashboard", nil)
}
