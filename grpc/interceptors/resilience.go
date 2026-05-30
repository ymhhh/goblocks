package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ymhhh/goblocks/metrics"
	"github.com/ymhhh/goblocks/resilience"
)

const grpcProtocol = "grpc"

// UnaryServerInterceptor returns a gRPC unary server interceptor with resilience.
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

		if err := policy.Allow(); err != nil {
			if err == resilience.ErrRateLimited {
				if m != nil {
					m.RecordRateLimitRejected(grpcProtocol)
				}
				return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}
			return nil, status.Error(codes.Unavailable, err.Error())
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

		if err := policy.Allow(); err != nil {
			if err == resilience.ErrRateLimited {
				if m != nil {
					m.RecordRateLimitRejected(grpcProtocol)
				}
				return status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}
			return status.Error(codes.Unavailable, err.Error())
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
