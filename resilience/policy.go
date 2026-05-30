package resilience

import (
	"errors"

	"github.com/sony/gobreaker"
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Policy combines rate limiting and circuit breaking.
type Policy struct {
	Breaker *gobreaker.CircuitBreaker
	Limiter *Limiter
}

// NewPolicy creates a Policy from optional breaker and limiter.
func NewPolicy(breaker *gobreaker.CircuitBreaker, limiter *Limiter) *Policy {
	return &Policy{
		Breaker: breaker,
		Limiter: limiter,
	}
}

// Allow checks rate limit before processing a request.
func (p *Policy) Allow() error {
	if p == nil || p.Limiter == nil {
		return nil
	}
	if !p.Limiter.Allow() {
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
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
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
