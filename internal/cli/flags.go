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
	flag.BoolVar(&update, "update", false, helpMsg)
	flag.BoolVar(&update, "u", false, helpMsg)

	flag.Parse()

	if help {
		fmt.Println("Welcome to ipcheq2.")
		fmt.Println("Invoke without arguments to launch the web UI.")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Println("Optional flags:")
		fmt.Println("  --mode <mode>    Set serving mode: webui | api | headless")
		fmt.Println("                     webui    - serves the web UI only (default)")
		fmt.Println("                     api      - serves web UI and exposes API")
		fmt.Println("                     headless - exposes API only, no web UI")
		fmt.Println("  --update -u      Update data sources.")
		fmt.Println("  --help -h        Show this help message.")
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
		return Config{Mode: ModeQuery, QueryIP: query, Update: update}, nil
	case mode == "api":
		return Config{Mode: ModeAPI, Update: update}, nil
	case mode == "headless":
		return Config{Mode: ModeHeadless, Update: update}, nil
	case mode == "" || mode == "webui":
		return Config{Mode: ModeWebUI, Update: update}, nil
	default:
		return Config{}, fmt.Errorf("unknown mode %q: must be webui, api, or headless", mode)
	}
}
