package src

import (
	"html/template"
	"net"
	"os"
	"time"
)

type Result struct {
	IP net.IP
	// abuseipdb data
	IsPub           bool
	AbuseConfidence int
	Country         string
	UsageType       string
	ISP             string
	Domain          string
	TotalReports    int
	Users           int
	LastReported    time.Time
	ThreatRisk      template.HTML

	//spur
	ParsedRes string // is / is not vpn
}

var Results []Result
var AbIPDBKey string

// InitializeAPIKey sets the AbuseIPDB API key from environment
func InitializeAPIKey() {
	AbIPDBKey = os.Getenv("ABIPDBKEY")
	if AbIPDBKey == "" {
		panic("ABIPDBKEY environment variable is not set")
	}
}

type abuseIPDBResponse struct {
	Data struct {
		IPAddress            string `json:"ipAddress"`
		IsPublic             bool   `json:"isPublic"`
		AbuseConfidenceScore int    `json:"abuseConfidenceScore"`
		CountryCode          string `json:"countryCode"`
		UsageType            string `json:"usageType"`
		ISP                  string `json:"isp"`
		Domain               string `json:"domain"`
		TotalReports         int    `json:"totalReports"`
		NumDistinctUsers     int    `json:"numDistinctUsers"`
		LastReportedAt       string `json:"lastReportedAt"`
	} `json:"data"`
}
