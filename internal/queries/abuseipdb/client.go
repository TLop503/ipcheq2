package abuseipdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"time"
)

// QueryAbuseIPDB returns a struct of the full API response
func QueryAbuseIPDB(ip netip.Addr) (ABIPDBResponse, error) {
	// fail non-destructively incase key intentionally not supplied
	// TODO: check for keys upstream!
	if abIPDBKey == "" {
		return ABIPDBResponse{}, fmt.Errorf("No ABIPDB API Key Supplied!")
	}

	resp, err := queryHelper(ip.String())
	if err != nil {
		return ABIPDBResponse{}, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ABIPDBResponse{}, fmt.Errorf("error reading response body for IP %s: %v", ip, err)
	}

	var result ABIPDBResponse
	var payload struct {
		Data ABIPDBResponse `json:"data"`
	}
	if err = json.Unmarshal(body, &payload); err != nil {
		return ABIPDBResponse{}, fmt.Errorf("failed to parse JSON for ABIPDB %s: %v", ip, err)
	}
	result = payload.Data

	return result, nil
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
