package src

import (
	"log"
	"net"
	"net/http"
	"time"
)

func HandleIPPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// read IP
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm error: %v", err)
		http.Error(w, "Missing IP addr", 500)
		return
	}
	ip := r.Form.Get("ip")
	if ip == "" {
		http.Error(w, "Missing IP address", 500)
		return
	}
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		http.Error(w, "Invalid IP address", 500)
		return
	}

	result := Result{
		IP:              parsedIP,
		IsPub:           true, // example
		AbuseConfidence: 42,
		Country:         "US",
		UsageType:       "Residential",
		ISP:             "MockISP",
		Domain:          "mock.domain",
		TotalReports:    7,
		Users:           3,
		LastReported:    time.Now().Add(-24 * time.Hour),
		ThreatRisk:      "Medium",
		ParsedRes:       "is vpn",
	}

	Results = append(Results, result)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
