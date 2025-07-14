package src

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"time"
)

// Fetch and parse the AbuseIPDB result
func checkAbuseIPDB(ip string) (Result, error) {
	url := "https://api.abuseipdb.com/api/v2/check"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return Result{}, fmt.Errorf("failed to create request: %v", err)
	}

	// Add query parameters
	query := req.URL.Query()
	query.Add("ipAddress", ip)
	query.Add("maxAgeInDays", "30")
	query.Add("verbose", "")
	req.URL.RawQuery = query.Encode()

	// Add headers
	req.Header.Set("Key", AbIPDBKey)
	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("request error for IP %s: %v", ip, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("non-200 status code for IP %s: %s", ip, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("error reading response body for IP %s: %v", ip, err)
	}

	// Parse response into intermediate struct
	var raw abuseIPDBResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return Result{}, fmt.Errorf("failed to parse JSON for IP %s: %v", ip, err)
	}

	// Convert timestamp
	var lastReported time.Time
	if raw.Data.LastReportedAt != "" {
		lastReported, err = time.Parse(time.RFC3339, raw.Data.LastReportedAt)
		if err != nil {
			return Result{}, fmt.Errorf("invalid time format: %v", err)
		}
	}

	confidence := raw.Data.AbuseConfidenceScore
	var risk template.HTML
	if confidence == 0 {
		risk = template.HTML(`<span style="background-color:lightgreen; padding:1px 2px; border-radius:2px;">Clean</span>`)
	} else if confidence < 26 {
		risk = template.HTML(`<span style="background-color:yellow; padding:1px 2px; border-radius:2px;">Low Risk</span>`)
	} else if confidence < 51 {
		risk = template.HTML(`<span style="background-color:orange; padding:1px 2px; border-radius:2px;">Medium Risk</span>`)
	} else {
		risk = template.HTML(`<span style="background-color:red; padding:1px 2px; border-radius:2px;">High Risk</span>`)
	}

	// Populate Result
	return Result{
		IP:              net.ParseIP(raw.Data.IPAddress),
		IsPub:           raw.Data.IsPublic,
		AbuseConfidence: raw.Data.AbuseConfidenceScore,
		Country:         raw.Data.CountryName,
		CountryCode:     raw.Data.CountryCode,
		UsageType:       raw.Data.UsageType,
		ISP:             raw.Data.ISP,
		Domain:          raw.Data.Domain,
		TotalReports:    raw.Data.TotalReports,
		Users:           raw.Data.NumDistinctUsers,
		LastReported:    lastReported,
		ThreatRisk:      risk, // Optional logic can go here
		ParsedRes:       "Not Anonymous",
	}, nil
}
