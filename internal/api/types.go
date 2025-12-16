package api

import (
	"fmt"
	"github.com/tlop503/ipcheq2/internal/vpnid"
	"github.com/yl2chen/cidranger"
)

// InitializeVpnID initializes the VPN identification ranger from config file
func InitializeVpnID() {
	ranger, err := vpnid.Initialize("vpnid_config.txt")
	if err != nil {
		panic("Failed to initialize VPN ID: " + err.Error())
	}
	VpnIDRanger = ranger
	fmt.Println("VPN ID ranger initialized successfully")
}

var VpnIDRanger cidranger.Ranger
