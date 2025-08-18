package src

import (
	"log"
	"net"
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
	ip = strings.TrimSpace(ip)
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
		vpnResults, err := checkVPN(ip)
		if err != nil {
			log.Fatal(err)
		} else if !strings.Contains(vpnResults, "VPN") && CheckICloudIP(parsedIP) {
			result.ParsedRes = "iCloud Private Relay"
		} else {
			result.ParsedRes = vpnResults
		}
	}

	Results = append([]Result{result}, Results...)
	if len(Results) > 5 {
		Results = Results[:5] // truncate for prettiness on screen.
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// HandleVPNPost is used to directly query VPN information without going through the regular IP check flow.
// This is useful when an IP was already checked but didn't trigger the regular VPN query.
// The user can then manually query VPN information for that IP.
func HandleVPNPost(w http.ResponseWriter, r *http.Request) {
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
			vpnResults, err := checkVPN(ip)
			if err != nil {
				log.Printf("VPN query error: %v", err)
				http.Error(w, "VPN query failed", 500)
				return
			}
			Results[i].ParsedRes = vpnResults
			break
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
