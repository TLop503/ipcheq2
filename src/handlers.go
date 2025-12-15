package src

import (
	"github.com/tlop503/ipcheq2/src/vpnid"
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

	rawIP := r.Form.Get("ip")

	rawIP = strings.TrimSpace(rawIP)

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

	// Check for VPN, iCloud, etc.

	queryRes, _, err := vpnid.Query(ip, VpnIDRanger)
	if err != nil {
		log.Fatal(err)
	} else {
		result.ParsedRes = queryRes
	}

	// result.ParsedRes now stores vpn/etc status

	Results = append([]Result{result}, Results...)
	if len(Results) > 5 {
		Results = Results[:5] // truncate for prettiness on screen.
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
