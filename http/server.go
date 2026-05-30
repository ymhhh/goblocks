package http

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go/http3"
)

// Config holds HTTP server configuration.
type Config struct {
	Addr     string
	TLS      TLSOptions
	H3       H3Options
}

// TLSOptions holds TLS settings.
type TLSOptions struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

// H3Options holds HTTP/3 settings.
type H3Options struct {
	Enabled bool
	Addr    string
}

// Server wraps Gin with HTTP/1, HTTP/2, and optional HTTP/3 support.
type Server struct {
	engine   *gin.Engine
	config   Config
	httpSrv  *http.Server
	h3Srv    *http3.Server
	tlsConfig *tls.Config
}

// NewServer creates a new HTTP server with the given Gin engine and config.
func NewServer(engine *gin.Engine, cfg Config) *Server {
	if engine == nil {
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
	}
	return &Server{
		engine: engine,
		config: cfg,
	}
}

// Engine returns the underlying Gin engine for route registration.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Start begins listening for HTTP/1, HTTP/2 (with TLS), and optionally HTTP/3.
func (s *Server) Start() error {
	if s.config.Addr == "" {
		s.config.Addr = ":8080"
	}

	handler := s.engine

	if s.config.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(s.config.TLS.CertFile, s.config.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("load tls cert: %w", err)
		}
		s.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"h2", "http/1.1"},
		}
		s.httpSrv = &http.Server{
			Addr:      s.config.Addr,
			Handler:   handler,
			TLSConfig: s.tlsConfig,
		}

		if s.config.H3.Enabled {
			h3Addr := s.config.H3.Addr
			if h3Addr == "" {
				h3Addr = ":8443"
			}
			s.h3Srv = &http3.Server{
				Addr:      h3Addr,
				Handler:   handler,
				TLSConfig: s.tlsConfig,
			}
		}

		go func() {
			if err := s.httpSrv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				panic(fmt.Sprintf("http server: %v", err))
			}
		}()

		if s.h3Srv != nil {
			go func() {
				if err := s.h3Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					panic(fmt.Sprintf("http3 server: %v", err))
				}
			}()
		}
		return nil
	}

	s.httpSrv = &http.Server{
		Addr:    s.config.Addr,
		Handler: handler,
	}

	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Sprintf("http server: %v", err))
		}
	}()

	return nil
}

// Shutdown gracefully stops all HTTP listeners.
func (s *Server) Shutdown() error {
	var firstErr error

	if s.httpSrv != nil {
		if err := s.httpSrv.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if s.h3Srv != nil {
		if err := s.h3Srv.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Addr returns the configured listen address.
func (s *Server) Addr() string {
	return s.config.Addr
}
