package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ikm/models"
	"ikm/utils"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
)

// Home Page Handler

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	gallery, media, err := app.GalleryModel.GetByTitle("Home")

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("❌ Error fetching featured gallery: %v", err)
	}

	data := map[string]interface{}{
		"Title":      "Home",
		"Gallery":    gallery,
		"Media":      media,
		"ActiveLink": "home",
	}

	app.render(w, r, "index.html", data)
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
		app.render(w, r, "login.html", map[string]interface{}{
			"Title":       "Login",
			"HideSidebar": true,
		})
		return
	}

	// Parse form values
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Form error", http.StatusBadRequest)
		return
	}

	form := utils.NewForm(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		app.render(w, r, "login.html", map[string]interface{}{
			"Title":       "Login",
			"HideSidebar": true,
			"Form":        form,
		})
		http.Error(w, "Invalid login", http.StatusUnauthorized)
		fmt.Printf("❌ Error authenticating user: %v", form.NonFieldErrors)
		return
	}

	// Authenticate user
	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := app.UserModel.Authenticate(email, password)
	if err != nil {
		form.NonFieldErrors = append(form.NonFieldErrors, "Invalid email or password")
		app.render(w, r, "login.html", map[string]interface{}{
			"Title":       "Login",
			"HideSidebar": true,
			"Form":        form,
		})
		http.Error(w, "Invalid login", http.StatusUnauthorized)
		fmt.Printf("❌ Error authenticating user: %v", form.NonFieldErrors)
		return
	}

	SetSession(user.ID, w)
	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
}

// Logout Handler
func (app *Application) Logout(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// Admin Dashboard Handler
func (app *Application) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	galleryCount, err := app.GalleryModel.Count()
	if err != nil {
		log.Printf("❌ Error fetching gallery count: %v", err)
		http.Error(w, "Error fetching gallery count", http.StatusInternalServerError)
		return
	}

	// get latest galleries
	latestGalleries, err := app.GalleryModel.GetLatest(5)
	if err != nil {
		log.Printf("❌ Error fetching latest galleries: %v", err)
		http.Error(w, "Error fetching latest galleries", http.StatusInternalServerError)
		return
	}

	projectCount, err := app.ProjectModel.Count()
	if err != nil {
		log.Printf("❌ Error fetching project count: %v", err)
		http.Error(w, "Error fetching project count", http.StatusInternalServerError)
		return
	}

	// get latest projects
	latestProjects, err := app.ProjectModel.GetLatest(5)
	if err != nil {
		log.Printf("❌ Error fetching latest projects: %v", err)
		http.Error(w, "Error fetching latest projects", http.StatusInternalServerError)
		return
	}

	// get count of users
	userCount, err := app.UserModel.Count()
	if err != nil {
		log.Printf("❌ Error fetching user count: %v", err)
		http.Error(w, "Error fetching user count", http.StatusInternalServerError)
		return
	}

	// Get Media Count
	mediaCount, err := app.MediaModel.Count()
	if err != nil {
		log.Printf("❌ Error fetching media count: %v", err)
		http.Error(w, "Error fetching media count", http.StatusInternalServerError)
		return
	}

	// Get Latest Media
	latestMedia, err := app.MediaModel.GetLatest(5)
	if err != nil {
		log.Printf("❌ Error fetching latest media: %v", err)
		http.Error(w, "Error fetching latest media", http.StatusInternalServerError)
		return
	}

	// Get Contacts Count
	contactCount, err := app.ContactModel.Count()
	if err != nil {
		log.Printf("❌ Error fetching contact count: %v", err)
		http.Error(w, "Error fetching contact count", http.StatusInternalServerError)
		return
	}
	// Get latest contacts
	contacts, err := app.ContactModel.GetLatest(5)
	if err != nil {
		log.Printf("❌ Error fetching latest contacts: %v", err)
		http.Error(w, "Error fetching latest contacts", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":        "Admin Dashboard",
		"ActiveLink":   "dashboard",
		"GalleryCount": galleryCount,

		"LatestGalleries": latestGalleries,
		"ProjectCount":    projectCount,
		"LatestProjects":  latestProjects,
		"UserCount":       userCount,
		"MediaCount":      mediaCount,
		"LatestMedia":     latestMedia,
		"ContactCount":    contactCount,
		"LatestContacts":  contacts,
	}
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
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Save all key/value settings

	for key, values := range r.MultipartForm.Value {
		if key == "about_me_image" {
			continue // ✅ skip this one — it’s handled separately
		}
		if len(values) == 0 {
			continue
		}
		value := values[0]
		if err := app.SettingsModel.Set(key, value); err != nil {
			http.Error(w, "Failed to update setting: "+key, http.StatusInternalServerError)
			return
		}
	}

	// Handle optional About Me image upload
	file, handler, err := r.FormFile("about_me_image")
	if err == nil {
		defer file.Close()

		// Decode image
		img, err := imaging.Decode(file, imaging.AutoOrientation(true))
		if err != nil {
			log.Printf("❌ Failed to decode image: %v", err)
			http.Error(w, "Invalid image format", http.StatusBadRequest)
			return
		}

		// Resize to max 1500x900 (preserving aspect ratio)
		resized := imaging.Fit(img, 1000, 900, imaging.Lanczos)

		// Encode to buffer
		var buf bytes.Buffer
		if err := imaging.Encode(&buf, resized, imaging.JPEG); err != nil {
			log.Printf("❌ Failed to encode resized image: %v", err)
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}

		// Generate S3 object name

		objectName := fmt.Sprintf("settings/about_me_image_%d_%s", time.Now().UnixNano(), handler.Filename)

		// Upload to S3
		_, err = app.S3Client.PutObject(r.Context(), app.S3Bucket, objectName, &buf, int64(buf.Len()), minio.PutObjectOptions{
			ContentType: "image/jpeg",
			UserMetadata: map[string]string{
				"x-amz-acl": "public-read",
			},
		})
		if err != nil {
			log.Printf("❌ Failed to upload image to S3: %v", err)
			http.Error(w, "S3 upload failed", http.StatusInternalServerError)
			return
		}

		// Save public URL to settings
		imageURL := fmt.Sprintf("https://%s/%s/%s", os.Getenv("VULTR_S3_ENDPOINT"), app.S3Bucket, objectName)
		if err := app.SettingsModel.Set("about_me_image", imageURL); err != nil {
			log.Printf("❌ Failed to save setting: %v", err)
			http.Error(w, "Error saving setting", http.StatusInternalServerError)
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
		"Title":        "Projects",
		"Projects":     projects,
		"CanonicalURL": utils.BuildCanonicalURL(r, "/projects"),
		"ActiveLink":   "projects",
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
	slug := chi.URLParam(r, "slug")

	project, err := app.ProjectModel.GetBySlug(slug)
	if err != nil {
		log.Printf("❌ Project not found: %v", err)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	media, err := app.ProjectModel.GetMediaPaginated(project.ID, 100, 0)
	if err != nil {
		log.Printf("❌ Failed to get media for project %d: %v", project.ID, err)
		http.Error(w, "Unable to load media", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Project %d media count: %d", project.ID, len(media))

	var heroMedia []*models.Media
	var restMedia []*models.Media

	if len(media) > 1 {
		heroMedia = media[:1]
		restMedia = media[1:]
	} else {
		heroMedia = media
	}

	canonicalURL := fmt.Sprintf("https://%s/projects/%s", os.Getenv("BASE_URL"), project.Slug)

	data := map[string]interface{}{
		"Title":        project.Title,
		"ActiveLink":   "projects",
		"Project":      project,
		"ProjectID":    project.ID,
		"CanonicalURL": canonicalURL,
		"Description":  project.Description,
		"OGImage":      project.CoverImageURL,
		"HeroMedia":    heroMedia,
		"Media":        restMedia, // remaining media
		"ParentTitle":  "Projects",
		"ParentURL":    "/projects",
		"CurrentLabel": project.Title,
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

	data := map[string]interface{}{
		"Title":      "Manage Galleries",
		"Galleries":  galleries,
		"ActiveLink": "galleries",
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

	pageStr := r.URL.Query().Get("page")
	page := 0
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}
	limit := 20
	offset := page * limit

	log.Printf("🧪 GalleryID: %d | Page: %d | Limit: %d | Offset: %d", id, page, limit, offset)

	// Fetch gallery info
	gallery, err := app.GalleryModel.GetByID(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Load limit+1 items to check if there’s more
	media, err := app.GalleryModel.GetMediaPaginated(id, limit+1, offset)
	if err != nil {
		http.Error(w, "Failed to get media", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(gallery.MediaCount) / float64(limit)))

	hasNext := false
	if len(media) > limit {
		hasNext = true
		media = media[:limit] // trim the extra item
	}

	log.Printf("🧪 Page: %d | Offset: %d | Media: %d | HasNext: %v", page, offset, len(media), hasNext)

	data := map[string]interface{}{
		"Title":             "Edit Gallery",
		"Media":             media,
		"MediaCount":        gallery.MediaCount,
		"GalleryID":         id,
		"Gallery":           gallery,
		"Page":              page,
		"Limit":             limit,
		"TotalPages":        totalPages,
		"HasNext":           hasNext,
		"PaginationBaseURL": fmt.Sprintf("/admin/gallery/%d", id),
		"Target":            "#sortableGrid",
	}

	// If HTMX, render just the sortable media grid block
	if utils.IsHTMX(r) {
		app.renderPartialHTMX(w, "admin_media_grid", data)
		return
	}

	// Full page render
	app.render(w, r, "admin/edit_gallery.html", data)
}

// Update Gallery Title

func (app *Application) UpdateGallery(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	title := r.FormValue("title")
	description := r.FormValue("description")

	err := app.GalleryModel.Update(id, title, description, utils.Slugify(title))
	if err != nil {
		http.Error(w, "Error updating gallery", http.StatusInternalServerError)
		return
	}

	// Get updated gallery to pass to partial
	gallery, err := app.GalleryModel.GetByID(id)
	if err != nil {
		http.Error(w, "Error fetching updated gallery", http.StatusInternalServerError)
		return
	}

	app.renderPartialHTMX(w, "partials/gallery_info_static.html", map[string]interface{}{
		"Gallery": gallery,
	})
}

// Create a New Gallery (POST)
func (app *Application) CreateGallery(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	description := r.FormValue("description")

	err := app.GalleryModel.Create(title, description, utils.Slugify(title))
	if err != nil {
		http.Error(w, "Error creating gallery", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusFound)
}

func (app *Application) DeleteGallery(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	// Delete from database
	err = app.GalleryModel.Delete(id)
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

	err := app.ProjectModel.Create(title, description, utils.Slugify(title))
	if err != nil {
		http.Error(w, "Error creating project", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/projects", http.StatusSeeOther)
}

// Delete Project
func (app *Application) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Delete from database
	err = app.ProjectModel.Delete(id)
	if err != nil {
		http.Error(w, "Error deleting project", http.StatusInternalServerError)
		return
	}
	// HTMX: Remove the row without reloading
	w.WriteHeader(http.StatusOK)
}

func (app *Application) EditProjectForm(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	project, err := app.ProjectModel.GetByID(id)
	if err != nil {
		log.Printf("❌ GetByID: Failed to find project %d: %v", id, err)
		http.NotFound(w, r)
		return
	}

	// 🔢 Pagination Params
	pageStr := r.URL.Query().Get("page")
	page := 0
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}
	limit := 20
	offset := page * limit

	// 📦 Paginated Media
	media, err := app.ProjectModel.GetMediaPaginated(id, limit+1, offset)
	if err != nil {
		log.Printf("❌ Error fetching media for project %d: %v", id, err)
		http.Error(w, "Error loading media", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(project.MediaCount) / float64(limit)))

	hasNext := false
	if len(media) > limit {
		hasNext = true
		media = media[:limit] // trim the extra item
	}

	log.Printf("🧪 EditProjectForm: Page=%d | Limit=%d | TotalMedia=%d | HasNext=%t", page, limit, project.MediaCount, hasNext)

	data := map[string]interface{}{
		"Title":             "Edit Project",
		"Media":             media,
		"MediaCount":        project.MediaCount,
		"ProjectID":         id,
		"Project":           project,
		"Page":              page,
		"Limit":             limit,
		"TotalPages":        totalPages,
		"HasNext":           hasNext,
		"PaginationBaseURL": fmt.Sprintf("/admin/project/edit/%d", id),
		"Target":            "#sortableGrid",
	}

	// 🧠 HTMX: If HTMX, only re-render sortableGrid block
	if utils.IsHTMX(r) {
		app.renderPartialHTMX(w, "admin_media_grid", data)
		return
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

	thumbURL := media.ThumbnailURL

	app.renderPartialHTMX(w, "partials/cover_preview.html", map[string]interface{}{
		"ProjectID":     projectID,
		"CoverImageURL": thumbURL,
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
	pageStr := r.URL.Query().Get("page")
	page := 0
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}
	limit := 15
	offset := page * limit

	totalMedia, err := app.MediaModel.Count()
	if err != nil {
		log.Printf("❌ Failed to count media: %v", err)
		http.Error(w, "Unable to count media", http.StatusInternalServerError)
		return
	}

	media, err := app.MediaModel.GetPaginated(limit, offset)
	if err != nil {
		log.Printf("❌ Failed to load paginated media: %v", err)
		http.Error(w, "Unable to load media", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalMedia) / float64(limit)))
	hasNext := (page+1)*limit < totalMedia

	data := map[string]interface{}{
		"Title":             "Manage Media",
		"Media":             media,
		"MediaCount":        totalMedia,
		"Page":              page,
		"Limit":             limit,
		"HasNext":           hasNext,
		"TotalPages":        totalPages,
		"PaginationBaseURL": "/admin/media",
		"Target":            "#sortable",
		"ActiveLink":        "media",
	}

	if utils.IsHTMX(r) {
		app.renderPartialHTMX(w, "admin_media_grid", data)
		log.Printf("HTMX: Rendered media grid for page %d", page)
		return
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

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		files = r.MultipartForm.File["files[]"] // fallback to match frontend
	}

	var outputBuffer bytes.Buffer
	ctx := context.Background()

	projectID, _ := strconv.Atoi(r.FormValue("project_id"))
	galleryID, _ := strconv.Atoi(r.FormValue("gallery_id"))

	isProject := projectID > 0
	isGallery := galleryID > 0

	var position int
	var errPos error

	if isProject {
		position, errPos = app.MediaModel.GetNextProjectPosition(projectID)
	} else if isGallery {
		position, errPos = app.GalleryModel.GetNextPosition(galleryID)
	}

	if errPos != nil && (isProject || isGallery) {
		log.Printf("❌ Error getting next position: %v", errPos)
		http.Error(w, "Error getting next position", http.StatusInternalServerError)
		return
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("❌ Error opening file: %v", err)
			continue
		}
		defer file.Close()

		base := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileHeader.Filename)
		fileName := base
		thumbName := "thumb_" + base

		fileKey := "Uploads/" + fileName
		thumbKey := "Uploads/" + thumbName

		img, err := imaging.Decode(file, imaging.AutoOrientation(true))
		if err != nil {
			log.Printf("❌ Error decoding image: %v", err)
			continue
		}

		thumbnailImg := imaging.Resize(img, 500, 0, imaging.Lanczos)
		file.Seek(0, 0)

		fileSize := fileHeader.Size
		contentType := fileHeader.Header.Get("Content-Type")

		// Upload original
		_, err = app.S3Client.PutObject(ctx, app.S3Bucket, fileKey, io.LimitReader(file, fileSize), fileSize, minio.PutObjectOptions{
			ContentType:  contentType,
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
		})
		if err != nil {
			log.Printf("❌ Error uploading original: %v", err)
			continue
		}

		// Upload thumbnail
		var thumbBuf bytes.Buffer
		if err := imaging.Encode(&thumbBuf, thumbnailImg, imaging.JPEG); err == nil {
			app.S3Client.PutObject(ctx, app.S3Bucket, thumbKey, bytes.NewReader(thumbBuf.Bytes()), int64(thumbBuf.Len()), minio.PutObjectOptions{
				ContentType:  "image/jpeg",
				UserMetadata: map[string]string{"x-amz-acl": "public-read"},
			})
		}

		// Create URLs
		fullURL := "https://" + os.Getenv("VULTR_S3_ENDPOINT") + "/" + app.S3Bucket + "/" + fileKey
		thumbURL := "https://" + os.Getenv("VULTR_S3_ENDPOINT") + "/" + app.S3Bucket + "/" + thumbKey

		// Insert into media table
		mediaID, err := app.MediaModel.InsertAndReturnID(fileName, fullURL, thumbURL)
		if err != nil {
			log.Printf("❌ DB insert failed: %v", err)
			continue
		}

		// Attach to project or gallery if needed
		if isProject {

			err = app.MediaModel.AttachToProject(projectID, mediaID, position)
			if err != nil {
				log.Printf("❌ Failed to attach to project: %v", err)
				continue
			}
			position++
		} else if isGallery {
			err = app.GalleryModel.AttachMedia(galleryID, mediaID, position)
			if err != nil {
				log.Printf("❌ Failed to attach to gallery: %v", err)
				continue
			}
			position++
		}

		// Render media item partial
		media := &models.Media{
			ID:           mediaID,
			FileName:     fileName,
			ThumbnailURL: thumbURL,
			FullURL:      fullURL,
		}
		fmt.Printf("Rendering media item: %+v\n", media)

		app.renderPartialHTMX(w, "partials/media_item.html", map[string]any{
			"Media":     media,
			"ProjectID": projectID,
			"GalleryID": galleryID,
		})
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write(outputBuffer.Bytes())
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

	app.renderPartialHTMX(w, "partials/alert_toast.html", map[string]any{
		"Heading":  "Message Sent!",
		"Subtitle": "Your message has been submitted successfully.",
		"Variant":  "success", // options: success, error, warning, info
	})

}

// Get All Galleries
func (app *Application) PublicGalleriesList(w http.ResponseWriter, r *http.Request) {
	galleries, err := app.GalleryModel.GetAllPublic()
	log.Printf("Galleries: %v", galleries)

	if err != nil {
		http.Error(w, "Error fetching galleries", http.StatusInternalServerError)
		return
	}
	app.render(w, r, "galleries.html", map[string]interface{}{
		"Title":        "Galleries",
		"Galleries":    galleries,
		"CanonicalURL": utils.BuildCanonicalURL(r, "/galleries"),
		"ActiveLink":   "galleries",
	})
}

// View Gallery Page

func (app *Application) GalleryView(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	gallery, err := app.GalleryModel.GetBySlug(slug)

	log.Printf("✅ GalleryView requested for ID: %d", gallery.ID)

	// Fetch media
	media, err := app.GalleryModel.GetMediaPaginated(gallery.ID, 25, 0)
	if err != nil {
		log.Printf("❌ Error fetching media: %v", err)
		http.Error(w, "Error retrieving media", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Fetched %d media items for Gallery ID: %d", len(media), gallery.ID)

	// Canonical URL for SEO
	canonical := utils.BuildCanonicalURL(r, fmt.Sprintf("/gallery/%s", gallery.Slug))

	if r.Header.Get("HX-Request") != "" {
		log.Println("🔄 HTMX request detected, rendering media partial")
		app.render(w, r, "partials/gallery_component.html", map[string]interface{}{
			"Gallery": gallery,
			"Media":   media,
		})
		return
	}

	app.render(w, r, "gallery.html", map[string]interface{}{
		"Title":        gallery.Title,
		"Description":  gallery.Description,
		"CanonicalURL": canonical,
		"OGImage":      gallery.CoverImageURL,
		"ActiveLink":   "galleries",
		"Gallery":      gallery,
		"Media":        media,
		"ParentURL":    "/galleries",
		"CurrentLabel": gallery.Title,
		"ParentTitle":  "Galleries",
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
	app.renderPartialHTMX(w, "partials/cover_preview.html", map[string]interface{}{
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
	var payload struct {
		Order     []int `json:"order"`
		GalleryID int   `json:"gallery_id,omitempty"`
		ProjectID int   `json:"project_id,omitempty"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		log.Printf("❌ Unmarshal error: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("🧪 GalleryID: %d | ProjectID: %d | Order: %v", payload.GalleryID, payload.ProjectID, payload.Order)

	if payload.GalleryID != 0 {
		err = app.MediaModel.UpdatePositionsForGallery(payload.GalleryID, payload.Order)
	} else if payload.ProjectID != 0 {
		err = app.MediaModel.UpdatePositionsForProject(payload.ProjectID, payload.Order)
	}

	if err != nil {
		log.Printf("❌ Failed to bulk update media positions: %v", err)
		http.Error(w, "Failed to update positions", http.StatusInternalServerError)
		return
	}

	log.Printf("📦 Received reorder: GalleryID=%d | ProjectID=%d | Order=%v", payload.GalleryID, payload.ProjectID, payload.Order)

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

func (app *Application) GetAboutMeImageModal(w http.ResponseWriter, r *http.Request) {
	media, err := app.MediaModel.GetAll()
	if err != nil {
		http.Error(w, "Error loading media", http.StatusInternalServerError)
		return
	}

	app.renderPartialHTMX(w, "partials/about_me_image_modal.html", map[string]interface{}{
		"Media":   media,
		"Context": "settings",
	})
}

func (app *Application) SetAboutMeImage(w http.ResponseWriter, r *http.Request) {
	idStr := r.FormValue("media_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	media, err := app.MediaModel.GetByID(id)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	err = app.SettingsModel.Set("about_me_image", media.ThumbnailURL)
	if err != nil {
		http.Error(w, "Failed to save setting", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"ImageURL": media.ThumbnailURL,
	}
	app.renderPartialHTMX(w, "partials/about_image_preview.html", data)

}

func (app *Application) UploadMediaModal(w http.ResponseWriter, r *http.Request) {
	projectIDStr := r.URL.Query().Get("project_id")
	galleryIDStr := r.URL.Query().Get("gallery_id")
	contextParam := r.URL.Query().Get("context")

	pageStr := r.URL.Query().Get("page")
	page := 0
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p >= 0 {
			page = p
		}
	}
	limit := 20
	offset := page * limit

	var existingMedia []*models.Media
	var mediaCount int
	var err error

	if projectIDStr != "" {
		projectID, _ := strconv.Atoi(projectIDStr)
		existingMedia, mediaCount, err = app.MediaModel.GetUnlinkedMediaPaginated("project_media", "project_id", projectID, limit+1, offset)
	} else if galleryIDStr != "" {
		galleryID, _ := strconv.Atoi(galleryIDStr)
		existingMedia, mediaCount, err = app.MediaModel.GetUnlinkedMediaPaginated("gallery_media", "gallery_id", galleryID, limit+1, offset)
	} else if contextParam == "standalone" {
		existingMedia, err = app.MediaModel.GetPaginated(limit+1, offset)
		if err != nil {
			log.Printf("❌ Failed to load media: %v", err)
			http.Error(w, "Error loading media", http.StatusInternalServerError)
			return
		}
		mediaCount, err = app.MediaModel.Count()
		if err != nil {
			log.Printf("❌ Failed to count media: %v", err)
			http.Error(w, "Error counting media", http.StatusInternalServerError)
			return
		}
	} else {
		existingMedia, err = app.MediaModel.GetPaginated(limit+1, offset)
		if err != nil {
			log.Printf("❌ Failed to load media: %v", err)
			http.Error(w, "Error loading media", http.StatusInternalServerError)
			return
		}
		mediaCount, err = app.MediaModel.Count()
		if err != nil {
			log.Printf("❌ Failed to count media: %v", err)
			http.Error(w, "Error counting media", http.StatusInternalServerError)
			return
		}
	}
	if err != nil {
		log.Printf("❌ Failed to load media: %v", err)
		http.Error(w, "Error loading media", http.StatusInternalServerError)
		return
	}

	// ---- existing pagination calculation ----
	hasNext := len(existingMedia) > limit
	if hasNext {
		existingMedia = existingMedia[:limit]
	}

	totalPages := int(math.Ceil(float64(mediaCount) / float64(limit)))

	context := "standalone"
	if galleryIDStr != "" {
		context = "gallery"
	} else if projectIDStr != "" {
		context = "project"
	} else if contextParam == "settings" {
		context = "settings"
	}

	// ---- NEW Pagination BaseURL logic ----
	q := r.URL.Query()
	q.Del("page") // Remove existing page param
	encodedParams := q.Encode()

	paginationBaseURL := r.URL.Path
	if encodedParams != "" {
		paginationBaseURL += "?" + encodedParams
	}

	app.renderPartialHTMX(w, "partials/upload_media_modal.html", map[string]any{
		"ExistingMedia":     existingMedia,
		"Context":           context,
		"ProjectID":         projectIDStr,
		"GalleryID":         galleryIDStr,
		"ActiveTab":         "existing",
		"Page":              page,
		"Limit":             limit,
		"HasNext":           hasNext,
		"TotalPages":        totalPages,
		"MediaCount":        mediaCount,
		"PaginationBaseURL": paginationBaseURL,
		"Target":            "#upload-tab-existing",
	})
}

func (app *Application) ProjectInfoView(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	project, err := app.ProjectModel.GetByID(id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Project": project,
	}

	app.renderPartialHTMX(w, "partials/project_info_view.html", data)
}

func (app *Application) ProjectInfoEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	project, err := app.ProjectModel.GetByID(id)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	app.renderPartialHTMX(w, "partials/project_info_form.html", project)
}

func (app *Application) ProjectInfoUpdate(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")

	err := app.ProjectModel.UpdateBasicInfo(id, title, description, utils.Slugify(title))
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	// Return updated view
	project, _ := app.ProjectModel.GetByID(id)

	data := map[string]interface{}{
		"Project": project,
	}

	app.renderPartialHTMX(w, "partials/project_info_view.html", data)
}

func (app *Application) AttachMediaToItem(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.Atoi(r.FormValue("media_id"))
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	// Handle attaching to project
	if projectIDStr := r.FormValue("project_id"); projectIDStr != "" {
		projectID, err := strconv.Atoi(projectIDStr)
		if err != nil {
			http.Error(w, "Invalid project ID", http.StatusBadRequest)
			return
		}

		err = app.MediaModel.InsertProjectMedia(projectID, mediaID)
		if err != nil {
			log.Printf("❌ Failed to link media %d to project %d: %v", mediaID, projectID, err)
			http.Error(w, "Failed to attach media to project", http.StatusInternalServerError)
			return
		}

		media, err := app.MediaModel.GetByIDUnsafe(mediaID)
		if err != nil {
			http.Error(w, "Failed to load media", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Trigger", "refresh-admin-grid")
		w.Header().Set("HX-Trigger", fmt.Sprintf("media-attached-%d", mediaID))
		w.Header().Set("HX-Trigger-After-Settle", "show-toast")

		// ✅ Render media_grid block inside edit_project.html
		app.renderPartialHTMX(w, "partials/media_item.html", map[string]interface{}{
			"Media":     media,
			"ProjectID": projectID,
		})
		return
	}

	// Handle attaching to gallery

	if galleryIDStr := r.FormValue("gallery_id"); galleryIDStr != "" {
		galleryID, err := strconv.Atoi(galleryIDStr)
		if err != nil {
			log.Printf("❌ Invalid gallery ID: %v", galleryIDStr)
			http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
			return
		}

		log.Printf("📎 Linking media_id=%d to gallery_id=%d", mediaID, galleryID)

		err = app.MediaModel.InsertGalleryMedia(galleryID, mediaID)
		if err != nil {
			log.Printf("❌ Failed to link media %d to gallery %d: %v", mediaID, galleryID, err)
			http.Error(w, "Failed to attach media to gallery", http.StatusInternalServerError)
			return
		}

		log.Println("🔍 Attempting GetByIDUnsafe...")
		media, err := app.MediaModel.GetByIDUnsafe(mediaID)
		if err != nil {
			log.Printf("❌ GetByIDUnsafe failed for media_id=%d: %v", mediaID, err)
			http.Error(w, "Media not found", http.StatusNotFound)
			return
		}

		log.Printf("✅ Found media: ID=%d, File=%s", media.ID, media.FileName)

		w.Header().Set("HX-Trigger", fmt.Sprintf("media-attached-%d", mediaID))
		w.Header().Set("HX-Trigger-After-Settle", "show-toast")

		app.renderPartialHTMX(w, "partials/media_item.html", map[string]interface{}{
			"Media":     media,
			"GalleryID": galleryID,
		})

		return
	}

	http.Error(w, "Missing project_id or gallery_id", http.StatusBadRequest)
}

func (app *Application) UnlinkMediaFromItem(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.Atoi(r.FormValue("media_id"))
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	// Project unlink
	if projectIDStr := r.FormValue("project_id"); projectIDStr != "" {
		projectID, _ := strconv.Atoi(projectIDStr)
		err := app.MediaModel.UnlinkMediaFromProject(projectID, mediaID)
		if err != nil {
			http.Error(w, "Failed to unlink media from project", http.StatusInternalServerError)
			return
		}
		w.Header().Set("HX-Trigger-After-Settle", "show-toast-unlinked")
		w.Header().Set("HX-Trigger", "refresh-admin-grid")
		w.WriteHeader(http.StatusOK)

		return
	}

	// Gallery unlink
	if galleryIDStr := r.FormValue("gallery_id"); galleryIDStr != "" {
		galleryID, _ := strconv.Atoi(galleryIDStr)
		err := app.MediaModel.UnlinkMediaFromGallery(galleryID, mediaID)
		if err != nil {
			http.Error(w, "Failed to unlink media from gallery", http.StatusInternalServerError)
			return
		}
		w.Header().Set("HX-Trigger-After-Settle", "show-toast-unlinked")
		w.Header().Set("HX-Trigger", "refresh-admin-grid")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Missing project_id or gallery_id", http.StatusBadRequest)
}

func (app *Application) Toast(w http.ResponseWriter, r *http.Request) {
	variant := r.URL.Query().Get("variant")
	heading := r.URL.Query().Get("heading")
	subtitle := r.URL.Query().Get("subtitle")

	app.renderPartialHTMX(w, "partials/alert_toast.html", map[string]interface{}{
		"Heading":  heading,
		"Subtitle": subtitle,
		"Variant":  variant,
		"Timeout":  5000,
	})
}

func (app *Application) AdminGalleryInfoView(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	gallery, _ := app.GalleryModel.GetByID(id)
	app.renderPartialHTMX(w, "partials/gallery_info_static.html", map[string]interface{}{
		"Gallery": gallery,
	})
}

func (app *Application) AdminGalleryInfoEdit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	gallery, _ := app.GalleryModel.GetByID(id)
	app.renderPartialHTMX(w, "partials/gallery_info_form.html", map[string]interface{}{
		"Gallery": gallery,
	})
}

func (app *Application) deleteFromS3(key string) error {
	err := app.S3Client.RemoveObject(context.Background(), app.S3Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("❌ S3 delete failed for %s: %w", key, err)
	}
	return nil
}

// Delete Media File (DELETE)

func (app *Application) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	mediaIDStr := r.FormValue("media_id")
	mediaID, err := strconv.Atoi(mediaIDStr)
	if err != nil || mediaID <= 0 {
		log.Printf("❌ Invalid media ID: %v", err)
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	media, err := app.MediaModel.GetByID(mediaID)
	if err != nil {
		log.Printf("❌ Media not found: %v", err)
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	if err := app.MediaModel.Delete(mediaID); err != nil {
		log.Printf("❌ Failed to delete media from DB: %v", err)
		http.Error(w, "Failed to delete media", http.StatusInternalServerError)
		return
	}

	// Delete thumbnail (same folder, with "thumb_" prefix)
	thumbKey := "Uploads/thumb_" + media.FileName

	// 🧹 Delete from S3 (MinIO)
	if err := app.deleteFromS3("Uploads/" + media.FileName); err != nil {
		log.Printf("⚠️ Failed to delete full image: %v", err)
	}
	if err := app.deleteFromS3(thumbKey); err != nil {
		log.Printf("⚠️ Failed to delete thumbnail: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}
