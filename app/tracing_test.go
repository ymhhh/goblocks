package app

import (
	"testing"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func TestWithHTTPTracing(t *testing.T) {
	called := false
	a := New(nil).WithHTTPTracing(func(c *gin.Context) {
		called = true
		c.Next()
	})
	if len(a.httpTracing) != 1 {
		t.Fatalf("expected 1 tracing middleware, got %d", len(a.httpTracing))
	}
	_ = called
}

func TestWithGRPCTracing(t *testing.T) {
	a := New(nil).WithGRPCTracing(grpc.MaxRecvMsgSize(1024))
	if len(a.grpcTracing) != 1 {
		t.Fatalf("expected 1 grpc tracing option, got %d", len(a.grpcTracing))
	}
}
