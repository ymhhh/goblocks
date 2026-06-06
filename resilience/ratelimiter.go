package resilience

import (
	"context"
	"errors"
)

// ErrRateLimited is returned when the rate limiter rejects a request.
var ErrRateLimited = errors.New("rate limit exceeded")

// Scope identifies which rate-limit layer rejected a request.
type Scope string

const (
	ScopeGlobal Scope = "global"
	ScopeUser   Scope = "user"
	ScopeRoute  Scope = "route"
)

// LimitRule describes token-bucket parameters for one limit key.
type LimitRule struct {
	RPS   float64
	Burst int
}

// RateLimiter checks allowance for a logical key (global, user:id, route:POST:/path).
type RateLimiter interface {
	Allow(ctx context.Context, key string, rule LimitRule) (bool, error)
}

// KeyFunc extracts a rate-limit key from request context (HTTP/gRPC adapters provide wrappers).
type KeyFunc func(ctx context.Context) string
