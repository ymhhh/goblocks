package resilience

import (
	"github.com/ymhhh/goblocks/config"
)

// NewPolicyFromConfig creates a Policy from application config.
func NewPolicyFromConfig(cfg config.ResilienceConfig) *Policy {
	breaker := NewBreaker(BreakerSettings{
		Name:        "default",
		MaxRequests: cfg.Breaker.MaxRequests,
		Interval:    cfg.Breaker.Interval,
		Timeout:     cfg.Breaker.Timeout,
	})
	limiter := NewLimiter(cfg.RateLimit.RPS, cfg.RateLimit.Burst)
	return NewPolicy(breaker, limiter)
}
