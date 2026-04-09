package main

import (
	"github.com/tlop503/ipcheq2/internal/cli"
	"github.com/tlop503/ipcheq2/internal/queries"
)

func main() {
	queries.InitConnectors()
	cli.InitServer()
}
