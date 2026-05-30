package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitLoggerWithFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
logger:
  level: warn
  format: text
  output: discard
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.InitLogger(); err != nil {
		t.Fatal(err)
	}
}

func TestInitLoggerWithoutFile(t *testing.T) {
	cfg := Default()
	cfg.Logger.Output = "discard"
	if err := cfg.InitLogger(); err != nil {
		t.Fatal(err)
	}
}

func TestInitLoggerNilConfig(t *testing.T) {
	var cfg *Config
	if err := cfg.InitLogger(); err == nil {
		t.Fatal("expected error for nil config")
	}
}
