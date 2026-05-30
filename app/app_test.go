package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ymhhh/goblocks/config"
	"github.com/ymhhh/goblocks/resilience"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAppHTTPRegistration(t *testing.T) {
	cfg := config.Default()
	engine := gin.New()
	engine.Use(gin.Recovery())

	registered := false
	register := func(e *gin.Engine, _ *resilience.Policy) {
		registered = true
		e.GET("/health", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})
	}

	policy, err := resilience.NewPolicyFromConfig(cfg.Resilience)
	if err != nil {
		t.Fatal(err)
	}
	register(engine, policy)
	if !registered {
		t.Fatal("expected route registration")
	}

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAppShutdown(t *testing.T) {
	cfg := config.Default()
	cfg.Server.GRPC.Enabled = false
	a := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	err := a.Shutdown(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("unexpected shutdown error: %v", err)
	}
}

func TestAppAIClient(t *testing.T) {
	cfg := config.Default()
	cfg.AI.Enabled = true
	cfg.AI.APIKey = "test"

	a := New(cfg)
	client := a.AIClient()
	if client == nil {
		t.Fatal("expected ai client")
	}
}

func TestAppPolicy(t *testing.T) {
	cfg := config.Default()
	a := New(cfg)
	if a.Policy() == nil {
		t.Fatal("expected policy")
	}
}

func TestAppConfig(t *testing.T) {
	cfg := config.Default()
	a := New(cfg)
	if a.Config() != cfg {
		t.Fatal("expected config reference")
	}
}
