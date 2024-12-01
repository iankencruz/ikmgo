// internal/viewdata/homepage.go
package viewdata

import "net/http"

type HomepageData struct {
	TemplateData      // Embedding the generic TemplateData struct
	WelcomeMessage    string
	FeaturedGalleries []string
}

func NewHomepageData(r *http.Request, welcomeMessage string) HomepageData {
	templateData := NewTemplateData()

	return HomepageData{
		TemplateData:      templateData,
		WelcomeMessage:    welcomeMessage,
		FeaturedGalleries: []string{},
	}
}
