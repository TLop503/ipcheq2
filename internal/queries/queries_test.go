package queries

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"net/netip"

	"github.com/tlop503/ipcheq2/v2/internal/data"
	"github.com/tlop503/ipcheq2/v2/internal/queries/abuseipdb"
	"github.com/tlop503/ipcheq2/v2/internal/queries/virustotal"
	"github.com/tlop503/ipcheq2/v2/internal/queries/vpnid"
)

func TestFirstPartyQuery_WhenRangerNotInitialized(t *testing.T) {
	prevRanger := vpnid.VpnIDRanger
	vpnid.VpnIDRanger = nil
	t.Cleanup(func() {
		vpnid.VpnIDRanger = prevRanger
	})

	addr := netip.MustParseAddr("1.1.1.1")
	_, err := FirstPartyQuery(addr)
	if err == nil {
		t.Fatalf("FirstPartyQuery failed to return expected error")
	}
	if err.Error() != "VPNIDRanger not initialized" {
		t.Fatalf("FirstPartyQuery failed to return expected error, instead got %s", err.Error())
	}
}

func TestFirstPartyQuery_WithInitializedRanger(t *testing.T) {
	prevRanger := vpnid.VpnIDRanger
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	// InitializeVpnID expects vpnid_config.txt relative to repo root.
	repoRoot := filepath.Clean(filepath.Join(prevWD, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to chdir to repo root %q: %v", repoRoot, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
		vpnid.VpnIDRanger = prevRanger
	})

	if _, err := data.EnsureDataDir(); err != nil {
		t.Fatalf("failed to hydrate cache data for test: %v", err)
	}

	vpnid.InitializeVpnID()

	addr := netip.MustParseAddr("191.101.210.110")
	body, err := FirstPartyQuery(addr)
	if err != nil {
		t.Fatalf("FirstPartyQuery returned unexpected error: %v", err)
	}

	var got FirstPartyResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("failed to unmarshal FirstPartyQuery response: %v", err)
	}

	if got.IPAddress != addr.String() {
		t.Fatalf("IPAddress = %q, want %q", got.IPAddress, addr.String())
	}

	if len(got.VPNIDMatches) == 0 {
		t.Fatalf("VPNIDMatches is empty, expected at least one provider")
	}

	if !contains(got.VPNIDMatches, "cyberghost") {
		t.Fatalf("VPNIDMatches = %v, want to contain %q", got.VPNIDMatches, "cyberghost")
	}
}

func TestThirdPartyQuery_NoExternalKeys(t *testing.T) {
	prevVTKeyPresent := virustotal.VTKeyPresent
	virustotal.VTKeyPresent = false
	t.Cleanup(func() {
		virustotal.VTKeyPresent = prevVTKeyPresent
	})

	addr := netip.MustParseAddr("8.8.8.8")
	body, err := ThirdPartyQuery(addr)
	if err != nil {
		t.Fatalf("ThirdPartyQuery returned unexpected error: %v", err)
	}

	var got ThirdPartyResponse
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("failed to unmarshal ThirdPartyQuery response: %v", err)
	}

	if got.IPAddress != addr.String() {
		t.Fatalf("IPAddress = %q, want %q", got.IPAddress, addr.String())
	}

	if !reflect.DeepEqual(got.ABIPDBResponse, abuseipdb.ABIPDBResponse{}) {
		t.Fatalf("ABIPDBResponse = %#v, want zero-value response when key is absent", got.ABIPDBResponse)
	}

	if !reflect.DeepEqual(got.VirusTotalResponse, virustotal.VirusTotalResponse{}) {
		t.Fatalf("VirusTotalResponse = %#v, want zero-value response when key is absent", got.VirusTotalResponse)
	}
}

func TestThirdPartyQuery_JSONKeys(t *testing.T) {
	prevVTKeyPresent := virustotal.VTKeyPresent
	virustotal.VTKeyPresent = false
	t.Cleanup(func() {
		virustotal.VTKeyPresent = prevVTKeyPresent
	})

	addr := netip.MustParseAddr("9.9.9.9")
	body, err := ThirdPartyQuery(addr)
	if err != nil {
		t.Fatalf("ThirdPartyQuery returned unexpected error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("failed to unmarshal response as generic JSON object: %v", err)
	}

	for _, key := range []string{"ipAddress", "ABIPDBResponse", "VirusTotalResponse"} {
		if _, ok := got[key]; !ok {
			t.Fatalf("expected JSON key %q to be present, got keys: %v", key, mapKeys(got))
		}
	}
}

func TestThirdPartyQuery_AbuseIPDBKeyPresent_NoVTKey(t *testing.T) {
	const helperEnv = "IPCHEQ2_THIRDPARTY_HELPER"

	if os.Getenv(helperEnv) == "1" {
		if err := abuseipdb.InitializeAPIKey(); err != nil {
			t.Fatalf("InitializeAPIKey returned unexpected error: %v", err)
		}
		virustotal.VTKeyPresent = false

		addr := netip.MustParseAddr("8.8.4.4")
		body, err := ThirdPartyQuery(addr)
		if err != nil {
			t.Fatalf("ThirdPartyQuery returned unexpected error: %v", err)
		}

		var got ThirdPartyResponse
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("failed to unmarshal ThirdPartyQuery response: %v", err)
		}

		if got.IPAddress != addr.String() {
			t.Fatalf("IPAddress = %q, want %q", got.IPAddress, addr.String())
		}

		if !reflect.DeepEqual(got.VirusTotalResponse, virustotal.VirusTotalResponse{}) {
			t.Fatalf("VirusTotalResponse = %#v, want zero-value response when VT key is absent", got.VirusTotalResponse)
		}

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run", "^TestThirdPartyQuery_AbuseIPDBKeyPresent_NoVTKey$")
	cmd.Env = append(os.Environ(), helperEnv+"=1", "ABIPDBKEY=dummy-test-key")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helper process failed: %v\noutput:\n%s", err, string(output))
	}
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if strings.EqualFold(item, want) {
			return true
		}
	}
	return false
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
