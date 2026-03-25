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

func TestQueryFlag(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2", "-i", "1.2.3.4"}

	cfg, err := InitFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != ModeQuery {
		t.Errorf("expected ModeQuery, got %v", cfg.Mode)
	}
	if cfg.QueryIP != "1.2.3.4" {
		t.Errorf("expected QueryIP 1.2.3.4, got %q", cfg.QueryIP)
	}
}

func TestQueryIPv6(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2", "-i", "2001:db8::1"}

	cfg, err := InitFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != ModeQuery {
		t.Errorf("expected ModeQuery, got %v", cfg.Mode)
	}
	if cfg.QueryIP != "2001:db8::1" {
		t.Errorf("expected QueryIP 2001:db8::1, got %q", cfg.QueryIP)
	}
}

func TestQueryInvalidIP(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2", "-i", "not-an-ip"}

	_, err := InitFlags()
	if err == nil {
		t.Error("expected error for invalid IP, got nil")
	}
}

func TestMutualExclusion(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2", "-i", "1.2.3.4", "--mode", "api"}

	_, err := InitFlags()
	if err == nil {
		t.Error("expected error when both -i and --mode are set, got nil")
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

func TestQueryLoopback(t *testing.T) {
	resetFlags()
	os.Args = []string{"ipcheq2", "-i", "127.0.0.1"}

	cfg, err := InitFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.QueryIP != "127.0.0.1" {
		t.Errorf("expected 127.0.0.1, got %q", cfg.QueryIP)
	}
}
