package virustotal

import (
	"fmt"
	"log"
	"net/netip"
	"os"
	"strings"

	vt "github.com/VirusTotal/vt-go"
)

// per "https://github.com/VirusTotal/vt-go", vt-go is the official Go client library for VirusTotal.
// With this library you can interact with the VirusTotal REST API v3 without having to send plain HTTP requests with the standard "http" package.
var VTKeyPresent bool = false
var vtKey string
var vtClient *vt.Client

// InitializeAPIKey sets the VirusTotal API key from environment
func InitializeVTAPIKey() {
	InitializeVTAPIKeyFromValue(os.Getenv("VTKEY"))
}

// InitializeVTAPIKeyFromValue sets the VirusTotal API key from an explicit value.
func InitializeVTAPIKeyFromValue(key string) {
	VTKeyPresent = false
	vtClient = nil

	vtKey = strings.TrimSpace(key)
	if vtKey == "" {
		log.Println("Warning: VTKEY not set; VirusTotal lookups are disabled")
	} else {
		VTKeyPresent = true
		log.Println("VT Key loaded")
		vtClient = vt.NewClient(vtKey)
	}
}

// QueryVirusTotal returns a raw struct of VT data for use in ipcheq API third-party queries
func QueryVirusTotal(ip netip.Addr) (VirusTotalResponse, error) {
	result, err := vtClient.GetObject(vt.URL("ip_addresses/%s", ip.String()))
	if err != nil {
		return VirusTotalResponse{}, fmt.Errorf("failed to query virustotal: %v", err)
	}

	var resp VirusTotalResponse

	// Simple fields
	resp.Tags = getStringSlice(result, "tags")
	resp.LastModificationDate = getInt(result, "last_modification_date")
	resp.LastAnalysisDate = getInt(result, "last_analysis_date")

	// Stats
	resp.LastAnalysisStats.Malicious = getInt(result, "last_analysis_stats.malicious")
	resp.LastAnalysisStats.Suspicious = getInt(result, "last_analysis_stats.suspicious")
	resp.LastAnalysisStats.Undetected = getInt(result, "last_analysis_stats.undetected")
	resp.LastAnalysisStats.Harmless = getInt(result, "last_analysis_stats.harmless")
	resp.LastAnalysisStats.Timeout = getInt(result, "last_analysis_stats.timeout")

	// RDAP
	resp.RDAP.Handle = getString(result, "rdap.handle")
	resp.RDAP.StartAddress = getString(result, "rdap.start_address")
	resp.RDAP.EndAddress = getString(result, "rdap.end_address")
	resp.RDAP.ParentHandle = getString(result, "rdap.parent_handle")

	return resp, nil
}

type VirusTotalResponse struct {
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
}
