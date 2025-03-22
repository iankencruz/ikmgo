package main

import (
	"fmt"
	"html/template"
	"io"
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
		log.Printf("✅ Cached partial: %s", templateName)
	}

	return nil
}

// Render function to display templates

func (app *Application) render(w http.ResponseWriter, r *http.Request, tmpl string, data map[string]interface{}) {
	// Ensure template cache is properly populated
	if TemplateCache == nil {
		TemplateCache = make(map[string]*template.Template)
	}

	// Check if the template is already in cache
	t, ok := TemplateCache[tmpl]
	if !ok {
		// If not found, reload all templates and attempt again
		err := LoadTemplates()
		if err != nil {
			log.Printf("❌ Error loading templates: %v", err)
			http.Error(w, "Error loading templates", http.StatusInternalServerError)
			return
		}

		t, ok = TemplateCache[tmpl]
		if !ok {
			log.Printf("❌ Template still not found in cache: %s", tmpl)
			http.Error(w, "Template not found", http.StatusInternalServerError)
			return
		}
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	// Optionally, pass the logged-in user (if any)
	userID, _ := GetSession(r)
	if userID > 0 {
		user, err := app.UserModel.GetUserByID(userID)
		if err == nil {
			data["User"] = user
		}
	}

	var err error

	// Load settings
	settings, err := app.SettingsModel.GetAll()
	if err == nil {
		data["Settings"] = settings
	}

	// Determine if this is a partial and extract the template name
	if strings.HasPrefix(tmpl, "partials/") {
		// partialName := strings.TrimPrefix(tmpl, "partials/")
		err = t.ExecuteTemplate(w, tmpl, data)
	} else {
		// Full-page template rendering
		err = t.Execute(w, data)
	}

	if err != nil {
		log.Printf("❌ Error executing template %s: %v", tmpl, err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func (app *Application) renderToWriter(w io.Writer, r *http.Request, tmpl string, data map[string]interface{}) error {
	t, ok := TemplateCache[tmpl]
	if !ok {
		return fmt.Errorf("template not found in cache: %s", tmpl)
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	// Optionally include the logged-in user if session is active
	userID, _ := GetSession(r)
	if userID > 0 {
		user, err := app.UserModel.GetUserByID(userID)
		if err == nil {
			data["User"] = user
		}
	}

	return t.ExecuteTemplate(w, tmpl, data)

}
