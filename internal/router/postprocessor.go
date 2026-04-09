package router

import (
	"log"
	"net/netip"
	"slices"
	"strings"

	"github.com/tlop503/ipcheq2/internal/queries"
	"github.com/tlop503/ipcheq2/internal/queries/virustotal"
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

	// make pretty string for vpnid hits
	if len(data.VPNIDMatches) > 0 {
		// sort and dedupe
		slices.Sort(data.VPNIDMatches)
		slices.Compact(data.VPNIDMatches)
		moveToEnd(data.VPNIDMatches, "Generic VPN from ASN Data") // as it's least verbose
		fed.VpnidParsedResults = strings.Join(data.VPNIDMatches, ", ")
	} else {
		fed.VpnidParsedResults = "Not found in VPNID"
	}

	return fed
}

// Size of the result buffer declared here!
var Results = NewResultsBuffer(8)

type FrontEndData struct {
	FQ                 queries.FullQueryResponse
	VpnidParsedResults string `default:"Not found in VPNID"`
	VtTotalDetections  int    `default:"0"`
	VtTotalEngines     int    `default:"0"`
	ShowAbuseLinks     bool   `default:"false"`
}
