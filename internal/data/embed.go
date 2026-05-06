package data

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//go:embed *.txt
var embeddedData embed.FS

// getDataDir determines correct cache path
func getDataDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "ipcheq2", "data"), nil
}

// getHashDir determines correct cache path
func getHashDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "ipcheq2", "hashes"), nil
}

// copyEmbeddedData writes embedded txt files to disk.
func copyEmbeddedData(dstDir string) error {
	entries, err := embeddedData.ReadDir(".")
	if err != nil {
		return fmt.Errorf("reading embedded data entries: %w", err)
	}

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("creating data directory %q: %w", dstDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") {
			continue
		}

		srcFile, err := embeddedData.Open(entry.Name())
		if err != nil {
			return fmt.Errorf("opening embedded file %q: %w", entry.Name(), err)
		}

		target := filepath.Join(dstDir, entry.Name())
		dstFile, err := os.Create(target)
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("creating destination file %q: %w", target, err)
		}

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			dstFile.Close()
			srcFile.Close()
			return fmt.Errorf("copying embedded file %q to %q: %w", entry.Name(), target, err)
		}

		if err := dstFile.Close(); err != nil {
			srcFile.Close()
			return fmt.Errorf("closing destination file %q: %w", target, err)
		}
		if err := srcFile.Close(); err != nil {
			return fmt.Errorf("closing embedded file %q: %w", entry.Name(), err)
		}
	}

	return nil
}

// EnsureDataDir confirms the user's cache dir exists
func EnsureDataDir() (string, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := copyEmbeddedData(dataDir); err != nil {
			return "", err
		}
	}

	return dataDir, nil
}

// EnsureHashDir creates the hash directory if it does not exist
func EnsureHashDir() (string, error) {
	hashDir, err := getHashDir()
	if err != nil {
		return "", err
	}

	// create dir
	if _, err := os.Stat(hashDir); os.IsNotExist(err) {
		if err := os.MkdirAll(hashDir, 0755); err != nil {
			return "", fmt.Errorf("creating data directory %q: %w", hashDir, err)
		}
	}

	return hashDir, nil
}
