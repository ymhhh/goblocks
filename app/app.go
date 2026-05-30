package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/ymhhh/goblocks/ai"
	"github.com/ymhhh/goblocks/config"
	gblocksgrpc "github.com/ymhhh/goblocks/grpc"
	grpcinterceptors "github.com/ymhhh/goblocks/grpc/interceptors"
	gblockshttp "github.com/ymhhh/goblocks/http"
	httpmiddleware "github.com/ymhhh/goblocks/http/middleware"
	"github.com/ymhhh/goblocks/metrics"
	"github.com/ymhhh/goblocks/resilience"
)

// HTTPRegisterFunc registers routes on the Gin engine.
type HTTPRegisterFunc func(engine *gin.Engine, policy *resilience.Policy)

// GRPCRegisterFunc registers services on the gRPC server.
type GRPCRegisterFunc func(server *grpc.Server, policy *resilience.Policy)

// App orchestrates HTTP, gRPC, and AI client lifecycle.
type App struct {
	cfg          *config.Config
	policy       *resilience.Policy
	metrics      *metrics.Registry
	httpServer   *gblockshttp.Server
	grpcServer   *gblocksgrpc.Server
	aiClient     ai.Client
	httpRegister HTTPRegisterFunc
	grpcRegister GRPCRegisterFunc
}

// New creates a new App from configuration.
func New(cfg *config.Config) *App {
	if cfg == nil {
		cfg = config.Default()
	}

	var reg *metrics.Registry
	if cfg.Metrics.Enabled {
		reg = metrics.NewRegistry()
	}

	policy := resilience.NewPolicyFromConfig(cfg.Resilience, resilience.WithBreakerStateChange(func(name string, from, to gobreaker.State) {
		if reg != nil {
			reg.SetCircuitBreakerState(name, to)
		}
	}))
	if reg != nil && policy.Breaker != nil {
		reg.SetCircuitBreakerState("default", policy.Breaker.State())
	}

	return &App{
		cfg:     cfg,
		policy:  policy,
		metrics: reg,
	}
}

// WithHTTP sets the HTTP route registration function.
func (a *App) WithHTTP(fn HTTPRegisterFunc) *App {
	a.httpRegister = fn
	return a
}

// WithGRPC sets the gRPC service registration function.
func (a *App) WithGRPC(fn GRPCRegisterFunc) *App {
	a.grpcRegister = fn
	return a
}

// Policy returns the shared resilience policy.
func (a *App) Policy() *resilience.Policy {
	return a.policy
}

// Metrics returns the metrics registry, or nil when disabled.
func (a *App) Metrics() *metrics.Registry {
	return a.metrics
}

// AIClient returns the AI client, initializing it if enabled in config.
func (a *App) AIClient() ai.Client {
	if a.aiClient == nil && a.cfg.AI.Enabled {
		a.aiClient = ai.NewOpenAIClient(ai.OpenAIConfig{
			BaseURL: a.cfg.AI.BaseURL,
			APIKey:  a.cfg.AI.APIKey,
			Model:   a.cfg.AI.Model,
			Policy:  a.policy,
			Metrics: a.metrics,
		})
	}
	return a.aiClient
}

// Run starts all servers and blocks until shutdown signal.
func (a *App) Run(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	if a.httpRegister != nil {
		gin.SetMode(gin.ReleaseMode)
		engine := gin.New()
		engine.Use(gin.Recovery())
		if a.metrics != nil {
			engine.Use(a.metrics.HTTPMiddleware())
		}
		engine.Use(httpmiddleware.ResilienceWithBreaker(a.policy, a.metrics))
		a.httpRegister(engine, a.policy)

		if a.metrics != nil {
			path := a.cfg.Metrics.Path
			if path == "" {
				path = "/metrics"
			}
			engine.GET(path, gin.WrapH(a.metrics.Handler()))
			slog.Info("metrics endpoint enabled", "path", path)
		}

		a.httpServer = gblockshttp.NewServer(engine, gblockshttp.Config{
			Addr: a.cfg.Server.HTTP.Addr,
			TLS: gblockshttp.TLSOptions{
				Enabled:  a.cfg.Server.HTTP.TLS.Enabled,
				CertFile: a.cfg.Server.HTTP.TLS.CertFile,
				KeyFile:  a.cfg.Server.HTTP.TLS.KeyFile,
			},
			H3: gblockshttp.H3Options{
				Enabled: a.cfg.Server.HTTP.H3.Enabled,
				Addr:    a.cfg.Server.HTTP.H3.Addr,
			},
		})

		if err := a.httpServer.Start(); err != nil {
			return fmt.Errorf("start http: %w", err)
		}
		slog.Info("http server started", "addr", a.httpServer.Addr())
	}

	if a.cfg.Server.GRPC.Enabled {
		if a.grpcRegister == nil {
			return fmt.Errorf("grpc is enabled but no handler registered: call app.WithGRPC(registerGRPC) in infrastructure/run.go")
		}

		interceptors := []grpc.UnaryServerInterceptor{
			grpcinterceptors.UnaryServerInterceptor(a.policy, a.metrics),
		}
		if a.metrics != nil {
			interceptors = append([]grpc.UnaryServerInterceptor{a.metrics.GRPCUnaryServerInterceptor()}, interceptors...)
		}

		opts := []grpc.ServerOption{
			grpc.ChainUnaryInterceptor(interceptors...),
		}
		a.grpcServer = gblocksgrpc.NewServer(gblocksgrpc.Config{
			Addr: a.cfg.Server.GRPC.Addr,
		}, opts...)
		a.grpcRegister(a.grpcServer.GRPCServer(), a.policy)

		if err := a.grpcServer.Start(); err != nil {
			return fmt.Errorf("start grpc: %w", err)
		}
		slog.Info("grpc server started", "addr", a.grpcServer.Addr())
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-sigCh:
			slog.Info("received shutdown signal", "signal", sig.String())
			cancel()
			return nil
		}
	})

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	return a.Shutdown(shutdownCtx)
}

// Shutdown gracefully stops all servers.
func (a *App) Shutdown(ctx context.Context) error {
	var firstErr error

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(); err != nil && firstErr == nil {
			firstErr = err
		}
		slog.Info("http server stopped")
	}

	if a.grpcServer != nil {
		a.grpcServer.Shutdown()
		slog.Info("grpc server stopped")
	}

	select {
	case <-ctx.Done():
		if firstErr == nil {
			firstErr = ctx.Err()
		}
	default:
	}

	return firstErr
}

// Config returns the application configuration.
func (a *App) Config() *config.Config {
	return a.cfg
}
