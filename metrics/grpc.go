package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// GRPCUnaryServerInterceptor records gRPC request count and latency.
func (r *Registry) GRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if r == nil {
			return handler(ctx, req)
		}

		start := time.Now()
		resp, err := handler(ctx, req)

		code := status.Code(err).String()
		r.GRPCRequestsTotal.WithLabelValues(info.FullMethod, code).Inc()
		r.GRPCRequestDuration.WithLabelValues(info.FullMethod).Observe(time.Since(start).Seconds())

		return resp, err
	}
}
