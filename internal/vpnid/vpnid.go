package vpnid

import (
	"bufio"
	"fmt"
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

	var entries []configEntry
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

		entries = append(entries, configEntry{
			Name: name,
			Path: filePath,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading config file: %w", err)
	}

	return entries, nil
}

func initialize(path string) (cidranger.Ranger, error) {
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

// addToTree reads every line from a file at 'path', collapses IPs into CIDR ranges,
// and inserts them into the cidranger tree using the given provider/source name.
func addToTree(tree cidranger.Ranger, path string, provider string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %q: %w", path, err)
	}
	defer file.Close()

	var ipv4s []net.IP
	var prefixes []netip.Prefix

	scanner := bufio.NewScanner(file)
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

		return fmt.Errorf("invalid IP or CIDR on line %d of %q: %q", lineNum, path, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %q: %w", path, err)
	}

	// Insert any CIDR prefixes directly
	for _, p := range prefixes {
		tree.Insert(treeEntry{Prefix: p, Provider: provider})
	}

	// collapse IPv4 IPs into ranges (IPv6 already handled above)
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

func Query(ip netip.Addr, ranger cidranger.Ranger) (string, error) {
	if !ip.Is4() { // IPv6
		return "IPv6 Support Coming Soon", nil
	}

	// Lookup all prefixes containing this IP
	entries, err := ranger.ContainingNetworks(ip.AsSlice())
	if err != nil {
		return "Query Error!", err
	}
	if len(entries) == 0 {
		return fmt.Sprintf("%s not found in dataset", ip), nil
	}

	// Collect provider names (in case of overlap)
	providers := []string{}
	for _, e := range entries {
		if te, ok := e.(treeEntry); ok {
			providers = append(providers, te.Provider)
		}
	}

	return fmt.Sprintf("%s is used by %v", ip, providers), nil
}

// sortIPs sorts a slice of net.IP addresses
func sortIPs(ips []net.IP) {
	sort.Slice(ips, func(i, j int) bool {
		return ipLess(ips[i], ips[j])
	})
}

// ipLess compares two IP addresses
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
