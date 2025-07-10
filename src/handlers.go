package src

import (
	"log"
	"net/http"
	"net/netip"
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
	parsedIP, err := netip.ParseAddr(ip)
	if err != nil {
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
		} else if !strings.Contains(spurResults, "VPN") && CheckICloudIP(parsedIP) {
			result.ParsedRes = "iCloud Private Relay"
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
