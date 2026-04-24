package data

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tlop503/ipcheq2/internal/iputils"
	"gopkg.in/yaml.v2"
)

// bufferIfHashDiffers compares a remote file to an expected hash. If they differ, the file and hash are returned
// otherwise, nil
func bufferIfHashDiffers(url, expectedHashPath string) ([]byte, string, error) {
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

	var expectedHashStr string
	if _, err := os.Stat(expectedHashPath); errors.Is(err, os.ErrNotExist) {
		expectedHashStr = ""
	} else {
		expectedHash, err := os.ReadFile(expectedHashPath)
		if err != nil {
			return nil, "", err
		}
		expectedHashStr = string(expectedHash)
	}

	// If hashes differ, return the data
	if actualHash != expectedHashStr {
		log.Println("Hashes differ, loading data...")
		log.Println("Actual hash:", actualHash)
		log.Println("Expected hash:", expectedHashStr)
		return buf.Bytes(), actualHash, nil
	}

	// If same, no work needed
	log.Println("Hashes do not differ, continuing...")
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

func writeHashToFile(hash string, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(hash)
	return err
}

type bulkCompactSource struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type bulkCompactConfig struct {
	Sources []bulkCompactSource `yaml:"sources"`
}

func bulk_compact() {
	cfg, err := loadBulkCompactConfig()
	if err != nil {
		log.Fatal(err)
	}

	for _, source := range cfg.Sources {
		if source.Path == "" {
			continue
		}

		log.Printf("Compacting %s (%s)", source.Name, source.Path)

		beforeInfo, err := os.Stat(source.Path)
		if err != nil {
			log.Fatalf("stat before compact (%s): %v", source.Path, err)
		}

		f, err := os.Open(source.Path)
		if err != nil {
			log.Fatalf("open source (%s): %v", source.Path, err)
		}

		rawIPs := iputils.DataToIPNetSlice(bufio.NewScanner(f))
		if err := f.Close(); err != nil {
			log.Fatalf("close source (%s): %v", source.Path, err)
		}

		compacted := iputils.Compact(rawIPs)
		if err := WriteNormalizedIPNets(compacted, source.Path); err != nil {
			log.Fatalf("write compacted file (%s): %v", source.Path, err)
		}

		afterInfo, err := os.Stat(source.Path)
		if err != nil {
			log.Fatalf("stat after compact (%s): %v", source.Path, err)
		}

		deltaKB := float64(beforeInfo.Size()-afterInfo.Size()) / 1024
		log.Printf("Compacted %s: %.2f KB saved", source.Path, deltaKB)
	}

	log.Println("Finished compacting configured data files")
}

func loadBulkCompactConfig() (*bulkCompactConfig, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "ipcheq2", "ipcheq2.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("missing config file; run a normal query once to initialize config")
		}
		return nil, err
	}

	var cfg bulkCompactConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	cacheRoot := filepath.Join(cacheDir, "ipcheq2")
	for i := range cfg.Sources {
		cfg.Sources[i].Name = strings.TrimSpace(cfg.Sources[i].Name)
		cfg.Sources[i].Path = strings.TrimSpace(cfg.Sources[i].Path)
		if cfg.Sources[i].Path == "" {
			continue
		}
		if !filepath.IsAbs(cfg.Sources[i].Path) {
			cfg.Sources[i].Path = filepath.Clean(filepath.Join(cacheRoot, cfg.Sources[i].Path))
		} else {
			cfg.Sources[i].Path = filepath.Clean(cfg.Sources[i].Path)
		}
	}

	return &cfg, nil
}
