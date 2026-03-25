package router

import (
	"github.com/tlop503/ipcheq2/internal/abuseipdb"
	"log"
	"net/http"
)

func RouteWebui() {
	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "index.html", abuseipdb.Results)
	})
	http.HandleFunc("/ip", handleIPPost)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("web/js"))))

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
