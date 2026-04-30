package vpnid

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tlop503/ipcheq2/v2/internal/config"
	"gopkg.in/yaml.v2"
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

func writeInitConfig(t *testing.T, cfg config.Config) string {
	t.Helper()

	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("APPDATA", configHome)

	configDir := filepath.Join(configHome, "ipcheq2")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "ipcheq2.yaml")
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		t.Fatalf("failed to marshal config yaml: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	return configPath
}

func TestInitialize(t *testing.T) {
	t.Run("loads configured sources", func(t *testing.T) {
		tmp := t.TempDir()

		dataFileA := makeTempFile(t, tmp, "provider-a.txt", "1.1.1.1\n2.2.2.0/24\n")
		dataFileB := makeTempFile(t, tmp, "provider-b.txt", "3.3.3.3\n")

		writeInitConfig(t, config.Config{
			Sources: []config.Source{
				{Name: "provider-a", Path: dataFileA},
				{Name: "provider-b", Path: dataFileB},
			},
		})

		ranger, err := initialize()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if ranger == nil {
			t.Fatal("expected initialized ranger, got nil")
		}
	})

	t.Run("fails when source contains invalid line", func(t *testing.T) {
		tmp := t.TempDir()

		invalidDataFile := makeTempFile(t, tmp, "invalid-provider.txt", "not-an-ip\n")

		writeInitConfig(t, config.Config{
			Sources: []config.Source{
				{Name: "broken-provider", Path: invalidDataFile},
			},
		})

		if _, err := initialize(); err == nil {
			t.Fatal("expected initialize to fail for invalid source data, got nil")
		}
	})
}
