package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"ikm/internal/models"
	"ikm/internal/session"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	logger        *log.Logger
	templateCache map[string]*template.Template
	db            *pgxpool.Pool
	Gallery       *models.GalleryModel
	Media         *models.MediaModel
	User          *models.UserModel
	S3Uploader    *s3manager.Uploader
	session       *session.Manager
}

// New creates a new Handlers instance
func New(logger *log.Logger, templateCache map[string]*template.Template, db *pgxpool.Pool, galleryModel *models.GalleryModel, mediaModel *models.MediaModel, userModel *models.UserModel, sessionManager *session.Manager) *Handlers {
	return &Handlers{
		logger:        logger,
		templateCache: templateCache,
		db:            db,
		Gallery:       galleryModel,
		Media:         mediaModel,
		User:          userModel,
		session:       sessionManager,
	}
}

// isHTMX is a helper function to check if the request is an HTMX request
func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") != ""
}

func (h *Handlers) Render(w http.ResponseWriter, r *http.Request, name string, data interface{}, options ...func(*template.Template, *bytes.Buffer) error) {
	// Append .html suffix to the template name
	templateName := name + ".html"

	ts, ok := h.templateCache[templateName]
	if !ok {
		h.logger.Printf("Template %s does not exist", templateName)
		http.Error(w, "The template does not exist", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	// Use the isHTMX helper function to determine if the request is an HTMX request
	if isHTMX(r) {
		// Render only the specified template (usually a partial)
		err := ts.ExecuteTemplate(buf, name, data)
		if err != nil {
			h.logger.Printf("Unable to render HTMX template: %v", err)
			http.Error(w, "Unable to render HTMX template", http.StatusInternalServerError)
			return
		}
	} else {
		// Render the full page using the base layout
		err := ts.ExecuteTemplate(buf, "base", data)
		if err != nil {
			h.logger.Printf("Unable to render template: %v", err)
			http.Error(w, "Unable to render template", http.StatusInternalServerError)
			return
		}
	}

	// Apply any additional options (for custom behavior)
	for _, option := range options {
		if err := option(ts, buf); err != nil {
			h.logger.Printf("Error applying custom render option: %v", err)
			http.Error(w, "Unable to render template", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}

// Helpers for error handling
func (h *Handlers) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	h.logger.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
