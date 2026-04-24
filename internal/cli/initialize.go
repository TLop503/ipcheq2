package cli

import (
	"github.com/tlop503/ipcheq2/internal/data"
	"github.com/tlop503/ipcheq2/internal/queries"
	"github.com/tlop503/ipcheq2/internal/router"
	"log"
	"os"
)

// InitServer parses CLI flags and calls corresponding router hooks
func InitServer() {
	cfg, err := InitFlags()
	if err != nil {
		log.Fatalf("InitFlags: %v\n", err)
	}

	queries.InitConnectors()

	if cfg.Update {
		data.Update()
		log.Println("Update complete!")
		os.Exit(0)
	}

	switch cfg.Mode {
	case ModeWebUI:
		router.RouteWebui()
		router.StartServing()
	case ModeAPI:
		router.RouteWebui()
		router.RouteAPI()
		router.StartServing()
	case ModeHeadless:
		router.RouteAPI()
		router.StartServing()

	default:
		log.Fatalf("Unknown mode: %v\n", cfg.Mode)
	}
}
