package queries

import (
	"log"

	"github.com/joho/godotenv"
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

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables directly")
	}
	// Initialize API keys in internal package -- at minimum, abuseIPDB key is required
	abuseipdb.InitializeAPIKey()
	virustotal.InitializeVTAPIKey()
	// Initialize VPN ID ranger
	vpnid.InitializeVpnID()
}
