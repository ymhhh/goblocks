package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHTTPMiddlewareRecordsMetrics(t *testing.T) {
	reg := NewRegistry()
	r := gin.New()
	r.Use(reg.HTTPMiddleware())
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	if v := testutil.ToFloat64(reg.HTTPRequestsTotal.WithLabelValues("GET", "/health", "200")); v != 1 {
		t.Fatalf("expected 1 request, got %v", v)
	}
}

func TestResilienceMetrics(t *testing.T) {
	reg := NewRegistry()
	reg.RecordRateLimitRejected("http", "global")
	reg.RecordCircuitBreakerRejected("grpc")

	if v := testutil.ToFloat64(reg.RateLimitRejectedTotal.WithLabelValues("http", "global")); v != 1 {
		t.Fatalf("expected 1 rate limit rejection, got %v", v)
	}
	if v := testutil.ToFloat64(reg.CircuitBreakerRejectedTotal.WithLabelValues("grpc")); v != 1 {
		t.Fatalf("expected 1 breaker rejection, got %v", v)
	}
}

func TestObserveAIRequest(t *testing.T) {
	reg := NewRegistry()
	reg.ObserveAIRequest("gpt-4o-mini", "success", 0)

	if v := testutil.ToFloat64(reg.AIRequestsTotal.WithLabelValues("gpt-4o-mini", "success")); v != 1 {
		t.Fatalf("expected 1 ai request, got %v", v)
	}
}

func TestMetricsHandler(t *testing.T) {
	reg := NewRegistry()
	reg.HTTPRequestsTotal.WithLabelValues("GET", "/health", "200").Inc()

	w := httptest.NewRecorder()
	reg.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !contains(w.Body.String(), "goblocks_http_requests_total") {
		t.Fatal("expected metrics output")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && findSub(s, sub)))
}

func findSub(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
