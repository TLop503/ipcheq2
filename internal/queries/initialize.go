package queries

import (
	"log"
	"os"
	"strings"

	"github.com/tlop503/ipcheq2/v2/internal/config"
	"github.com/tlop503/ipcheq2/v2/internal/data"
	"github.com/tlop503/ipcheq2/v2/internal/queries/abuseipdb"
	"github.com/tlop503/ipcheq2/v2/internal/queries/virustotal"
	"github.com/tlop503/ipcheq2/v2/internal/queries/vpnid"
)

// InitConnectors reads API keys for upstream services and calls vpnid's init
func InitConnectors() {
	if _, err := data.EnsureDataDir(); err != nil {
		log.Panicf("Error hydrating local data to cache: %v", err)
	}

	if _, err := data.EnsureHashDir(); err != nil {
		log.Panicf("Error creating hash dir in cache: %v", err)
	}

	if _, err := os.Stat(".env"); err == nil {
		log.Println("Warning: .env loading is deprecated and ignored. Use user config keys file or environment variables instead.")
	}

	// load from env
	abuseIPDBKey := strings.TrimSpace(os.Getenv("ABIPDBKEY"))
	vtKey := strings.TrimSpace(os.Getenv("VTKEY"))

	// and from disk
	disk, err := config.LoadKeys()
	if err != nil {
		log.Println("Error loading keys, falling back to env vars: ", err)
		if err == os.ErrNotExist {
			path, created, err := config.EnsureKeysFile()
			if err != nil {
				log.Fatalf("Unable to initialize keys file: %v", err)
			} else if created {
				log.Printf("Created blank keys file at %s", path)
			}
		}
	}

	// prioritize env variables over config
	if abuseIPDBKey == "" {
		abuseIPDBKey = disk.ABIPDBKey
	}
	if vtKey == "" {
		vtKey = disk.VTKey
	}

	// Initialize API keys in internal package -- at minimum, abuseIPDB key is required
	if err := abuseipdb.InitializeAPIKeyFromValue(abuseIPDBKey); err != nil {
		log.Printf("Warning: AbuseIPDBKey not loaded. InitalizeAPIKeyFromValue: %v", err)
	}
	virustotal.InitializeVTAPIKeyFromValue(vtKey)
	// Initialize VPN ID ranger
	vpnid.InitializeVpnID()
}
