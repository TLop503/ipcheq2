package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKeysMissingFileReturnsEmpty(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	keys, err := LoadKeys()
	if err != nil {
		t.Fatalf("LoadKeys returned unexpected error: %v", err)
	}

	if keys.ABIPDBKey != "" || keys.VTKey != "" {
		t.Fatalf("expected empty keys on missing keys file, got %+v", keys)
	}
}

func TestLoadKeysParsesYAML(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	keysPath := filepath.Join(configHome, "ipcheq2", "keys.yaml")
	if err := os.MkdirAll(filepath.Dir(keysPath), 0755); err != nil {
		t.Fatalf("failed to create keys directory: %v", err)
	}

	content := []byte("abipdbKey: test-ab-key\nvtKey: test-vt-key\n")
	if err := os.WriteFile(keysPath, content, 0644); err != nil {
		t.Fatalf("failed to write keys file: %v", err)
	}

	keys, err := LoadKeys()
	if err != nil {
		t.Fatalf("LoadKeys returned unexpected error: %v", err)
	}

	if keys.ABIPDBKey != "test-ab-key" {
		t.Fatalf("ABIPDBKey = %q, want %q", keys.ABIPDBKey, "test-ab-key")
	}
	if keys.VTKey != "test-vt-key" {
		t.Fatalf("VTKey = %q, want %q", keys.VTKey, "test-vt-key")
	}
}

func TestEnsureKeysFileCreatesWhenMissing(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	path, created, err := EnsureKeysFile()
	if err != nil {
		t.Fatalf("EnsureKeysFile returned unexpected error: %v", err)
	}
	if !created {
		t.Fatal("expected created=true when keys file is missing")
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected keys file to exist at %q: %v", path, err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading created keys file: %v", err)
	}
	if len(content) != 0 {
		t.Fatalf("expected blank keys file, got %q", string(content))
	}
}

func TestEnsureKeysFileDoesNotOverwriteExisting(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)

	keysPath := filepath.Join(configHome, "ipcheq2", "keys.yaml")
	if err := os.MkdirAll(filepath.Dir(keysPath), 0755); err != nil {
		t.Fatalf("failed to create keys directory: %v", err)
	}
	const existing = "abipdbKey: keep-me\n"
	if err := os.WriteFile(keysPath, []byte(existing), 0644); err != nil {
		t.Fatalf("failed to seed keys file: %v", err)
	}

	path, created, err := EnsureKeysFile()
	if err != nil {
		t.Fatalf("EnsureKeysFile returned unexpected error: %v", err)
	}
	if created {
		t.Fatal("expected created=false when keys file already exists")
	}
	if path != keysPath {
		t.Fatalf("path = %q, want %q", path, keysPath)
	}

	content, err := os.ReadFile(keysPath)
	if err != nil {
		t.Fatalf("failed reading keys file: %v", err)
	}
	if string(content) != existing {
		t.Fatalf("keys file was overwritten: got %q, want %q", string(content), existing)
	}
}
