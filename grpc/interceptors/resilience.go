package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/ymhhh/goblocks/metrics"
	"github.com/ymhhh/goblocks/resilience"
)

const grpcProtocol = "grpc"

const metadataUserID = "x-user-id"

// UnaryServerInterceptor returns a gRPC unary server interceptor with L1 global rate limit and breaker.
func UnaryServerInterceptor(policy *resilience.Policy, m *metrics.Registry) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if policy == nil {
			return handler(ctx, req)
		}

		if err := policy.AllowGlobal(ctx); err != nil {
			return nil, rateLimitError(m, err, resilience.ScopeGlobal)
		}

		result, err := policy.Execute(func() (any, error) {
			return handler(ctx, req)
		})
		if err != nil {
			if err == resilience.ErrCircuitOpen {
				if m != nil {
					m.RecordCircuitBreakerRejected(grpcProtocol)
				}
				return nil, status.Error(codes.Unavailable, "circuit breaker is open")
			}
			return nil, err
		}
		return result, nil
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor with L1 global rate limit and breaker.
func StreamServerInterceptor(policy *resilience.Policy, m *metrics.Registry) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if policy == nil {
			return handler(srv, ss)
		}

		if err := policy.AllowGlobal(ss.Context()); err != nil {
			return rateLimitError(m, err, resilience.ScopeGlobal)
		}

		_, err := policy.Execute(func() (any, error) {
			return nil, handler(srv, ss)
		})
		if err == resilience.ErrCircuitOpen {
			if m != nil {
				m.RecordCircuitBreakerRejected(grpcProtocol)
			}
			return status.Error(codes.Unavailable, "circuit breaker is open")
		}
		return err
	}
}

// UserUnaryServerInterceptor applies L2 per-user rate limiting (requires x-user-id metadata or context user id).
func UserUnaryServerInterceptor(policy *resilience.Policy, m *metrics.Registry) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if policy == nil || !policy.RateLimits.UserEnabled {
			return handler(ctx, req)
		}
		ctx = userIDFromMetadata(ctx)
		if err := policy.AllowUser(ctx, ""); err != nil {
			return nil, rateLimitError(m, err, resilience.ScopeUser)
		}
		return handler(ctx, req)
	}
}

// UserStreamServerInterceptor applies L2 per-user rate limiting to streams.
func UserStreamServerInterceptor(policy *resilience.Policy, m *metrics.Registry) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if policy == nil || !policy.RateLimits.UserEnabled {
			return handler(srv, ss)
		}
		ctx := userIDFromMetadata(ss.Context())
		if err := policy.AllowUser(ctx, ""); err != nil {
			return rateLimitError(m, err, resilience.ScopeUser)
		}
		return handler(srv, ss)
	}
}

// RouteUnaryServerInterceptor applies L3 per-method rate limiting when configured.
func RouteUnaryServerInterceptor(policy *resilience.Policy, m *metrics.Registry) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if policy == nil || len(policy.RateLimits.RouteRules) == 0 {
			return handler(ctx, req)
		}
		if err := policy.AllowRoute(ctx, "", info.FullMethod); err != nil {
			return nil, rateLimitError(m, err, resilience.ScopeRoute)
		}
		return handler(ctx, req)
	}
}

// RouteStreamServerInterceptor applies L3 per-method rate limiting to streams.
func RouteStreamServerInterceptor(policy *resilience.Policy, m *metrics.Registry) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if policy == nil || len(policy.RateLimits.RouteRules) == 0 {
			return handler(srv, ss)
		}
		if err := policy.AllowRoute(ss.Context(), "", info.FullMethod); err != nil {
			return rateLimitError(m, err, resilience.ScopeRoute)
		}
		return handler(srv, ss)
	}
}

func userIDFromMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	vals := md.Get(metadataUserID)
	if len(vals) == 0 || vals[0] == "" {
		return ctx
	}
	return resilience.ContextWithUserID(ctx, vals[0])
}

func rateLimitError(m *metrics.Registry, err error, scope resilience.Scope) error {
	if err == resilience.ErrRateLimited {
		if m != nil {
			m.RecordRateLimitRejected(grpcProtocol, string(scope))
		}
		return status.Error(codes.ResourceExhausted, "rate limit exceeded")
	}
	return status.Error(codes.Unavailable, err.Error())
}

// UnaryClientInterceptor returns a gRPC unary client interceptor with resilience.
func UnaryClientInterceptor(policy *resilience.Policy, m *metrics.Registry) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if policy == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		if err := policy.AllowGlobal(ctx); err != nil {
			return rateLimitError(m, err, resilience.ScopeGlobal)
		}

		_, err := policy.Execute(func() (any, error) {
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})
		if err != nil {
			if err == resilience.ErrCircuitOpen {
				if m != nil {
					m.RecordCircuitBreakerRejected(grpcProtocol)
				}
				return status.Error(codes.Unavailable, "circuit breaker is open")
			}
		}
		return err
	}
}

// StreamClientInterceptor returns a gRPC stream client interceptor with resilience.
func StreamClientInterceptor(policy *resilience.Policy, m *metrics.Registry) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if policy == nil {
			return streamer(ctx, desc, cc, method, opts...)
		}

		if err := policy.AllowGlobal(ctx); err != nil {
			return nil, rateLimitError(m, err, resilience.ScopeGlobal)
		}

		result, err := policy.Execute(func() (any, error) {
			return streamer(ctx, desc, cc, method, opts...)
		})
		if err != nil {
			if err == resilience.ErrCircuitOpen {
				if m != nil {
					m.RecordCircuitBreakerRejected(grpcProtocol)
				}
				return nil, status.Error(codes.Unavailable, "circuit breaker is open")
			}
			return nil, err
		}
		return result.(grpc.ClientStream), nil
	}
}
