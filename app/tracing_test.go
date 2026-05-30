package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ymhhh/go-common/logger"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

func TestWithHTTPTracing(t *testing.T) {
	a := New(nil).WithHTTPTracing(func(c *gin.Context) {
		c.Next()
	})
	if len(a.httpTracing) != 1 {
		t.Fatalf("expected 1 tracing middleware, got %d", len(a.httpTracing))
	}
}

func TestWithGRPCTracing(t *testing.T) {
	a := New(nil).WithGRPCTracing(grpc.MaxRecvMsgSize(1024))
	if len(a.grpcTracing) != 1 {
		t.Fatalf("expected 1 grpc tracing option, got %d", len(a.grpcTracing))
	}
}

func TestHTTPTracingMiddlewareOrder(t *testing.T) {
	var order []string

	a := New(nil).WithHTTPTracing(func(c *gin.Context) {
		order = append(order, "tracing")
		c.Next()
	})

	r := gin.New()
	for _, mw := range a.httpTracing {
		r.Use(mw)
	}
	r.GET("/test", func(c *gin.Context) {
		order = append(order, "handler")
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if len(order) != 2 || order[0] != "tracing" || order[1] != "handler" {
		t.Fatalf("unexpected middleware order: %v", order)
	}
}

func TestLFromContextInHTTPHandler(t *testing.T) {
	traceID, err := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	if err != nil {
		t.Fatal(err)
	}
	spanID, err := trace.SpanIDFromHex("0102030405060708")
	if err != nil {
		t.Fatal(err)
	}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	var logTraceID, logSpanID string
	r := gin.New()
	r.Use(func(c *gin.Context) {
		ctx := trace.ContextWithSpanContext(c.Request.Context(), sc)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	r.GET("/test", func(c *gin.Context) {
		entry := logger.LFromContext(c.Request.Context())
		logTraceID, _ = entry.Data["trace_id"].(string)
		logSpanID, _ = entry.Data["span_id"].(string)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if logTraceID != traceID.String() {
		t.Fatalf("trace_id = %q, want %q", logTraceID, traceID.String())
	}
	if logSpanID != spanID.String() {
		t.Fatalf("span_id = %q, want %q", logSpanID, spanID.String())
	}
}

func TestLFromContextWithoutSpan(t *testing.T) {
	var hasTrace bool
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		entry := logger.LFromContext(c.Request.Context())
		_, hasTrace = entry.Data["trace_id"]
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if hasTrace {
		t.Fatal("expected no trace_id without OTel span in context")
	}
}
