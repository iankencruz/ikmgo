package handlers

import (
	"bytes"
	"html/template"
	"ikm/internal/models"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	logger        *log.Logger
	templateCache map[string]*template.Template
	db            *pgxpool.Pool
	Gallery       *models.GalleryModel
	Media         *models.MediaModel
}

// New creates a new Handlers instance
func New(logger *log.Logger, templateCache map[string]*template.Template, db *pgxpool.Pool, galleryModel *models.GalleryModel, mediaModel *models.MediaModel) *Handlers {
	return &Handlers{
		logger:        logger,
		templateCache: templateCache,
		db:            db,
		Gallery:       galleryModel,
		Media:         mediaModel,
	}
}

// Render method for rendering templates
func (h *Handlers) Render(w http.ResponseWriter, r *http.Request, name string, data interface{}) {
	// Append .html suffix to the template name
	templateName := name + ".html"

	ts, ok := h.templateCache[templateName]
	if !ok {
		h.logger.Printf("Template %s does not exist", templateName)
		http.Error(w, "The template does not exist", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.Execute(buf, data)
	if err != nil {
		h.logger.Printf("Unable to render template: %v", err)
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}
