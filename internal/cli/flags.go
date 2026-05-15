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
		fmt.Println("  --update -u      Update data sources")
		fmt.Println("                   	currently only updates iCloud relays")
		fmt.Println("  --compact -c     Compress data to minimum spanning subnets")
		fmt.Println("                   	compression may take a few min")
		fmt.Println("                   	note: bundled data is already compacted, but updates are raw")
		fmt.Println("  						and should be compressed")
		fmt.Println("  --help -h        Show this help message.")
		fmt.Println()
		fmt.Println("-----------------------------------------------------------------")
		os.Exit(0)
	}

	switch {
	case query != "":
		return Config{Mode: ModeQuery, Update: update, Compact: compact}, nil
	case mode == "api":
		return Config{Mode: ModeAPI, Update: update, Compact: compact}, nil
	case mode == "headless":
		return Config{Mode: ModeHeadless, Update: update, Compact: compact}, nil
	case mode == "" || mode == "webui":
		return Config{Mode: ModeWebUI, Update: update, Compact: compact}, nil
	default:
		return Config{}, fmt.Errorf("unknown mode %q: must be webui, api, or headless", mode)
	}
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
		fmt.Println("  -m --mode <mode>    		Set query mode: first | third | full")
		fmt.Println("                     				 	first    - only use local data")
		fmt.Println("                     				 	third    - only query remote sources")
		fmt.Println("										full     - query local and remote sources (DEFAULT)")
		fmt.Println("  -a --addr <ip address>		IP address to query (v4 or v6)")
		fmt.Println("  -H --human				Print human-friendly summary rather than formatted JSON")
		fmt.Println("  -h --help                   Show this help message.")
		fmt.Println()
		fmt.Println("--------------------------------------------------------------------------------------------")
		//fmt.Println("NOTE: -i and --mode are mutually exclusive.")
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
