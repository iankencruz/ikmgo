// internal/viewdata/gallery.go
package viewdata

import (
	"ikm/internal/models"
	"ikm/internal/session"
	"net/http"
)

type GalleryPageData struct {
	TemplateData // Embedding the generic TemplateData struct
	Title        string
	Description  string
	Images       []string
}

func NewGalleryPageData(r *http.Request, session *session.Manager, userModel *models.UserModel, title, description string) GalleryPageData {
	templateData := NewTemplateData(r, session, userModel)

	return GalleryPageData{
		TemplateData: templateData,
		Title:        title,
		Description:  description,
		Images:       []string{},
	}
}
