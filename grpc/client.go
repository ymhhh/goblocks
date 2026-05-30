package grpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ClientConfig holds gRPC client configuration.
type ClientConfig struct {
	Addr string
}

// Dial creates a gRPC client connection with the given options.
func Dial(cfg ClientConfig, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("grpc client addr is required")
	}

	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	defaultOpts = append(defaultOpts, opts...)

	conn, err := grpc.NewClient(cfg.Addr, defaultOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial grpc: %w", err)
	}
	return conn, nil
}
