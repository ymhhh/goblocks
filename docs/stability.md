# API stability

## v0.x (current)

- Breaking changes allowed but must be noted in [CHANGELOG](../CHANGELOG.md)
- YAML keys and env var changes are breaking changes
- Patch releases (v0.2.x) prioritize bugfixes and backward-compatible improvements

## Frozen (best-effort compatibility)

| API | Description |
|-----|-------------|
| `config.Load(path)` | Returns `*Config, error` |
| `app.New(cfg)` | Construct application |
| `app.(*App).WithHTTP` / `WithGRPC` | Register routes and services |
| `app.(*App).Run(ctx)` | Start and graceful shutdown |
| `resilience.NewPolicyFromConfig` | Build policy from config; returns `(*Policy, error)` |

### Layered rate limiting (v0.3+)

| API | Layer | Mount location |
|-----|-------|----------------|
| `http/middleware.GlobalRateLimit` | L1 | `app.Run` default |
| `http/middleware.UserRateLimit` | L2 | Business `infrastructure` |
| `http/middleware.RouteRateLimit` | L3 | `app.Run` (when config has `routes`) |
| `grpc/interceptors.UnaryServerInterceptor` | L1 | `app.Run` default |
| `grpc/interceptors.UserUnaryServerInterceptor` | L2 | Business infrastructure |
| `grpc/interceptors.RouteUnaryServerInterceptor` | L3 | `app.Run` (when config has `routes`) |

Config `resilience.rate_limit` adds `global` / `user` / `routes` / `backend`; top-level `rps`/`burst` remain compatible.

## v1.0 goals

- Freeze public API signatures above
- Strict semver MAJOR.MINOR.PATCH
- Config schema migration guide and compatibility period

## Experimental APIs

May change in minor releases:

- `WithHTTPTracing` / `WithGRPCTracing`
- `server.http.health.*` probe behavior
- `metrics.addr` / `metrics.auth_token`
