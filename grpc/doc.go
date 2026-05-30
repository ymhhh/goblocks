// Package grpc provides gRPC server and client wrappers.
//
// Resilience interceptors live in grpc/interceptors:
//
//   - UnaryServerInterceptor: L1 global rate limit + circuit breaker (app default)
//   - UserUnaryServerInterceptor: L2 per-user (chain in infrastructure)
//   - RouteUnaryServerInterceptor: L3 per-method when configured
//   - UnaryClientInterceptor: outbound L1 + breaker for clients
//
// gRPC user identity for L2: incoming metadata key x-user-id.
package grpc
