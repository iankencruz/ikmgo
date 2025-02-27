// internal/viewdata/templatedata.go
package viewdata

import (
	"ikm/internal/models"
	"ikm/internal/session"
	"net/http"
	"time"
)

type NavLink struct {
	Path     string
	Label    string
	IsActive bool
}

type TemplateData struct {
	CurrentPath     string
	NavLinks        []NavLink
	FlashMessages   []string
	Flash           string
	CurrentYear     int
	IsAuthenticated bool
	User            *models.SanitizedUser
	Title           string
	Data            interface{}
	FieldErrors     map[string]string
	Form            any
}

// internal/viewdata/templatedata.go
func NewTemplateData(r *http.Request, sessionManager *session.Manager, userModel *models.UserModel) TemplateData {
	isAuthenticated := false
	var user *models.SanitizedUser

	// Retrieve the userID from the session
	session, err := sessionManager.Get(r, "user-session")
	if err == nil {
		userID, ok := session.Values["userID"].(int64)
		if ok && userID != 0 {
			isAuthenticated = true

			// Fetch the user from the database using the userID
			dbUser, err := userModel.GetByID(userID)
			if err == nil {
				user = &models.SanitizedUser{
					ID:        dbUser.ID,
					FirstName: dbUser.FirstName,
					LastName:  dbUser.LastName,
					Email:     dbUser.Email,
					Role:      dbUser.Role,
				}

			}
		}
	}

	return TemplateData{
		CurrentYear:     time.Now().Year(),
		IsAuthenticated: isAuthenticated,
		User:            user,
		FlashMessages:   []string{},
		FieldErrors:     make(map[string]string),
	}
}
