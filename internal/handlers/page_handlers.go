// ./internal/handlers/page_handlers.go
package handlers

import (
	vw "ikm/internal/viewdata"
	"net/http"
)

// HomePageHandler handles requests for the homepage
// HomePageHandler handles the homepage requests
func (h *Handlers) HomePageHandler(w http.ResponseWriter, r *http.Request) {
	data := vw.NewHomepageData(r, h.session, h.User, "Welcome to the Image Gallery")

	data.CurrentPath = r.URL.Path

	h.Render(w, r, "homePage", data)

	// Render the updated navigation as OOB swap (render it only if it's HTMX)
	if isHTMX(r) {
		// Optionally apply custom behavior to the OOB navigation template
		h.Render(w, r, "nav_oob", data)
	}
}

// AboutPageHandler handles requests for the about page
func (h *Handlers) AboutPageHandler(w http.ResponseWriter, r *http.Request) {
	data := vw.NewTemplateData(r, h.session, h.User)
	data.CurrentPath = r.URL.Path
	h.Render(w, r, "aboutPage", data)

	// Render the updated navigation as OOB swap (render it only if it's HTMX)
	if isHTMX(r) {
		// Optionally apply custom behavior to the OOB navigation template
		h.Render(w, r, "nav_oob", data)
	}
}

func (h *Handlers) ContactPageHandler(w http.ResponseWriter, r *http.Request) {
	data := vw.NewTemplateData(r, h.session, h.User)
	data.CurrentPath = r.URL.Path
	h.Render(w, r, "contactPage", data)

	// Render the updated navigation as OOB swap (render it only if it's HTMX)
	if isHTMX(r) {
		// Optionally apply custom behavior to the OOB navigation template
		h.Render(w, r, "nav_oob", data)
	}
}
