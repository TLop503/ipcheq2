package virustotal

import (
	"log"
	"net/netip"
	"os"

	vt "github.com/VirusTotal/vt-go"
)

// per "https://github.com/VirusTotal/vt-go", vt-go is the official Go client library for VirusTotal.
// With this library you can interact with the VirusTotal REST API v3 without having to send plain HTTP requests with the standard "http" package.
var VTKeyPresent bool = false
var vtKey string

// InitializeAPIKey sets the VirusTotal API key from environment
func InitializeVTAPIKey() {
	vtKey = os.Getenv("VTKEY")
	if vtKey == "" {
		log.Println("Warning: VTKEY environment variable is not set")
	} else {
		VTKeyPresent = true
		log.Println("VT Key loaded")
	}
}

// Queries VirusTotal for an IP and returns number of malicious detections + total num of engines
func CheckVirusTotal(ip netip.Addr) (int, int, error) {

	client := vt.NewClient(vtKey)

	result, err := client.GetObject(vt.URL("ip_addresses/%s", ip.String()))
	if err != nil {
		return 0, 0, err
	}

	// fields in the VT IP scan result
	fields := []string{"harmless", "malicious", "suspicious", "undetected", "timeout"}
	var total, malicious int

	for _, field := range fields {
		fieldValue, err := result.GetInt64("last_analysis_stats." + field)
		if err != nil {
			continue // if there's any missing fields, skip it
		}
		if field == "malicious" {
			malicious = int(fieldValue)
		}
		total += int(fieldValue)
	}

	return malicious, total, nil

}
