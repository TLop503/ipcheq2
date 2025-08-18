package vpnid

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

// Verify that each file specified in the config is real and readable
func validateConfig(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

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
			return fmt.Errorf("invalid config format on line %d: %q", lineNum, line)
		}

		filePath := strings.TrimSpace(parts[1])

		// Check file exists and readable
		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("file %q does not exist: %w", filePath, err)
		}
		if info.IsDir() {
			return fmt.Errorf("file %q is a directory", filePath)
		}
		
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed reading config file: %w", err)
	}

	return nil
}

func Initialize(path string) error {
	return errors.New("Not implemented")
}

func Query(addr net.IPAddr) (string, error) {
	return "", errors.New("Not implemented")
}
