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

// LoadTemplates dynamically loads all templates with support for multiple base layouts
func LoadTemplates() error {
	basePath := "templates/"
	baseTemplates := map[string]string{
		"public": filepath.Join(basePath, "base.html"),
		"admin":  filepath.Join(basePath, "admin_base.html"),
	}

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

	// ✅ First, parse all partials and store them separately
	partialsTemplate := template.New("")
	for _, partialPath := range partials {
		_, err := partialsTemplate.ParseFiles(partialPath)
		if err != nil {
			log.Printf("❌ Error loading partial template %s: %v", partialPath, err)
			return err
		}
		templateName := strings.TrimPrefix(partialPath, basePath)
		TemplateCache[templateName] = partialsTemplate
		log.Printf("✅ Loaded partial: %s", templateName)
	}

	// ✅ Load full-page templates with respective base layouts
	for _, tmplPath := range templates {
		var selectedBase string
		if strings.Contains(tmplPath, "admin/") {
			selectedBase = baseTemplates["admin"]
		} else {
			selectedBase = baseTemplates["public"]
		}

		// ✅ Parse the full template, including all partials
		t, err := template.ParseFiles(append([]string{selectedBase, tmplPath}, partials...)...)
		if err != nil {
			log.Printf("❌ Error loading template %s: %v", tmplPath, err)
			return err
		}

		// Extract filename without path
		templateName := strings.TrimPrefix(tmplPath, basePath)
		TemplateCache[templateName] = t
		log.Printf("✅ Loaded template: %s", templateName)
	}

	return nil
}

// Render function to display templates

func (app *Application) render(w http.ResponseWriter, r *http.Request, tmpl string, data map[string]interface{}) {
	t, ok := TemplateCache[tmpl]
	if !ok {
		log.Printf("❌ Template not found in cache: %s", tmpl)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	// Optionally, pass the logged-in user (if any).
	userID, _ := GetSession(r)
	if userID > 0 {
		user, err := app.UserModel.GetUserByID(userID)
		if err == nil {
			data["User"] = user
		}
	}

	var err error

	// For partials in "partials/" we use ExecuteTemplate with the name from `define "..."`.
	if strings.HasPrefix(tmpl, "partials/") {
		// "cover_image" is the define name we used in cover_image.html
		err = t.ExecuteTemplate(w, "cover_image", data)
	} else {
		// For full-page templates, just render the entire file normally
		err = t.Execute(w, data)
	}

	if err != nil {
		log.Printf("❌ Error executing template %s: %v", tmpl, err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
