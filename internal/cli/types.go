package cli

import "net/netip"

type RunMode int

const (
	ModeWebUI RunMode = iota
	ModeAPI
	ModeHeadless
	ModeQuery
)

type Config struct {
	Mode   RunMode
	Update bool
}

var (
	help   bool
	mode   string
	query  string
	update bool
)

type CliMode int

const (
	ModeFirst CliMode = iota
	ModeThird
	ModeFull
)

type CliConfig struct {
	Mode    CliMode
	QueryIP netip.Addr
}
