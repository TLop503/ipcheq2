package vpnid

import (
	"fmt"
	"github.com/yl2chen/cidranger"
	"net"
	"net/netip"
)

// ConfigEntry represents one line in the config file: name → path
type ConfigEntry struct {
	Name string
	Path string
}

type TreeEntry struct {
	Prefix   netip.Prefix
	Provider string
}

// Network implements RangerEntry using net.IPNet
func (t TreeEntry) Network() net.IPNet {
	return net.IPNet{
		IP:   t.Prefix.Addr().AsSlice(),                               // starting IP
		Mask: net.CIDRMask(t.Prefix.Bits(), t.Prefix.Addr().BitLen()), // mask
	}
}

var VpnIDRanger cidranger.Ranger

// InitializeVpnID initializes the VPN identification ranger from config file
func InitializeVpnID() {
	ranger, err := initialize("vpnid_config.txt")
	if err != nil {
		panic("Failed to initialize VPN ID: " + err.Error())
	}
	VpnIDRanger = ranger
	fmt.Println("VPN ID ranger initialized successfully")
}
