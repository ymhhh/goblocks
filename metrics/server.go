package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// Server exposes Prometheus metrics on a dedicated HTTP listener.
type Server struct {
	httpSrv *http.Server
}

// NewServer creates a standalone metrics HTTP server.
func NewServer(addr, path string, handler http.Handler, authToken string) *Server {
	mux := http.NewServeMux()
	mux.Handle(path, AuthWrap(authToken, handler))
	return &Server{
		httpSrv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

// Start listens and serves metrics. Listen errors return synchronously.
func (s *Server) Start() (<-chan error, error) {
	lis, err := net.Listen("tcp", s.httpSrv.Addr)
	if err != nil {
		return nil, fmt.Errorf("listen metrics: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.Serve(lis); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("metrics server: %w", err)
		}
	}()
	return errCh, nil
}

// Shutdown gracefully stops the metrics server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}
	return s.httpSrv.Shutdown(ctx)
}

// AuthWrap optionally protects a handler with Bearer token auth.
func AuthWrap(token string, next http.Handler) http.Handler {
	if token == "" || next == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
