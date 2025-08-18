package vpnid

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

// ConfigEntry represents one line in the config file: name → path
type ConfigEntry struct {
	Name string
	Path string
}

// validateConfig reads the config file, parses it into entries, and returns them.
// Returns an error if the file cannot be read or any line is invalid.
func validateConfig(path string) ([]ConfigEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var entries []ConfigEntry
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Split line into "name : path"
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid config format on line %d: %q", lineNum, line)
		}

		name := strings.TrimSpace(parts[0])
		filePath := strings.TrimSpace(parts[1])

		if name == "" || filePath == "" {
			return nil, fmt.Errorf("empty name or path on line %d: %q", lineNum, line)
		}

		// Check file exists and readable
		info, err := os.Stat(filePath)
		if err != nil {
			return nil, fmt.Errorf("file %q does not exist: %w", filePath, err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("file %q is a directory", filePath)
		}

		entries = append(entries, ConfigEntry{
			Name: name,
			Path: filePath,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading config file: %w", err)
	}

	return entries, nil
}

func Initialize(path string) error {
	_, err := validateConfig(path)
	if err != nil {
		return fmt.Errorf("config validation error: %w", err)
	}

	return errors.New("WIP")

}

func Query(addr net.IPAddr) (string, error) {
	return "", errors.New("Not implemented")
}
