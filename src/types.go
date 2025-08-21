package src

import (
	"fmt"
	"html/template"
	"net/netip"
	"os"
	"time"

	"ipcheq2/src/vpnid"

	"github.com/yl2chen/cidranger"
)

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

var Results []Result
var AbIPDBKey string
var VpnIDRanger cidranger.Ranger

// InitializeAPIKey sets the AbuseIPDB API key from environment
func InitializeAPIKey() {
	AbIPDBKey = os.Getenv("ABIPDBKEY")
	if AbIPDBKey == "" {
		panic("ABIPDBKEY environment variable is not set")
	}
}

// InitializeVpnID initializes the VPN identification ranger from config file
func InitializeVpnID() {
	ranger, err := vpnid.Initialize("vpnid_config.txt")
	if err != nil {
		panic("Failed to initialize VPN ID: " + err.Error())
	}
	VpnIDRanger = ranger
	fmt.Println("VPN ID ranger initialized successfully")
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
