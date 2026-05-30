package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Server.HTTP.Addr != ":8080" {
		t.Fatalf("expected :8080, got %s", cfg.Server.HTTP.Addr)
	}
	if cfg.Server.GRPC.Addr != ":9090" {
		t.Fatalf("expected :9090, got %s", cfg.Server.GRPC.Addr)
	}
	normalized := cfg.Resilience.RateLimit.Normalized()
	if normalized.Global.RPS != 100 {
		t.Fatalf("expected global rps 100, got %f", normalized.Global.RPS)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
server:
  http:
    addr: ":3000"
  grpc:
    enabled: false
    addr: ":4000"
resilience:
  rate_limit:
    rps: 50
    burst: 100
ai:
  enabled: true
  base_url: "http://localhost:11434/v1"
  model: "llama3"
logger:
  level: debug
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.HTTP.Addr != ":3000" {
		t.Fatalf("expected :3000, got %s", cfg.Server.HTTP.Addr)
	}
	if !cfg.Server.GRPC.Enabled == false {
		t.Fatal("expected grpc disabled")
	}
	normalized := cfg.Resilience.RateLimit.Normalized()
	if normalized.Global.RPS != 50 {
		t.Fatalf("expected global rps 50, got %f", normalized.Global.RPS)
	}
	if cfg.AI.Model != "llama3" {
		t.Fatalf("expected llama3, got %s", cfg.AI.Model)
	}
	if cfg.Logger.Level != "debug" {
		t.Fatalf("expected debug, got %s", cfg.Logger.Level)
	}
}

func TestEnvOverride(t *testing.T) {
	t.Setenv("GOBLOCKS_HTTP_ADDR", ":9999")
	t.Setenv("GOBLOCKS_AI_API_KEY", "test-key")

	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.HTTP.Addr != ":9999" {
		t.Fatalf("expected :9999, got %s", cfg.Server.HTTP.Addr)
	}
	if cfg.AI.APIKey != "test-key" {
		t.Fatalf("expected test-key, got %s", cfg.AI.APIKey)
	}
}

func TestLoggerLevelEnvOverride(t *testing.T) {
	t.Setenv("GOBLOCKS_LOGGER_LEVEL", "warn")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Logger.Level != "warn" {
		t.Fatalf("expected warn, got %s", cfg.Logger.Level)
	}
}

func TestLegacyLogLevelEnvOverride(t *testing.T) {
	t.Setenv("GOBLOCKS_LOG_LEVEL", "error")
	cfg, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Logger.Level != "error" {
		t.Fatalf("expected error, got %s", cfg.Logger.Level)
	}
}

func TestExpandEnvVar(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test")
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
ai:
  api_key: "${OPENAI_API_KEY}"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AI.APIKey != "sk-test" {
		t.Fatalf("expected sk-test, got %s", cfg.AI.APIKey)
	}
}

func TestBreakerDurationParsing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
resilience:
  breaker:
    max_requests: 5
    interval: 30s
    timeout: 10s
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Resilience.Breaker.MaxRequests != 5 {
		t.Fatalf("expected 5, got %d", cfg.Resilience.Breaker.MaxRequests)
	}
	if cfg.Resilience.Breaker.Interval != 30*time.Second {
		t.Fatalf("expected 30s, got %v", cfg.Resilience.Breaker.Interval)
	}
}

func TestLoadInclude(t *testing.T) {
	dir := t.TempDir()
	incPath := filepath.Join(dir, "inc.yaml")
	if err := os.WriteFile(incPath, []byte(`
server:
  http:
    addr: ":7000"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	mainPath := filepath.Join(dir, "config.yaml")
	content := `
#include inc.yaml
server:
  grpc:
    addr: ":7001"
`
	if err := os.WriteFile(mainPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.HTTP.Addr != ":7000" {
		t.Fatalf("expected :7000 from include, got %s", cfg.Server.HTTP.Addr)
	}
	if cfg.Server.GRPC.Addr != ":7001" {
		t.Fatalf("expected :7001, got %s", cfg.Server.GRPC.Addr)
	}
}
