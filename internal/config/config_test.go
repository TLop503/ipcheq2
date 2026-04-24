package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if len(cfg.Sources) == 0 {
		t.Fatal("defaultConfig returned no sources")
	}

	if got := len(cfg.Sources); got != 14 {
		t.Fatalf("len(cfg.Sources) = %d, want 14", got)
	}
}

func TestEnsureConfigCreatesDefaultWhenMissing(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "ipcheq2.yaml")

	if err := ensureConfig(configPath); err != nil {
		t.Fatalf("ensureConfig returned unexpected error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read created config file: %v", err)
	}

	var got Config
	if err := yaml.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to parse created config yaml: %v", err)
	}

	if len(got.Sources) != len(defaultConfig().Sources) {
		t.Fatalf("created config sources length = %d, want %d", len(got.Sources), len(defaultConfig().Sources))
	}
}

func TestEnsureConfigDoesNotOverwriteExisting(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "ipcheq2.yaml")

	original := Config{Sources: []Source{{Name: "Custom", Path: "/tmp/custom.txt"}}}
	if err := writeConfig(configPath, original); err != nil {
		t.Fatalf("failed to write initial config: %v", err)
	}

	if err := ensureConfig(configPath); err != nil {
		t.Fatalf("ensureConfig returned unexpected error: %v", err)
	}

	got, err := LoadAndValidateConfig(configPath)
	if err == nil {
		t.Fatalf("expected validation error because %q does not exist", original.Sources[0].Path)
	}

	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("failed to read config file after ensureConfig: %v", readErr)
	}

	var onDisk Config
	if unmarshalErr := yaml.Unmarshal(data, &onDisk); unmarshalErr != nil {
		t.Fatalf("failed to unmarshal config from disk: %v", unmarshalErr)
	}

	if len(onDisk.Sources) != 1 || onDisk.Sources[0].Name != "Custom" {
		t.Fatalf("config appears overwritten: got %+v", onDisk.Sources)
	}

	if got != nil {
		t.Fatalf("expected nil config on validation error, got %+v", got)
	}
}

func TestLoadAndValidateConfigSuccess(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.txt")
	if err := os.WriteFile(sourcePath, []byte("1.2.3.4\n"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	cfgOnDisk := Config{Sources: []Source{{Name: "Test Source", Path: sourcePath}}}
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := writeConfig(configPath, cfgOnDisk); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	got, err := LoadAndValidateConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAndValidateConfig returned unexpected error: %v", err)
	}

	if len(got.Sources) != 1 {
		t.Fatalf("len(got.Sources) = %d, want 1", len(got.Sources))
	}

	if got.Sources[0].Name != "Test Source" || got.Sources[0].Path != sourcePath {
		t.Fatalf("got source = %+v, want name/path preserved", got.Sources[0])
	}
}

func TestLoadAndValidateConfigResolvesRelativePathFromCacheRoot(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	cacheRoot := filepath.Join(cacheHome, "ipcheq2")
	relPath := filepath.Join("data", "custom.txt")
	absPath := filepath.Join(cacheRoot, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}
	if err := os.WriteFile(absPath, []byte("ok\n"), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	cfgOnDisk := Config{Sources: []Source{{Name: "Relative Source", Path: relPath}}}
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := writeConfig(configPath, cfgOnDisk); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	got, err := LoadAndValidateConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAndValidateConfig returned unexpected error: %v", err)
	}

	if got.Sources[0].Path != absPath {
		t.Fatalf("resolved path = %q, want %q", got.Sources[0].Path, absPath)
	}
}

func TestLoadAndValidateConfigPreservesAbsoluteCustomPath(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	absPath := filepath.Join(t.TempDir(), "external", "custom.txt")
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}
	if err := os.WriteFile(absPath, []byte("ok\n"), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	cfgOnDisk := Config{Sources: []Source{{Name: "Absolute Source", Path: absPath}}}
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := writeConfig(configPath, cfgOnDisk); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	got, err := LoadAndValidateConfig(configPath)
	if err != nil {
		t.Fatalf("LoadAndValidateConfig returned unexpected error: %v", err)
	}

	if got.Sources[0].Path != filepath.Clean(absPath) {
		t.Fatalf("resolved path = %q, want %q", got.Sources[0].Path, filepath.Clean(absPath))
	}
}

func TestLoadAndValidateConfigErrors(t *testing.T) {
	tempDir := t.TempDir()
	validSource := filepath.Join(tempDir, "valid.txt")
	if err := os.WriteFile(validSource, []byte("ok\n"), 0644); err != nil {
		t.Fatalf("failed creating valid source file: %v", err)
	}

	tests := []struct {
		name       string
		configBody string
		path       string
		wantErr    string
	}{
		{
			name:       "missing config file",
			path:       filepath.Join(tempDir, "does-not-exist.yaml"),
			wantErr:    "reading config",
			configBody: "",
		},
		{
			name: "invalid yaml",
			path: filepath.Join(tempDir, "invalid.yaml"),
			configBody: `sources:
  - name: "A"
    path: [not-a-string`,
			wantErr: "parsing config",
		},
		{
			name: "no sources",
			path: filepath.Join(tempDir, "no-sources.yaml"),
			configBody: `sources: []
`,
			wantErr: "no sources defined",
		},
		{
			name: "empty source name",
			path: filepath.Join(tempDir, "empty-name.yaml"),
			configBody: `sources:
  - name: ""
    path: "` + validSource + `"
`,
			wantErr: "name is empty",
		},
		{
			name: "empty source path",
			path: filepath.Join(tempDir, "empty-path.yaml"),
			configBody: `sources:
  - name: "A"
    path: ""
`,
			wantErr: "path is empty",
		},
		{
			name: "duplicate source name",
			path: filepath.Join(tempDir, "duplicate.yaml"),
			configBody: `sources:
  - name: "A"
    path: "` + validSource + `"
  - name: "A"
    path: "` + validSource + `"
`,
			wantErr: "duplicate source name",
		},
		{
			name: "source file missing",
			path: filepath.Join(tempDir, "missing-source.yaml"),
			configBody: `sources:
  - name: "A"
    path: "` + filepath.Join(tempDir, "missing.txt") + `"
`,
			wantErr: "cannot open file",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.configBody != "" {
				if err := os.WriteFile(tt.path, []byte(tt.configBody), 0644); err != nil {
					t.Fatalf("failed to write test config %q: %v", tt.path, err)
				}
			}

			_, err := LoadAndValidateConfig(tt.path)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestEnsureConfigCreatesMissingParentDir(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "missing-parent", "ipcheq2.yaml")

	err := ensureConfig(configPath)
	if err != nil {
		t.Fatalf("expected success when parent directory is missing, got error: %v", err)
	}

	if _, statErr := os.Stat(configPath); statErr != nil {
		t.Fatalf("expected config file to be created, stat error: %v", statErr)
	}
}
