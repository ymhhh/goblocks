// Package middleware provides Gin middleware for goblocks HTTP servers.
//
// Rate limiting middleware (see docs/architecture.md):
//
//   - GlobalRateLimit: L1 service-wide (mounted by app.Run)
//   - BreakerCheck: reject when circuit breaker is open
//   - UserRateLimit: L2 per-user (mount in infrastructure after auth)
//   - RouteRateLimit: L3 per-route when rules exist in config
//   - RateLimitByKey: custom key/rule for L2 or L3
//
// User identity for L2: set via GinContextWithUserID or resilience.ContextWithUserID
// on the request context before UserRateLimit runs.
package middleware
