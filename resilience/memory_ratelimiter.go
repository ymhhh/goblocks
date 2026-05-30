package resilience

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type bucketEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// MemoryRateLimiter implements RateLimiter with in-process keyed token buckets.
type MemoryRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucketEntry
}

// NewMemoryRateLimiter creates an in-memory keyed rate limiter.
func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		buckets: make(map[string]*bucketEntry),
	}
}

// Allow reports whether the request for key is allowed under rule.
func (m *MemoryRateLimiter) Allow(_ context.Context, key string, rule LimitRule) (bool, error) {
	rps, burst := normalizeRule(rule)
	l := m.getLimiter(key, rps, burst)
	return l.Allow(), nil
}

func (m *MemoryRateLimiter) getLimiter(key string, rps float64, burst int) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.buckets[key]
	if !ok || entry.limiter.Burst() != burst || entry.limiter.Limit() != rate.Limit(rps) {
		entry = &bucketEntry{
			limiter:  rate.NewLimiter(rate.Limit(rps), burst),
			lastSeen: time.Now(),
		}
		m.buckets[key] = entry
	} else {
		entry.lastSeen = time.Now()
	}
	return entry.limiter
}

func normalizeRule(rule LimitRule) (float64, int) {
	rps := rule.RPS
	if rps <= 0 {
		rps = 100
	}
	burst := rule.Burst
	if burst <= 0 {
		burst = int(rps * 2)
		if burst < 1 {
			burst = 1
		}
	}
	return rps, burst
}
