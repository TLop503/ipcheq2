package data

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

// bufferIfHashDiffers compares a remote file to an expected hash. If they differ, the file and hash are returned
// otherwise, nil
func bufferIfHashDiffers(url, expectedHash string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("bad status: %s", resp.Status)
	}

	hasher := sha256.New()
	var buf bytes.Buffer

	// Stream into both hash and buffer
	writer := io.MultiWriter(hasher, &buf)

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return nil, "", err
	}

	actualHash := fmt.Sprintf("%x", hasher.Sum(nil))

	// If hashes differ, return the data
	if actualHash != expectedHash {
		return buf.Bytes(), actualHash, nil
	}

	// If same, no work needed
	return nil, "", nil
}

// WriteNormalizedIPNets writes IPs and ranges to a file, one per line
// /32 and /128 addresses are written with no cidr suffix
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

		// normalize representation
		if ip.To4() != nil {
			ip = ip.To4()
		} else {
			ip = ip.To16()
		}

		ones, bits := n.Mask.Size()

		switch {
		// IPv4 host route (/32)
		case ip.To4() != nil && ones == 32 && bits == 32:
			fmt.Fprintln(w, ip.String())

		// IPv6 host route (/128)
		case ip.To4() == nil && ones == 128 && bits == 128:
			fmt.Fprintln(w, ip.String())

		// everything else stays CIDR
		default:
			fmt.Fprintf(w, "%s/%d\n", ip.String(), ones)
		}
	}

	return nil
}
