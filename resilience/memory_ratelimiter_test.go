package resilience

import (
	"context"
	"testing"
	"time"

	"github.com/ymhhh/goblocks/config"
)

func TestMemoryRateLimiterAllow(t *testing.T) {
	lim := NewMemoryRateLimiter()
	defer lim.Close()

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

func TestMemoryRateLimiterCleanup(t *testing.T) {
	lim := NewMemoryRateLimiter(
		WithEntryTTL(50*time.Millisecond),
		WithCleanupInterval(50*time.Millisecond),
	)
	defer lim.Close()

	rule := LimitRule{RPS: 100, Burst: 200}

	ok, _ := lim.Allow(context.Background(), "key1", rule)
	if !ok {
		t.Fatal("first allow should succeed")
	}

	time.Sleep(150 * time.Millisecond)

	lim.mu.Lock()
	_, exists := lim.buckets["key1"]
	lim.mu.Unlock()
	if exists {
		t.Fatal("stale bucket should have been evicted")
	}
}

func TestMemoryRateLimiterClose(t *testing.T) {
	lim := NewMemoryRateLimiter()
	if err := lim.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := lim.Close(); err != nil {
		t.Fatalf("second close should be idempotent: %v", err)
	}
}

func TestBuildRateLimitRulesLegacy(t *testing.T) {
	cfg := config.RateLimitConfig{RPS: 50, Burst: 80}
	global, _, _, _ := BuildRateLimitRules(cfg)
	if global.RPS != 50 || global.Burst != 80 {
		t.Fatalf("expected 50/80, got %v", global)
	}
}
