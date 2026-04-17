package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tlop503/ipcheq2/internal/data"
)

func main() {
	dataDir := flag.String("data-dir", ".", "Directory containing iCloud prefix files and hash cache")
	flag.Parse()

	changed, err := data.UpdateICloudRelays(*dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if changed {
		fmt.Println("hashes differ, updating prefix files")
		return
	}

	fmt.Println("skipping update since list hasn't changed")
}
