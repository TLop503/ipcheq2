package cli

import "net/netip"

type RunMode int
type CliMode int

const (
	ModeWebUI RunMode = iota
	ModeAPI
	ModeHeadless
	ModeQuery
	ModeREPL
)

const (
	ModeFull CliMode = iota
	ModeFirst
	ModeThird
)

type Config struct {
	Mode    RunMode
	QueryIP string
}

type CliConfig struct {
	Mode    CliMode
	QueryIP netip.Addr
}

var (
	help    bool
	mode    string
	query   string
	cliMode string
)
