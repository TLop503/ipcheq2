package queries

import (
	"fmt"
	"github.com/tlop503/ipcheq2/internal"
	"github.com/tlop503/ipcheq2/internal/cli"
	"log"
	"net/netip"
)

func PrettyPrint(ip netip.Addr, queryType cli.CliMode) {
	var res []byte
	var err error

	switch queryType {
	case cli.ModeFirst:
		res, err = FirstPartyQuery(ip)
	case cli.ModeThird:
		res, err = ThirdPartyQuery(ip)
	default:
		res, err = FullQuery(ip)
	}

	if err != nil {
		log.Fatalf("FullQuery: %v\n", err)
	}
	pretty, err := internal.PrettyJSON(res)
	if err != nil {
		log.Fatalf("FullQuery: %v\n", err)
	}
	fmt.Println(string(pretty))
}
