package cli

type RunMode int

const (
	ModeWebUI RunMode = iota
	ModeAPI
	ModeHeadless
	ModeQuery
	ModeREPL
)

type Config struct {
	Mode    RunMode
	QueryIP string
}

var (
	help  bool
	mode  string
	query string
)
