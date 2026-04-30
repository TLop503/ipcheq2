package router

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/netip"
	"strings"

	"github.com/tlop503/ipcheq2/v2/internal/queries"
	"github.com/tlop503/ipcheq2/v2/internal/web"
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

	Results.Add(QueryAndStyle(ip))

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// renderTemplate executes templates into a single HTML to serve to the client
func renderTemplate(w http.ResponseWriter, pagePath string, data any) {
	// Read template files from embedded FS
	mainTemplateBytes, err := fs.ReadFile(web.FS, pagePath)
	if err != nil {
		log.Printf("Template file read error (%s): %v", pagePath, err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	historyBytes, err := fs.ReadFile(web.FS, "templates/history.html")
	if err != nil {
		log.Printf("Template file read error (templates/history.html): %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	titleBytes, err := fs.ReadFile(web.FS, "templates/title.html")
	if err != nil {
		log.Printf("Template file read error (templates/title.html): %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Create and parse templates
	t, err := template.New(pagePath).Parse(string(mainTemplateBytes))
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Parse history and title templates
	_, err = t.New("history.html").Parse(string(historyBytes))
	if err != nil {
		log.Printf("Template parsing error (history): %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	_, err = t.New("title.html").Parse(string(titleBytes))
	if err != nil {
		log.Printf("Template parsing error (title): %v", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Execute template
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template Error", 500)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(buf.Bytes())
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
