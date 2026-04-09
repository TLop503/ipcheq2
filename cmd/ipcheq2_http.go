package main

import (
	"github.com/tlop503/ipcheq2/internal"
	"log"

	"github.com/tlop503/ipcheq2/internal/cli"
	"github.com/tlop503/ipcheq2/internal/router"
)

func main() {

	cfg, err := cli.InitFlags()
	if err != nil {
		log.Fatalf("InitFlags: %v\n", err)
	}

	internal.SharedStartup()

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
