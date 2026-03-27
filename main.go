package main

import (
	"log"

	"github.com/tlop503/ipcheq2/internal/cli"
	"github.com/tlop503/ipcheq2/internal/queries/abuseipdb"
	"github.com/tlop503/ipcheq2/internal/queries/virustotal"
	"github.com/tlop503/ipcheq2/internal/queries/vpnid"
	"github.com/tlop503/ipcheq2/internal/router"

	"github.com/joho/godotenv"
)

func main() {

	cfg, err := cli.InitFlags()
	if err != nil {
		log.Fatalf("InitFlags: %v\n", err)
	}

	// Try to load .env file (optional for container deployment)
	err = godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables directly")
	}

	// Initialize API keys in internal package -- at minimum, abuseIPDB key is required
	abuseipdb.InitializeAPIKey()
	virustotal.InitializeVTAPIKey()

	// Initialize VPN ID ranger
	vpnid.InitializeVpnID()

	switch cfg.Mode {
	case cli.ModeWebUI:
		router.RouteWebui()
		router.StartServing()
	case cli.ModeAPI:
		router.RouteWebui()
		router.RouteAPI()
		router.StartServing()
	case cli.ModeHeadless:
		router.RouteAPI()
		router.StartServing()
	//case cli.ModeQuery:
	//	log.Println("Query mode not yet implemented!")
	default:
		log.Fatalf("Unknown mode: %v\n", cfg.Mode)
	}
}
