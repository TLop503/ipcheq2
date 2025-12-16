package vpnid

import (
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestQueryIPs(t *testing.T) {
	tmp := t.TempDir()

	// Example IPs for each provider
	files := map[string]string{
		"cyberghost.txt": "1.1.1.1\n2.2.2.0/24\n",
		"express.txt":    "3.3.3.3\n",
		"nord.txt":       "4.4.4.4/30\n",
	}

	// Create data files
	for fname, content := range files {
		makeTempFile(t, tmp, fname, content)
	}

	// Create config file
	configContent := ""
	for fname := range files {
		configContent += fmt.Sprintf("%s : %s\n", fname, filepath.Join(tmp, fname))
	}
	configPath := makeTempFile(t, tmp, "config.txt", configContent)

	// Initialize ranger
	ranger, err := initialize(configPath)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Test cases: IP → expected provider(s)
	tests := []struct {
		ip       string
		expected string
	}{
		{"1.1.1.1", "cyberghost.txt"},
		{"2.2.2.5", "cyberghost.txt"}, // within 2.2.2.0/24
		{"3.3.3.3", "express.txt"},
		{"4.4.4.5", "nord.txt"}, // within 4.4.4.4/30 (range is 4.4.4.4-4.4.4.7)
		{"5.5.5.5", ""},         // not in any file
	}

	for _, tt := range tests {
		addr, err := netip.ParseAddr(tt.ip)
		if err != nil {
			t.Errorf("ParseAddr(%s) failed: %v", tt.ip, err)
			continue
		}
		result, err := Query(addr, ranger)
		if err != nil {
			t.Errorf("Query(%s) returned error: %v", tt.ip, err)
			continue
		}

		if tt.expected != "" && !containsProvider(result, tt.expected) {
			t.Errorf("Query(%s) = %q; want contains %q", tt.ip, result, tt.expected)
		} else if tt.expected == "" && result != fmt.Sprintf("%s not found in dataset", addr) {
			t.Errorf("Query(%s) = %q; want not found message", tt.ip, result)
		}
	}
}

func TestWithProdData(t *testing.T) {
	dataDir := filepath.Join("..", "..", "data")

	providers := []string{
		"cyberghost", "express", "mullvad", "nord",
		"pia", "proton", "surfshark", "torguard", "tunnelbear",
	}

	configPath := filepath.Join(t.TempDir(), "vpnid_config.txt")
	var config string
	for _, p := range providers {
		path := filepath.Join(dataDir, p+".txt")
		config += p + " : " + path + "\n"
	}

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	start := time.Now() // ⏱ start timing
	ranger, err := initialize(configPath)
	duration := time.Since(start)
	t.Logf("Init took %s", duration.String())

	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	tests := []struct {
		ip       string
		expected string
	}{
		{"191.101.210.110", "cyberghost"},
		{"191.101.61.121", "pia"},
		{"146.70.49.90", "torguard"},
		{"103.214.20.130", "mullvad"},
		{"50.118.143.28", "express"},
	}

	for _, tt := range tests {
		addr, err := netip.ParseAddr(tt.ip)
		if err != nil {
			t.Errorf("ParseAddr(%s) failed: %v", tt.ip, err)
			continue
		}

		start := time.Now() // ⏱ start timing
		result, err := Query(addr, ranger)
		duration := time.Since(start)

		t.Logf("Query(%s) took %d ns", tt.ip, duration.Nanoseconds())

		if err != nil {
			t.Errorf("Query(%s) returned error: %v", tt.ip, err)
			continue
		}

		if tt.expected != "" && !containsProvider(result, tt.expected) {
			t.Errorf("Query(%s) = %q; want contains %q", tt.ip, result, tt.expected)
		} else if tt.expected == "" && result != fmt.Sprintf("%s not found in dataset", addr) {
			t.Errorf("Query(%s) = %q; want not found message", tt.ip, result)
		}
	}
}

// Helper to check if result string contains the provider name
func containsProvider(result, provider string) bool {
	return strings.Contains(result, provider)
}
