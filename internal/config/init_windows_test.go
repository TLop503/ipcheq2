//go:build windows

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitWindowsCreatesAndLoadsConfig(t *testing.T) {
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(prevWD, "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed to chdir to repo root %q: %v", repoRoot, err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevWD)
	})

	appData := t.TempDir()
	t.Setenv("APPDATA", appData)

	cfg := Init()
	if cfg == nil {
		t.Fatal("Init returned nil config")
	}

	if len(cfg.Sources) != len(defaultConfig().Sources) {
		t.Fatalf("len(cfg.Sources) = %d, want %d", len(cfg.Sources), len(defaultConfig().Sources))
	}

	configPath := filepath.Join(appData, "ipcheq2", "ipcheq2.json")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file at %q, stat error: %v", configPath, err)
	}

	loaded, err := LoadAndValidateConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAndValidateConfig failed for Init-created config: %v", err)
	}

	if len(loaded.Sources) != len(defaultConfig().Sources) {
		t.Fatalf("len(loaded.Sources) = %d, want %d", len(loaded.Sources), len(defaultConfig().Sources))
	}
}
