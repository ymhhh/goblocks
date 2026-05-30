package interceptors

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ymhhh/goblocks/resilience"
)

func echoHandler(ctx context.Context, req any) (any, error) {
	return req, nil
}

func TestUnaryServerInterceptor(t *testing.T) {
	policy := resilience.NewPolicy(nil, resilience.NewLimiter(100, 200))
	interceptor := UnaryServerInterceptor(policy, nil)

	handler := func(ctx context.Context, req any) (any, error) {
		return "hello", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test/Echo"}
	result, err := interceptor(context.Background(), "world", info, handler)
	if err != nil {
		t.Fatal(err)
	}
	if result.(string) != "hello" {
		t.Fatalf("expected hello, got %v", result)
	}
}

func TestUnaryServerInterceptorRateLimit(t *testing.T) {
	policy := resilience.NewPolicy(nil, resilience.NewLimiter(1, 1))
	interceptor := UnaryServerInterceptor(policy, nil)

	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Echo"}

	_, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatal(err)
	}

	_, err = interceptor(context.Background(), nil, info, handler)
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted, got %v", err)
	}
}

func TestUnaryServerInterceptorNilPolicy(t *testing.T) {
	interceptor := UnaryServerInterceptor(nil, nil)
	handler := func(ctx context.Context, req any) (any, error) {
		return req, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test/Echo"}

	result, err := interceptor(context.Background(), "test", info, handler)
	if err != nil || result.(string) != "test" {
		t.Fatalf("unexpected: %v, %v", result, err)
	}
}

func TestRouteUnaryServerInterceptor(t *testing.T) {
	policy := &resilience.Policy{
		RateLimits: resilience.RateLimits{
			Backend: resilience.NewMemoryRateLimiter(),
			RouteRules: map[string]resilience.LimitRule{
				"GRPC:/test.Echo": {RPS: 1, Burst: 1},
			},
		},
	}
	interceptor := RouteUnaryServerInterceptor(policy, nil)
	handler := func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Echo"}

	_, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatal(err)
	}
	_, err = interceptor(context.Background(), nil, info, handler)
	if status.Code(err) != codes.ResourceExhausted {
		t.Fatalf("expected ResourceExhausted, got %v", err)
	}

	infoOther := &grpc.UnaryServerInfo{FullMethod: "/test.Other"}
	_, err = interceptor(context.Background(), nil, infoOther, handler)
	if err != nil {
		t.Fatalf("unconfigured method should pass, got %v", err)
	}
}

func TestUnaryClientInterceptor(t *testing.T) {
	policy := resilience.NewPolicy(nil, nil)
	interceptor := UnaryClientInterceptor(policy, nil)

	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		*(reply.(*string)) = "response"
		return nil
	}

	var reply string
	err := interceptor(context.Background(), "/test/Echo", "req", &reply, nil, invoker)
	if err != nil {
		t.Fatal(err)
	}
	if reply != "response" {
		t.Fatalf("expected response, got %s", reply)
	}
}
