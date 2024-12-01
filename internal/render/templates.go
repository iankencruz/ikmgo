// internal/render/templates.go
package render

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
	"time"
)

// HumanDate formats a time.Time object into a human-readable string
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// Check if the given path is active
func isActive(currentPath, basePath string) bool {
	return strings.HasPrefix(currentPath, basePath)
}

// Define a FuncMap to hold the functions that will be available in the templates
var functions = template.FuncMap{
	"humanDate": humanDate,
	"isActive":  isActive,
}

func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		// Use the global FuncMap when creating the template
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html", page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		cache[name] = ts
		fmt.Printf("Added template %s to cache\n", name)
	}

	fmt.Println("Template cache contents:")
	for name := range cache {
		fmt.Println(name)
	}

	return cache, nil
}
