package main

import (
	"github.com/tlop503/ipcheq2/internal"
	"github.com/tlop503/ipcheq2/internal/cli"
	"github.com/tlop503/ipcheq2/internal/queries"
	"log"
)

func main() {
	cfg, err := cli.InitCliFlags()
	if err != nil {
		log.Fatalf("InitCliFlags: %v\n", err)
	}

	internal.SharedStartup()

	queries.PrettyPrint(cfg.QueryIP, cfg.Mode)

}
