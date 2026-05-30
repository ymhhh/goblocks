package config

import (
	"fmt"
	"os"
	"time"

	commonconfig "github.com/ymhhh/go-common/config"
)

const envPrefix = "GOBLOCKS_"

// Config holds the full application configuration.
type Config struct {
	Server     ServerConfig     `yaml:"server" json:"server"`
	Resilience ResilienceConfig `yaml:"resilience" json:"resilience"`
	AI         AIConfig         `yaml:"ai" json:"ai"`
	Logger     LoggerConfig     `yaml:"logger" json:"logger"`
	Metrics    MetricsConfig    `yaml:"metrics" json:"metrics"`

	source commonconfig.Config
}

// ServerConfig holds HTTP and gRPC server settings.
type ServerConfig struct {
	HTTP HTTPConfig `yaml:"http" json:"http"`
	GRPC GRPCConfig `yaml:"grpc" json:"grpc"`
}

// HTTPConfig holds HTTP server settings.
type HTTPConfig struct {
	Addr   string       `yaml:"addr" json:"addr"`
	TLS    TLSConfig    `yaml:"tls" json:"tls"`
	H3     H3Config     `yaml:"h3" json:"h3"`
	Health HealthConfig `yaml:"health" json:"health"`
}

// HealthConfig holds liveness/readiness probe settings.
type HealthConfig struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	LivenessPath  string `yaml:"liveness_path" json:"liveness_path"`
	ReadinessPath string `yaml:"readiness_path" json:"readiness_path"`
}

// TLSConfig holds TLS certificate settings.
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	CertFile string `yaml:"cert_file" json:"cert_file"`
	KeyFile  string `yaml:"key_file" json:"key_file"`
}

// H3Config holds HTTP/3 settings.
type H3Config struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Addr    string `yaml:"addr" json:"addr"`
}

// GRPCConfig holds gRPC server settings.
type GRPCConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Addr    string `yaml:"addr" json:"addr"`
}

// ResilienceConfig holds circuit breaker and rate limit settings.
type ResilienceConfig struct {
	Breaker   BreakerConfig   `yaml:"breaker" json:"breaker"`
	RateLimit RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
}

// BreakerConfig holds circuit breaker settings.
type BreakerConfig struct {
	MaxRequests          uint32        `yaml:"max_requests" json:"max_requests"`
	ConsecutiveFailures  uint32        `yaml:"consecutive_failures" json:"consecutive_failures"`
	Interval             time.Duration `yaml:"interval" json:"interval"`
	Timeout              time.Duration `yaml:"timeout" json:"timeout"`
}

// RateLimitConfig holds rate limiter settings.
type RateLimitConfig struct {
	RPS   float64 `yaml:"rps" json:"rps"`
	Burst int     `yaml:"burst" json:"burst"`
}

// AIConfig holds OpenAI-compatible API settings.
type AIConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	BaseURL string `yaml:"base_url" json:"base_url"`
	APIKey  string `yaml:"api_key" json:"api_key"`
	Model   string `yaml:"model" json:"model"`
}

// MetricsConfig holds Prometheus metrics settings.
type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	Path      string `yaml:"path" json:"path"`
	Addr      string `yaml:"addr" json:"addr"`
	AuthToken string `yaml:"auth_token" json:"auth_token"`
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
				Health: HealthConfig{
					Enabled:       true,
					LivenessPath:  "/health",
					ReadinessPath: "/ready",
				},
			},
			GRPC: GRPCConfig{
				Enabled: true,
				Addr:    ":9090",
			},
		},
		Resilience: ResilienceConfig{
			Breaker: BreakerConfig{
				MaxRequests:         3,
				ConsecutiveFailures: 3,
				Interval:            60 * time.Second,
				Timeout:             30 * time.Second,
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
		Logger: LoggerConfig{
			Level:  "info",
			Format: "text",
			Output: "stderr",
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
		tree, err := commonconfig.Load(path)
		if err != nil {
			return nil, fmt.Errorf("load config file: %w", err)
		}
		if err := tree.Object(cfg); err != nil {
			return nil, fmt.Errorf("decode config: %w", err)
		}
		cfg.source = tree
	}

	applyEnvOverrides(cfg)
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
	if v := os.Getenv(envPrefix + "LOGGER_LEVEL"); v != "" {
		cfg.Logger.Level = v
	} else if v := os.Getenv(envPrefix + "LOG_LEVEL"); v != "" {
		cfg.Logger.Level = v
	}
	if v := os.Getenv(envPrefix + "METRICS_ENABLED"); v != "" {
		cfg.Metrics.Enabled = v == "true" || v == "1"
	}
}
