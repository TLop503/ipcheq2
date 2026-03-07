package virustotal

import (
	"fmt"
	"net/netip"
	"os"

	vt "github.com/VirusTotal/vt-go"
)

// per "https://github.com/VirusTotal/vt-go", vt-go is the official Go client library for VirusTotal.
// With this library you can interact with the VirusTotal REST API v3 without having to send plain HTTP requests with the standard "http" package.

var VTKey string

// InitializeAPIKey sets the VirusTotal API key from environment
func InitializeVTAPIKey() {
	VTKey = os.Getenv("VTKEY")
	if VTKey == "" {
		panic("VTKEY environment variable is not set")
	}
}

// Queries VirusTotal for an IP and returns "malicious_count/total_engines"
func VTQuery(ip netip.Addr) (string, error) {

	client := vt.NewClient(VTKey)

	result, err := client.GetObject(vt.URL("ip_addresses/%s", ip.String()))
	if err != nil {
		return "", err
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

	return fmt.Sprintf("%d/%d", malicious, total), nil
}
