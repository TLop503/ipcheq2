package cli

import (
	"flag"
	"fmt"
	"net/netip"
	"os"
)

func InitFlags() (Config, error) {
	flag.StringVar(&mode, "mode", "", modeMsg)
	flag.StringVar(&query, "i", "", queryMsg)
	flag.BoolVar(&help, "h", false, helpMsg)
	flag.BoolVar(&help, "help", false, helpMsg)

	flag.Parse()

	if help {
		fmt.Println("Welcome to ipcheq2.")
		fmt.Println("Invoke without arguments to launch the web UI.")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("Optional flags:")
		//fmt.Println("  -i <ip>          Query a single IP and exit")
		fmt.Println("  --mode <mode>    Set serving mode: webui | api | headless")
		fmt.Println("                     webui    - serves the web UI only (default)")
		fmt.Println("                     api      - serves web UI and exposes API")
		fmt.Println("                     headless - exposes API only, no web UI")
		fmt.Println()
		fmt.Println("-----------------------------------------------------------------")
		//fmt.Println("NOTE: -i and --mode are mutually exclusive.")
		os.Exit(0)
	}

	// if user set mode AND query
	if mode != "" && query != "" {
		return Config{}, fmt.Errorf("-i and --mode are mutually exclusive")
	}

	// otherwise if user only set query
	if query != "" {
		if _, err := netip.ParseAddr(query); err != nil {
			return Config{}, fmt.Errorf("invalid IP address %q: %w", query, err)
		}
	}

	switch {
	case query != "":
		return Config{Mode: ModeQuery, QueryIP: query}, nil
	case mode == "api":
		return Config{Mode: ModeAPI}, nil
	case mode == "headless":
		return Config{Mode: ModeHeadless}, nil
	case mode == "" || mode == "webui":
		return Config{Mode: ModeWebUI}, nil
	default:
		return Config{}, fmt.Errorf("unknown mode %q: must be webui, api, or headless", mode)
	}
}

// InitCliFlags parses arguments for the cli mode.
func InitCliFlags() (CliConfig, error) {
	flag.StringVar(&cliMode, "mode", "", "")
	flag.StringVar(&cliMode, "m", "", "")
	flag.StringVar(&query, "a", "127.0.0.1", "")
	flag.StringVar(&query, "addr", "", "")
	flag.BoolVar(&help, "h", false, helpMsg)
	flag.BoolVar(&help, "help", false, helpMsg)

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
	case cliMode == "first":
		return CliConfig{Mode: ModeFirst, QueryIP: addr}, nil
	case cliMode == "third":
		return CliConfig{Mode: ModeThird, QueryIP: addr}, nil
	case cliMode == "" || cliMode == "full":
		return CliConfig{Mode: ModeFull, QueryIP: addr}, nil
	default:
		return CliConfig{}, fmt.Errorf("unknown mode %q: must be first, third, or full", cliMode)
	}
}
