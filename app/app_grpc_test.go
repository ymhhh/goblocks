package app

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/ymhhh/goblocks/config"
	"github.com/ymhhh/goblocks/resilience"
)

func TestGRPCRequiresRegistration(t *testing.T) {
	cfg := config.Default()
	cfg.Server.GRPC.Enabled = true

	a, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = a.Run(context.Background())
	if err == nil {
		t.Fatal("expected error when grpc enabled without WithGRPC")
	}
}

func TestGRPCWithExplicitRegistration(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()

	cfg := config.Default()
	cfg.Server.GRPC.Enabled = true
	cfg.Server.GRPC.Addr = addr

	a, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a.WithGRPC(func(server *grpc.Server, _ *resilience.Policy) {
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(server, healthServer)
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Run(ctx)
	}()

	deadline := time.Now().Add(2 * time.Second)
	var conn *grpc.ClientConn
	for time.Now().Before(deadline) {
		conn, err = grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if conn == nil {
		t.Fatalf("dial grpc: %v", err)
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("expected SERVING, got %v", resp.GetStatus())
	}

	cancel()
	<-errCh
}
