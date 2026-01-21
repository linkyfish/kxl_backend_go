package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_WithConfigFileInRepoRoot(t *testing.T) {
	// Load() expects to find config/config.yaml relative to the working directory.
	// When running `go test ./...`, the CWD for this package is internal/config, so
	// we chdir to the repository root to exercise YAML parsing.
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir to repo root (%s): %v", root, err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	// Ensure a valid port even if the environment is polluted.
	_ = os.Setenv("SERVER_PORT", "8787")
	t.Cleanup(func() { _ = os.Unsetenv("SERVER_PORT") })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Session.Prefix == "" {
		t.Fatalf("expected non-empty session prefix")
	}
}

