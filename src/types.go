package src

import (
	"net"
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
	ThreatRisk      string

	//spur
	ParsedRes string // is / is not vpn
}

var Results []Result
