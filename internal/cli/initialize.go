package cli

import (
	"github.com/tlop503/ipcheq2/internal/router"
	"log"
)

// InitServer parses CLI flags and calls corresponding router hooks
func InitServer() {
	cfg, err := InitFlags()
	if err != nil {
		log.Fatalf("InitFlags: %v\n", err)
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
	//case cli.ModeQuery:
	//	log.Println("Query mode not yet implemented!")
	default:
		log.Fatalf("Unknown mode: %v\n", cfg.Mode)
	}
}
