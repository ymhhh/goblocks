package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const envPrefix = "GOBLOCKS_"

// Config holds the full application configuration.
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Resilience ResilienceConfig `yaml:"resilience"`
	AI         AIConfig         `yaml:"ai"`
	Log        LogConfig        `yaml:"log"`
	Metrics    MetricsConfig    `yaml:"metrics"`
}

// ServerConfig holds HTTP and gRPC server settings.
type ServerConfig struct {
	HTTP HTTPConfig `yaml:"http"`
	GRPC GRPCConfig `yaml:"grpc"`
}

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	Addr string    `yaml:"addr"`
	TLS  TLSConfig `yaml:"tls"`
	H3   H3Config  `yaml:"h3"`
}

// TLSConfig holds TLS certificate settings.
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// H3Config holds HTTP/3 settings.
type H3Config struct {
	Enabled bool   `yaml:"enabled"`
	Addr    string `yaml:"addr"`
}

// GRPCConfig holds gRPC server settings.
type GRPCConfig struct {
	Enabled bool   `yaml:"enabled"`
	Addr    string `yaml:"addr"`
}

// ResilienceConfig holds circuit breaker and rate limit settings.
type ResilienceConfig struct {
	Breaker   BreakerConfig   `yaml:"breaker"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

// BreakerConfig holds circuit breaker settings.
type BreakerConfig struct {
	MaxRequests uint32        `yaml:"max_requests"`
	Interval    time.Duration `yaml:"interval"`
	Timeout     time.Duration `yaml:"timeout"`
}

// RateLimitConfig holds rate limiter settings.
type RateLimitConfig struct {
	RPS   float64 `yaml:"rps"`
	Burst int     `yaml:"burst"`
}

// AIConfig holds OpenAI-compatible API settings.
type AIConfig struct {
	Enabled bool   `yaml:"enabled"`
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level string `yaml:"level"`
}

// MetricsConfig holds Prometheus metrics settings.
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			HTTP: HTTPConfig{
				Addr: ":8080",
				TLS: TLSConfig{
					Enabled: false,
				},
				H3: H3Config{
					Enabled: false,
					Addr:    ":8443",
				},
			},
			GRPC: GRPCConfig{
				Enabled: true,
				Addr:    ":9090",
			},
		},
		Resilience: ResilienceConfig{
			Breaker: BreakerConfig{
				MaxRequests: 3,
				Interval:    60 * time.Second,
				Timeout:     30 * time.Second,
			},
			RateLimit: RateLimitConfig{
				RPS:   100,
				Burst: 200,
			},
		},
		AI: AIConfig{
			Enabled: false,
			BaseURL: "https://api.openai.com/v1",
			APIKey:  "",
			Model:   "gpt-4o-mini",
		},
		Log: LogConfig{
			Level: "info",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
}

// Load reads configuration from a YAML file and applies environment overrides.
func Load(path string) (*Config, error) {
	cfg := Default()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config file: %w", err)
		}
	}

	applyEnvOverrides(cfg)
	expandEnvVars(cfg)
	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv(envPrefix + "HTTP_ADDR"); v != "" {
		cfg.Server.HTTP.Addr = v
	}
	if v := os.Getenv(envPrefix + "GRPC_ADDR"); v != "" {
		cfg.Server.GRPC.Addr = v
	}
	if v := os.Getenv(envPrefix + "AI_API_KEY"); v != "" {
		cfg.AI.APIKey = v
	}
	if v := os.Getenv(envPrefix + "AI_BASE_URL"); v != "" {
		cfg.AI.BaseURL = v
	}
	if v := os.Getenv(envPrefix + "LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv(envPrefix + "METRICS_ENABLED"); v != "" {
		cfg.Metrics.Enabled = v == "true" || v == "1"
	}
}

func expandEnvVars(cfg *Config) {
	cfg.AI.APIKey = expandString(cfg.AI.APIKey)
	cfg.AI.BaseURL = expandString(cfg.AI.BaseURL)
}

func expandString(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		key := s[2 : len(s)-1]
		if val := os.Getenv(key); val != "" {
			return val
		}
	}
	return os.ExpandEnv(s)
}
