package resilience

import (
	"context"
	"errors"
	"strings"

	"github.com/sony/gobreaker"
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// RateLimits holds layered rate limit state for a Policy.
type RateLimits struct {
	Backend     RateLimiter
	GlobalRule  LimitRule
	GlobalKey   string
	UserRule    LimitRule
	UserEnabled bool
	RouteRules  map[string]LimitRule
}

// Policy combines rate limiting and circuit breaking.
type Policy struct {
	Breaker    *gobreaker.CircuitBreaker
	Limiter    *Limiter // deprecated: single-bucket in-memory helper for tests
	RateLimits RateLimits
}

// NewPolicy creates a Policy from optional breaker and legacy limiter.
func NewPolicy(breaker *gobreaker.CircuitBreaker, limiter *Limiter) *Policy {
	p := &Policy{
		Breaker: breaker,
		Limiter: limiter,
	}
	if limiter != nil {
		p.RateLimits.Backend = &legacyLimiterAdapter{limiter: limiter}
		p.RateLimits.GlobalRule = LimitRule{RPS: 100, Burst: 200}
		p.RateLimits.GlobalKey = GlobalKey("")
	}
	return p
}

type legacyLimiterAdapter struct {
	limiter *Limiter
}

func (a *legacyLimiterAdapter) Allow(_ context.Context, _ string, _ LimitRule) (bool, error) {
	return a.limiter.Allow(), nil
}

// Allow checks the global (L1) rate limit for backward compatibility.
func (p *Policy) Allow() error {
	return p.AllowGlobal(context.Background())
}

// AllowGlobal checks L1 service-wide rate limit.
func (p *Policy) AllowGlobal(ctx context.Context) error {
	if p == nil {
		return nil
	}
	return p.allow(ctx, p.RateLimits.GlobalKey, p.RateLimits.GlobalRule, ScopeGlobal)
}

// AllowUser checks L2 per-user rate limit when enabled.
func (p *Policy) AllowUser(ctx context.Context, userKey string) error {
	if !p.RateLimits.UserEnabled || p.RateLimits.Backend == nil {
		return nil
	}
	if userKey == "" {
		userKey = UserKeyFromContext(ctx)
	}
	return p.allow(ctx, userKey, p.RateLimits.UserRule, ScopeUser)
}

// AllowRoute checks L3 per-route rate limit when a rule exists.
func (p *Policy) AllowRoute(ctx context.Context, method, path string) error {
	if p.RateLimits.Backend == nil || len(p.RateLimits.RouteRules) == 0 {
		return nil
	}
	key := strings.ToUpper(method) + ":" + path
	rule, ok := p.RateLimits.RouteRules[key]
	if !ok {
		return nil
	}
	return p.allow(ctx, RouteKey(method, path), rule, ScopeRoute)
}

func (p *Policy) allow(ctx context.Context, key string, rule LimitRule, _ Scope) error {
	if p == nil || p.RateLimits.Backend == nil {
		return nil
	}
	if key == "" {
		key = GlobalKey("")
	}
	ok, err := p.RateLimits.Backend.Allow(ctx, key, rule)
	if err != nil {
		return err
	}
	if !ok {
		return ErrRateLimited
	}
	return nil
}

// Execute runs fn through the circuit breaker if configured.
func (p *Policy) Execute(fn func() (any, error)) (any, error) {
	if p == nil || p.Breaker == nil {
		return fn()
	}
	result, err := p.Breaker.Execute(fn)
	if err != nil {
		if err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests {
			return nil, ErrCircuitOpen
		}
		return nil, err
	}
	return result, nil
}

// ExecuteVoid runs a void function through the circuit breaker.
func (p *Policy) ExecuteVoid(fn func() error) error {
	_, err := p.Execute(func() (any, error) {
		return nil, fn()
	})
	return err
}
