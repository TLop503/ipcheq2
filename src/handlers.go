package src

import (
	"ipcheq2/src/vpnid"
	"log"
	"net/http"
	"net/netip"
	"strings"
)

func HandleIPPost2(w http.ResponseWriter, r *http.Request) {
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
	rawIP := r.Form.Get("ip")

	// Parse string -> netip.Addr
	ip, err := netip.ParseAddr(rawIP)
	if err != nil {
		log.Printf("ParseAddr error: %v", err)
		http.Error(w, "Missing or invalid IP addr", 500)
		return
	}

	// Abuseipdb query
	// This code is only reachable if the IP was parsed validly, so we know
	// we can safely use rawIP here
	result, err := checkAbuseIPDB(rawIP)
	if err != nil {
		log.Fatal(err)
	}

	// Check for VPN, iCloud, etc. if datacenter IP
	if strings.Contains(result.UsageType, "Data Center") {
		queryRes, found, err := vpnid.Query(ip, VpnIDRanger)
		if err != nil {
			log.Fatal(err)
		} else if !found && CheckICloudIP(ip) {
			result.ParsedRes = "iCloud Private Relay"
		} else {
			result.ParsedRes = queryRes
		}
	}
	// result.ParsedRes now stores vpn/etc status

	Results = append([]Result{result}, result)
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
	rawIP := r.Form.Get("ip")

	// Parse string -> netip.Addr
	ip, err := netip.ParseAddr(rawIP)
	if err != nil {
		log.Printf("ParseAddr error: %v", err)
		http.Error(w, "Missing or invalid IP addr", 500)
		return
	}

	for _, result := range Results {
		if result.IP == ip {
			queryRes, _, err := vpnid.Query(ip, VpnIDRanger)
			if err != nil {
				log.Printf("Query error: %v", err)
				http.Error(w, "VPN query failed", 500)
			}
			result.ParsedRes = queryRes
			break
		}
	}
}
