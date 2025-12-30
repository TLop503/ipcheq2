package abuseipdb

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/netip"
	"time"
)

// Fetch and parse the AbuseIPDB result
func CheckAbuseIPDB(ip netip.Addr) (Result, error) {
	resp, err := queryHelper(ip.String())
	if err != nil {
		return Result{}, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("error reading response body for IP %s: %v", ip, err)
	}

	// Parse response into intermediate struct
	var raw abuseIPDBResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return Result{}, fmt.Errorf("failed to parse JSON for IP %s: %v", ip, err)
	}

	// Populate Result
	return Result{
		IP:              ip,
		IsPub:           raw.Data.IsPublic,
		AbuseConfidence: raw.Data.AbuseConfidenceScore,
		Country:         raw.Data.CountryName,
		CountryCode:     raw.Data.CountryCode,
		UsageType:       raw.Data.UsageType,
		ISP:             raw.Data.ISP,
		Domain:          raw.Data.Domain,
		TotalReports:    raw.Data.TotalReports,
		Users:           raw.Data.NumDistinctUsers,
		LastReported:    timeHelper(raw.Data.LastReportedAt),
		ThreatRisk:      confidenceHelper(raw.Data.AbuseConfidenceScore), // Optional logic can go here
		AbuseLinks: 	 raw.Data.AbuseConfidenceScore > 0, // whether or not to show ABIPDP and OTX links
		ParsedRes:       "Not Anonymous",
	}, nil
}

// confidenceHelper is used to map images to confidence levels in the UI
func confidenceHelper(c int) template.HTML {
	switch {
	case c == 0:
		return `<span style="padding:1px 2px; border-radius:2px;">Clean</span> <img style="vertical-align: middle;" src="/assets/icons8-checkmark-50.png" alt="Green Checkmark" width="20" height="20"/>` // test with 212.102.51.57
	case c < 26:
		return `<span style="padding:1px 2px; border-radius:2px;">Low Risk</span> <img style="vertical-align: middle;" src="/assets/icons8-yield-sign-50.png" alt="Yellow Yield" width="17" height="17"/>` // test with 54.204.34.130
	case c < 51:
		return `<span style="padding:1px 2px; border-radius:2px;">Medium Risk</span> <img style="vertical-align: middle;" src="/assets/icons8-caution-50.png" alt="Orange Caution" width="18" height="18"/>` // test with 209.85.221.176
	}
	return `<span style="padding:1px 2px; border-radius:2px;">High Risk</span> <img style="vertical-align: middle;" src="/assets/icons8-unavailable-48.png" alt="Red Unavailable" width="20" height="20"/>` // test with 111.26.184.29
}

// timeHelper formats time in a human-readable format, or returns blank conversion fails
func timeHelper(t string) time.Time {
	prettyTime, err := time.Parse(time.RFC3339, t)
	if err != nil {
		// return dummy if fail
		return time.Time{}
	}
	return prettyTime
}

// queryHelper queries abuseipdb and returns errors if invalid data is received or query fails
func queryHelper(ip string) (*http.Response, error) {
	url := "https://api.abuseipdb.com/api/v2/check"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add query parameters
	query := req.URL.Query()
	query.Add("ipAddress", ip)
	query.Add("maxAgeInDays", "30")
	query.Add("verbose", "")
	req.URL.RawQuery = query.Encode()

	// Add headers
	req.Header.Set("Key", abIPDBKey)
	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error for IP %s: %v", ip, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status code for IP %s: %s", ip, resp.Status)
	}

	return resp, nil
}
