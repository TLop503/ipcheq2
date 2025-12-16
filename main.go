package main

import (
	"github.com/tlop503/ipcheq2/internal/abuseipdb"
	"github.com/tlop503/ipcheq2/internal/api"
	"github.com/tlop503/ipcheq2/internal/vpnid"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {

	// Try to load .env file (optional for container deployment)
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables directly")
	}

	// Initialize API key in internal package
	abuseipdb.InitializeAPIKey()

	// Initialize VPN ID ranger
	vpnid.InitializeVpnID()

	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		api.RenderTemplate(w, "index.html", abuseipdb.Results)
	})
	http.HandleFunc("/ip", api.HandleIPPost)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
