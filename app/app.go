package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"github.com/ymhhh/go-common/logger"
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
	httpServer    *gblockshttp.Server
	grpcServer    *gblocksgrpc.Server
	metricsServer *metrics.Server
	aiClient      ai.Client
	httpRegister  HTTPRegisterFunc
	grpcRegister  GRPCRegisterFunc
	httpTracing   []gin.HandlerFunc
	grpcTracing   []grpc.ServerOption
}

// New creates a new App from configuration.
func New(cfg *config.Config) (*App, error) {
	if cfg == nil {
		cfg = config.Default()
	}

	var reg *metrics.Registry
	if cfg.Metrics.Enabled {
		reg = metrics.NewRegistry()
	}

	policy, err := resilience.NewPolicyFromConfig(cfg.Resilience, resilience.WithBreakerStateChange(func(name string, from, to gobreaker.State) {
		if reg != nil {
			reg.SetCircuitBreakerState(name, to)
		}
	}))
	if err != nil {
		return nil, fmt.Errorf("init policy: %w", err)
	}
	if reg != nil && policy.Breaker != nil {
		reg.SetCircuitBreakerState("default", policy.Breaker.State())
	}

	return &App{
		cfg:     cfg,
		policy:  policy,
		metrics: reg,
	}, nil
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

	if err := a.cfg.InitLogger(); err != nil {
		return fmt.Errorf("init logger: %w", err)
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
		for _, mw := range a.httpTracing {
			engine.Use(mw)
		}
		engine.Use(httpmiddleware.GlobalRateLimit(a.policy, a.metrics))
		engine.Use(httpmiddleware.BreakerCheck(a.policy, a.metrics))
		if len(a.policy.RateLimits.RouteRules) > 0 {
			engine.Use(httpmiddleware.RouteRateLimit(a.policy, a.metrics))
		}
		a.httpRegister(engine, a.policy)
		registerHealthRoutes(engine, a)

		if a.metrics != nil {
			path := a.cfg.Metrics.Path
			if path == "" {
				path = "/metrics"
			}
			if a.cfg.Metrics.Addr != "" {
				a.metricsServer = metrics.NewServer(
					a.cfg.Metrics.Addr,
					path,
					a.metrics.Handler(),
					a.cfg.Metrics.AuthToken,
				)
				metricsErrCh, err := a.metricsServer.Start()
				if err != nil {
					return fmt.Errorf("start metrics: %w", err)
				}
				logger.L().WithFields(logger.Fields{
					"addr": a.cfg.Metrics.Addr,
					"path": path,
				}).Info("metrics server started")
				g.Go(func() error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case err, ok := <-metricsErrCh:
						if !ok {
							return nil
						}
						return err
					}
				})
			} else {
				engine.GET(path, gin.WrapH(metrics.AuthWrap(a.cfg.Metrics.AuthToken, a.metrics.Handler())))
				logger.L().WithField("path", path).Info("metrics endpoint enabled")
			}
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

		httpErrCh, err := a.httpServer.Start()
		if err != nil {
			return fmt.Errorf("start http: %w", err)
		}
		logger.L().WithField("addr", a.httpServer.Addr()).Info("http server started")

		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err, ok := <-httpErrCh:
				if !ok {
					return nil
				}
				return err
			}
		})
	}

	if a.cfg.Server.GRPC.Enabled {
		if a.grpcRegister == nil {
			return fmt.Errorf("grpc is enabled but no handler registered: call app.WithGRPC(registerGRPC) in infrastructure/run.go")
		}

		unaryInterceptors := []grpc.UnaryServerInterceptor{
			grpcinterceptors.UnaryServerInterceptor(a.policy, a.metrics),
		}
		if len(a.policy.RateLimits.RouteRules) > 0 {
			unaryInterceptors = append(unaryInterceptors, grpcinterceptors.RouteUnaryServerInterceptor(a.policy, a.metrics))
		}
		if a.metrics != nil {
			unaryInterceptors = append([]grpc.UnaryServerInterceptor{a.metrics.GRPCUnaryServerInterceptor()}, unaryInterceptors...)
		}

		streamInterceptors := []grpc.StreamServerInterceptor{
			grpcinterceptors.StreamServerInterceptor(a.policy, a.metrics),
		}
		if len(a.policy.RateLimits.RouteRules) > 0 {
			streamInterceptors = append(streamInterceptors, grpcinterceptors.RouteStreamServerInterceptor(a.policy, a.metrics))
		}

		opts := append([]grpc.ServerOption{}, a.grpcTracing...)
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
		opts = append(opts, grpc.ChainStreamInterceptor(streamInterceptors...))
		a.grpcServer = gblocksgrpc.NewServer(gblocksgrpc.Config{
			Addr: a.cfg.Server.GRPC.Addr,
		}, opts...)
		a.grpcRegister(a.grpcServer.GRPCServer(), a.policy)

		grpcErrCh, err := a.grpcServer.Start()
		if err != nil {
			return fmt.Errorf("start grpc: %w", err)
		}
		logger.L().WithField("addr", a.grpcServer.Addr()).Info("grpc server started")

		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case err, ok := <-grpcErrCh:
				if !ok {
					return nil
				}
				return err
			}
		})
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case sig := <-sigCh:
			logger.L().WithField("signal", sig.String()).Info("received shutdown signal")
			cancel()
			return nil
		}
	})

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	shutdownErr := a.Shutdown(shutdownCtx)
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return shutdownErr
}

// Shutdown gracefully stops all servers.
func (a *App) Shutdown(ctx context.Context) error {
	var firstErr error

	if a.metricsServer != nil {
		if err := a.metricsServer.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		logger.L().Info("metrics server stopped")
	}

	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		logger.L().Info("http server stopped")
	}

	if a.grpcServer != nil {
		a.grpcServer.Shutdown()
		logger.L().Info("grpc server stopped")
	}

	if a.policy != nil && a.policy.RateLimits.Backend != nil {
		if closer, ok := a.policy.RateLimits.Backend.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
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
