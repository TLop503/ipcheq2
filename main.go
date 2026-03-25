package main

import (
	"github.com/tlop503/ipcheq2/internal/abuseipdb"
	"github.com/tlop503/ipcheq2/internal/cli"
	"github.com/tlop503/ipcheq2/internal/router"
	"github.com/tlop503/ipcheq2/internal/virustotal"
	"github.com/tlop503/ipcheq2/internal/vpnid"
	"log"

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
	case cli.ModeAPI:
		router.RouteWebui()
		log.Println("API mode not yet implemented!")
	case cli.ModeHeadless:
		log.Println("Headless mode not yet implemented!")
	case cli.ModeQuery:
		log.Println("Query mode not yet implemented!")
	case cli.ModeREPL:
		log.Println("REPL mode not yet implemented!")
	default:
		log.Fatalf("Unknown mode: %v\n", cfg.Mode)
	}
}
