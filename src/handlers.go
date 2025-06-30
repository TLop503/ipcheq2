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
		spur_results, err := checkSpur(ip)
		if err != nil {
			log.Fatal(err)
		} else {
			result.ParsedRes = spur_results
		}
	}

	Results = append(Results, result)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
