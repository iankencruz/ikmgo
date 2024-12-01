// ./internal/handlers/page_handlers.go
package handlers

import (
	vw "ikm/internal/viewdata"
	"net/http"
)

// HomePageHandler handles requests for the homepage
// HomePageHandler handles the homepage requests
func (h *Handlers) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title          string
		WelcomeMessage string
	}{
		Title:          "Home",
		WelcomeMessage: "Welcome to the home page!",
	}
	h.Render(w, r, "home", data)
}

// AboutPageHandler handles requests for the about page
func (h *Handlers) AboutPageHandler(w http.ResponseWriter, r *http.Request) {
	data := vw.NewTemplateData()
	data.CurrentPath = r.URL.Path
	h.Render(w, r, "about", data)
}
