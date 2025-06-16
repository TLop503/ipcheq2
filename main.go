package main

import (
	"log"
	"net/http"

	"ipcheq2/src"
)

func main() {

	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		src.RenderTemplate(w, "index.html", src.Results)
	})
	http.HandleFunc("/ip", src.HandleIPPost)

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
