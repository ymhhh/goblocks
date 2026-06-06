package resilience

import (
	"errors"
	"testing"
	"time"
)

func TestPolicyAllow(t *testing.T) {
	p := &Policy{
		RateLimits: RateLimits{
			Backend:    NewMemoryRateLimiter(),
			GlobalRule: LimitRule{RPS: 1, Burst: 1},
			GlobalKey:  GlobalKey(""),
		},
	}
	if err := p.Allow(); err != nil {
		t.Fatal(err)
	}
	if err := p.Allow(); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestPolicyExecute(t *testing.T) {
	p := &Policy{}
	result, err := p.Execute(func() (any, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.(string) != "ok" {
		t.Fatalf("expected ok, got %v", result)
	}
}

func TestBreakerOpensAfterFailures(t *testing.T) {
	breaker := NewBreaker(BreakerSettings{
		Name:        "test",
		MaxRequests: 1,
		Interval:    time.Second,
		Timeout:     time.Second,
	})
	p := &Policy{Breaker: breaker}

	fail := func() error {
		return p.ExecuteVoid(func() error {
			return errors.New("downstream error")
		})
	}

	for i := 0; i < 3; i++ {
		_ = fail()
	}

	err := p.ExecuteVoid(func() error {
		return nil
	})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected circuit open, got %v", err)
	}
}

func TestPolicyAllowRoute(t *testing.T) {
	p := &Policy{
		RateLimits: RateLimits{
			Backend: NewMemoryRateLimiter(),
			RouteRules: map[string]LimitRule{
				"POST:/ai/chat": {RPS: 1, Burst: 1},
			},
		},
	}

	if err := p.AllowRoute(t.Context(), "POST", "/ai/chat"); err != nil {
		t.Fatal(err)
	}
	if err := p.AllowRoute(t.Context(), "POST", "/ai/chat"); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
	if err := p.AllowRoute(t.Context(), "GET", "/other"); err != nil {
		t.Fatalf("unconfigured route should pass, got %v", err)
	}
}

func TestNilPolicy(t *testing.T) {
	var p *Policy
	if err := p.Allow(); err != nil {
		t.Fatal("nil policy should allow")
	}
	result, err := p.Execute(func() (any, error) {
		return 42, nil
	})
	if err != nil || result.(int) != 42 {
		t.Fatalf("unexpected result: %v, %v", result, err)
	}
}
