package resilience

import (
	"errors"
	"testing"
	"time"

	"github.com/ymhhh/goblocks/config"
)

func TestBreakerConsecutiveFailuresFromConfig(t *testing.T) {
	cfg := config.Default()
	cfg.Resilience.Breaker.ConsecutiveFailures = 2

	p, err := NewPolicyFromConfig(cfg.Resilience)
	if err != nil {
		t.Fatal(err)
	}
	fail := func() error {
		return p.ExecuteVoid(func() error {
			return errors.New("downstream error")
		})
	}

	for i := 0; i < 2; i++ {
		_ = fail()
	}

	err = p.ExecuteVoid(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected circuit open after 2 failures, got %v", err)
	}
}

func TestBreakerDefaultConsecutiveFailures(t *testing.T) {
	breaker := NewBreaker(BreakerSettings{
		Name:     "test",
		Interval: time.Second,
		Timeout:  time.Second,
	})
	if breaker == nil {
		t.Fatal("expected breaker")
	}
}
