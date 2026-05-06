package vpnid

import (
	"bufio"
	"fmt"
	"github.com/tlop503/ipcheq2/v2/internal/iputils"
	"net"
	"net/netip"
	"os"
	"strings"

	"github.com/tlop503/ipcheq2/v2/internal/config"
	"github.com/yl2chen/cidranger"
)

// initialize builds a new CIDRanger from configured sources.
func initialize() (cidranger.Ranger, error) {
	var ranger = cidranger.NewPCTrieRanger()
	cfg, err := config.Init()

	if err != nil {
		return nil, err
	}

	// iterate through idx:sources, only looking at values
	for _, source := range cfg.Sources {
		if err := addToTree(ranger, source.Path, source.Name); err != nil {
			return nil, fmt.Errorf("failed loading source %q from %q: %w", source.Name, source.Path, err)
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
			// Convert to net.IP for Collapse function
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

	// Collapse IPv4 IPs into ranges (IPv6 not yet collapsed)
	if len(ipv4s) > 0 {
		// Sort IPv4 IPs for Collapse function
		iputils.SortIPs(ipv4s)

		// Collapse into CIDR ranges
		cidrs := iputils.CollapseIPsToNets(ipv4s)

		// Insert collapsed ranges
		for _, cidr := range cidrs {
			// Convert net.IPNet back to netip.Prefix
			prefix, err := iputils.NetipFromIPNet(cidr)
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
