package ipcheq2

import "embed"

// BundledDataFS contains the VPN ID config and data sources for portable binaries.
//
//go:embed vpnid_config.txt internal/data/*.txt
var BundledDataFS embed.FS
