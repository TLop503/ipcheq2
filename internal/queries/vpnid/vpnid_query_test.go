package vpnid

import (
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/tlop503/ipcheq2/internal/config"
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

	var sources []config.Source
	for fname := range files {
		sources = append(sources, config.Source{Name: fname, Path: filepath.Join(tmp, fname)})
	}
	writeInitConfig(t, config.Config{Sources: sources})

	// Initialize ranger
	ranger, err := initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	prev := VpnIDRanger
	VpnIDRanger = ranger
	t.Cleanup(func() {
		VpnIDRanger = prev
	})

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
		result, err := QueryToSlice(addr)
		if err != nil {
			t.Errorf("Query(%s) returned error: %v", tt.ip, err)
			continue
		}

		if tt.expected != "" {
			if !slices.Contains(result, tt.expected) {
				t.Errorf("Query(%s) = %q; want contains %q", tt.ip, result, tt.expected)
			}
		} else {
			if len(result) != 0 {
				t.Errorf("Query(%s) = %q; Should be an empty slice", tt.ip, result)
			}
		}
	}
}

func TestWithProdData(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	dataDir := filepath.Join(repoRoot, "data")

	providers := []string{
		"cyberghost", "express", "mullvad", "nord",
		"pia", "proton", "surfshark", "torguard", "tunnelbear",
	}

	var sources []config.Source
	for _, p := range providers {
		path := filepath.Join(dataDir, p+".txt")
		sources = append(sources, config.Source{Name: p, Path: path})
	}
	writeInitConfig(t, config.Config{Sources: sources})

	start := time.Now() // ⏱ start timing
	ranger, err := initialize()
	duration := time.Since(start)
	t.Logf("Init took %s", duration.String())

	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	prev := VpnIDRanger
	VpnIDRanger = ranger
	t.Cleanup(func() {
		VpnIDRanger = prev
	})

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
		result, err := QueryToSlice(addr)
		duration := time.Since(start)

		t.Logf("Query(%s) took %d ns", tt.ip, duration.Nanoseconds())

		if err != nil {
			t.Errorf("Query(%s) returned error: %v", tt.ip, err)
			continue
		}

		if tt.expected != "" {
			if !slices.Contains(result, tt.expected) {
				t.Errorf("Query(%s) = %q; want contains %q", tt.ip, result, tt.expected)
			}
		} else {
			if len(result) != 0 {
				t.Errorf("Query(%s) = %q; Should be an empty slice", tt.ip, result)
			}
		}
	}
}

func TestQueryWithoutInitialize(t *testing.T) {
	prev := VpnIDRanger
	VpnIDRanger = nil
	t.Cleanup(func() {
		VpnIDRanger = prev
	})

	addr, err := netip.ParseAddr("1.1.1.1")
	if err != nil {
		t.Fatalf("ParseAddr failed: %v", err)
	}

	_, err = QueryToSlice(addr)
	if err == nil {
		t.Fatalf("expected error when querying before initialization, got nil")
	}

	if err.Error() != "VPNIDRanger not initialized" {
		t.Fatalf("unexpected error: %v", err)
	}
}
