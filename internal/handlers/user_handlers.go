// internal/handlers/user_handlers.go
package handlers

import (
	"ikm/internal/validation"
	"ikm/internal/viewdata"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func (h *Handlers) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render the registration form
		data := viewdata.NewTemplateData(r, h.session, h.User)
		h.Render(w, r, "registerPage", data)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to process form", http.StatusBadRequest)
			return
		}

		fname := r.FormValue("first-name")
		lname := r.FormValue("last-name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		role := "admin"

		errors := validation.ValidateForm(
			validation.ValidateRequired("first_name", fname),
			validation.ValidateRequired("last_name", lname),
			validation.ValidateEmail("email", email),
			validation.ValidateMinLength("password", password, 8),
		)

		if len(errors) > 0 {
			data := viewdata.NewTemplateData(r, h.session, h.User)
			data.FieldErrors = errors
			h.Render(w, r, "registerPage", data)
			return
		}

		// Use named parameters to insert a new user
		err = h.User.Insert(fname, lname, email, password, role)
		if err != nil {
			http.Error(w, "Unable to create user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func (h *Handlers) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render the login form
		data := viewdata.NewTemplateData(r, h.session, h.User)
		h.Render(w, r, "loginPage", data)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to process form", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := h.User.GetByEmail(email)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)) != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Set a session or cookie for the authenticated user
		session, _ := h.session.Get(r, "user-session")
		session.Values["userID"] = user.ID
		session.Values["userRole"] = strings.ToLower(user.Role)
		if err := session.Save(r, w); err != nil {
			http.Error(w, "Failed to save session", http.StatusInternalServerError)
			return
		}

		h.logger.Printf("Session user Role: %v", user.Role)

		http.Redirect(w, r, "/galleries", http.StatusSeeOther)
	}
}

func (h *Handlers) LogoutUserHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the user session
	session, err := h.session.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Unable to retrieve session", http.StatusInternalServerError)
		return
	}

	// Remove isAuthenticated from the session
	delete(session.Values, "isAuthenticated")

	// Mark the session as expired
	session.Options.MaxAge = -1

	// Save the expired session to effectively log out the user
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to destroy session", http.StatusInternalServerError)
		return
	}

	// Handle HTMX request for redirect
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("HX-Redirect", "/")
		return
	}

	// Redirect to the homepage or login page after logout
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) ValidateEmailHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	err := validation.ValidateEmail("email", email)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(err.Message))
		return
	}

	w.Write([]byte("Email is valid"))
}
