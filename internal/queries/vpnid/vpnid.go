package vpnid

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/netip"
	"os"
	"sort"
	"strings"

	"github.com/yl2chen/cidranger"
)

// validateConfig reads the config file, parses it into entries, and returns them.
// Returns an error if the file cannot be read or any line is invalid.
func validateConfig(path string) ([]configEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	return parseConfigEntries(file, func(filePath string) error {
		info, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("file %q does not exist: %w", filePath, err)
		}
		if info.IsDir() {
			return fmt.Errorf("file %q is a directory", filePath)
		}
		return nil
	})
}

func validateConfigFromFS(fsys fs.FS, path string) ([]configEntry, error) {
	file, err := fsys.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	return parseConfigEntries(file, func(filePath string) error {
		info, err := fs.Stat(fsys, filePath)
		if err != nil {
			return fmt.Errorf("file %q does not exist: %w", filePath, err)
		}
		if info.IsDir() {
			return fmt.Errorf("file %q is a directory", filePath)
		}
		return nil
	})
}

func parseConfigEntries(reader io.Reader, validatePath func(filePath string) error) ([]configEntry, error) {
	var entries []configEntry
	scanner := bufio.NewScanner(reader)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid config format on line %d: %q", lineNum, line)
		}

		name := strings.TrimSpace(parts[0])
		filePath := strings.TrimSpace(parts[1])
		if name == "" || filePath == "" {
			return nil, fmt.Errorf("empty name or path on line %d: %q", lineNum, line)
		}

		if err := validatePath(filePath); err != nil {
			return nil, err
		}

		entries = append(entries, configEntry{Name: name, Path: filePath})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading config file: %w", err)
	}

	return entries, nil
}

// initialize returns a new CIDRanger from a given config, passed as a file path
func initialize(path string) (cidranger.Ranger, error) {
	configEntries, err := validateConfig(path)
	if err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return initializeWithEntries(configEntries, func(tree cidranger.Ranger, path, provider string) error {
		return addToTree(tree, path, provider)
	})
}

func initializeFromFS(fsys fs.FS, path string) (cidranger.Ranger, error) {
	configEntries, err := validateConfigFromFS(fsys, path)
	if err != nil {
		return nil, fmt.Errorf("config validation error: %w", err)
	}

	return initializeWithEntries(configEntries, func(tree cidranger.Ranger, path, provider string) error {
		return addToTreeFromFS(tree, fsys, path, provider)
	})
}

func initializeWithEntries(configEntries []configEntry, add func(tree cidranger.Ranger, path, provider string) error) (cidranger.Ranger, error) {

	var ranger = cidranger.NewPCTrieRanger()

	for _, entry := range configEntries {
		err := add(ranger, entry.Path, entry.Name)
		if err != nil {
			return nil, err
		}
	}

	return ranger, nil
}

// addToTree reads every line from a file at 'path', collapses IPs into CIDR ranges,
// and inserts them into the cidranger tree using the given provider/source name.
func addToTree(tree cidranger.Ranger, path string, provider string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer file.Close()

	return addToTreeFromReader(tree, file, path, provider)
}

func addToTreeFromFS(tree cidranger.Ranger, fsys fs.FS, path string, provider string) error {
	file, err := fsys.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer file.Close()

	return addToTreeFromReader(tree, file, path, provider)
}

func addToTreeFromReader(tree cidranger.Ranger, reader io.Reader, source string, provider string) error {

	var ipv4s []net.IP
	var prefixes []netip.Prefix

	scanner := bufio.NewScanner(reader)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try parse as CIDR first
		if p, err := netip.ParsePrefix(line); err == nil {
			prefixes = append(prefixes, p)
			continue
		}

		// Try parse as single IP
		if ip, err := netip.ParseAddr(line); err == nil {
			// Convert to net.IP for collapse function
			netIP := ip.AsSlice()
			if ip.Is4() {
				ipv4s = append(ipv4s, netIP)
			} else {
				// For IPv6, just add as individual /128 prefix for now
				p := netip.PrefixFrom(ip, ip.BitLen())
				prefixes = append(prefixes, p)
			}
			continue
		}

		return fmt.Errorf("invalid IP or CIDR on line %d of %q: %q", lineNum, source, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %q: %w", source, err)
	}

	// Insert any CIDR prefixes directly
	for _, p := range prefixes {
		tree.Insert(treeEntry{Prefix: p, Provider: provider})
	}

	// collapse IPv4 IPs into ranges (IPv6 not yet collapsed)
	if len(ipv4s) > 0 {
		// Sort IPv4 IPs for collapse function
		sortIPs(ipv4s)

		// collapse into CIDR ranges
		cidrs := collapse(ipv4s)

		// Insert collapsed ranges
		for _, cidr := range cidrs {
			// Convert net.IPNet back to netip.Prefix
			prefix, err := netipFromIPNet(cidr)
			if err != nil {
				return fmt.Errorf("failed to convert CIDR %s: %w", cidr, err)
			}
			tree.Insert(treeEntry{Prefix: prefix, Provider: provider})
		}
	}

	return nil
}

// query to slice checks for db matches and returns a slice of strings or nil
func QueryToSlice(ip netip.Addr) ([]string, error) {
	if VpnIDRanger == nil {
		return nil, fmt.Errorf("VPNIDRanger not initialized")
	}

	// Lookup all prefixes containing this IP
	entries, err := VpnIDRanger.ContainingNetworks(ip.AsSlice())
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return []string{}, nil
	}

	seen := make(map[string]struct{})
	var providers []string

	for _, e := range entries {
		te := e.(treeEntry)

		if _, ok := seen[te.Provider]; !ok {
			seen[te.Provider] = struct{}{}
			providers = append(providers, te.Provider)
		}
	}

	return providers, nil
}

// sortIPs sorts a slice of net.IP addresses
func sortIPs(ips []net.IP) {
	sort.Slice(ips, func(i, j int) bool {
		return ipLess(ips[i], ips[j])
	})
}

// ipLess compares two IP addresses and returns true if the first is less than (as 16B) the second
func ipLess(a, b net.IP) bool {
	// Ensure both are same format
	a = a.To16()
	b = b.To16()
	for i := 0; i < len(a); i++ {
		if a[i] < b[i] {
			return true
		} else if a[i] > b[i] {
			return false
		}
	}
	return false // they're equal
}

// netipFromIPNet converts net.IPNet to netip.Prefix
func netipFromIPNet(ipnet *net.IPNet) (netip.Prefix, error) {
	addr, ok := netip.AddrFromSlice(ipnet.IP)
	if !ok {
		return netip.Prefix{}, fmt.Errorf("invalid IP address")
	}

	ones, _ := ipnet.Mask.Size()
	return netip.PrefixFrom(addr, ones), nil
}
