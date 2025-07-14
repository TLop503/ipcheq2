package src

import (
	"log"
	"net"
	"net/http"
	"strings"
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

	// Fire abuseipdb query
	result, err := checkAbuseIPDB(ip)
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(result.UsageType, "Data Center") {
		spurResults, err := checkSpur(ip)
		if err != nil {
			log.Fatal(err)
		} else {
			result.ParsedRes = spurResults
		}
	}

	Results = append([]Result{result}, Results...)
	if len(Results) > 5 {
		Results = Results[:5] // truncate for prettiness on screen.
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleSpurPost(w http.ResponseWriter, r *http.Request) {
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

	// Find the result in the Results slice and update it
	for i := range Results {
		if Results[i].IP.String() == parsedIP.String() {
			spurResults, err := checkSpur(ip)
			if err != nil {
				log.Printf("Spur query error: %v", err)
				http.Error(w, "Spur query failed", 500)
				return
			}
			Results[i].ParsedRes = spurResults
			break
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
