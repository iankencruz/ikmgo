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
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid dict call: must pass key-value pairs")
		}
		d := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			d[key] = values[i+1]
		}
		return d, nil
	},
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
	"seq": func(start, end int) []int {
		arr := make([]int, end-start+1)
		for i := range arr {
			arr[i] = start + i
		}
		return arr
	},
	"min": func(a, b int) int {
		if a < b {
			return a
		}
		return b
	},

	"mul": func(a, b int) int {
		return a * b
	},
	"hasSuffix": strings.HasSuffix,
	"coalesce": func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	},
	"split": strings.Split,
}

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

	// ✅ 1. Load full-page templates with base layout
	for _, tmplPath := range templates {
		var selectedBase string
		if strings.Contains(tmplPath, "admin/") {
			selectedBase = baseTemplates["admin"]
		} else {
			selectedBase = baseTemplates["public"]
		}

		files := append([]string{selectedBase, tmplPath}, partials...)
		t, err := template.New(filepath.Base(tmplPath)).Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			log.Printf("❌ Error loading template %s: %v", tmplPath, err)
			return err
		}

		templateName := strings.TrimPrefix(tmplPath, basePath)
		TemplateCache[templateName] = t
		// logging templates
		// log.Printf("✅ Cached template: %s", templateName)
		// for _, tmpl := range t.Templates() {
		// 	log.Printf("  └─ contains block: %s", tmpl.Name())
		// }
	}

	// ✅ 2. Load partials as root templates (individually)
	for _, partialPath := range partials {
		t, err := template.New(filepath.Base(partialPath)).Funcs(funcMap).ParseFiles(partialPath)
		if err != nil {
			log.Printf("❌ Error loading partial %s: %v", partialPath, err)
			continue
		}

		partialName := strings.TrimPrefix(partialPath, basePath)
		TemplateCache[partialName] = t

		// Logging Partials
		// log.Printf("✅ Cached partial: %s", partialName)
		// for _, tmpl := range t.Templates() {
		// 	log.Printf("  └─ contains block: %s", tmpl.Name())
		// }
	}

	return nil
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

	// ✅ Inject request path
	data["CurrentPath"] = r.URL.Path

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

func (app *Application) renderHTMX(w http.ResponseWriter, baseTemplate string, block string, data interface{}) {
	tmpl, ok := TemplateCache[baseTemplate]
	if !ok {
		http.Error(w, "Template not found: "+baseTemplate, http.StatusInternalServerError)
		return
	}

	err := tmpl.ExecuteTemplate(w, block, data)
	if err != nil {
		log.Printf("❌ Error rendering HTMX block '%s' from %s: %v", block, baseTemplate, err)
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (app *Application) renderPartialHTMX(w io.Writer, block string, data interface{}) {
	// Exact match: if the partial is loaded as a root template
	if tmpl, ok := TemplateCache[block]; ok {
		err := tmpl.ExecuteTemplate(w, block, data)
		if err != nil {
			log.Printf("❌ Render error (exact): %s - %v", block, err)
			if rw, ok := w.(http.ResponseWriter); ok {
				http.Error(rw, "Render error", http.StatusInternalServerError)
			}
		}
		return
	}

	// Fallback: search for block inside all templates
	for name, tmpl := range TemplateCache {
		if tmpl.Lookup(block) != nil {
			err := tmpl.ExecuteTemplate(w, block, data)
			if err != nil {
				log.Printf("❌ Render error (fallback in %s -> %s): %v", name, block, err)
				if rw, ok := w.(http.ResponseWriter); ok {
					http.Error(rw, "Render error", http.StatusInternalServerError)
				}
			}
			return
		}
	}

	log.Printf("❌ HTMX partial not found: %s", block)
	if rw, ok := w.(http.ResponseWriter); ok {
		http.Error(rw, "Partial not found", http.StatusInternalServerError)
	}
}
