package resilience

import (
	"github.com/sony/gobreaker"

	"github.com/ymhhh/goblocks/config"
)

// NewPolicyFromConfig creates a Policy from application config.
func NewPolicyFromConfig(cfg config.ResilienceConfig, opts ...PolicyOption) (*Policy, error) {
	build := &policyBuildConfig{}
	for _, opt := range opts {
		opt(build)
	}

	var breaker *gobreaker.CircuitBreaker
	if cfg.Breaker.Enabled {
		breaker = NewBreaker(BreakerSettings{
			Name:                "default",
			MaxRequests:         cfg.Breaker.MaxRequests,
			ConsecutiveFailures: cfg.Breaker.ConsecutiveFailures,
			Interval:            cfg.Breaker.Interval,
			Timeout:             cfg.Breaker.Timeout,
			OnStateChange:       build.onBreakerStateChange,
		})
	}

	backend, err := NewRateLimiterBackend(cfg.RateLimit)
	if err != nil {
		return nil, err
	}
	globalRule, userRule, userEnabled, routes := BuildRateLimitRules(cfg.RateLimit)

	return &Policy{
		Breaker: breaker,
		RateLimits: RateLimits{
			Backend:     backend,
			GlobalRule:  globalRule,
			GlobalKey:   GlobalKey("default"),
			UserRule:    userRule,
			UserEnabled: userEnabled,
			RouteRules:  routes,
		},
	}, nil
}
