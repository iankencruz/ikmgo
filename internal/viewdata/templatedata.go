// internal/viewdata/templatedata.go
package viewdata

import (
	"time"
)

type NavLink struct {
	Path     string
	Label    string
	IsActive bool
}

type TemplateData struct {
	CurrentPath   string
	NavLinks      []NavLink
	FlashMessages []string
	CurrentYear   int
	IsLoggedIn    bool
	Title         string
	Data          interface{}
}

func NewTemplateData() TemplateData {
	return TemplateData{
		CurrentYear:   time.Now().Year(),
		IsLoggedIn:    false,
		FlashMessages: []string{},
	}
}
