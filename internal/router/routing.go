package router

import (
	"log"
	"net/http"

	"github.com/tlop503/ipcheq2/internal/queries/abuseipdb"
)

func RouteWebui() {
	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "index.html", abuseipdb.Results)
	})
	http.HandleFunc("/ip", handleIPPost)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("web/js"))))
}

func RouteAPI() {
	http.HandleFunc("/api/firstparty", handleFirstPartyGet)
	http.HandleFunc("/api/thirdparty", handleThirdPartyGet)
}

func StartServing() {
	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
