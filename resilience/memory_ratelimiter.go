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
	mu              sync.Mutex
	buckets         map[string]*bucketEntry
	entryTTL        time.Duration
	cleanupInterval time.Duration
	done            chan struct{}
}

// MemoryRateLimiterOption configures a MemoryRateLimiter.
type MemoryRateLimiterOption func(*MemoryRateLimiter)

// WithEntryTTL sets how long a bucket lives without being accessed.
func WithEntryTTL(d time.Duration) MemoryRateLimiterOption {
	return func(m *MemoryRateLimiter) {
		if d > 0 {
			m.entryTTL = d
		}
	}
}

// WithCleanupInterval sets how often stale buckets are evicted.
func WithCleanupInterval(d time.Duration) MemoryRateLimiterOption {
	return func(m *MemoryRateLimiter) {
		if d > 0 {
			m.cleanupInterval = d
		}
	}
}

const (
	defaultEntryTTL        = 5 * time.Minute
	defaultCleanupInterval = 1 * time.Minute
)

// NewMemoryRateLimiter creates an in-memory keyed rate limiter with background cleanup.
func NewMemoryRateLimiter(opts ...MemoryRateLimiterOption) *MemoryRateLimiter {
	m := &MemoryRateLimiter{
		buckets:         make(map[string]*bucketEntry),
		entryTTL:        defaultEntryTTL,
		cleanupInterval: defaultCleanupInterval,
		done:            make(chan struct{}),
	}
	for _, opt := range opts {
		opt(m)
	}
	go m.cleanup()
	return m
}

// Allow reports whether the request for key is allowed under rule.
func (m *MemoryRateLimiter) Allow(_ context.Context, key string, rule LimitRule) (bool, error) {
	rps, burst := normalizeRule(rule)
	l := m.getLimiter(key, rps, burst)
	return l.Allow(), nil
}

// Close stops the background cleanup goroutine.
func (m *MemoryRateLimiter) Close() error {
	select {
	case <-m.done:
	default:
		close(m.done)
	}
	return nil
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

func (m *MemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			m.evictStale()
		}
	}
}

func (m *MemoryRateLimiter) evictStale() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for key, entry := range m.buckets {
		if now.Sub(entry.lastSeen) > m.entryTTL {
			delete(m.buckets, key)
		}
	}
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
