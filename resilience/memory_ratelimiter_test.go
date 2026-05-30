package resilience

import (
	"context"
	"testing"

	"github.com/ymhhh/goblocks/config"
)

func TestMemoryRateLimiterAllow(t *testing.T) {
	lim := NewMemoryRateLimiter()
	rule := LimitRule{RPS: 1, Burst: 1}

	ok, err := lim.Allow(context.Background(), "user:1", rule)
	if err != nil || !ok {
		t.Fatalf("first allow: ok=%v err=%v", ok, err)
	}
	ok, err = lim.Allow(context.Background(), "user:1", rule)
	if err != nil || ok {
		t.Fatalf("second allow should reject: ok=%v err=%v", ok, err)
	}

	ok, err = lim.Allow(context.Background(), "user:2", rule)
	if err != nil || !ok {
		t.Fatalf("different key should allow: ok=%v err=%v", ok, err)
	}
}

func TestBuildRateLimitRulesLegacy(t *testing.T) {
	cfg := config.RateLimitConfig{RPS: 50, Burst: 80}
	global, _, _, _ := BuildRateLimitRules(cfg)
	if global.RPS != 50 || global.Burst != 80 {
		t.Fatalf("expected 50/80, got %v", global)
	}
}
