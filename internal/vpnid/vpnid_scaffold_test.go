package vpnid

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// helper to create temp file with given content
func makeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		tmp := t.TempDir()

		// files with content
		file1 := makeTempFile(t, tmp, "f1.txt", "foo")
		file2 := makeTempFile(t, tmp, "f2.txt", "bar")

		// config content
		config := `foo : ` + file1 + "\nbar : " + file2 + "\n"
		configPath := makeTempFile(t, tmp, "config.txt", config)

		entries, err := validateConfig(configPath)
		if err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}

		// optional: check parsed entries
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(entries))
		} else {
			if entries[0].Name != "foo" || entries[0].Path != file1 {
				t.Errorf("first entry mismatch: %+v", entries[0])
			}
			if entries[1].Name != "bar" || entries[1].Path != file2 {
				t.Errorf("second entry mismatch: %+v", entries[1])
			}
		}
	})

	t.Run("bad format line", func(t *testing.T) {
		tmp := t.TempDir()
		config := "this-is-wrong\n"
		configPath := makeTempFile(t, tmp, "config.txt", config)

		if _, err := validateConfig(configPath); err == nil {
			t.Errorf("expected error for bad format line, got nil")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		tmp := t.TempDir()
		config := "foo : /nonexistent/path\n"
		configPath := makeTempFile(t, tmp, "config.txt", config)

		if _, err := validateConfig(configPath); err == nil {
			t.Errorf("expected error for missing file, got nil")
		}
	})
}

func TestInitialize(t *testing.T) {
	t.Run("no error expected", func(t *testing.T) {
		tmp := t.TempDir()

		// Create a data file first
		dataFile := makeTempFile(t, tmp, "data.txt", "1.1.1.1\n2.2.2.0/24\n")

		// Create config that references the actual data file
		config := fmt.Sprintf("provider1 : %s\n", dataFile)
		configPath := makeTempFile(t, tmp, "config.txt", config)

		if _, err := initialize(configPath); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
