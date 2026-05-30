// Package resilience provides circuit breaking and layered rate limiting.
//
// Rate limiting uses three scopes (see docs/architecture.md):
//
//   - L1 global: service/cluster protection (app mounts by default)
//   - L2 user: per-user quotas (business infrastructure mounts after auth)
//   - L3 route: per-API limits (business route groups or config routes)
//
// Backends implement RateLimiter: MemoryRateLimiter (in-process) and
// RedisRateLimiter (distributed, GCRA via go-redis/redis_rate).
//
// Policy combines Breaker, RateLimits bundle, and AllowGlobal/AllowUser/AllowRoute.
// NewPolicyFromConfig builds from config.ResilienceConfig.
package resilience
