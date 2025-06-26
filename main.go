package main

import (
	"log"
	"net/http"

	"ipcheq2/src"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize API key in src package
	src.InitializeAPIKey()

	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		src.RenderTemplate(w, "index.html", src.Results)
	})
	http.HandleFunc("/ip", src.HandleIPPost)

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
