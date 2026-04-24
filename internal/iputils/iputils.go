package iputils

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tlop503/ipcheq2/internal/config"
	"github.com/yl2chen/cidranger"
)

// SortIPs sorts a slice of net.IP addresses as though they were just bytestrings
func SortIPs(ips []net.IP) {
	sort.Slice(ips, func(i, j int) bool {
		// true if first IP less than second
		return -1 == bytes.Compare(ips[i], ips[j])
	})
}

// NetipFromIPNet converts net.IPNet to netip.Prefix
func NetipFromIPNet(ipnet *net.IPNet) (netip.Prefix, error) {
	addr, ok := netip.AddrFromSlice(ipnet.IP)
	if !ok {
		return netip.Prefix{}, fmt.Errorf("invalid IP address")
	}

	ones, _ := ipnet.Mask.Size()
	return netip.PrefixFrom(addr, ones), nil
}

// CollapseIPsToNets takes a sorted list of IPs and returns minimal CIDR subnets
func CollapseIPsToNets(ips []net.IP) []*net.IPNet {
	var results []*net.IPNet
	if len(ips) == 0 {
		return results
	}

	start := ips[0]
	end := ips[0]

	for i := 1; i < len(ips); i++ {
		// If current is exactly one greater than end, extend range
		if ipIncrementN(end, 1).Equal(ips[i]) {
			end = ips[i]
		} else {
			// finalize current block
			results = append(results, rangeToIPNetCidr(start, end)...)
			start = ips[i]
			end = ips[i]
		}
	}
	// finalize last block
	results = append(results, rangeToIPNetCidr(start, end)...)

	return results
}

// ipIncrementN increases an address by N bytes
func ipIncrementN(ip net.IP, n int) net.IP {
	ip = ip.To4()
	val := ipToUint32(ip)
	val += uint32(n)
	return uint32ToIP(val)
}

func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		// This should not happen in our IPv4-only processing
		panic("ipToUint32: got nil IP after To4() conversion - this indicates an IPv6 address in IPv4 processing")
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func uint32ToIP(v uint32) net.IP {
	return net.IPv4(
		byte(v>>24),
		byte((v>>16)&0xFF),
		byte((v>>8)&0xFF),
		byte(v&0xFF),
	)
}

// rangeToIPNetCidr takes a start and end IP (inclusive) and returns minimal CIDRs covering the range.
func rangeToIPNetCidr(start, end net.IP) []*net.IPNet {
	var nets []*net.IPNet
	s := ipToUint32(start)
	e := ipToUint32(end)

	for s <= e {
		// Find the largest CIDR block that:
		// 1. Starts at the current position (s)
		// 2. Doesn't exceed the end (e)
		// 3. Is properly aligned

		var prefixLen uint32 = 32

		// Find the largest block size that fits
		for prefixLen > 0 {
			blockSize := uint32(1) << (32 - prefixLen)

			// Check alignment: s must be divisible by blockSize
			if s&(blockSize-1) != 0 {
				prefixLen++
				break
			}

			// Check if block would exceed the end
			if s+blockSize-1 > e {
				prefixLen++
				break
			}

			prefixLen--
		}

		// Create the CIDR
		ip := uint32ToIP(s)
		mask := net.CIDRMask(int(prefixLen), 32)
		ipnet := &net.IPNet{
			IP:   ip.Mask(mask),
			Mask: mask,
		}

		nets = append(nets, ipnet)

		// Move to the next block
		blockSize := uint32(1) << (32 - prefixLen)
		s += blockSize
	}

	return nets
}

// DataToIPNetSlice reads a bufio scanner containing 1 addr or subnet per line
// and outputs a slice of all addresses as ipnet types
// Note: each line is parsed until a comma is seen, if present
func DataToIPNetSlice(scanner *bufio.Scanner) []*net.IPNet {
	var rawIPs []*net.IPNet

	for scanner.Scan() {
		line := untilComma(strings.TrimSpace(scanner.Text()))

		// Try CIDR first
		if _, ipNet, err := net.ParseCIDR(line); err == nil && ipNet != nil {
			ipNet.IP = ipNet.IP.To16()
			rawIPs = append(rawIPs, ipNet)
			continue
		}

		// Otherwise attempt to parse single IP
		if ip := net.ParseIP(line); ip != nil {
			ip = ip.To16()

			var mask net.IPMask
			if ip.To4() != nil {
				mask = net.CIDRMask(32, 32)
			} else {
				mask = net.CIDRMask(128, 128)
			}

			rawIPs = append(rawIPs, &net.IPNet{
				IP:   ip,
				Mask: mask,
			})
			continue
		}

		log.Printf("Failed to parse line: %s", line)
	}

	return rawIPs
}

func untilComma(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return s[:i]
		}
	}
	return s
}

// normalize takes a raw []*net.IPNet and returns clean, deduplicated,
// host-bit-normalized entries split by address family.
func normalize(raw []*net.IPNet) (v4, v6 []*net.IPNet) {
	seen := make(map[string]struct{})

	for _, n := range raw {
		if n == nil {
			continue
		}

		// Zero host bits: e.g. 192.168.1.1/24 → 192.168.1.0/24
		masked := &net.IPNet{
			IP:   n.IP.Mask(n.Mask),
			Mask: n.Mask,
		}

		key := masked.String()
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}

		if masked.IP.To4() != nil {
			v4 = append(v4, masked)
		} else {
			v6 = append(v6, masked)
		}
	}
	return v4, v6
}

// sortBySize sorts prefixes broadest-first (smallest mask length = largest network).
// Ties broken by IP for determinism.
func sortBySize(nets []*net.IPNet) {
	sort.Slice(nets, func(i, j int) bool {
		mi, _ := nets[i].Mask.Size()
		mj, _ := nets[j].Mask.Size()
		if mi != mj {
			return mi < mj // broader first
		}
		return nets[i].IP.String() < nets[j].IP.String()
	})
}

// filterContained inserts prefixes broadest-first; skips any prefix
// already covered by something in the ranger.
func filterContained(nets []*net.IPNet) ([]*net.IPNet, error) {
	sortBySize(nets)

	ranger := cidranger.NewPCTrieRanger()
	var kept []*net.IPNet

	for _, n := range nets {
		containing, err := ranger.ContainingNetworks(n.IP)
		if err != nil {
			return nil, err
		}
		// Also check if a broader prefix we already added covers this one.
		// ContainingNetworks checks a single IP — use the network address,
		// which is sufficient since we normalized host bits already.
		alreadyCovered := false
		for _, entry := range containing {
			existing := entry.Network()
			if existing.Contains(n.IP) {
				// Check mask: if existing is broader or equal, n is redundant
				eOnes, _ := existing.Mask.Size()
				nOnes, _ := n.Mask.Size()
				if eOnes <= nOnes {
					alreadyCovered = true
					break
				}
			}
		}
		if !alreadyCovered {
			if err := ranger.Insert(cidranger.NewBasicRangerEntry(*n)); err != nil {
				return nil, err
			}
			kept = append(kept, n)
		}
	}
	return kept, nil
}

// mergeSiblings repeatedly merges adjacent same-size sibling prefixes
// (e.g. 10.0.0.0/25 + 10.0.0.128/25 → 10.0.0.0/24) until stable.
func mergeSiblings(nets []*net.IPNet) []*net.IPNet {
	for {
		merged, changed := mergePass(nets)
		if !changed {
			return merged
		}
		nets = merged
	}
}

func mergePass(nets []*net.IPNet) ([]*net.IPNet, bool) {
	sortBySize(nets) // ensure consistent order for pairing

	used := make([]bool, len(nets))
	var result []*net.IPNet
	changed := false

	for i := 0; i < len(nets); i++ {
		if used[i] {
			continue
		}
		onesI, bitsI := nets[i].Mask.Size()
		if onesI == 0 {
			// Default route — can't merge up
			result = append(result, nets[i])
			continue
		}

		merged := false
		for j := i + 1; j < len(nets); j++ {
			if used[j] {
				continue
			}
			onesJ, bitsJ := nets[j].Mask.Size()
			if onesI != onesJ || bitsI != bitsJ {
				continue // different prefix lengths
			}

			parent := parentPrefix(nets[i])
			if parent == nil {
				continue
			}
			if parent.Contains(nets[j].IP) {
				// Check nets[j] is also directly under parent (not just contained)
				onesP, _ := parent.Mask.Size()
				if onesI == onesP+1 {
					result = append(result, parent)
					used[i] = true
					used[j] = true
					changed = true
					merged = true
					break
				}
			}
		}
		if !merged {
			result = append(result, nets[i])
		}
	}
	return result, changed
}

// parentPrefix returns the next-broader prefix containing n,
// or nil if already at /0.
func parentPrefix(n *net.IPNet) *net.IPNet {
	ones, bits := n.Mask.Size()
	if ones == 0 {
		return nil
	}
	parentMask := net.CIDRMask(ones-1, bits)
	parentIP := n.IP.Mask(parentMask)
	return &net.IPNet{IP: parentIP, Mask: parentMask}
}

// Compact calculates the minimum spanning subnets for a list of potentially overlapping data
func Compact(rawIPs []*net.IPNet) []*net.IPNet {
	log.Println("Attempting to compact IPs")
	v4, v6 := normalize(rawIPs)
	log.Println("Sorting...")
	sortBySize(v4)
	sortBySize(v6)
	log.Println("Sorted!")

	log.Println("Filtering...")

	var err error
	v4, err = filterContained(v4)
	if err != nil {
		log.Fatal(err)
	}
	v6, err = filterContained(v6)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Filtered!")

	log.Println("Merging...")
	v4 = mergeSiblings(v4)
	v6 = mergeSiblings(v6)
	log.Println("Merged!")

	output := append(v4, v6...)
	return output
}

// WriteNormalizedIPNets writes IPs and ranges to a file, one per line.
// /32 and /128 addresses are written with no CIDR suffix.
func WriteNormalizedIPNets(nets []*net.IPNet, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for _, n := range nets {
		if n == nil {
			continue
		}

		ip := n.IP

		if ip.To4() != nil {
			ip = ip.To4()
		} else {
			ip = ip.To16()
		}

		ones, bits := n.Mask.Size()

		switch {
		case ip.To4() != nil && ones == 32 && bits == 32:
			fmt.Fprintln(w, ip.String())
		case ip.To4() == nil && ones == 128 && bits == 128:
			fmt.Fprintln(w, ip.String())
		default:
			fmt.Fprintf(w, "%s/%d\n", ip.String(), ones)
		}
	}

	return nil
}

// BulkCompact compacts each configured file in place and logs size reduction in KB.
func BulkCompact() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Error getting user config dir: %v", err)
		log.Printf("Using default config instead...")
	}

	configDir = filepath.Join(configDir, "ipcheq2")
	configFile := filepath.Join(configDir, "ipcheq2.yaml")
	sources, err := config.LoadAndValidateConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	for _, source := range sources.Sources {
		log.Printf("Compacting %s (%s)", source.Name, source.Path)

		beforeInfo, err := os.Stat(source.Path)
		if err != nil {
			return fmt.Errorf("stat before compact (%s): %w", source.Path, err)
		}

		f, err := os.Open(source.Path)
		if err != nil {
			return fmt.Errorf("open source (%s): %w", source.Path, err)
		}

		rawIPs := DataToIPNetSlice(bufio.NewScanner(f))
		if err := f.Close(); err != nil {
			return fmt.Errorf("close source (%s): %w", source.Path, err)
		}

		compacted := Compact(rawIPs)
		if err := WriteNormalizedIPNets(compacted, source.Path); err != nil {
			return fmt.Errorf("write compacted file (%s): %w", source.Path, err)
		}

		afterInfo, err := os.Stat(source.Path)
		if err != nil {
			return fmt.Errorf("stat after compact (%s): %w", source.Path, err)
		}

		deltaKB := float64(beforeInfo.Size()-afterInfo.Size()) / 1024
		log.Printf("Compacted %s: %.2f KB saved", source.Path, deltaKB)
	}

	log.Println("Finished compacting configured data files")
	return nil
}
