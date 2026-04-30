//go:build windows

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tlop503/ipcheq2/v2/internal/data"
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
	cacheHome := t.TempDir()
	t.Setenv("APPDATA", appData)
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	if _, err := data.EnsureDataDir(); err != nil {
		t.Fatalf("failed to hydrate cache data for Init test: %v", err)
	}

	cfg, err := Init()
	if err != nil {
		t.Errorf("failed to init test config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Init returned nil config")
	}

	if len(cfg.Sources) != len(defaultConfig().Sources) {
		t.Fatalf("len(cfg.Sources) = %d, want %d", len(cfg.Sources), len(defaultConfig().Sources))
	}

	configPath := filepath.Join(appData, "ipcheq2", "ipcheq2.yaml")
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
