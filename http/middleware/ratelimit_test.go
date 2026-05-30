package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
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

func TestGlobalRateLimit(t *testing.T) {
	policy := resilience.NewPolicy(nil, resilience.NewLimiter(1, 1))

	r := gin.New()
	r.Use(GlobalRateLimit(policy, nil))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", w2.Code)
	}
}

func TestUserRateLimit(t *testing.T) {
	policy := &resilience.Policy{
		RateLimits: resilience.RateLimits{
			Backend:     resilience.NewMemoryRateLimiter(),
			UserEnabled: true,
			UserRule:    resilience.LimitRule{RPS: 1, Burst: 1},
		},
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		GinContextWithUserID(c, "alice")
	})
	r.Use(UserRateLimit(policy, nil))
	r.GET("/users/:id", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/users/alice", nil)
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
}

func TestUserRateLimitDisabled(t *testing.T) {
	policy := &resilience.Policy{
		RateLimits: resilience.RateLimits{
			Backend:     resilience.NewMemoryRateLimiter(),
			UserEnabled: false,
			UserRule:    resilience.LimitRule{RPS: 1, Burst: 1},
		},
	}

	r := gin.New()
	r.Use(UserRateLimit(policy, nil))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestBreakerCheckOpen(t *testing.T) {
	breaker := resilience.NewBreaker(resilience.BreakerSettings{
		Name:                "test",
		ConsecutiveFailures: 1,
		Interval:            time.Minute,
		Timeout:             time.Minute,
	})
	policy := resilience.NewPolicy(breaker, nil)
	if err := policy.ExecuteVoid(func() error {
		return errors.New("downstream error")
	}); err == nil {
		t.Fatal("expected execute error")
	}
	if policy.Breaker.State() != gobreaker.StateOpen {
		t.Fatalf("expected breaker open, got %v", policy.Breaker.State())
	}

	r := gin.New()
	r.Use(BreakerCheck(policy, nil))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when breaker open, got %d", w.Code)
	}
}

func TestGinContextWithUserID(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		GinContextWithUserID(c, "user-1")
		if got := GinUserKey(c); got != "user:user-1" {
			t.Fatalf("GinUserKey = %q", got)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/test", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
