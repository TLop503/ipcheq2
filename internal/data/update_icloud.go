package data

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/tlop503/ipcheq2/internal/iputils"
)

var prefixes_url = "https://mask-api.icloud.com/egress-ip-ranges.csv"

func UpdateiCloud(dataPath string, hashPath string) error {
	// check our local hash of remote data against remote data
	file, hash, err := bufferIfHashDiffers(prefixes_url, hashPath)
	if err != nil {
		return err
	}

	if hash == "" || file == nil {
		return errors.New("no error from fetching remote, but hash or file are empty!")
	}

	// parse write IPs + cidr
	scanner := bufio.NewScanner(bytes.NewReader(file))
	IPs := iputils.DataToIPNetSlice(scanner)
	err = WriteNormalizedIPNets(IPs, dataPath)
	if err != nil {
		return err
	}
	return nil
}
