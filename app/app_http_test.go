package app

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ymhhh/goblocks/config"
	"github.com/ymhhh/goblocks/resilience"
)

func TestHTTPL3RouteRateLimitAutoMount(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()

	cfg := config.Default()
	cfg.Server.GRPC.Enabled = false
	cfg.Server.HTTP.Addr = addr
	cfg.Resilience.RateLimit.Routes = []config.RouteRateLimitConfig{
		{Method: "POST", Path: "/ai/chat", RPS: 1, Burst: 1},
	}

	a, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	a.WithHTTP(func(e *gin.Engine, _ *resilience.Policy) {
		e.POST("/ai/chat", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Run(ctx)
	}()

	client := &http.Client{Timeout: time.Second}
	baseURL := "http://" + addr
	chatURL := baseURL + "/ai/chat"

	deadline := time.Now().Add(2 * time.Second)
	var ready bool
	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server did not become ready")
	}

	resp1, err := client.Post(chatURL, "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", resp1.StatusCode)
	}

	resp2, err := client.Post(chatURL, "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", resp2.StatusCode)
	}

	resp3, err := client.Get(baseURL + "/other")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	if resp3.StatusCode != http.StatusNotFound {
		t.Fatalf("unconfigured route: expected 404, got %d body=%q", resp3.StatusCode, body)
	}

	cancel()
	<-errCh
}
