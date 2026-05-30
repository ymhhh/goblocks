package resilience

import "github.com/sony/gobreaker"

// PolicyOption configures policy construction.
type PolicyOption func(*policyBuildConfig)

type policyBuildConfig struct {
	onBreakerStateChange func(name string, from, to gobreaker.State)
}

// WithBreakerStateChange registers a callback when circuit breaker state changes.
func WithBreakerStateChange(fn func(name string, from, to gobreaker.State)) PolicyOption {
	return func(cfg *policyBuildConfig) {
		cfg.onBreakerStateChange = fn
	}
}
