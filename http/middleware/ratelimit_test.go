package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ymhhh/goblocks/resilience"
)

func TestRouteRateLimit(t *testing.T) {
	backend := resilience.NewMemoryRateLimiter()
	policy := &resilience.Policy{
		RateLimits: resilience.RateLimits{
			Backend: backend,
			RouteRules: map[string]resilience.LimitRule{
				"POST:/ai/chat": {RPS: 1, Burst: 1},
			},
		},
	}

	r := gin.New()
	r.Use(RouteRateLimit(policy, nil))
	r.POST("/ai/chat", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.GET("/other", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/ai/chat", nil)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", w2.Code)
	}

	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, httptest.NewRequest(http.MethodGet, "/other", nil))
	if w3.Code != http.StatusOK {
		t.Fatalf("unconfigured route should pass: got %d", w3.Code)
	}
}

func TestRouteRateLimitNoRules(t *testing.T) {
	policy := &resilience.Policy{
		RateLimits: resilience.RateLimits{
			Backend:    resilience.NewMemoryRateLimiter(),
			RouteRules: nil,
		},
	}

	r := gin.New()
	r.Use(RouteRateLimit(policy, nil))
	r.POST("/ai/chat", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/ai/chat", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}
