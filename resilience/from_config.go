package resilience

import (
	"github.com/sony/gobreaker"
	"github.com/ymhhh/goblocks/config"
)

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

// NewPolicyFromConfig creates a Policy from application config.
func NewPolicyFromConfig(cfg config.ResilienceConfig, opts ...PolicyOption) *Policy {
	build := &policyBuildConfig{}
	for _, opt := range opts {
		opt(build)
	}

	breaker := NewBreaker(BreakerSettings{
		Name:                "default",
		MaxRequests:         cfg.Breaker.MaxRequests,
		ConsecutiveFailures: cfg.Breaker.ConsecutiveFailures,
		Interval:            cfg.Breaker.Interval,
		Timeout:             cfg.Breaker.Timeout,
		OnStateChange:       build.onBreakerStateChange,
	})
	limiter := NewLimiter(cfg.RateLimit.RPS, cfg.RateLimit.Burst)
	return NewPolicy(breaker, limiter)
}
