package queries

import (
	"log"
	"os"
	"strings"

	"github.com/tlop503/ipcheq2/internal/config"
	"github.com/tlop503/ipcheq2/internal/data"
	"github.com/tlop503/ipcheq2/internal/queries/abuseipdb"
	"github.com/tlop503/ipcheq2/internal/queries/virustotal"
	"github.com/tlop503/ipcheq2/internal/queries/vpnid"
)

// InitConnectors reads API keys for upstream services and calls vpnid's init
func InitConnectors() {
	if _, err := data.EnsureDataDir(); err != nil {
		log.Panicf("Error hydrating local data to cache: %v", err)
	}

	if _, err := os.Stat(".env"); err == nil {
		log.Println("Warning: .env loading is deprecated and ignored. Use user config keys file or environment variables instead.")
	}

	abuseIPDBKey := strings.TrimSpace(os.Getenv("ABIPDBKEY"))
	vtKey := strings.TrimSpace(os.Getenv("VTKEY"))

	if abuseIPDBKey == "" && vtKey == "" {
		path, created, err := config.EnsureKeysFile()
		if err != nil {
			log.Fatalf("Unable to initialize keys file: %v", err)
		}
		if created {
			log.Printf("Created blank keys file at %s", path)
		}
	}

	keys, err := config.LoadKeys()
	if err != nil {
		log.Fatalf("Unable to load API keys configuration: %v", err)
	}

	if abuseIPDBKey == "" {
		abuseIPDBKey = keys.ABIPDBKey
	}

	if vtKey == "" {
		vtKey = keys.VTKey
	}

	// Initialize API keys in internal package -- at minimum, abuseIPDB key is required
	if err := abuseipdb.InitializeAPIKeyFromValue(abuseIPDBKey); err != nil {
		log.Fatalf("Missing required AbuseIPDB key. Set ABIPDBKEY env var or abipdbKey in keys.yaml: %v", err)
	}
	virustotal.InitializeVTAPIKeyFromValue(vtKey)
	// Initialize VPN ID ranger
	vpnid.InitializeVpnID()
}
