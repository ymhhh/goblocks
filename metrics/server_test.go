package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthWrapAllowsValidToken(t *testing.T) {
	called := false
	h := AuthWrap("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthWrapRejectsMissingToken(t *testing.T) {
	h := AuthWrap("secret", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMetricsServerStartShutdown(t *testing.T) {
	reg := NewRegistry()
	srv := NewServer("127.0.0.1:0", "/metrics", reg.Handler(), "")

	errCh, err := srv.Start()
	if err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected serve error: %v", err)
		}
	default:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func TestAuthWrapEmptyTokenPassthrough(t *testing.T) {
	h := AuthWrap("", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
