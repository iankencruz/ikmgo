package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"ikm/models"
	"ikm/utils"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-chi/chi/v5"
	"github.com/minio/minio-go/v7"
)

// Home Page Handler
func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	gallery, media, err := app.GalleryModel.GetByTitle("Japan")
	if err != nil {
		log.Printf("❌ Error fetching featured gallery: %v", err)
	}
	app.render(w, r, "index.html", map[string]interface{}{
		"Title":      "Home",
		"Gallery":    gallery,
		"Media":      media,
		"ActiveLink": "home",
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
	http.Redirect(w, r, "/", http.StatusFound)
}

// Admin Dashboard Handler
func (app *Application) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{} // ✅ Ensure a map is passed
	data["Title"] = "Admin Dashboard"
	data["ActiveLink"] = "dashboard"
	app.render(w, r, "admin/dashboard.html", data)
}

func (app *Application) AdminSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := app.SettingsModel.GetAll()
	if err != nil {
		http.Error(w, "Error fetching settings", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":    "Settings",
		"Settings": settings,
	}

	app.render(w, r, "admin/settings.html", data)
}

// Update settings (POST)

func (app *Application) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	for key, values := range r.PostForm {
		if len(values) == 0 {
			continue
		}
		value := values[0]

		if err := app.SettingsModel.Set(key, value); err != nil {
			http.Error(w, "Failed to update setting: "+key, http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/admin/settings", http.StatusSeeOther)
}

func (app *Application) PublicProjectsList(w http.ResponseWriter, r *http.Request) {
	projects, err := app.ProjectModel.GetAllPublic()
	if err != nil {
		log.Printf("❌ Error fetching public projects: %v", err)
		http.Error(w, "Unable to load projects", http.StatusInternalServerError)
		return
	}

	log.Printf("Project Publics: %v", projects)

	data := map[string]interface{}{
		"Title":      "Projects",
		"Projects":   projects,
		"ActiveLink": "projects",
	}

	app.render(w, r, "projects.html", data)
}

func (app *Application) SetProjectVisibility(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	published := r.FormValue("published") == "on"

	err = app.ProjectModel.SetPublished(id, published)
	if err != nil {
		http.Error(w, "Error updating project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (app *Application) PublicProjectView(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil || projectID <= 0 {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	project, err := app.ProjectModel.GetByID(projectID)
	if err != nil {
		log.Printf("❌ Project not found: %v", err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	media, err := app.ProjectModel.GetMedia(projectID)
	if err != nil {
		log.Printf("❌ Failed to get media for project %d: %v", projectID, err)
		http.Error(w, "Unable to load media", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Project %d media count: %d", projectID, len(media))

	var heroMedia []*models.Media
	var restMedia []*models.Media

	if len(media) > 4 {
		heroMedia = media[:4]
		restMedia = media[4:]
	} else {
		heroMedia = media
	}

	data := map[string]interface{}{
		"Title":     project.Title,
		"Project":   project,
		"HeroMedia": heroMedia,
		"Media":     restMedia, // remaining media
	}

	log.Printf("Project Media Count: %d", len(media))

	app.render(w, r, "project_view.html", data)
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
		// ✅ Use map indexing instead of dot notation
		galleryID, ok := gallery["ID"].(int)
		if !ok {
			log.Printf("❌ Invalid gallery ID format: %v", gallery["ID"])
			continue
		}

		media, err := app.GalleryModel.GetMedia(galleryID)
		if err != nil {
			log.Printf("❌ Error fetching media for gallery %d: %v", galleryID, err)
			continue
		}
		galleryMedia[galleryID] = media
	}

	data := map[string]interface{}{
		"Title":        "Manage Galleries",
		"Galleries":    galleries,    // ✅ Passes proper `[]map[string]interface{}` slice
		"GalleryMedia": galleryMedia, // ✅ Ensures media is available per gallery
		"ActiveLink":   "galleries",
	}

	app.render(w, r, "admin/galleries.html", data)
}

// Form to Create a New Gallery
func (app *Application) CreateGalleryForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "admin/create_gallery.html", nil)
}

// Edit Gallery Page
func (app *Application) EditGalleryForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	gallery, err := app.GalleryModel.GetByID(id)
	if err != nil {
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	media, err := app.GalleryModel.GetMedia(id)
	if err != nil {
		log.Printf("❌ Error fetching media for gallery %d: %v", id, err)
	}

	data := map[string]interface{}{
		"Title":   "Edit Gallery",
		"Gallery": gallery,
		"Media":   media,
	}

	app.render(w, r, "admin/edit_gallery.html", data)
}

// Update Gallery Title
func (app *Application) UpdateGallery(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	title := r.FormValue("title")

	err := app.GalleryModel.Update(id, title)
	if err != nil {
		http.Error(w, "Error updating gallery", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/galleries", http.StatusSeeOther)
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

func (app *Application) AdminContacts(w http.ResponseWriter, r *http.Request) {
	contacts, err := app.ContactModel.GetAll()
	if err != nil {
		log.Printf("❌ Error fetching contacts: %v", err)
		http.Error(w, "Error fetching contacts", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":      "Manage Contacts",
		"Contacts":   contacts,
		"ActiveLink": "contacts",
	}

	app.render(w, r, "admin/contacts.html", data)
}

func (app *Application) AdminProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := app.ProjectModel.GetAll()
	if err != nil {
		http.Error(w, "Error fetching projects", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":      "Manage Projects",
		"Projects":   projects,
		"ActiveLink": "projects",
	}

	app.render(w, r, "admin/projects.html", data)
}

func (app *Application) CreateProjectForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, "admin/create_project.html", map[string]interface{}{
		"Title": "Create Project",
	})
}

func (app *Application) CreateProject(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	description := r.FormValue("description")

	err := app.ProjectModel.Create(title, description)
	if err != nil {
		http.Error(w, "Error creating project", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

func (app *Application) EditProjectForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	project, err := app.ProjectModel.GetByID(id)
	if err != nil {
		log.Printf("❌ GetByID: Failed to find project %d: %v", id, err)
	}

	media, err := app.ProjectModel.GetMedia(id)
	if err != nil {
		log.Printf("❌ Error fetching media for project %d: %v", id, err)
		http.Error(w, "Error loading media", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":   "Edit Project",
		"Project": project,
		"Media":   media,
	}

	app.render(w, r, "admin/edit_project.html", data)
}

func (app *Application) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	title := r.FormValue("title")
	description := r.FormValue("description")

	_, err := app.DB.Exec(context.Background(),
		`UPDATE projects SET title = $1, description = $2 WHERE id = $3`,
		title, description, id)

	if err != nil {
		http.Error(w, "Error updating project", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

func (app *Application) SetProjectCoverImage(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.Atoi(r.FormValue("media_id"))
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	err = app.ProjectModel.SetCoverImage(projectID, mediaID)
	if err != nil {
		log.Printf("❌ Error setting project cover image: %v", err)
		http.Error(w, "Error updating project", http.StatusInternalServerError)
		return
	}

	media, err := app.MediaModel.GetByID(mediaID)
	if err != nil {
		log.Printf("❌ Cover image not found: %v", err)
		http.Error(w, "Cover image not found", http.StatusInternalServerError)
		return
	}

	app.render(w, r, "partials/cover_preview.html", map[string]interface{}{
		"ProjectID":     projectID,
		"CoverImageURL": media.ThumbnailURL,
	})
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
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	galleryID, err := strconv.Atoi(r.FormValue("gallery_id"))
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	displayMode := r.FormValue("display_mode")
	if displayMode == "" {
		displayMode = "grid"
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	var outputBuffer bytes.Buffer
	ctx := context.Background()

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("❌ Error opening file: %v", err)
			continue
		}
		defer file.Close()

		base := fmt.Sprintf("%d_%d_%s", time.Now().UnixNano(), galleryID, fileHeader.Filename)
		fileName := base
		thumbName := "thumb_" + base
		mediumName := "medium_" + base

		img, err := imaging.Decode(file, imaging.AutoOrientation(true))
		if err != nil {
			log.Printf("❌ Error decoding image: %v", err)
			continue
		}

		thumbnailImg := imaging.Resize(img, 500, 0, imaging.Lanczos)
		mediumImg := imaging.Resize(img, 1024, 0, imaging.Lanczos)
		file.Seek(0, 0)

		fileKey := "Uploads/" + fileName
		thumbKey := "Uploads/" + thumbName
		mediumKey := "Uploads/" + mediumName

		fileSize := fileHeader.Size
		contentType := fileHeader.Header.Get("Content-Type")

		_, err = app.S3Client.PutObject(ctx, app.S3Bucket, fileKey, io.LimitReader(file, fileSize), fileSize, minio.PutObjectOptions{
			ContentType:  contentType,
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
		})
		if err != nil {
			log.Printf("❌ Error uploading original: %v", err)
			continue
		}

		var mediumBuf, thumbBuf bytes.Buffer

		if err := imaging.Encode(&mediumBuf, mediumImg, imaging.JPEG); err == nil {
			app.S3Client.PutObject(ctx, app.S3Bucket, mediumKey, bytes.NewReader(mediumBuf.Bytes()), int64(mediumBuf.Len()), minio.PutObjectOptions{
				ContentType:  "image/jpeg",
				UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			})
		}

		if err := imaging.Encode(&thumbBuf, thumbnailImg, imaging.JPEG); err == nil {
			app.S3Client.PutObject(ctx, app.S3Bucket, thumbKey, bytes.NewReader(thumbBuf.Bytes()), int64(thumbBuf.Len()), minio.PutObjectOptions{
				ContentType:  "image/jpeg",
				UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			})
		}

		fullURL := "https://" + os.Getenv("VULTR_S3_ENDPOINT") + "/" + app.S3Bucket + "/" + fileKey
		thumbURL := "https://" + os.Getenv("VULTR_S3_ENDPOINT") + "/" + app.S3Bucket + "/" + thumbKey

		// Insert media (no galleryID)
		mediaID, err := app.MediaModel.InsertAndReturnID(fileName, fullURL, thumbURL)
		if err != nil {
			log.Printf("❌ DB insert failed: %v", err)
			continue
		}

		// Get next position from join table
		position, err := app.GalleryModel.GetNextPosition(galleryID)
		if err != nil {
			log.Printf("❌ Failed to get next position: %v", err)
			continue
		}

		// Attach media to gallery
		err = app.GalleryModel.AttachMedia(galleryID, mediaID, position)
		if err != nil {
			log.Printf("❌ Failed to attach media: %v", err)
			continue
		}

		media := &models.Media{
			ID:           mediaID,
			FileName:     fileName,
			ThumbnailURL: thumbURL,
			FullURL:      fullURL,
		}

		app.renderPartialHTMX(&outputBuffer, "partials/media_item.html", media)

	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(outputBuffer.Bytes())
}

// Upload Project Media (POST)

func (app *Application) UploadProjectMedia(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	projectID, err := strconv.Atoi(r.FormValue("project_id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	var outputBuffer bytes.Buffer
	ctx := context.Background()

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("❌ Error opening file: %v", err)
			continue
		}
		defer file.Close()

		base := fmt.Sprintf("%d_%d_%s", time.Now().UnixNano(), projectID, fileHeader.Filename)
		fileName := base
		thumbName := "thumb_" + base
		mediumName := "medium_" + base

		img, err := imaging.Decode(file, imaging.AutoOrientation(true))
		if err != nil {
			log.Printf("❌ Error decoding image: %v", err)
			continue
		}

		thumbnailImg := imaging.Resize(img, 500, 0, imaging.Lanczos)
		mediumImg := imaging.Resize(img, 1024, 0, imaging.Lanczos)
		file.Seek(0, 0)

		fileKey := "Uploads/" + fileName
		thumbKey := "Uploads/" + thumbName
		mediumKey := "Uploads/" + mediumName

		fileSize := fileHeader.Size
		contentType := fileHeader.Header.Get("Content-Type")

		_, err = app.S3Client.PutObject(ctx, app.S3Bucket, fileKey, io.LimitReader(file, fileSize), fileSize, minio.PutObjectOptions{
			ContentType:  contentType,
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
		})
		if err != nil {
			log.Printf("❌ Error uploading original: %v", err)
			continue
		}

		var mediumBuf, thumbBuf bytes.Buffer

		if err := imaging.Encode(&mediumBuf, mediumImg, imaging.JPEG); err == nil {
			app.S3Client.PutObject(ctx, app.S3Bucket, mediumKey, bytes.NewReader(mediumBuf.Bytes()), int64(mediumBuf.Len()), minio.PutObjectOptions{
				ContentType:  "image/jpeg",
				UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			})
		}

		if err := imaging.Encode(&thumbBuf, thumbnailImg, imaging.JPEG); err == nil {
			app.S3Client.PutObject(ctx, app.S3Bucket, thumbKey, bytes.NewReader(thumbBuf.Bytes()), int64(thumbBuf.Len()), minio.PutObjectOptions{
				ContentType:  "image/jpeg",
				UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			})
		}

		fullURL := "https://" + os.Getenv("VULTR_S3_ENDPOINT") + "/" + app.S3Bucket + "/" + fileKey
		thumbURL := "https://" + os.Getenv("VULTR_S3_ENDPOINT") + "/" + app.S3Bucket + "/" + thumbKey

		position, err := app.MediaModel.GetNextProjectPosition(projectID)
		if err != nil {
			log.Printf("❌ Failed to get project position: %v", err)
			continue
		}

		mediaID, err := app.MediaModel.InsertProjectMedia(fileName, fullURL, thumbURL, projectID, position)
		if err != nil {
			log.Printf("❌ DB Project insert failed: %v", err)
			continue
		}

		media := &models.Media{
			ID:           mediaID,
			FileName:     fileName,
			FullURL:      fullURL,
			ThumbnailURL: thumbURL,
		}

		app.renderPartialHTMX(&outputBuffer, "partials/media_item.html", media)
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(outputBuffer.Bytes())
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
		"Title":      "About Me",
		"ActiveLink": "about",
	})
}

// Contact Handler (GET + POST)
func (app *Application) Contact(w http.ResponseWriter, r *http.Request) {
	// 1. Handle GET request - show contact form
	if r.Method == http.MethodGet {
		app.render(w, r, "contact.html", map[string]interface{}{
			"Title":      "Contact",
			"ActiveLink": "contact",
		})
		return
	}

	// 2. Parse POST form
	if err := r.ParseForm(); err != nil {
		log.Printf("❌ Failed to parse contact form: %v", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// 3. Populate ContactForm struct
	form := models.ContactForm{
		FirstName:      r.FormValue("first_name"),
		LastName:       r.FormValue("last_name"),
		Email:          r.FormValue("email"),
		Subject:        r.FormValue("subject"),
		Message:        r.FormValue("message"),
		RecaptchaToken: r.FormValue("g-recaptcha-response"),
	}
	log.Printf("🔍 Form submission: %+v", form)

	// 4. Validate form fields
	if err := utils.ValidateStruct(&form); err != nil {
		log.Printf("❌ Contact form validation failed: %v", err)
		app.render(w, r, "contact.html", map[string]interface{}{
			"Title":  "Contact",
			"Errors": err,
			"Form":   form,
		})
		return
	}

	// 5. Verify reCAPTCHA
	if !verifyRecaptcha(form.RecaptchaToken, os.Getenv("RECAPTCHA_SECRET")) {
		log.Printf("❌ reCAPTCHA verification failed for email: %s", form.Email)
		http.Error(w, "reCAPTCHA verification failed", http.StatusForbidden)
		return
	}

	// 6. Save contact info to DB
	if err := app.ContactModel.Insert(form.FirstName, form.LastName, form.Email, form.Subject, form.Message); err != nil {
		log.Printf("❌ Error saving contact form submission: %v", err)
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		return
	}
	fmt.Println("Insert success. starting SendEmail method")

	// 7. Send email with Gomail via utils.SendEmail
	err := utils.SendEmail(
		os.Getenv("CONTACT_EMAIL"),              // From
		os.Getenv("CONTACT_EMAIL"),              // To
		form.Subject,                            // Subject
		"./templates/emails/contact_email.html", // Template path
		form,                                    // Data for template
	)
	if err != nil {
		log.Printf("❌ Email sending failed: %v", err)
		http.Error(w, "Failed to send email notification", http.StatusInternalServerError)
		return
	}

	// 8. Everything succeeded
	log.Printf("✅ Contact form submitted successfully by %s %s (%s)", form.FirstName, form.LastName, form.Email)

	// app.render(w, r, "partials/contact_success_modal.html", nil) // ✅ Correct

	app.renderPartialHTMX(w, "partials/contact_success_modal", nil)

}

// Get All Galleries
func (app *Application) Galleries(w http.ResponseWriter, r *http.Request) {
	galleries, err := app.GalleryModel.GetAllPublic()
	if err != nil {
		http.Error(w, "Error fetching galleries", http.StatusInternalServerError)
		return
	}
	app.render(w, r, "galleries.html", map[string]interface{}{
		"Title":      "Galleries",
		"Galleries":  galleries,
		"ActiveLink": "galleries",
	})
}

// View Gallery Page

func (app *Application) GalleryView(w http.ResponseWriter, r *http.Request) {
	galleryIDStr := chi.URLParam(r, "id")

	// ✅ Check if galleryID is null or invalid
	if galleryIDStr == "" || galleryIDStr == "null" {
		log.Println("❌ Invalid gallery ID: received 'null'")
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	galleryID, err := strconv.Atoi(galleryIDStr)
	if err != nil {
		log.Printf("❌ Invalid gallery ID: %v", err)
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	log.Printf("✅ GalleryView requested for ID: %d", galleryID)

	// Fetch gallery
	gallery, err := app.GalleryModel.GetByID(galleryID)
	if err != nil {
		log.Printf("❌ Error fetching gallery: %v", err)
		http.Error(w, "Gallery not found", http.StatusNotFound)
		return
	}

	// Fetch media
	media, err := app.GalleryModel.GetMedia(galleryID)
	if err != nil {
		log.Printf("❌ Error fetching media: %v", err)
		http.Error(w, "Error retrieving media", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Fetched %d media items for Gallery ID: %d", len(media), galleryID)

	if r.Header.Get("HX-Request") != "" {
		log.Println("🔄 HTMX request detected, rendering media partial")
		app.render(w, r, "partials/gallery_component.html", map[string]interface{}{
			"Gallery": gallery,
			"Media":   media,
		})
		return
	}

	app.render(w, r, "gallery.html", map[string]interface{}{
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

// Set Cover Image for Gallery

func (app *Application) SetCoverImage(w http.ResponseWriter, r *http.Request) {
	galleryID, err := strconv.Atoi(chi.URLParam(r, "galleryID"))
	if err != nil {
		log.Printf("❌ Invalid gallery ID: %v", err)
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	mediaIDStr := r.FormValue("media_id")
	if mediaIDStr == "" {
		log.Println("❌ No media_id sent in request")
		http.Error(w, "Missing media ID", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.Atoi(mediaIDStr)
	if err != nil {
		log.Printf("❌ Invalid media ID: %v", err)
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	// Set the cover image
	err = app.GalleryModel.SetCoverImage(galleryID, mediaID)
	if err != nil {
		log.Printf("❌ Error setting cover image: %v", err)
		http.Error(w, "Error setting cover image", http.StatusInternalServerError)
		return
	}

	// Fetch media for rendering the new preview
	media, err := app.MediaModel.GetByIDAndGallery(mediaID, galleryID)
	if err != nil {
		log.Printf("❌ Cover image not found: %v", err)
		http.Error(w, "Media not found in this gallery", http.StatusNotFound)
		return
	}

	thumbURL := "https://" + os.Getenv("VULTR_S3_ENDPOINT") + "/" + app.S3Bucket + "/Uploads/" + media.FileName

	// ✅ Render the updated preview container
	app.render(w, r, "partials/cover_preview.html", map[string]interface{}{
		"GalleryID":     galleryID,
		"CoverImageURL": thumbURL,
	})
}

func (app *Application) SetGalleryVisibility(w http.ResponseWriter, r *http.Request) {
	// Get gallery id from URL parameter
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		log.Printf("❌ Invalid gallery ID: %v", err)
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	// The checkbox sends "on" if checked, empty if not.
	published := r.FormValue("published") == "on"

	// Update the published status in the database.
	err = app.GalleryModel.SetPublished(id, published)
	if err != nil {
		log.Printf("❌ Error updating gallery visibility: %v", err)
		http.Error(w, "Error updating gallery visibility", http.StatusInternalServerError)
		return
	}

	// Respond with a 200 OK (no content is needed for HTMX)
	w.WriteHeader(http.StatusOK)
}

func (app *Application) UpdateProjectMediaOrder(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ProjectID int   `json:"project_id"`
		Order     []int `json:"order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("❌ Invalid JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := app.MediaModel.UpdatePositionsForProject(payload.ProjectID, payload.Order)
	if err != nil {
		log.Printf("❌ Failed to update project media order: %v", err)
		http.Error(w, "Failed to update media order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Sort Media
func (app *Application) UpdateMediaOrderBulk(w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)

	var payload struct {
		Order     []int `json:"order"`
		GalleryID int   `json:"gallery_id"`
	}

	err := json.Unmarshal(body, &payload)
	if err != nil {
		log.Printf("❌ Unmarshal error: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = app.MediaModel.UpdatePositionsInBulk(payload.GalleryID, payload.Order)
	if err != nil {
		log.Printf("❌ Failed to bulk update media positions: %v", err)
		http.Error(w, "Failed to update positions", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper function to verify recaptcha

func verifyRecaptcha(token, secret string) bool {
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{
			"secret":   {secret},
			"response": {token},
		},
	)

	if err != nil {
		log.Printf("❌ reCAPTCHA request failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	var result struct {
		Success    bool     `json:"success"`
		Score      float64  `json:"score"`
		Action     string   `json:"action"`
		Hostname   string   `json:"hostname"`
		ErrorCodes []string `json:"error-codes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("❌ Failed to parse reCAPTCHA response: %v", err)
		return false
	}

	// Log the full reCAPTCHA response for debugging
	log.Printf("🔍 reCAPTCHA Response: %+v", result)

	// If verification failed, log errors
	if !result.Success {
		log.Printf("❌ reCAPTCHA verification failed: %v", result.ErrorCodes)
		return false
	}

	// Require a minimum score to prevent spam bots
	if result.Score < 0.5 {
		log.Printf("⚠️ Low reCAPTCHA score (%f) - possible bot", result.Score)
		return false
	}

	return true
}
