// Package app orchestrates HTTP, gRPC, AI client, and graceful shutdown.
//
// Run mounts framework middleware: Recovery, metrics, tracing (optional),
// L1 GlobalRateLimit, BreakerCheck, and L3 RouteRateLimit when config routes exist.
// L2 user rate limits are registered in WithHTTP/WithGRPC from business infrastructure.
package app
