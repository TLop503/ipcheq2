package api

import (
	"github.com/tlop503/ipcheq2/internal/abuseipdb"
	"github.com/tlop503/ipcheq2/internal/vpnid"
	"log"
	"net/http"
	"net/netip"
	"strings"
)

// HandleIPPost parses out IP and queries abuseipdb and vpnid
func HandleIPPost(w http.ResponseWriter, r *http.Request) {
	// verify method, parsability
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Printf("ParseForm error: %v", err)
		http.Error(w, "Missing IP addr", 500)
		return
	}

	// Parse IP to netip.addr after trimming whitespace
	ip, err := netip.ParseAddr(strings.TrimSpace(r.Form.Get("ip")))
	if err != nil {
		log.Printf("ParseAddr error: %v", err)
		http.Error(w, "Missing or invalid IP addr", 500)
		return
	}

	// Abuseipdb query
	result, err := abuseipdb.CheckAbuseIPDB(ip)
	if err != nil {
		log.Fatal(err)
	}

	// Check for VPN, iCloud, etc.
	result.ParsedRes, err = vpnid.Query(ip, VpnIDRanger)
	if err != nil {
		log.Print(err)
	}

	abuseipdb.Results = append([]abuseipdb.Result{result}, abuseipdb.Results...)
	if len(abuseipdb.Results) > 5 {
		abuseipdb.Results = abuseipdb.Results[:5] // truncate for prettiness on screen.
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
