package router

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/netip"
	"path/filepath"
	"slices"
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
	db_hits, err := vpnid.QueryToSlice(ip)
	if err != nil {
		log.Println(err)
	}

	slices.Sort(db_hits)
	slices.Compact(db_hits) //dedupe

	if len(db_hits) == 1 {
		result.ParsedRes = db_hits[0]
	} else if len(db_hits) == 0 {
		result.ParsedRes = "Not found in dataset"
	} else {
		moveToEnd(db_hits, "Generic VPN from ASN Data")
		result.ParsedRes = strings.Join(db_hits, ", ")
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

// handleFirstPartyGet funnels api queries to vpnid
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

// handleThirdPartyGet returns abuseipdb and vt data
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

// handleFullQuery is a combination of first and third party data
func handleFullQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	ip, ok := parseAPIIP(w, r)
	if !ok {
		return
	}

	response, err := queries.FullQuery(ip)
	if err != nil {
		log.Printf("FullQuery error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to run full query")
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

// moveToEnd is a helper function to nicely format results when multiple db hits occur
func moveToEnd(slice []string, target string) {
	for i, v := range slice {
		if v == target {
			// Shift elements left (overwrite target)
			copy(slice[i:], slice[i+1:])
			// Place target at the end
			slice[len(slice)-1] = target
			return
		}
	}
}
