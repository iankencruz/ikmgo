// internal/viewdata/homepage.go
package viewdata

import (
	"ikm/internal/models"
	"ikm/internal/session"
	"net/http"
)

type HomepageData struct {
	TemplateData      // Embedding the generic TemplateData struct
	WelcomeMessage    string
	FeaturedGalleries []string
}

// NewHomepageData initializes HomepageData with dynamic template data and additional fields.
func NewHomepageData(
	r *http.Request,
	sessionManager *session.Manager,
	userModel *models.UserModel,
	welcomeMessage string,
) HomepageData {
	// Generate TemplateData dynamically
	templateData := NewTemplateData(r, sessionManager, userModel)

	return HomepageData{
		TemplateData:      templateData, // Embed TemplateData
		WelcomeMessage:    welcomeMessage,
		FeaturedGalleries: []string{}, // Placeholder for additional data
	}
}
