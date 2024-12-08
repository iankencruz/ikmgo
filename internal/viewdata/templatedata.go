// internal/viewdata/templatedata.go
package viewdata

import (
	"fmt"
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
	CurrentYear     int
	IsAuthenticated bool
	UserName        string
	Title           string
	Data            interface{}
}

// internal/viewdata/templatedata.go
func NewTemplateData(r *http.Request, sessionManager *session.Manager, userModel *models.UserModel) TemplateData {
	isAuthenticated := false
	userName := ""

	// Retrieve the userID from the session
	session, err := sessionManager.Get(r, "user-session")
	if err == nil {
		userID, ok := session.Values["userID"].(int64)
		if ok && userID != 0 {
			isAuthenticated = true

			// Fetch the user from the database using the userID
			user, err := userModel.GetByID(userID)
			if err == nil {
				userName = user.Name
			}
		}
	}

	fmt.Printf("UserName: %s, IsAuthenticated: %v\n", userName, isAuthenticated)

	return TemplateData{
		CurrentYear:     time.Now().Year(),
		IsAuthenticated: isAuthenticated,
		UserName:        userName,
		FlashMessages:   []string{},
	}
}
