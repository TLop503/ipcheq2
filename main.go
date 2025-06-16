package main

import (
	"html/template"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"
)

type Result struct {
	IP net.Addr
	// abuseipdb data
	IsPub           bool
	AbuseConfidence int
	Country         string
	UsageType       string
	ISP             string
	Domain          string
	TotalReports    int
	Users           int
	LastReported    time.Time
	ThreatRisk      string

	//spur
	ParsedRes string // is / is not vpn
}

func renderTemplate(w http.ResponseWriter, pagePath string, data any) {
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

func main() {

	var Results []Result

	// dummy test data
	Results = []Result{
		{
			IP:              &net.IPAddr{IP: net.ParseIP("192.0.2.1")},
			IsPub:           true,
			AbuseConfidence: 90,
			Country:         "US",
			UsageType:       "Data Center/Web Hosting/Transit",
			ISP:             "ExampleISP",
			Domain:          "example.com",
			TotalReports:    120,
			Users:           45,
			LastReported:    time.Now().Add(-48 * time.Hour),
			ThreatRisk:      "High",
			ParsedRes:       "is vpn",
		},
		{
			IP:              &net.IPAddr{IP: net.ParseIP("203.0.113.5")},
			IsPub:           false,
			AbuseConfidence: 0,
			Country:         "DE",
			UsageType:       "Residential",
			ISP:             "HomeISP",
			Domain:          "",
			TotalReports:    0,
			Users:           0,
			LastReported:    time.Time{},
			ThreatRisk:      "Low",
			ParsedRes:       "is not vpn",
		},
	}

	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "index.html", Results)
	})

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
