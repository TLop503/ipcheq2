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
	Update  bool
}

var (
	help   bool
	mode   string
	query  string
	update bool
)
