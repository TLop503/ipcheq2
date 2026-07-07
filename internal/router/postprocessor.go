package router

import (
	"log"
	"net/netip"
	"slices"
	"strings"

	"github.com/tlop503/ipcheq2/v2/internal/queries/abuseipdb"

	"github.com/tlop503/ipcheq2/v2/internal/queries"
	"github.com/tlop503/ipcheq2/v2/internal/queries/virustotal"
)

func QueryAndStyle(ip netip.Addr) FrontEndData {
	// check first and third party sources
	// this covers abipdb, vt (if present), and vpnid
	data, err := queries.FullQueryToStruct(ip)
	if err != nil {
		log.Printf("FullQueryToStruct error: %v", err)
	}

	var fed FrontEndData
	fed.FQ = data

	fed.AbKeyPresent = abuseipdb.ABIPKeyPresent

	// populate VT data if present
	if virustotal.VTKeyPresent {
		fed.VtTotalDetections = data.VirusTotalResponse.LastAnalysisStats.Malicious +
			data.VirusTotalResponse.LastAnalysisStats.Suspicious

		fed.VtTotalEngines = fed.VtTotalDetections +
			data.VirusTotalResponse.LastAnalysisStats.Undetected +
			data.VirusTotalResponse.LastAnalysisStats.Harmless
	}

	// toggle on external links if abuse reported
	if len(data.VPNIDMatches) > 0 || fed.VtTotalDetections > 0 {
		fed.ShowAbuseLinks = true
	}

	if len(data.VPNIDMatches) > 0 {
		fed.VpnColorClass = classifyVPN(data.VPNIDMatches) // classify before reordering
		slices.Sort(data.VPNIDMatches)
		slices.Compact(data.VPNIDMatches)
		moveToEnd(data.VPNIDMatches, "Generic VPN from ASN Data V4")
		moveToEnd(data.VPNIDMatches, "Generic VPN from ASN Data V6")
		fed.VpnidParsedResults = strings.Join(data.VPNIDMatches, ", ")
		fed.VPNidHasMatches = true
	} else {
		fed.VpnidParsedResults = "Not found in VPNID"
		fed.VPNidHasMatches = false
		fed.VpnColorClass = VpnClassNone
	}

	return fed
}

// classifyVPN inspects all VPNIDMatches for an IP and returns a single CSS
// class based on priority: tor > icloud > vpn > generic.
func classifyVPN(matches []string) string {
	switch {
	case len(matches) == 0:
		return VpnClassNone
	case anyContains(matches, torKeywords):
		return VpnClassTor
	case anyContains(matches, icloudKeywords):
		return VpnClassICloud
	default:
		// anything else with matches present is treated as a generic VPN hit
		return VpnClassVPN
	}
}

// anyContains reports whether any of "matches" contains any of "keywords"
// (case-insensitive substring match).
func anyContains(matches []string, keywords []string) bool {
	for _, m := range matches {
		lower := strings.ToLower(m)
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return true
			}
		}
	}
	return false
}

// Size of the result buffer declared here!
var Results = NewResultsBuffer(8)

// VPN color classes
const (
	VpnClassNone   = "vpn-none"
	VpnClassTor    = "vpn-tor"
	VpnClassICloud = "vpn-icloud"
	VpnClassVPN    = "vpn-generic-provider" // known VPN, green
)

// torKeywords / icloudKeywords are substrings (checked case-insensitively)
// that identify a match string as belonging to that category.
var torKeywords = []string{"tor"}
var icloudKeywords = []string{"icloud", "private relay"}

type FrontEndData struct {
	FQ                 queries.FullQueryResponse
	VpnidParsedResults string `default:"Not found in VPNID"`
	VpnColorClass      string `default:"vpn-none"`
	VPNidHasMatches    bool   `default:"false"`
	VtTotalDetections  int    `default:"0"`
	VtTotalEngines     int    `default:"0"`
	ShowAbuseLinks     bool   `default:"false"`
	AbKeyPresent       bool   `default:"false"`
}
