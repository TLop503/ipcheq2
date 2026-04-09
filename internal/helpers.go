package internal

import (
	"bytes"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/tlop503/ipcheq2/internal/queries/abuseipdb"
	"github.com/tlop503/ipcheq2/internal/queries/virustotal"
	"github.com/tlop503/ipcheq2/internal/queries/vpnid"
	"log"
)

func PrettyJSON(b []byte) ([]byte, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "  "); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// SharedStartup includes behavior used by multiple binaries at start time
func SharedStartup() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables directly")
	}

	// Initialize API keys in internal package -- at minimum, abuseIPDB key is required
	abuseipdb.InitializeAPIKey()
	virustotal.InitializeVTAPIKey()
	vpnid.InitializeVpnID()
}
