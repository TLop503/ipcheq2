package cli

import (
	"flag"
	"fmt"
	"net/netip"
	"os"
)

// InitFlags acts as the entry for the main ipcheq binary
func InitFlags() (Config, error) {
	registerSharedFlags()
	flag.BoolVar(&update, "update", false, "")
	flag.BoolVar(&update, "u", false, "")
	flag.BoolVar(&compact, "compact", false, "")
	flag.BoolVar(&compact, "c", false, "")
	flag.IntVar(&port, "port", 8080, "")
	flag.IntVar(&port, "p", 8080, "")

	flag.Parse()

	if help {
		fmt.Println("Welcome to ipcheq2.")
		fmt.Println("Invoke without arguments to launch the web UI.")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("Optional flags:")
		fmt.Println("  --mode <mode>    Set serving mode: webui | api | headless")
		fmt.Println("                   	webui    - serves the web UI only (default)")
		fmt.Println("                   	api      - serves web UI and exposes API")
		fmt.Println("                   	headless - exposes API only, no web UI")
		fmt.Println("  --port -p        Port to serve on. Low ports may require root")
		fmt.Println("                      overrides value set in config")
		fmt.Println("  --update -u      Update data sources")
		fmt.Println("                   	currently only updates iCloud relays")
		fmt.Println("  --compact -c     Compress data to minimum spanning subnets")
		fmt.Println("                   	compression may take a few min")
		fmt.Println("                   	note: bundled data is already compacted,")
		fmt.Println("  						 but updates are uncompressed")
		fmt.Println("  --help -h        Show this help message.")
		fmt.Println()
		fmt.Println("-----------------------------------------------------------------")
		os.Exit(0)
	}

	var runMode RunMode

	switch {
	case mode == "api":
		runMode = ModeAPI
	case mode == "headless":
		runMode = ModeHeadless
	case mode == "" || mode == "webui":
		runMode = ModeWebUI
	default:
		return Config{}, fmt.Errorf("unknown mode %q: must be webui, api, or headless", mode)
	}

	return Config{Mode: runMode, Update: update, Compact: compact, Port: port}, nil
}

// InitCliFlags parses arguments for the cli binary
func InitCliFlags() (CliConfig, error) {
	registerSharedFlags()
	flag.StringVar(&query, "a", "127.0.0.1", "")
	flag.StringVar(&query, "addr", "", "")
	flag.BoolVar(&human, "H", false, "")
	flag.BoolVar(&human, "human", false, "")

	flag.Parse()

	if help {
		fmt.Println("Welcome to ipcheq2 cli")
		fmt.Println("Usage: ipc2c [OPTIONS] [ADDRESS]")
		fmt.Println("--------------------------------------------------------------------------------------------")
		fmt.Println("Optional flags:")
		fmt.Println("  -m --mode <mode>             Set query mode: first | third | full")
		fmt.Println("                                      first    - only use local data")
		fmt.Println("                                      third    - only query remote sources")
		fmt.Println("                                      full     - query local and remote sources (DEFAULT)")
		fmt.Println("  -a --addr <ip address>       IP address to query (v4 or v6)")
		fmt.Println("  -H --human                   Print human-friendly summary rather than formatted JSON")
		fmt.Println("  -h --help                    Show this help message.")
		fmt.Println()
		fmt.Println("--------------------------------------------------------------------------------------------")
		os.Exit(0)
	}

	// validate query
	addr, err := netip.ParseAddr(query)
	if err != nil {
		return CliConfig{}, fmt.Errorf("invalid IP address %q: %w", query, err)
	}

	switch {
	case mode == "first":
		return CliConfig{Mode: ModeFirst, QueryIP: addr, HumanReadable: human}, nil
	case mode == "third":
		return CliConfig{Mode: ModeThird, QueryIP: addr, HumanReadable: human}, nil
	case mode == "" || mode == "full":
		return CliConfig{Mode: ModeFull, QueryIP: addr, HumanReadable: human}, nil
	default:
		return CliConfig{}, fmt.Errorf("unknown mode %q: must be first, third, or full", mode)
	}
}

// registerSharedFlags establishes help and mode flags
func registerSharedFlags() {
	// mode maps to iotas for either binary
	flag.StringVar(&mode, "mode", "", "")
	flag.StringVar(&mode, "m", "", "")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "")
}
