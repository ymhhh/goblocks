package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// Config holds gRPC server configuration.
type Config struct {
	Addr string
}

// Server wraps a gRPC server.
type Server struct {
	config Config
	server *grpc.Server
	lis    net.Listener
}

// NewServer creates a gRPC server with the given options.
func NewServer(cfg Config, opts ...grpc.ServerOption) *Server {
	return &Server{
		config: cfg,
		server: grpc.NewServer(opts...),
	}
}

// GRPCServer returns the underlying grpc.Server for service registration.
func (s *Server) GRPCServer() *grpc.Server {
	return s.server
}

// Start begins listening for gRPC connections.
func (s *Server) Start() error {
	if s.config.Addr == "" {
		s.config.Addr = ":9090"
	}

	lis, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	s.lis = lis

	go func() {
		if err := s.server.Serve(lis); err != nil {
			panic(fmt.Sprintf("grpc server: %v", err))
		}
	}()

	return nil
}

// Shutdown gracefully stops the gRPC server.
func (s *Server) Shutdown() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// Addr returns the configured listen address.
func (s *Server) Addr() string {
	return s.config.Addr
}
