# Changelog

All notable changes to this project are documented in this file.

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
