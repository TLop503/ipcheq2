package src

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func RenderTemplate(w http.ResponseWriter, pagePath string, data any) {
	templatePath := filepath.Join("web", pagePath)
	historyPath := filepath.Join("web/templates", "history.html")
	titlePath := filepath.Join("web/templates", "title.html")
	t, err := template.ParseFiles(templatePath, historyPath, titlePath)
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template Error", 500)
	}
}
