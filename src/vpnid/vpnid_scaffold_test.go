package vpnid

import (
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

		// file with content
		file1 := makeTempFile(t, tmp, "f1.txt", "foo")
		file2 := makeTempFile(t, tmp, "f2.txt", "bar")

		// config content
		config := `foo : ` + file1 + "\nbar : " + file2 + "\n"
		configPath := makeTempFile(t, tmp, "config.txt", config)

		if err := validateConfig(configPath); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("bad format line", func(t *testing.T) {
		tmp := t.TempDir()
		config := "this-is-wrong\n"
		configPath := makeTempFile(t, tmp, "config.txt", config)

		if err := validateConfig(configPath); err == nil {
			t.Errorf("expected error for bad format line, got nil")
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		tmp := t.TempDir()
		config := "foo : /nonexistent/path\n"
		configPath := makeTempFile(t, tmp, "config.txt", config)

		if err := validateConfig(configPath); err == nil {
			t.Errorf("expected error for missing file, got nil")
		}
	})
}

func TestInitialize(t *testing.T) {
	t.Run("no error expected", func(t *testing.T) {
		tmp := t.TempDir()
		config := "foo : bar\n"
		configPath := makeTempFile(t, tmp, "config.txt", config)

		if err := Initialize(configPath); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
