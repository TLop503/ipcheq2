package src

import (
	"bufio"
	"log"
	"net/netip"
	"os"
	"strconv"
)

var (
	ipV4Prefixes []netip.Prefix
	ipV6Prefixes []netip.Prefix
)

func loadPrefixFile(filename string) []netip.Prefix {
	prefixFile, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Couldn't open %s: %s", filename, err)
	}

	scanner := bufio.NewScanner(prefixFile)
	scanner.Scan()

	// first line of file is number of prefixes
	ipCount, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatalf("Couldn't read count: %s", err)
	}

	prefixes := make([]netip.Prefix, 0, ipCount)

	// parse prefixes on remaining lines
	for scanner.Scan() {
		prefixStr := scanner.Text()
		prefix, err := netip.ParsePrefix(prefixStr)
		if err != nil {
			log.Fatalf("Invalid prefix: %s (error: %s)", prefixStr, err)
		}

		prefixes = append(prefixes, prefix.Masked())
	}

	return prefixes
}

func LoadICloudPrefixes() {
	ipV4Prefixes = loadPrefixFile("prefixes/ipv4.txt")
	ipV6Prefixes = loadPrefixFile("prefixes/ipv6.txt")
}

func CheckICloudIP(address netip.Addr) bool {
	// determine proper prefix list to use (IPv4/6)
	var checkList []netip.Prefix
	if address.Is4() {
		checkList = ipV4Prefixes
	} else {
		checkList = ipV6Prefixes
	}

	for _, prefix := range checkList {
		if prefix.Contains(address) {
			return true
		}
	}

	return false
}
