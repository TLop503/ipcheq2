package cli

import (
	"flag"
	"os"
	"testing"
)

// resetFlags tears down and recreates the flag.CommandLine between tests,
// since flag.Parse can only be called once per FlagSet.
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	mode = ""
	query = ""
	help = false
}

func TestDefaultMode(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2"}

	cfg, err := InitFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != ModeWebUI {
		t.Errorf("expected ModeWebUI, got %v", cfg.Mode)
	}
}

func TestModeWebUI(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2", "--mode", "webui"}

	cfg, err := InitFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != ModeWebUI {
		t.Errorf("expected ModeWebUI, got %v", cfg.Mode)
	}
}

func TestModeAPI(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2", "--mode", "api"}

	cfg, err := InitFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != ModeAPI {
		t.Errorf("expected ModeAPI, got %v", cfg.Mode)
	}
}

func TestModeHeadless(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2", "--mode", "headless"}

	cfg, err := InitFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != ModeHeadless {
		t.Errorf("expected ModeHeadless, got %v", cfg.Mode)
	}
}

func TestUnknownMode(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2", "--mode", "blah"}

	_, err := InitFlags()
	if err == nil {
		t.Error("expected error for unknown mode, got nil")
	}
}
