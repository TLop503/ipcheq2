package vpnid

import (
	"bufio"
	"fmt"
	"github.com/yl2chen/cidranger"
	"net"
	"net/netip"
	"os"
	"strings"
)

// ConfigEntry represents one line in the config file: name → path
type ConfigEntry struct {
	Name string
	Path string
}

type TreeEntry struct {
	Prefix   netip.Prefix
	Provider string
}

// Network implements RangerEntry using net.IPNet
func (t TreeEntry) Network() net.IPNet {
	return net.IPNet{
		IP:   t.Prefix.Addr().AsSlice(),                               // starting IP
		Mask: net.CIDRMask(t.Prefix.Bits(), t.Prefix.Addr().BitLen()), // mask
	}
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

func Initialize(path string) (cidranger.Ranger, error) {
	configEntries, err := validateConfig(path)

	if err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	var ranger = cidranger.NewPCTrieRanger()

	for _, entry := range configEntries {
		err = addToTree(ranger, entry.Path, entry.Name)
		if err != nil {
			return nil, err
		}
	}

	return ranger, nil
}

// addToTree reads every line from a file at 'path' and inserts it into the cidranger tree
// using the given provider/source name.
func addToTree(tree cidranger.Ranger, path string, provider string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try parse as CIDR
		if p, err := netip.ParsePrefix(line); err == nil {
			tree.Insert(TreeEntry{Prefix: p, Provider: provider})
			continue
		}

		// Try parse as single IP
		if ip, err := netip.ParseAddr(line); err == nil {
			p := netip.PrefixFrom(ip, ip.BitLen()) // /32 or /128
			tree.Insert(TreeEntry{Prefix: p, Provider: provider})
			continue
		}

		return fmt.Errorf("invalid IP or CIDR on line %d of %q: %q", lineNum, path, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %q: %w", path, err)
	}

	return nil
}

func Query(addr net.IPAddr, ranger cidranger.Ranger) (string, error) {
	ip := addr.IP

	if ip.To4() == nil { // IPv6
		return "IPv6 Support Coming Soon", nil
	}

	// Lookup all prefixes containing this IP
	entries, err := ranger.ContainingNetworks(ip)
	if err != nil {
		return "", err
	}

	if len(entries) == 0 {
		return fmt.Sprintf("%s not found in dataset", ip), nil
	}

	// Collect provider names (in case of overlap)
	providers := []string{}
	for _, e := range entries {
		if te, ok := e.(TreeEntry); ok {
			providers = append(providers, te.Provider)
		}
	}

	return fmt.Sprintf("%s is owned by %v", ip, providers), nil
}
