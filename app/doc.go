// Package app orchestrates HTTP, gRPC, AI client, and graceful shutdown.
//
// Run mounts framework middleware only: Recovery, metrics, tracing (optional),
// L1 GlobalRateLimit, and BreakerCheck. L2 user and L3 route rate limits are
// not mounted here; register them in WithHTTP/WithGRPC callbacks from
// business infrastructure after authentication.
package app
