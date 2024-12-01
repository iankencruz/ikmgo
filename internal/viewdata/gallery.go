// internal/viewdata/gallery.go
package viewdata

import "net/http"

type GalleryPageData struct {
	TemplateData // Embedding the generic TemplateData struct
	Title        string
	Description  string
	Images       []string
}

func NewGalleryPageData(r *http.Request, title, description string) GalleryPageData {
	templateData := NewTemplateData()

	return GalleryPageData{
		TemplateData: templateData,
		Title:        title,
		Description:  description,
		Images:       []string{},
	}
}
