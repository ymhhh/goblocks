package resilience

import (
	"fmt"
	"strings"

	"github.com/ymhhh/goblocks/config"
)

// NewRateLimiterBackend creates the configured rate limit backend.
func NewRateLimiterBackend(cfg config.RateLimitConfig) (RateLimiter, error) {
	backend := strings.ToLower(strings.TrimSpace(cfg.Backend))
	if backend == "" {
		backend = "memory"
	}
	switch backend {
	case "memory":
		return NewMemoryRateLimiter(), nil
	case "redis":
		return NewRedisRateLimiter(cfg.Redis.Addr, cfg.Redis.KeyPrefix)
	default:
		return nil, fmt.Errorf("resilience: unsupported rate_limit backend %q", cfg.Backend)
	}
}

// BuildRateLimitRules normalizes config into limit rules and route map.
func BuildRateLimitRules(cfg config.RateLimitConfig) (global, user LimitRule, userEnabled bool, routes map[string]LimitRule) {
	cfg = cfg.Normalized()

	global = LimitRule{RPS: cfg.Global.RPS, Burst: cfg.Global.Burst}
	userEnabled = cfg.User.Enabled
	user = LimitRule{RPS: cfg.User.DefaultRPS, Burst: cfg.User.Burst}

	routes = make(map[string]LimitRule)
	for _, r := range cfg.Routes {
		key := strings.ToUpper(r.Method) + ":" + r.Path
		routes[key] = LimitRule{RPS: r.RPS, Burst: r.Burst}
	}
	return global, user, userEnabled, routes
}
