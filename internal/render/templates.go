package render

import (
	"fmt"
	"html/template"
	"path/filepath"
)

func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.New(name).ParseFiles("./ui/html/base.html", page)
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
