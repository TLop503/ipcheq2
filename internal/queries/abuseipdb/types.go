package abuseipdb

import (
	"os"
	"strings"
)

var abIPDBKey string

// InitializeAPIKey sets the AbuseIPDB API key from environment
func InitializeAPIKey() {
	InitializeAPIKeyFromValue(os.Getenv("ABIPDBKEY"))
	if abIPDBKey == "" {
		panic("ABIPDBKEY environment variable is not set")
	}
}

// InitializeAPIKeyFromValue sets the AbuseIPDB API key from an explicit value.
func InitializeAPIKeyFromValue(key string) {
	abIPDBKey = strings.TrimSpace(key)
	if abIPDBKey == "" {
		panic("ABIPDB key is not set (keys file and ABIPDBKEY environment variable are both empty)")
	}
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
