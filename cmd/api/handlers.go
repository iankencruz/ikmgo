package main

import (
	"context"
	"encoding/json"
	"fmt"
	"ikm/models"
	"ikm/utils"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

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
		"Masonry":    true,
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
		// ✅ Use map indexing instead of dot notation
		galleryID, ok := gallery["ID"].(int)
		if !ok {
			log.Printf("❌ Invalid gallery ID format: %v", gallery["ID"])
			continue
		}

		media, err := app.MediaModel.GetByGalleryID(galleryID)
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

	media, err := app.MediaModel.GetByGalleryID(id)
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
	if err := r.ParseMultipartForm(10 << 20); err != nil { // ✅ Parse only once
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

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

	// ✅ Fetch the next available position
	nextPosition, err := app.MediaModel.GetNextPosition(galleryID)
	if err != nil {
		http.Error(w, "Error retrieving next position", http.StatusInternalServerError)
		return
	}

	fileName := strconv.Itoa(galleryID) + "_" + handler.Filename
	fileKey := "Uploads/" + fileName

	ctx := context.Background()
	fileSize := handler.Size
	contentType := handler.Header.Get("Content-Type")

	_, err = app.S3Client.PutObject(
		ctx, app.S3Bucket, fileKey, file, fileSize,
		minio.PutObjectOptions{
			ContentType:  contentType,
			UserMetadata: map[string]string{"x-amz-acl": "public-read"},
		})

	if err != nil {
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}

	fileURL := "https://" + os.Getenv("VULTR_S3_ENDPOINT") + "/" + app.S3Bucket + "/" + fileKey

	err = app.MediaModel.Insert(fileName, fileURL, galleryID, nextPosition)
	if err != nil {
		http.Error(w, "Error saving media", http.StatusInternalServerError)
		return
	}

	// ✅ Return partial template with new image (HTMX live update)
	app.render(w, r, "partials/media_item.html", map[string]interface{}{
		"ID":       galleryID,
		"FileName": fileName,
		"URL":      fileURL,
	})
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

// Email sending utility
// func (app *Application) sendMail(from, to, subject, body string) error {
// 	smtpHost := os.Getenv("SMTP_HOST")
// 	smtpPort := os.Getenv("SMTP_PORT")
// 	smtpUser := os.Getenv("SMTP_USER")
// 	smtpPass := os.Getenv("SMTP_PASS")
//
// 	msg := []byte(fmt.Sprintf("Subject: %s\r\nFrom: %s\r\nTo: %s\r\n\r\n%s",
// 		subject, from, to, body))
//
// 	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
// 	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
//
// 	return smtp.SendMail(addr, auth, from, []string{to}, msg)
// }

// Contact Handler (GET + POST)

func (app *Application) Contact(w http.ResponseWriter, r *http.Request) {
	// 1. Handle GET request - show contact form
	if r.Method == http.MethodGet {
		app.render(w, r, "contact.html", map[string]interface{}{
			"Title": "Contact",
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

	app.render(w, r, "partials/contact_success_modal.html", nil) // ✅ Correct

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

	// 🔑 Fetch the updated media record
	media, err := app.MediaModel.GetByID(mediaID)
	if err != nil {
		log.Printf("❌ Cover image not found: %v", err)
		http.Error(w, "Cover image not found", http.StatusInternalServerError)
		return
	}

	// ✅ Use the S3 URL from the `media` record
	coverImageURL := media.URL

	// ✅ Return the updated partial to HTMX
	//    "partials/cover_image.html" is the path to your partial.
	//    If you're using a named block in cover_image.html, make sure to call ExecuteTemplate accordingly in your render method.
	app.render(w, r, "partials/cover_image.html", map[string]interface{}{
		"GalleryID":     galleryID,
		"CoverImageURL": coverImageURL,
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

func (app *Application) UpdateMediaOrder(w http.ResponseWriter, r *http.Request) {
	mediaID, err := strconv.Atoi(r.FormValue("media_id"))
	if err != nil {
		log.Printf("❌ Invalid media ID: %v", err)
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	newPosition, err := strconv.Atoi(r.FormValue("position"))
	if err != nil {
		log.Printf("❌ Invalid position: %v", err)
		http.Error(w, "Invalid position", http.StatusBadRequest)
		return
	}

	galleryID, err := strconv.Atoi(r.FormValue("gallery_id"))
	if err != nil {
		log.Printf("❌ Invalid gallery ID: %v", err)
		http.Error(w, "Invalid gallery ID", http.StatusBadRequest)
		return
	}

	// ✅ Ensure media exists before updating
	exists, err := app.MediaModel.MediaExists(mediaID, galleryID)
	if err != nil || !exists {
		log.Printf("❌ Media ID %d does not exist in Gallery ID %d", mediaID, galleryID)
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	// ✅ Update all media positions in the gallery
	err = app.MediaModel.ReorderPositions(galleryID, mediaID, newPosition)
	if err != nil {
		log.Printf("❌ Error updating media order: %v", err)
		http.Error(w, "Error updating media order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // ✅ Success response
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
