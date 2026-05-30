package resilience

import (
	"time"

	"github.com/sony/gobreaker"
)

// BreakerSettings configures a circuit breaker.
type BreakerSettings struct {
	Name        string
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
}

// NewBreaker creates a circuit breaker with the given settings.
func NewBreaker(s BreakerSettings) *gobreaker.CircuitBreaker {
	if s.Name == "" {
		s.Name = "default"
	}
	if s.MaxRequests == 0 {
		s.MaxRequests = 3
	}
	if s.Interval == 0 {
		s.Interval = 60 * time.Second
	}
	if s.Timeout == 0 {
		s.Timeout = 30 * time.Second
	}

	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        s.Name,
		MaxRequests: s.MaxRequests,
		Interval:    s.Interval,
		Timeout:     s.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	})
}
