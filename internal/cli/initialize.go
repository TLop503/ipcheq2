package cli

import (
	"fmt"
	"github.com/tlop503/ipcheq2/v2/internal/data"
	"github.com/tlop503/ipcheq2/v2/internal/queries"
	"github.com/tlop503/ipcheq2/v2/internal/router"
	"log"
)

// InitServer parses CLI flags and calls corresponding router hooks
func InitServer() {
	cfg, err := InitFlags()
	if err != nil {
		log.Fatalf("InitFlags: %v\n", err)
	}

	if cfg.Update {
		data.Update()
		log.Println("Update complete!")
	}
	if cfg.Compact {
		data.Bulk_compact()
		log.Println("Bulk compact complete!")
	}
	if cfg.Compact || cfg.Update {
		return
	}

	queries.InitConnectors()

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

// CliEntry parses flags and queries DB
func CliEntry() {
	cfg, err := InitCliFlags()
	if err != nil {
		log.Fatalf("InitFlags: %v\n", err)
	}

	queries.InitConnectors()
	fmt.Printf("\nQuery results for %v from ipcheq2:\n\n", cfg.QueryIP)
	queries.PrettyPrint(cfg.QueryIP, int(cfg.Mode), cfg.HumanReadable)
}
