package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ymhhh/goblocks/resilience"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor with resilience.
func UnaryServerInterceptor(policy *resilience.Policy) grpc.UnaryServerInterceptor {
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
				return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}
			return nil, status.Error(codes.Unavailable, err.Error())
		}

		result, err := policy.Execute(func() (any, error) {
			return handler(ctx, req)
		})
		if err != nil {
			if err == resilience.ErrCircuitOpen {
				return nil, status.Error(codes.Unavailable, "circuit breaker is open")
			}
			return nil, err
		}
		return result, nil
	}
}

// UnaryClientInterceptor returns a gRPC unary client interceptor with resilience.
func UnaryClientInterceptor(policy *resilience.Policy) grpc.UnaryClientInterceptor {
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
				return status.Error(codes.ResourceExhausted, "rate limit exceeded")
			}
			return status.Error(codes.Unavailable, err.Error())
		}

		_, err := policy.Execute(func() (any, error) {
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})
		if err != nil {
			if err == resilience.ErrCircuitOpen {
				return status.Error(codes.Unavailable, "circuit breaker is open")
			}
		}
		return err
	}
}
