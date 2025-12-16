package abuseipdb

import (
	"html/template"
	"net/netip"
	"os"
	"time"
)

var AbIPDBKey string
var Results []Result

type Result struct {
	IP netip.Addr
	// abuseipdb data
	IsPub           bool
	AbuseConfidence int
	Country         string
	CountryCode     string
	UsageType       string
	ISP             string
	Domain          string
	TotalReports    int
	Users           int
	LastReported    time.Time
	ThreatRisk      template.HTML

	//vpn status
	ParsedRes string // vpn provider or "not vpn"
}

// InitializeAPIKey sets the AbuseIPDB API key from environment
func InitializeAPIKey() {
	AbIPDBKey = os.Getenv("ABIPDBKEY")
	if AbIPDBKey == "" {
		panic("ABIPDBKEY environment variable is not set")
	}
}

type AbuseIPDBResponse struct {
	Data struct {
		IPAddress            string `json:"ipAddress"`
		IsPublic             bool   `json:"isPublic"`
		AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
		CountryName          string `json:"countryName"`
		CountryCode          string `json:"countryCode"`
		UsageType            string `json:"usageType"`
		ISP                  string `json:"isp"`
		Domain               string `json:"domain"`
		TotalReports         int    `json:"totalReports"`
		NumDistinctUsers     int    `json:"numDistinctUsers"`
		LastReportedAt       string `json:"lastReportedAt"`
	} `json:"data"`
}
