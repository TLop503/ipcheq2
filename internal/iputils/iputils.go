package iputils

import (
	"bytes"
	"fmt"
	"net"
	"net/netip"
	"sort"
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
