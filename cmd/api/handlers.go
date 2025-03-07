package main

import (
	"context"
	"ikm/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

// Home Page Handler
func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	featuredGallery, images, err := app.GalleryModel.GetFeatured()
	if err != nil {
		log.Printf("❌ Error fetching featured gallery: %v", err)
	}
	app.render(w, r, "index.html", map[string]interface{}{
		"Title":           "Home",
		"FeaturedGallery": featuredGallery,
		"Images":          images,
	})
}

// Register Handler (GET + POST)
func (app *Application) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := map[string]interface{}{
			"Title":       "Register",
			"HideSidebar": true, // Prevents the sidebar from rendering
		}
		app.render(w, r, "register.html", data)
		return
	}

	// Process registration form
	fname := strings.TrimSpace(r.FormValue("fname"))
	lname := strings.TrimSpace(r.FormValue("lname"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	// Validate input
	if fname == "" || lname == "" || email == "" || password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Create user
	err := app.UserModel.Create(fname, lname, email, password)
	if err != nil {
		log.Printf("❌ Error creating user: %v", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Login Handler (GET + POST)
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data := map[string]interface{}{
			"Title":       "Login",
			"HideSidebar": true, // Prevents the sidebar from rendering
		}
		app.render(w, r, "login.html", data)
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
	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
}

// Logout Handler
func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// Admin Dashboard Handler
func (app *Application) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{} // ✅ Ensure a map is passed
	data["Title"] = "Admin Dashboard"
	data["ActiveLink"] = "dashboard"
	app.render(w, r, "admin/dashboard.html", data)
}

func (app *Application) AdminGalleries(w http.ResponseWriter, r *http.Request) {
	galleries, err := app.GalleryModel.GetAll()
	if err != nil {
		log.Printf("❌ Error fetching galleries: %v", err)
		http.Error(w, "Error fetching galleries", http.StatusInternalServerError)
		return
	}

	// ✅ Ensure GalleryMedia is always initialized
	galleryMedia := make(map[int][]*models.Media)

	for _, gallery := range galleries {
		media, err := app.MediaModel.GetByGalleryID(gallery["ID"].(int)) // ✅ Type assert gallery ID
		if err != nil {
			log.Printf("❌ Error fetching media for gallery %d: %v", gallery["ID"], err)
			continue
		}
		galleryMedia[gallery["ID"].(int)] = media
	}

	data := map[string]interface{}{
		"Title":        "Manage Galleries",
		"Galleries":    galleries,
		"GalleryMedia": galleryMedia, // ✅ Ensures it is always available
	}

	app.render(w, r, "admin/galleries.html", data)
}

// Form to Create a New Gallery
func (app *Application) CreateGalleryForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "admin/create_gallery.html", nil)
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

func (app *Application) DeleteGallery(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Delete from database
	err := app.GalleryModel.Delete(id)
	if err != nil {
		http.Error(w, "Error deleting gallery", http.StatusInternalServerError)
		return
	}

	// HTMX: Remove the row without reloading
	w.WriteHeader(http.StatusOK)
}

func (app *Application) AdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.UserModel.GetAll()
	if err != nil {
		log.Printf("❌ Error fetching users: %v", err)
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":      "Manage Users",
		"Users":      users,
		"ActiveLink": "users",
	}

	app.render(w, r, "admin/users.html", data)
}

func (app *Application) EditUserForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	user, err := app.UserModel.GetUserByID(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title": "Edit User",
		"User":  user,
	}

	app.render(w, r, "admin/edit_user.html", data)
}

func (app *Application) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	fname := r.FormValue("fname")
	lname := r.FormValue("lname")
	email := r.FormValue("email")

	err := app.UserModel.Update(id, fname, lname, email)
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (app *Application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	err := app.UserModel.Delete(id)
	if err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	// HTMX: Remove the row without reloading
	w.WriteHeader(http.StatusOK)
}

func (app *Application) AdminMedia(w http.ResponseWriter, r *http.Request) {
	media, err := app.MediaModel.GetAll()
	if err != nil {
		log.Printf("❌ Error fetching media: %v", err)
		http.Error(w, "Error fetching media", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":      "Manage Media",
		"Media":      media,
		"ActiveLink": "media",
	}

	app.render(w, r, "admin/media.html", data)
}

// Upload Media Form (GET)
func (app *Application) UploadMediaForm(w http.ResponseWriter, r *http.Request) {
	galleries, _ := app.GalleryModel.GetAll() // Fetch galleries for selection
	data := map[string]interface{}{
		"Title":     "Upload Media",
		"Galleries": galleries,
	}

	app.render(w, r, "admin/upload_media.html", data)
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

	galleryID, err := strconv.Atoi(r.FormValue("gallery_id"))
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	fileName := strconv.Itoa(galleryID) + "_" + handler.Filename

	// 4) Build the key that includes our "Uploads/" folder prefix
	fileKey := "Uploads/" + fileName

	// Upload to s3
	fileSize := handler.Size
	contentType := handler.Header.Get("Content-Type")

	ctx := context.Background()
	_, err = app.S3Client.PutObject(
		ctx,
		app.S3Bucket,
		fileKey,
		file,
		fileSize,
		minio.PutObjectOptions{
			ContentType: contentType,
			// ACL for S3-compatible services (like Vultr)
			// This sets the file to be publicly readable
			UserMetadata: map[string]string{
				"x-amz-acl": "public-read",
			},
		})

	if err != nil {
		log.Printf("❌ Error uploading to S3: %v", err)
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}

	// Construct public URL
	fileURL := "https://" + os.Getenv("VULTR_S3_ENDPOINT") + "/" + app.S3Bucket + "/" + fileKey

	err = app.MediaModel.Insert(fileName, fileURL, galleryID)
	if err != nil {
		http.Error(w, "Error saving media", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/media", http.StatusFound)
}

// Delete Media File (DELETE)
func (app *Application) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	err := app.MediaModel.Delete(id)
	if err != nil {
		http.Error(w, "Error deleting media", http.StatusInternalServerError)
		return
	}

	// ✅ Return 200 OK without content so HTMX removes the item
	w.WriteHeader(http.StatusOK)
}

// About Page Handler
func (app *Application) About(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "about.html", map[string]interface{}{
		"Title": "About Me",
	})
}

// Contact Page Handler
func (app *Application) Contact(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "contact.html", map[string]interface{}{
		"Title": "Contact Me",
	})
}

// Get All Galleries
func (app *Application) Galleries(w http.ResponseWriter, r *http.Request) {
	galleries, err := app.GalleryModel.GetAll()
	if err != nil {
		http.Error(w, "Error fetching galleries", http.StatusInternalServerError)
		return
	}
	app.render(w, r, "galleries.html", map[string]interface{}{
		"Title":     "Galleries",
		"Galleries": galleries,
	})
}

// View Gallery Page
func (app *Application) GalleryView(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	gallery, _ := app.GalleryModel.GetByID(id)
	media, _ := app.MediaModel.GetByGalleryID(id)

	app.render(w, r, "gallery.html", map[string]interface{}{
		"Title":   gallery.Title,
		"Gallery": gallery,
		"Media":   media,
	})
}

func (app *Application) SetFeaturedGallery(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	err = app.GalleryModel.SetFeatured(id)
	if err != nil {
		http.Error(w, "Error updating featured gallery", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/galleries", http.StatusSeeOther)
}

func (app *Application) SetCoverImage(w http.ResponseWriter, r *http.Request) {
	galleryID, err := strconv.Atoi(chi.URLParam(r, "galleryID"))
	if err != nil {
		log.Printf("❌ Invalid gallery ID: %v", err)
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.Atoi(r.FormValue("media_id"))
	if err != nil {
		log.Printf("❌ Invalid media ID: %v", err)
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	err = app.GalleryModel.SetCoverImage(galleryID, mediaID)
	if err != nil {
		log.Printf("❌ Error setting cover image: %v", err)
		http.Error(w, "Error setting cover image", http.StatusInternalServerError)
		return
	}

	// Fetch the updated cover image
	media, err := app.MediaModel.GetByID(mediaID)
	if err != nil {
		log.Printf("❌ Cover image not found: %v", err)
		http.Error(w, "Cover image not found", http.StatusInternalServerError)
		return
	}

	coverImageURL := "/uploads/" + media.FileName // Adjust based on your storage location

	// ✅ Return the updated cover image partial for HTMX swap
	app.render(w, r, "partials/cover_image.html", map[string]interface{}{
		"GalleryID":     galleryID,
		"CoverImageURL": coverImageURL,
	})
}
