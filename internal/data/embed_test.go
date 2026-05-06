package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddedDataContainsExpectedFile(t *testing.T) {
	f, err := embeddedData.Open("cyberghost.txt")
	if err != nil {
		t.Fatalf("expected embedded file cyberghost.txt to be present: %v", err)
	}
	defer f.Close()
}

func TestEnsureDataDirHydratesOnFirstRun(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	dataDir, err := EnsureDataDir()
	if err != nil {
		t.Fatalf("EnsureDataDir returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dataDir, "cyberghost.txt")); err != nil {
		t.Fatalf("expected hydrated file to exist: %v", err)
	}
}

func TestEnsureDataDirDoesNotOverwriteExistingDiskData(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	dataDir, err := EnsureDataDir()
	if err != nil {
		t.Fatalf("EnsureDataDir first call returned error: %v", err)
	}

	target := filepath.Join(dataDir, "cyberghost.txt")
	const customContent = "custom-user-data\n"
	if err := os.WriteFile(target, []byte(customContent), 0644); err != nil {
		t.Fatalf("failed writing custom disk content: %v", err)
	}

	if _, err := EnsureDataDir(); err != nil {
		t.Fatalf("EnsureDataDir second call returned error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed reading custom disk content: %v", err)
	}

	if string(got) != customContent {
		t.Fatalf("expected disk content to be preserved, got %q", string(got))
	}
}
