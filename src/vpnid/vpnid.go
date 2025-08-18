package vpnid

import (
	"errors"
	"net"
)

func ValidateConfig(path string) error {
	return nil
}

func Initialize(path string) error {
	return errors.New("Not implemented")
}

func Query(addr net.IPAddr) (string, error) {
	return "", errors.New("Not implemented")
}
