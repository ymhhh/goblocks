package resilience

import (
	"errors"

	"golang.org/x/time/rate"
)

// ErrRateLimited is returned when the rate limiter rejects a request.
var ErrRateLimited = errors.New("rate limit exceeded")

// Limiter wraps a token bucket rate limiter.
type Limiter struct {
	limiter *rate.Limiter
}

// NewLimiter creates a rate limiter with the given RPS and burst.
func NewLimiter(rps float64, burst int) *Limiter {
	if rps <= 0 {
		rps = 100
	}
	if burst <= 0 {
		burst = int(rps * 2)
	}
	return &Limiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
	}
}

// Allow checks if a request is allowed under the rate limit.
func (l *Limiter) Allow() bool {
	if l == nil || l.limiter == nil {
		return true
	}
	return l.limiter.Allow()
}

// Wait blocks until a request is allowed or returns an error.
func (l *Limiter) Wait() error {
	if l == nil || l.limiter == nil {
		return nil
	}
	if !l.limiter.Allow() {
		return ErrRateLimited
	}
	return nil
}
