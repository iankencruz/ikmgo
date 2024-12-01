// internal/viewdata/templatedata.go
package viewdata

import "time"

type TemplateData struct {
	CurrentPath   string
	FlashMessages []string
	CurrentYear   int
	IsLoggedIn    bool
	Title         string
}

func NewTemplateData() TemplateData {
	return TemplateData{
		CurrentYear:   time.Now().Year(),
		IsLoggedIn:    false,
		FlashMessages: []string{},
	}
}
