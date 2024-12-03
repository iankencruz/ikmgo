// internal/render/templates.go
package render

import (
	"fmt"
	"html/template"
	"path/filepath"
	"time"
)

// HumanDate formats a time.Time object into a human-readable string
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// Define a FuncMap to hold the functions that will be available in the templates
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// NewTemplateCache creates a cache for all the templates
func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// Collect templates from pages, HTMX OOB, and partials directories
	templateDirs := []string{
		"./ui/html/pages/*.html",
		"./ui/html/htmx/*.html",
	}

	var templates []string
	for _, dir := range templateDirs {
		files, err := filepath.Glob(dir)
		if err != nil {
			return nil, fmt.Errorf("unable to collect templates from %s: %w", dir, err)
		}
		templates = append(templates, files...)
	}

	// Iterate through all templates and add them to the cache
	for _, tmpl := range templates {
		name := filepath.Base(tmpl)

		// Parse base layout, the current template, and partials
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html", tmpl)
		if err != nil {
			return nil, fmt.Errorf("unable to parse base and current template %s: %w", name, err)
		}

		// Parse partial templates
		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, fmt.Errorf("unable to parse partials for template %s: %w", name, err)
		}

		// Add template to cache
		cache[name] = ts
		fmt.Printf("Added template %s to cache\n", name)
	}

	fmt.Println("Template cache contents:")
	for name := range cache {
		fmt.Println(name)
	}

	return cache, nil
}
