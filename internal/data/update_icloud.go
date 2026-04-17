package data

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const (
	defaultICloudURL = "https://mask-api.icloud.com/egress-ip-ranges.csv"
	iCloudHashFile   = "upstream-icloud-list.hash"
	iCloudIPv4File   = "icloud_ipv4.txt"
	iCloudIPv6File   = "icloud_ipv6.txt"
)

type prefixRange struct {
	start []byte
	end   []byte
	bits  int
}

// UpdateICloudRelays refreshes the collapsed iCloud Private Relay prefix files in dataDir.
func UpdateICloudRelays(dataDir string) (bool, error) {
	return updateICloudRelays(dataDir, http.DefaultClient, defaultICloudURL)
}

func updateICloudRelays(dataDir string, client *http.Client, url string) (bool, error) {
	if client == nil {
		client = http.DefaultClient
	}

	response, err := client.Get(url)
	if err != nil {
		return false, fmt.Errorf("fetch iCloud prefixes: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, fmt.Errorf("fetch iCloud prefixes: unexpected status %s", response.Status)
	}

	prefixesCSV, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Errorf("read iCloud prefixes: %w", err)
	}

	prefixesHash := sha256.Sum256(prefixesCSV)
	hashPath := filepath.Join(dataDir, iCloudHashFile)
	previousHash, err := os.ReadFile(hashPath)
	if err != nil {
		return false, fmt.Errorf("read cached iCloud hash: %w", err)
	}

	if strings.TrimSpace(string(previousHash)) == fmt.Sprintf("%x", prefixesHash) {
		return false, nil
	}

	prefixesByVersion, err := parseICloudPrefixes(prefixesCSV)
	if err != nil {
		return false, err
	}

	ipv4, err := collapsePrefixes(prefixesByVersion[32])
	if err != nil {
		return false, err
	}
	ipv6, err := collapsePrefixes(prefixesByVersion[128])
	if err != nil {
		return false, err
	}

	if err := writePrefixes(filepath.Join(dataDir, iCloudIPv4File), ipv4); err != nil {
		return false, err
	}
	if err := writePrefixes(filepath.Join(dataDir, iCloudIPv6File), ipv6); err != nil {
		return false, err
	}

	if err := os.WriteFile(hashPath, []byte(fmt.Sprintf("%x", prefixesHash)), 0o644); err != nil {
		return false, fmt.Errorf("write cached iCloud hash: %w", err)
	}

	return true, nil
}

func parseICloudPrefixes(prefixesCSV []byte) (map[int][]netip.Prefix, error) {
	reader := csv.NewReader(bytes.NewReader(prefixesCSV))
	reader.FieldsPerRecord = -1

	prefixesByVersion := map[int][]netip.Prefix{
		32:  {},
		128: {},
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse iCloud CSV: %w", err)
		}
		if len(record) == 0 {
			continue
		}

		prefixText := strings.TrimSpace(record[0])
		if prefixText == "" {
			continue
		}

		prefix, err := netip.ParsePrefix(prefixText)
		if err != nil {
			if isLikelyHeader(prefixText) {
				continue
			}
			return nil, fmt.Errorf("parse iCloud prefix %q: %w", prefixText, err)
		}

		version := prefix.Addr().BitLen()
		prefixesByVersion[version] = append(prefixesByVersion[version], prefix.Masked())
	}

	return prefixesByVersion, nil
}

func isLikelyHeader(value string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func collapsePrefixes(prefixes []netip.Prefix) ([]netip.Prefix, error) {
	if len(prefixes) == 0 {
		return nil, nil
	}

	bitLen := prefixes[0].Addr().BitLen()
	ranges := make([]prefixRange, 0, len(prefixes))
	for _, prefix := range prefixes {
		if prefix.Addr().BitLen() != bitLen {
			return nil, fmt.Errorf("mixed IP versions are not supported")
		}

		start := append([]byte(nil), prefix.Masked().Addr().AsSlice()...)
		end, err := prefixRangeEnd(prefix.Masked())
		if err != nil {
			return nil, err
		}

		ranges = append(ranges, prefixRange{start: start, end: end, bits: bitLen})
	}

	sort.Slice(ranges, func(i, j int) bool {
		if cmp := compareBytes(ranges[i].start, ranges[j].start); cmp != 0 {
			return cmp < 0
		}
		return compareBytes(ranges[i].end, ranges[j].end) < 0
	})

	merged := mergeRanges(ranges)
	out := make([]netip.Prefix, 0, len(merged))
	for _, r := range merged {
		collapsed, err := rangeToPrefixes(r.start, r.end, r.bits)
		if err != nil {
			return nil, err
		}
		out = append(out, collapsed...)
	}

	return out, nil
}

func mergeRanges(ranges []prefixRange) []prefixRange {
	if len(ranges) == 0 {
		return nil
	}

	merged := []prefixRange{ranges[0]}
	for _, current := range ranges[1:] {
		last := &merged[len(merged)-1]
		lastPlusOne, overflow := incrementBytes(last.end)
		if overflow || compareBytes(current.start, lastPlusOne) <= 0 {
			if compareBytes(current.end, last.end) > 0 {
				last.end = append([]byte(nil), current.end...)
			}
			continue
		}
		merged = append(merged, current)
	}

	return merged
}

func rangeToPrefixes(start, end []byte, bitLen int) ([]netip.Prefix, error) {
	if compareBytes(start, end) > 0 {
		return nil, fmt.Errorf("invalid range")
	}

	current := new(big.Int).SetBytes(start)
	endInt := new(big.Int).SetBytes(end)
	one := big.NewInt(1)
	byteLen := len(start)
	var out []netip.Prefix

	for current.Cmp(endInt) <= 0 {
		currentBytes := bigIntToBytes(current, byteLen)
		alignmentBits := trailingZeroBits(currentBytes)

		remaining := new(big.Int).Sub(endInt, current)
		remaining.Add(remaining, one)
		maxRangeBits := remaining.BitLen() - 1

		blockBits := alignmentBits
		if maxRangeBits < blockBits {
			blockBits = maxRangeBits
		}

		addr, ok := netip.AddrFromSlice(currentBytes)
		if !ok {
			return nil, fmt.Errorf("invalid prefix address")
		}

		out = append(out, netip.PrefixFrom(addr, bitLen-blockBits).Masked())

		blockSize := new(big.Int).Lsh(one, uint(blockBits))
		current.Add(current, blockSize)
	}

	return out, nil
}

func prefixRangeEnd(prefix netip.Prefix) ([]byte, error) {
	addr := prefix.Masked().Addr()
	start := append([]byte(nil), addr.AsSlice()...)
	bitLen := addr.BitLen()
	hostBits := bitLen - prefix.Bits()
	if hostBits < 0 {
		return nil, fmt.Errorf("invalid prefix length")
	}
	if hostBits == 0 {
		return start, nil
	}

	end := append([]byte(nil), start...)
	fullBytes := hostBits / 8
	remBits := hostBits % 8
	for i := len(end) - fullBytes; i < len(end); i++ {
		end[i] = 0xFF
	}
	if remBits > 0 {
		boundary := len(end) - fullBytes - 1
		if boundary < 0 || boundary >= len(end) {
			return nil, fmt.Errorf("invalid prefix length")
		}
		end[boundary] |= byte((1 << remBits) - 1)
	}

	return end, nil
}

func writePrefixes(path string, prefixes []netip.Prefix) error {
	var builder strings.Builder
	for _, prefix := range prefixes {
		builder.WriteString(prefix.String())
		builder.WriteByte('\n')
	}

	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func compareBytes(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func incrementBytes(input []byte) ([]byte, bool) {
	output := append([]byte(nil), input...)
	for i := len(output) - 1; i >= 0; i-- {
		output[i]++
		if output[i] != 0 {
			return output, false
		}
	}
	return output, true
}

func trailingZeroBits(input []byte) int {
	bits := 0
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == 0 {
			bits += 8
			continue
		}
		bits += trailingZeroBitsInByte(input[i])
		break
	}
	return bits
}

func trailingZeroBitsInByte(value byte) int {
	for i := 0; i < 8; i++ {
		if value&(1<<i) != 0 {
			return i
		}
	}
	return 8
}

func bigIntToBytes(value *big.Int, length int) []byte {
	raw := value.Bytes()
	if len(raw) > length {
		raw = raw[len(raw)-length:]
	}
	if len(raw) == length {
		return raw
	}

	output := make([]byte, length)
	copy(output[length-len(raw):], raw)
	return output
}
