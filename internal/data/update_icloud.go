package data

import (
	"bufio"
	"bytes"

	"github.com/tlop503/ipcheq2/internal/iputils"
)

var prefixes_url = "https://mask-api.icloud.com/egress-ip-ranges.csv"

func updateiCloud(dataPath string, hashPath string) error {
	// check our local hash of remote data against remote data
	file, hash, err := bufferIfHashDiffers(prefixes_url, hashPath)
	if err != nil || (hash == "" && file == nil) {
		return err
	}

	// parse write IPs + cidr
	scanner := bufio.NewScanner(bytes.NewReader(file))
	IPs := iputils.DataToIPNetSlice(scanner)
	//IPs = iputils.Compact(IPs)
	err = WriteNormalizedIPNets(IPs, dataPath)
	if err != nil {
		return err
	}
	err = writeHashToFile(hash, hashPath)
	return nil
}
