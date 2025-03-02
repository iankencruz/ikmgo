package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// TemplateCache stores precompiled templates
var TemplateCache = make(map[string]*template.Template)

// LoadTemplates dynamically loads all templates
func LoadTemplates() error {
	basePath := "templates/"
	layout := filepath.Join(basePath, "base.html")

	var templates []string
	var partials []string

	// Walk through the templates folder and gather all template files
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore directories
		if info.IsDir() {
			return nil
		}

		// Skip base.html (it is included manually)
		if strings.HasSuffix(path, "base.html") {
			return nil
		}

		// Identify partials (inside templates/partials/)
		if strings.Contains(path, "/partials/") {
			partials = append(partials, path)
		} else {
			templates = append(templates, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Load full-page templates with partials
	for _, tmplPath := range templates {
		t, err := template.ParseFiles(append([]string{layout, tmplPath}, partials...)...)
		if err != nil {
			log.Printf("❌ Error loading template %s: %v", tmplPath, err)
			return err
		}

		// Extract filename without path
		templateName := strings.TrimPrefix(tmplPath, basePath)
		TemplateCache[templateName] = t

		log.Printf("✅ Loaded template: %s", templateName)
	}

	// Load partials separately (if needed)
	for _, partialPath := range partials {
		t, err := template.ParseFiles(partialPath)
		if err != nil {
			log.Printf("❌ Error loading partial template %s: %v", partialPath, err)
			return err
		}

		// Extract filename without path
		partialName := strings.TrimPrefix(partialPath, basePath)
		TemplateCache[partialName] = t

		log.Printf("✅ Loaded partial: %s", partialName)
	}

	return nil
}

// Render function to display templates
func render(w http.ResponseWriter, r *http.Request, tmpl string, data interface{}) {
	t, ok := TemplateCache[tmpl]
	if !ok {
		log.Printf("❌ Template not found in cache: %s", tmpl)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	err := t.Execute(w, data)
	if err != nil {
		log.Printf("❌ Error executing template %s: %v", tmpl, err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
