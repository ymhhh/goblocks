package app

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// WithHTTPTracing registers optional HTTP middleware (e.g. otelgin.Middleware).
// Middleware runs after Recovery and metrics, before resilience.
func (a *App) WithHTTPTracing(middleware ...gin.HandlerFunc) *App {
	a.httpTracing = append(a.httpTracing, middleware...)
	return a
}

// WithGRPCTracing registers optional gRPC server options (e.g. otelgrpc stats handler).
func (a *App) WithGRPCTracing(opts ...grpc.ServerOption) *App {
	a.grpcTracing = append(a.grpcTracing, opts...)
	return a
}
