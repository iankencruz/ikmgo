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

var TemplateCache = make(map[string]*template.Template)

var funcMap = template.FuncMap{
	"hasPrefix": func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	},
}

func LoadTemplates() error {
	basePath := "templates/"
	baseTemplates := map[string]string{
		"public": filepath.Join(basePath, "base.html"),
		"admin":  filepath.Join(basePath, "admin_base.html"),
	}

	partials, err := filepath.Glob(filepath.Join(basePath, "partials/*.html"))
	if err != nil {
		return fmt.Errorf("failed to glob partials: %w", err)
	}

	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || strings.Contains(path, "/partials/") {
			return err
		}

		var selectedBase string
		if strings.Contains(path, "admin/") {
			selectedBase = baseTemplates["admin"]
		} else {
			selectedBase = baseTemplates["public"]
		}

		files := append([]string{selectedBase, path}, partials...)
		t, err := template.New(filepath.Base(path)).Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			log.Printf("❌ Error loading template %s: %v", path, err)
			return err
		}

		templateName := strings.TrimPrefix(path, basePath)
		TemplateCache[templateName] = t
		log.Printf("✅ Cached template: %s", templateName)
		return nil
	})

	return err
}

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

	userID, _ := GetSession(r)
	if userID > 0 {
		user, err := app.UserModel.GetUserByID(userID)
		if err == nil {
			data["User"] = user
		}
	}

	if app.SettingsModel != nil {
		settings, err := app.SettingsModel.GetAll()
		if err == nil {
			data["Settings"] = settings
		}
	}

	layout := "base"
	if strings.Contains(tmpl, "admin/") {
		layout = "base_admin"
	}

	err := t.ExecuteTemplate(w, layout, data)
	if err != nil {
		log.Printf("❌ Template render error for %s: %v", tmpl, err)
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

	userID, _ := GetSession(r)
	if userID > 0 {
		user, err := app.UserModel.GetUserByID(userID)
		if err == nil {
			data["User"] = user
		}
	}

	layout := "base"
	if strings.Contains(tmpl, "admin/") {
		layout = "base_admin"
	}

	return t.ExecuteTemplate(w, layout, data)
}
