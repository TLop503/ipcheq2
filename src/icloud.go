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

	ipCount, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatalf("Couldn't read count: %s", err)
	}

	prefixes := make([]netip.Prefix, 0, ipCount)

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
	ipV4Prefixes = loadPrefixFile("ipv4.txt")
	ipV6Prefixes = loadPrefixFile("ipv6.txt")

	log.Printf("ipv4: %d\n", len(ipV4Prefixes))
	log.Printf("ipv6: %d\n", len(ipV6Prefixes))
}
