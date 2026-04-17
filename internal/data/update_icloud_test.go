package data

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollapsePrefixesIPv4(t *testing.T) {
	prefixes, err := collapsePrefixes(parsePrefixes("10.0.0.0/31", "10.0.0.2/31"))
	if err != nil {
		t.Fatalf("collapsePrefixes returned error: %v", err)
	}

	got := prefixesToStrings(prefixes)
	want := []string{"10.0.0.0/30"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("collapsePrefixes() = %v, want %v", got, want)
	}
}

func TestCollapsePrefixesIPv6(t *testing.T) {
	prefixes, err := collapsePrefixes(parsePrefixes("2001:db8::/127", "2001:db8::2/127"))
	if err != nil {
		t.Fatalf("collapsePrefixes returned error: %v", err)
	}

	got := prefixesToStrings(prefixes)
	want := []string{"2001:db8::/126"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("collapsePrefixes() = %v, want %v", got, want)
	}
}

func TestUpdateICloudRelays(t *testing.T) {
	responseCSV := strings.Join([]string{
		"10.0.0.0/31,US,CA,San Francisco,",
		"10.0.0.2/31,US,CA,San Francisco,",
		"2001:db8::/127,US,CA,San Francisco,",
		"2001:db8::2/127,US,CA,San Francisco,",
	}, "\n") + "\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, responseCSV)
	}))
	defer server.Close()

	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, iCloudHashFile), []byte("old-hash"), 0o644); err != nil {
		t.Fatalf("seed hash file: %v", err)
	}

	changed, err := updateICloudRelays(dataDir, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("updateICloudRelays returned error: %v", err)
	}
	if !changed {
		t.Fatal("updateICloudRelays returned false, want true")
	}

	ipv4Path := filepath.Join(dataDir, iCloudIPv4File)
	ipv6Path := filepath.Join(dataDir, iCloudIPv6File)
	ipv4Contents, err := os.ReadFile(ipv4Path)
	if err != nil {
		t.Fatalf("read ipv4 file: %v", err)
	}
	ipv6Contents, err := os.ReadFile(ipv6Path)
	if err != nil {
		t.Fatalf("read ipv6 file: %v", err)
	}

	if got, want := string(ipv4Contents), "10.0.0.0/30\n"; got != want {
		t.Fatalf("ipv4 file = %q, want %q", got, want)
	}
	if got, want := string(ipv6Contents), "2001:db8::/126\n"; got != want {
		t.Fatalf("ipv6 file = %q, want %q", got, want)
	}

	hashContents, err := os.ReadFile(filepath.Join(dataDir, iCloudHashFile))
	if err != nil {
		t.Fatalf("read hash file: %v", err)
	}
	wantHash := fmt.Sprintf("%x", sha256.Sum256([]byte(responseCSV)))
	if got := strings.TrimSpace(string(hashContents)); got != wantHash {
		t.Fatalf("hash file = %q, want %q", got, wantHash)
	}

	changed, err = updateICloudRelays(dataDir, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("second updateICloudRelays returned error: %v", err)
	}
	if changed {
		t.Fatal("updateICloudRelays returned true on unchanged data, want false")
	}
}

func parsePrefixes(values ...string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			panic(err)
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}

func prefixesToStrings(prefixes []netip.Prefix) []string {
	values := make([]string, 0, len(prefixes))
	for _, prefix := range prefixes {
		values = append(values, prefix.String())
	}
	return values
}
