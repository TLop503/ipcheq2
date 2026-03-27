package abuseipdb

import (
	"encoding/json"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

// TestQueryAbuseIPDB_PrintRawAndStruct is an opt-in integration test.
// It calls the live API, then logs both the raw JSON payload and parsed struct.
func TestQueryAbuseIPDB_PrintRawAndStruct(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping AbuseIPDB integration test in short mode")
	}

	key := envOrDotenv("ABIPDBKEY")
	if key == "" {
		t.Skip("ABIPDBKEY not set in environment or .env; skipping live AbuseIPDB integration test")
	}

	targetIP := envOrDotenv("ABIPDB_TEST_IP")
	if targetIP == "" {
		targetIP = "8.8.8.8"
	}

	ip, err := netip.ParseAddr(targetIP)
	if err != nil {
		t.Fatalf("invalid ABIPDB_TEST_IP %q: %v", targetIP, err)
	}

	abIPDBKey = key

	resp, err := queryHelper(ip.String())
	if err != nil {
		t.Fatalf("queryHelper failed for %s: %v", ip, err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed reading raw AbuseIPDB response body: %v", err)
	}
	t.Logf("raw AbuseIPDB response for %s:\n%s", ip, string(rawBody))

	parsed, err := QueryAbuseIPDB(ip)
	if err != nil {
		t.Fatalf("QueryAbuseIPDB failed for %s: %v", ip, err)
	}

	pretty, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal parsed struct for logging: %v", err)
	}
	t.Logf("parsed ABIPDBResponse struct for %s:\n%s", ip, string(pretty))

	if parsed.CountryCode == "" && parsed.CountryName == "" && parsed.Domain == "" && len(parsed.Hostnames) == 0 {
		t.Fatalf("parsed ABIPDBResponse appears empty; expected populated fields from API")
	}
}

func envOrDotenv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}

	loadDotenvFromCWDOrParents()
	return strings.TrimSpace(os.Getenv(key))
}

func loadDotenvFromCWDOrParents() {
	wd, err := os.Getwd()
	if err != nil {
		return
	}

	current := wd
	for {
		dotenvPath := filepath.Join(current, ".env")
		if _, statErr := os.Stat(dotenvPath); statErr == nil {
			_ = godotenv.Load(dotenvPath)
			return
		}

		parent := filepath.Dir(current)
		if parent == current {
			return
		}
		current = parent
	}
}
