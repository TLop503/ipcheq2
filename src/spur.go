package src

import (
	"net"
	"strings"

	"ipcheq2/src/vpnid"
)

func checkVPN(ip string) (string, error) {
	// Parse IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "Invalid IP address", nil
	}

	// Create net.IPAddr for vpnid.Query
	addr := net.IPAddr{IP: parsedIP}

	// Query using vpnid
	result, err := vpnid.Query(addr, VpnIDRanger)
	if err != nil {
		return "VPN query failed", err
	}

	// Format result for consistency
	if strings.Contains(result, "not found") {
		return "Not VPN", nil
	}

	return result, nil
}
