package queries

import (
	"encoding/json"
	"fmt"
	"github.com/tlop503/ipcheq2/internal/queries/vpnid"
	"net/netip"
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
	return nil, nil
}

type ThirdPartyResponse struct {
	IPAddress string `json:"ipAddress"`

	ABIPDBResponse struct {
		IsPublic             bool   `json:"isPublic"`
		IsWhitelisted        bool   `json:"isWhitelisted"`
		AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
		CountryCode          string `json:"countryCode"`
		CountryName          string `json:"countryName"`
		UsageType            string `json:"usageType"`
		ISP                  string `json:"isp"`
		Domain               string `json:"domain"`
		Hostnames            string `json:"hostnames"`
		IsTor                bool   `json:"isTor"`
		TotalReports         int    `json:"totalReports"`
		NumDistinctUsers     int    `json:"numDistinctUsers"`
		LastReportedAt       string `json:"lastReportedAt"`
	} `json:"ABIPDBResponse"`

	VirusTotalResponse struct {
		Tags                 []string `json:"tags"`
		LastModificationDate int      `json:"lastModificationDate"`
		LastAnalysisDate     int      `json:"lastAnalysisDate"`
		LastAnalysisStats    struct {
			Malicious  int `json:"malicious"`
			Suspicious int `json:"suspicious"`
			Undetected int `json:"undetected"`
			Harmless   int `json:"harmless"`
			Timeout    int `json:"timeout"`
		} `json:"lastAnalysisStats"`
		RDAP struct {
			Handle       string `json:"handle"`
			StartAddress string `json:"startAddress"`
			EndAddress   string `json:"endAddress"`
			ParentHandle string `json:"parentHandle"`
		} `json:"RDAP"`
	} `json:"VirusTotalResponse"`
}
