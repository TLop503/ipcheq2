package src

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func RenderTemplate(w http.ResponseWriter, pagePath string, data any) {
	templates := []string{
		filepath.Join("web/templates", "layout.html"),
		filepath.Join("web/templates", "history.html"),
		filepath.Join("web", pagePath),
	}

	t, err := template.ParseFiles(templates...)
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template Error", 500)
	}
}
