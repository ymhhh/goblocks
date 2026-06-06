package resilience

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter implements RateLimiter using Redis (cluster-safe).
type RedisRateLimiter struct {
	limiter *redis_rate.Limiter
	client  *redis.Client
	prefix  string
}

// NewRedisRateLimiter connects to Redis and returns a distributed rate limiter.
func NewRedisRateLimiter(addr, keyPrefix string) (*RedisRateLimiter, error) {
	if addr == "" {
		return nil, fmt.Errorf("resilience: redis addr is required")
	}
	opt, err := redis.ParseURL(addr)
	if err != nil {
		opt = &redis.Options{Addr: addr}
	}
	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("resilience: redis ping: %w", err)
	}
	if keyPrefix == "" {
		keyPrefix = "goblocks:rl:"
	}
	return &RedisRateLimiter{
		limiter: redis_rate.NewLimiter(client),
		client:  client,
		prefix:  keyPrefix,
	}, nil
}

// Allow reports whether the request for key is allowed under rule.
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, rule LimitRule) (bool, error) {
	rps, burst := normalizeRule(rule)
	res, err := r.limiter.Allow(ctx, r.prefix+key, redis_rate.Limit{
		Rate:   int(rps),
		Burst:  burst,
		Period: time.Second,
	})
	if err != nil {
		return false, err
	}
	return res.Allowed > 0, nil
}

// Close releases the Redis client connection pool.
func (r *RedisRateLimiter) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}
