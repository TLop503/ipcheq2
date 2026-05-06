package queries

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/netip"
)

// PrettyPrint nicely dumps first-party query results to the terminal
// to avoid circular dependencies, we take query type as int rather than iota name
// qT 0 = First
// qT 1 = Third
// qT 2 = Full
// Future versions may attempt to decouple flag config types from the cli parsing package,
// or general orchestration out of cli.init<thing> functions to avoid this hack
// and allow intuitive dependency usage
func PrettyPrint(ip netip.Addr, queryType int) {
	var res []byte
	var err error

	switch queryType {
	case 0:
		res, err = FirstPartyQuery(ip)
	case 1:
		res, err = ThirdPartyQuery(ip)
	case 2:
		res, err = FullQuery(ip)
	default:
		log.Fatalf("Unknown query type: %d", queryType)
	}
	if err != nil {
		log.Fatalf("Query type %d: %v\n", queryType, err)
	}

	pretty, err := prettyJSON(res)
	if err != nil {
		log.Fatalf("prettyJson: %v\n", err)
	}
	
	fmt.Println(string(pretty))
}

// prettyJSON indents JSON for readability
func prettyJSON(b []byte) ([]byte, error) {
	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "  "); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
