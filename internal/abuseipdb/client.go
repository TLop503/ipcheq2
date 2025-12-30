package abuseipdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"strings"
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
		CountryCode:     strings.ToLower(raw.Data.CountryCode),
		UsageType:       raw.Data.UsageType,
		ISP:             raw.Data.ISP,
		Domain:          raw.Data.Domain,
		TotalReports:    raw.Data.TotalReports,
		Users:           raw.Data.NumDistinctUsers,
		LastReported:    timeHelper(raw.Data.LastReportedAt),
		AbuseLinks:      raw.Data.AbuseConfidenceScore > 0, // whether or not to show ABIPDP and OTX links
		// while AbuseLinks is redundant to manually doing this check in the template, it is being used
		// for future-proofing incase more logic is used for displaying links, such as
		// a config or a list of providers.
		ParsedRes: "Not Anonymous",
	}, nil
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