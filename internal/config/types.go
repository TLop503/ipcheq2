package config

// file name to file path mapping
type Source struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// wrapper type for future config expansion
type Config struct {
	Sources []Source `yaml:"sources"`
}

func defaultConfig() Config {
	return Config{
		Sources: []Source{
			{Name: "Cyberghost", Path: "data/cyberghost.txt"},
			{Name: "Express VPN", Path: "data/express.txt"},
			{Name: "Mullvad", Path: "data/mullvad.txt"},
			{Name: "Nord VPN", Path: "data/nord.txt"},
			{Name: "PIA", Path: "data/pia.txt"},
			{Name: "Proton VPN", Path: "data/proton.txt"},
			{Name: "Surfshark", Path: "data/surfshark.txt"},
			{Name: "Torguard", Path: "data/torguard.txt"},
			{Name: "Tunnelbear", Path: "data/tunnelbear.txt"},
			{Name: "Tor Exit Nodes", Path: "data/torbulkexitlist.txt"},
			{Name: "iCloud Private Relay", Path: "data/icloud.txt"},
			{Name: "Hide.Me VPN", Path: "data/hide_me.txt"},
			{Name: "Generic VPN from ASN Data V4", Path: "data/vpn_by_asn.txt"},
			{Name: "Generic VPN from ASN Data V6", Path: "data/vpn_by_asnv6.txt"},
		},
	}
}
