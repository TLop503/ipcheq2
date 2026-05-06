package data

import (
	"log"
	"path/filepath"
)

func Update() {
	icloudWrapper()
	bulk_compact()
}

func icloudWrapper() {
	log.Println("Attempting to update iCloud relays")
	log.Println("This will take ~3 minutes")
	dataDir, err := EnsureDataDir()
	if err != nil {
		log.Fatal(err)
	}
	hashDir, err := EnsureHashDir()
	if err != nil {
		log.Fatal(err)
	}

	icloudData := filepath.Join(dataDir, "icloud.txt")
	icloudHash := filepath.Join(hashDir, "icloud.sha256")

	err = updateiCloud(icloudData, icloudHash)
	if err != nil {
		log.Fatalf("updateiCloud: %v\n", err)
	}
}
