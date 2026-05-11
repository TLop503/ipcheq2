package abuseipdb

import (
	"fmt"
	"log"
	"strings"
)

var ABIPKeyPresent = false
var abIPDBKey string

// InitializeAPIKeyFromValue sets the AbuseIPDB API key from an explicit value.
func InitializeAPIKeyFromValue(key string) error {
	ABIPKeyPresent = false
	abIPDBKey = strings.TrimSpace(key)
	if abIPDBKey == "" {
		return fmt.Errorf("ABIPDB key is not set (keys file and ABIPDBKEY environment variable are both empty)")
	} else {
		ABIPKeyPresent = true
		log.Println("ABIPDB Key loaded")
	}

	return nil
}

// API-specific data struct.
type ABIPDBResponse struct {
	IsPublic             bool     `json:"isPublic"`
	IsWhitelisted        bool     `json:"isWhitelisted"`
	AbuseConfidenceScore int      `json:"abuseConfidenceScore"`
	CountryCode          string   `json:"countryCode"`
	CountryName          string   `json:"countryName"`
	UsageType            string   `json:"usageType"`
	ISP                  string   `json:"isp"`
	Domain               string   `json:"domain"`
	Hostnames            []string `json:"hostnames"`
	IsTor                bool     `json:"isTor"`
	TotalReports         int      `json:"totalReports"`
	NumDistinctUsers     int      `json:"numDistinctUsers"`
	LastReportedAt       string   `json:"lastReportedAt"`
}
