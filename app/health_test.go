package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ymhhh/goblocks/config"
	"github.com/ymhhh/goblocks/resilience"
	"google.golang.org/grpc"
)

func TestHealthRoutes(t *testing.T) {
	cfg := config.Default()
	a := New(cfg)

	engine := gin.New()
	registerHealthRoutes(engine, a)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("liveness: expected 200, got %d", w.Code)
	}

	w = httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("readiness: expected 200, got %d", w.Code)
	}
}

func TestReadinessGRPCNotStarted(t *testing.T) {
	cfg := config.Default()
	cfg.Server.GRPC.Enabled = true
	a := New(cfg)
	a.grpcRegister = func(*grpc.Server, *resilience.Policy) {}

	engine := gin.New()
	registerHealthRoutes(engine, a)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when grpc enabled but not started, got %d", w.Code)
	}
}

func TestHealthDisabled(t *testing.T) {
	cfg := config.Default()
	cfg.Server.HTTP.Health.Enabled = false
	a := New(cfg)

	engine := gin.New()
	registerHealthRoutes(engine, a)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when health disabled, got %d", w.Code)
	}
}
