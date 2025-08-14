package main

import (
	"log"
	"net/http"

	"ipcheq2/src"

	"github.com/joho/godotenv"
)

func main() {

	// Try to load .env file (optional for container deployment)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables directly")
	}

	// Initialize API key in src package
	src.InitializeAPIKey()

	// Initialize Spur session cookie (mandatory)
	src.InitializeSpurCookie()

	// load iCloud private relay IPs
	src.LoadICloudPrefixes()

	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		src.RenderTemplate(w, "index.html", src.Results)
	})
	http.HandleFunc("/ip", src.HandleIPPost)
	http.HandleFunc("/spur", src.HandleSpurPost)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
