package config

import (
	"fmt"

	commonconfig "github.com/ymhhh/go-common/config"
	"github.com/ymhhh/go-common/logger"
	"gopkg.in/yaml.v3"
)

// LoggerConfig holds logging settings aligned with go-common/logger.
type LoggerConfig = logger.Config

// InitLogger initializes the global logger from the loaded configuration tree,
// or from the Logger struct when no config file was loaded.
func (c *Config) InitLogger() error {
	if c == nil {
		return fmt.Errorf("config: nil config")
	}
	if c.source != nil {
		return logger.InitGlobal(c.source, "logger")
	}
	return initLoggerFromStruct(c.Logger)
}

func initLoggerFromStruct(cfg LoggerConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("config: marshal logger: %w", err)
	}
	var subtree map[string]any
	if err := yaml.Unmarshal(data, &subtree); err != nil {
		return fmt.Errorf("config: unmarshal logger: %w", err)
	}
	opts := commonconfig.Options(map[string]any{
		"logger": subtree,
	})
	tree := (&opts).ToConfig()
	return logger.InitGlobal(tree, "logger")
}
