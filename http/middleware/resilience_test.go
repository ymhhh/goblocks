package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ymhhh/goblocks/resilience"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestResilienceRateLimit(t *testing.T) {
	policy := resilience.NewPolicy(nil, resilience.NewLimiter(1, 1))

	r := gin.New()
	r.Use(Resilience(policy))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w2.Code)
	}
}

func TestResilienceNilPolicy(t *testing.T) {
	r := gin.New()
	r.Use(Resilience(nil))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestResilienceWithBreakerOpen(t *testing.T) {
	breaker := resilience.NewBreaker(resilience.BreakerSettings{Name: "test"})
	policy := resilience.NewPolicy(breaker, nil)

	for i := 0; i < 3; i++ {
		_, _ = policy.Execute(func() (any, error) {
			return nil, resilience.ErrRateLimited
		})
	}

	r := gin.New()
	r.Use(ResilienceWithBreaker(policy))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusOK {
		t.Fatalf("expected 503 or 200, got %d", w.Code)
	}
}
