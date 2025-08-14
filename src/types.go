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
	CountryCode     string
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
var SpurCookie string

// InitializeAPIKey sets the AbuseIPDB API key from environment
func InitializeAPIKey() {
	AbIPDBKey = os.Getenv("ABIPDBKEY")
	if AbIPDBKey == "" {
		panic("ABIPDBKEY environment variable is not set")
	}
}

// InitializeSpurCookie sets the session cookie used for spur.us requests.
// This value is mandatory and must be provided via environment variable
// SPUR_SESSION_COOKIE. The value should be the full cookie header value
// you want sent (for example: "session=xxxxx" or "session=xxx; other=y").
func InitializeSpurCookie() {
	SpurCookie = os.Getenv("SPUR_SESSION_COOKIE")
	if SpurCookie == "" {
		panic("SPUR_SESSION_COOKIE environment variable is not set; spur session cookie is mandatory")
	}
}

type abuseIPDBResponse struct {
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
