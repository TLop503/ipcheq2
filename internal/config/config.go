package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Init loads and verifies the config file
func Init() *Config {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Error getting user config dir: %v", err)
		log.Printf("Using default config instead...")
	}

	configDir = filepath.Join(configDir, "ipcheq2")
	configFile := filepath.Join(configDir, "ipcheq2.yaml")
	err = ensureConfig(configFile)
	if err != nil {
		log.Panicf("Error ensuring config file: %v", err)
	}
	cfg, err := LoadAndValidateConfig(configFile)
	if err != nil {
		log.Panicf("Error loading config: %v", err)
	}
	return cfg
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

		// Check file exists + is readable
		f, err := os.Open(s.Path)
		if err != nil {
			return nil, fmt.Errorf("source %q: cannot open file %q: %w", s.Name, s.Path, err)
		}
		f.Close()
	}

	return &cfg, nil
}
