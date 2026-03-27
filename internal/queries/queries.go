package queries

import (
	"encoding/json"
	"fmt"
	"log"
	"net/netip"

	"github.com/tlop503/ipcheq2/internal/queries/abuseipdb"
	"github.com/tlop503/ipcheq2/internal/queries/virustotal"
	"github.com/tlop503/ipcheq2/internal/queries/vpnid"
)

// FirstPartyQuery only checks local data, currently just VPNID
func FirstPartyQuery(addr netip.Addr) ([]byte, error) {
	results, err := vpnid.QueryToSlice(addr)
	if err != nil {
		fmt.Println(err)
	}

	response := FirstPartyResponse{
		IPAddress:    addr.String(),
		VPNIDMatches: results,
	}

	return json.Marshal(response)
}

// FirstPartyResponse maps VPNIDMatches to db hits
type FirstPartyResponse struct {
	IPAddress    string   `json:"ipAddress"`
	VPNIDMatches []string `json:"vpnid_matches"`
}

// ThirdPartyQuery only checks external data sources (abipdb and optionally vt)
func ThirdPartyQuery(addr netip.Addr) ([]byte, error) {
	var result ThirdPartyResponse

	result.IPAddress = addr.String()

	// Query ABIPDB
	abResults, err := abuseipdb.QueryAbuseIPDB(addr)
	if err != nil {
		log.Printf("Issue with abuseipdb query in ThirdPartyQuery: %s", err)
	}
	result.ABIPDBResponse = abResults

	// Query VT (if key)
	if virustotal.VTKeyPresent {
		vtResults, err := virustotal.QueryVirusTotal(addr)
		if err != nil {
			log.Printf("Issue with virustotal query in ThirdPartyQuery: %s", err)
		}
		result.VirusTotalResponse = vtResults
	}

	return json.Marshal(result)
}

type ThirdPartyResponse struct {
	IPAddress          string                        `json:"ipAddress"`
	ABIPDBResponse     abuseipdb.ABIPDBResponse      `json:"ABIPDBResponse"`
	VirusTotalResponse virustotal.VirusTotalResponse `json:"VirusTotalResponse"`
}

func FullQuery(addr netip.Addr) ([]byte, error) {
	var result FullQueryResponse

	result.IPAddress = addr.String()

	fpResults, err := vpnid.QueryToSlice(addr)
	if err != nil {
		log.Printf("Issue with vpnid query in FullQuery: %s", err)
	}
	result.VPNIDMatches = fpResults

	abResults, err := abuseipdb.QueryAbuseIPDB(addr)
	if err != nil {
		log.Printf("Issue with abuseipdb query in FullQuery: %s", err)
	}
	result.ABIPDBResponse = abResults

	if virustotal.VTKeyPresent {
		vtResults, err := virustotal.QueryVirusTotal(addr)
		if err != nil {
			log.Printf("Issue with virustotal query in FullQuery: %s", err)
		}
		result.VirusTotalResponse = vtResults
	}

	return json.Marshal(result)
}

type FullQueryResponse struct {
	IPAddress          string                        `json:"ipAddress"`
	VPNIDMatches       []string                      `json:"vpnid_matches"`
	ABIPDBResponse     abuseipdb.ABIPDBResponse      `json:"ABIPDBResponse"`
	VirusTotalResponse virustotal.VirusTotalResponse `json:"VirusTotalResponse"`
}
