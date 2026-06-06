# Changelog

All notable changes to this project are documented in this file.

中文摘要见 [docs/zh/changelog.md](docs/zh/changelog.md).

## [v0.3.2] - 2026-06-06

### Breaking

- `app.New(cfg)` now returns `(*App, error)` instead of `panic` on policy init failure
- Removed deprecated `resilience.Limiter` type, `NewPolicy` constructor, `legacyLimiterAdapter`, and `Policy.Limiter` field; use `Policy.RateLimits` with `MemoryRateLimiter` directly

### Added

- gRPC stream interceptors: `StreamServerInterceptor`, `StreamClientInterceptor`, `UserStreamServerInterceptor`, `RouteStreamServerInterceptor` — streaming calls now get L1/L2/L3 rate limiting and circuit breaking
- gRPC L2 user interceptors (`UserUnaryServerInterceptor` / `UserStreamServerInterceptor`) are now auto-wired in `app.Run` when `resilience.rate_limit.user.enabled: true`
- Configurable graceful shutdown timeout via `server.shutdown_timeout` (default 30s)
- Optional circuit breaker: `resilience.breaker.enabled` (default true)
- `metrics.Registry.PromRegistry()` exposes the underlying `*prometheus.Registry` for custom collectors

### Fixed

- **Memory leak:** `MemoryRateLimiter` buckets grew without bound; added background eviction with configurable TTL (default 5m) and cleanup interval (default 1m)
- **Connection leak:** `RedisRateLimiter` now has `Close()`, wired into `App.Shutdown`
- **Data race:** `AIClient()` lazy init is now protected with `sync.Once`
- `Policy.AllowUser` and `AllowRoute` now safely return nil when receiver is nil (consistent with `AllowGlobal`)
- AI client now calls `AllowGlobal(ctx)` instead of `Allow()` to preserve trace context
- gRPC client interceptors emit `grpc_client` protocol label (not `grpc`) for rate limit metrics
- Redis rate limiter floors sub-1 RPS to 1 to avoid silent truncation (`int(0.5) == 0`)
- Burst defaults are now consistently `RPS * 2` in both `config.Normalized()` and `normalizeRule`
- gRPC route limit keys use the full method path directly (no more hardcoded `"GRPC"` pseudo-method)
- L3 route rate limit middleware skips unregistered routes (`FullPath()` is empty)

### Changed

- Metrics path default `/metrics` is now set only in `app.go` (removed duplicate in `metrics/server.go`)
- `Policy.allow()` no longer takes an unused `Scope` parameter
- `config.RateLimitConfig.Normalized()` computes default burst from RPS instead of hardcoded values

## [v0.3.1] - 2026-05-30

### Added

- Auto-mount L3 `RouteRateLimit` / `RouteUnaryServerInterceptor` when `resilience.rate_limit.routes` is configured
- Chinese documentation under `docs/zh/` (full translations of architecture, configuration, packages, metrics, tracing, etc.)
- Tracing docs: trace-log correlation via `logger.LFromContext` (go-common)
- Tests: L3 app auto-mount integration, tracing/`LFromContext`, HTTP L1/L2/L3 middleware, gRPC user interceptor, resilience key/factory helpers

### Changed

- Documentation: L3 mounting semantics, HTTP middleware order, configuration `routes` section, bilingual cross-links
- Dependency: bump `github.com/ymhhh/go-common` for `logger.LFromContext`

## [v0.3.0] - 2026-05-30

### Added

- **Layered rate limiting** (L1 global / L2 user / L3 route): `RateLimiter` interface with `MemoryRateLimiter` and `RedisRateLimiter` (GCRA via `go-redis/redis_rate`)
- HTTP middleware: `GlobalRateLimit`, `UserRateLimit`, `RouteRateLimit`, `RateLimitByKey`, `BreakerCheck`
- gRPC interceptors: `UserUnaryServerInterceptor`, `RouteUnaryServerInterceptor` (L1 remains in `UnaryServerInterceptor`)
- Config schema: `resilience.rate_limit` extended with `global`, `user`, `routes`, `backend` (`memory` | `redis`), `redis.addr`, `redis.key_prefix`
- Environment variables: `GOBLOCKS_REDIS_ADDR`, `GOBLOCKS_RATE_LIMIT_BACKEND`
- Metrics: `rate_limit_rejected_total` gains `scope` label (`global` | `user` | `route`)
- Key helpers: `ContextWithUserID`, `UserKeyFromContext`, `RouteKey`, `GlobalKey`

### Changed

- **Breaking:** `resilience.NewPolicyFromConfig` now returns `(*Policy, error)` (Redis backend requires valid `redis.addr`)
- **Breaking:** `app.Run` mounts **L1 global rate limit + breaker**; L3 auto-mounts when `routes` configured; L2 must be registered in business `infrastructure/registerHTTP`
- `resilience.rate_limit.rps` / `burst` deprecated in favor of `global.rps` / `global.burst` (legacy fields still mapped)
- HTTP resilience middleware split: removed `http/middleware/resilience.go`; use `ratelimit.go` APIs (`Resilience` / `ResilienceWithBreaker` deprecated)
- Documentation reorganized: layered rate-limit directory mapping in `docs/architecture.md`; README simplified

### Removed

- `examples/minimal` (use [goblocks-cli](https://github.com/ymhhh/goblocks-cli) to scaffold new projects)
- `docs/scaffold.md` (CLI usage documented in [goblocks-cli](https://github.com/ymhhh/goblocks-cli))

## [v0.2.2] - 2026-05-30

### Added

- Optional `/health` and `/ready` probes (`server.http.health`)
- Dedicated metrics listener (`metrics.addr`) and Bearer auth (`metrics.auth_token`)
- Configurable breaker `consecutive_failures`
- Optional tracing hooks: `WithHTTPTracing`, `WithGRPCTracing`
- `examples/minimal` runnable HTTP-only example
- Documentation: tracing, grpc-gateway, API stability policy

### Changed

- Go module minimum version set to 1.22 (CI matrix: 1.22 / 1.23)

## [v0.2.1] - 2026-05-30

### Added

- Prometheus metrics (`metrics/` package) with HTTP/gRPC middleware and AI client instrumentation
- Configuration loading via `github.com/ymhhh/go-common/config` (`#include`, `${ENV}`, cross-references)
- Logging via `github.com/ymhhh/go-common/logger` with `logger:` config key
- GitHub Actions CI (Go 1.22 / 1.23 matrix)

### Changed

- **Breaking:** YAML config key `log:` renamed to `logger:` with full go-common logger schema
- **Breaking:** gRPC requires explicit `app.WithGRPC(registerGRPC)` when `server.grpc.enabled: true`
- HTTP server uses graceful `Shutdown(ctx)` instead of `Close()`
- Server listen failures return errors from `app.Run()` instead of panicking
- CLI scaffold moved to separate repository [goblocks-cli](https://github.com/ymhhh/goblocks-cli)

### Deprecated

- `GOBLOCKS_LOG_LEVEL` — use `GOBLOCKS_LOGGER_LEVEL` (still honored for compatibility)

## [v0.2.0] - 2026-05-30

### Added

- Initial framework: HTTP (H1/H2/H3), gRPC, AI client, resilience (breaker + rate limit)
- Application lifecycle (`app` package)
- Documentation under `docs/`
