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
	if cfg.Resilience.RateLimit.RPS != 100 {
		t.Fatalf("expected rps 100, got %f", cfg.Resilience.RateLimit.RPS)
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
log:
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
	if cfg.Resilience.RateLimit.RPS != 50 {
		t.Fatalf("expected rps 50, got %f", cfg.Resilience.RateLimit.RPS)
	}
	if cfg.AI.Model != "llama3" {
		t.Fatalf("expected llama3, got %s", cfg.AI.Model)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("expected debug, got %s", cfg.Log.Level)
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
