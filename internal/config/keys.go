package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Keys struct {
	ABIPDBKey string `yaml:"abipdbKey"`
	VTKey     string `yaml:"vtKey"`
}

const keysTemplate = `# API keys for ipcheq2
# Place this file at:
#   Linux:   ~/.config/ipcheq2/keys.yaml (or $XDG_CONFIG_HOME/ipcheq2/keys.yaml)
#   Windows: %APPDATA%/ipcheq2/keys.yaml

abipdbKey: ""
vtKey: ""
`

// EnsureKeysFile creates a template keys file when one does not exist yet.
// Returns path, whether it was created, and any error.
func EnsureKeysFile() (string, bool, error) {
	path, err := keysFilePath()
	if err != nil {
		return "", false, fmt.Errorf("getting keys file path: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		return path, false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", false, fmt.Errorf("checking keys file %q: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", false, fmt.Errorf("creating keys directory %q: %w", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(keysTemplate), 0644); err != nil {
		return "", false, fmt.Errorf("creating keys file %q: %w", path, err)
	}

	return path, true, nil
}

func keysFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "ipcheq2", "keys.yaml"), nil
}

// LoadKeys loads API keys from the per-user config directory.
// Missing keys file is treated as "no keys configured" and returns a zero-value Keys struct.
func LoadKeys() (Keys, error) {
	path, err := keysFilePath()
	if err != nil {
		return Keys{}, fmt.Errorf("getting keys file path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Keys{}, nil
		}
		return Keys{}, fmt.Errorf("reading keys file %q: %w", path, err)
	}

	var keys Keys
	if err := yaml.Unmarshal(data, &keys); err != nil {
		return Keys{}, fmt.Errorf("parsing keys file %q: %w", path, err)
	}

	keys.ABIPDBKey = strings.TrimSpace(keys.ABIPDBKey)
	keys.VTKey = strings.TrimSpace(keys.VTKey)

	return keys, nil
}
