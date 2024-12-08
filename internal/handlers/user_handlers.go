// internal/handlers/user_handlers.go
package handlers

import (
	"ikm/internal/viewdata"
	"net/http"

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

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		role := "user"

		// Use named parameters to insert a new user
		err = h.User.Insert(name, email, password, role)
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
		session.Values["userRole"] = user.Role
		session.Save(r, w)

		h.logger.Printf("Session userID: %d", user.ID)

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

	// Mark the session as expired
	session.Options.MaxAge = -1

	// Save the expired session to effectively log out the user
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to destroy session", http.StatusInternalServerError)
		return
	}

	// Redirect to the homepage or login page after logout
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
