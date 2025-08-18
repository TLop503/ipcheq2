package vpnid

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"testing"
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
	ranger, err := Initialize(configPath)
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
		{"4.4.4.2", "nord.txt"}, // within 4.4.4.4/30
		{"5.5.5.5", ""},         // not in any file
	}

	for _, tt := range tests {
		addr := net.IPAddr{IP: net.ParseIP(tt.ip)}
		result, err := Query(addr, ranger)
		if err != nil {
			t.Errorf("Query(%s) returned error: %v", tt.ip, err)
			continue
		}

		if tt.expected != "" && !containsProvider(result, tt.expected) {
			t.Errorf("Query(%s) = %q; want contains %q", tt.ip, result, tt.expected)
		} else if tt.expected == "" && result != fmt.Sprintf("%s not found in ranger", addr.IP) {
			t.Errorf("Query(%s) = %q; want not found message", tt.ip, result)
		}
	}
}

// Helper to check if result string contains the provider name
func containsProvider(result, provider string) bool {
	return strings.Contains(result, provider)
}
