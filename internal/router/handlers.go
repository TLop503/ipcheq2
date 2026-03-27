package router

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/netip"
	"path/filepath"
	"strings"

	"github.com/tlop503/ipcheq2/internal/queries"
	"github.com/tlop503/ipcheq2/internal/queries/abuseipdb"
	"github.com/tlop503/ipcheq2/internal/queries/virustotal"
	"github.com/tlop503/ipcheq2/internal/queries/vpnid"
)

// handleIPPost parses out IP and queries abuseipdb, vpnid, and virustotal
func handleIPPost(w http.ResponseWriter, r *http.Request) {
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

	// vpnID query - Check for VPN, iCloud, etc.
	result.ParsedRes, err = vpnid.Query(ip)
	if err != nil {
		log.Print(err)
	}

	// VirusTotal query
	if virustotal.VTKeyPresent {
		result.VtDetections, result.VtNumEngines, err = virustotal.CheckVirusTotal(ip)
		if err != nil {
			log.Print(err)
		}
	}

	abuseipdb.Results = append([]abuseipdb.Result{result}, abuseipdb.Results...)
	if len(abuseipdb.Results) > 5 {
		abuseipdb.Results = abuseipdb.Results[:5] // truncate for prettiness on screen.
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// renderTemplate executes templates into a single HTML to serve to the client
func renderTemplate(w http.ResponseWriter, pagePath string, data any) {
	templatePath := filepath.Join("web", pagePath)
	historyPath := filepath.Join("web/templates", "history.html")
	titlePath := filepath.Join("web/templates", "title.html")
	t, err := template.ParseFiles(templatePath, historyPath, titlePath)
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template Error", 500)
	}
}

func handleFirstPartyGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	ip, ok := parseAPIIP(w, r)
	if !ok {
		return
	}

	response, err := queries.FirstPartyQuery(ip)
	if err != nil {
		log.Printf("FirstPartyQuery error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to query first-party sources")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func handleThirdPartyGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	ip, ok := parseAPIIP(w, r)
	if !ok {
		return
	}

	response, err := queries.ThirdPartyQuery(ip)
	if err != nil {
		log.Printf("ThirdPartyQuery error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to query third-party sources")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func parseAPIIP(w http.ResponseWriter, r *http.Request) (netip.Addr, bool) {
	ipRaw := strings.TrimSpace(r.URL.Query().Get("ip"))
	if ipRaw == "" {
		writeJSONError(w, http.StatusBadRequest, "missing required query parameter: ip")
		return netip.Addr{}, false
	}

	ip, err := netip.ParseAddr(ipRaw)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid ip query parameter")
		return netip.Addr{}, false
	}

	return ip, true
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errResponse := map[string]string{"error": message}
	if err := json.NewEncoder(w).Encode(errResponse); err != nil {
		log.Printf("writeJSONError encode failure: %v", err)
	}
}
