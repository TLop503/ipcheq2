package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// getCacheRootDir returns the user's cache directory + ipcheq2
func getCacheRootDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "ipcheq2"), nil
}

// resolveSourcePath adjusts relative paths to the cache directory, and laves absolute untouched
// for example, /foo -> /foo, but bar/buz -> $cachedir/bar/buz
func resolveSourcePath(cacheRoot, sourcePath string) string {
	if filepath.IsAbs(sourcePath) {
		return filepath.Clean(sourcePath)
	}

	return filepath.Clean(filepath.Join(cacheRoot, sourcePath))
}

// Init loads and verifies the config file
func Init() (*Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Error getting user config dir: %v", err)
		log.Printf("Using default config instead...")
	}

	configDir = filepath.Join(configDir, "ipcheq2")
	configFile := filepath.Join(configDir, "ipcheq2.yaml")
	err = ensureConfig(configFile)
	if err != nil {
		return nil, err
	}
	cfg, err := LoadAndValidateConfig(configFile)
	if err != nil {
		log.Printf("Error loading config: %v", err)
	}
	return cfg, nil
}

// ensureConfig verifies config exists, and if not creates the default
func ensureConfig(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Printf("Config file does not exist. Creating default at %s", path)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("Error creating config directory: %v", err)
		}
		cfg := defaultConfig()
		err = writeConfig(path, cfg)
		if err != nil {
			return fmt.Errorf("Error creating config file: %v", err)
		}
	}

	return nil
}

// writeConfig does
func writeConfig(path string, cfg Config) error {
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadAndValidateConfig makes sure each listed file exists and is readable
func LoadAndValidateConfig(path string) (*Config, error) {
	// load
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	// unmarshall
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// --- Validate sources ---
	if len(cfg.Sources) == 0 {
		return nil, errors.New("no sources defined")
	}

	cacheRoot, err := getCacheRootDir()
	if err != nil {
		return nil, fmt.Errorf("getting user cache dir: %w", err)
	}

	// verify each exists + is readable
	seen := make(map[string]struct{})
	for i, s := range cfg.Sources {
		if s.Name == "" {
			return nil, fmt.Errorf("sources[%d]: name is empty", i)
		}
		if s.Path == "" {
			return nil, fmt.Errorf("sources[%d]: path is empty", i)
		}

		// Check duplicate names
		if _, exists := seen[s.Name]; exists {
			return nil, fmt.Errorf("duplicate source name: %s", s.Name)
		}
		seen[s.Name] = struct{}{}

		cfg.Sources[i].Path = resolveSourcePath(cacheRoot, s.Path)

		// Check file exists + is readable
		f, err := os.Open(cfg.Sources[i].Path)
		if err != nil {
			return nil, fmt.Errorf("source %q: cannot open file %q: %w", s.Name, cfg.Sources[i].Path, err)
		}
		f.Close()
	}

	return &cfg, nil
}
