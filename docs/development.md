# Development guide

For **goblocks framework** contributors and maintainers. Scaffold CLI development: [goblocks-cli](https://github.com/ymhhh/goblocks-cli).

## Requirements

- Go >= 1.22 (CI on 1.22 / 1.23; `go.mod` minimum 1.22)
- Make (optional)

## Repository layout

```
goblocks/
├── app/              Application lifecycle (L1 + L3 by default)
├── config/           Config loading (layered rate_limit schema)
├── resilience/       Breaker + RateLimiter (memory/redis) + Policy
│   ├── ratelimiter.go / memory_ratelimiter.go / redis_ratelimiter.go
│   ├── keyed.go / factory.go / policy.go
│   └── limiter.go    Legacy single-bucket helper (tests)
├── http/             HTTP server
│   └── middleware/   GlobalRateLimit, UserRateLimit, RouteRateLimit
├── grpc/             gRPC server
│   └── interceptors/ L1/L2/L3 unary interceptors
├── ai/               AI client (outbound L1 + breaker)
├── metrics/          Prometheus (scope label)
├── docs/             Documentation
├── Makefile
└── README.md
```

## Common commands

```bash
make test     # go test ./... -race -count=1
make lint     # go vet ./... && go fmt ./...
```

## Testing

```bash
go test ./config/... -v
go test ./resilience/... -v
go test ./metrics/... -v
go test ./... -race -count=1
```

## Adding dependencies

```bash
go get example.com/foo
go mod tidy
```

Keep framework dependencies minimal; justify new deps in PRs.

## Release process

1. Ensure `go test ./... -race` passes
2. Update README / docs for API changes
3. Tag: `git tag v0.3.0`
4. Users reference:

```bash
go get github.com/ymhhh/goblocks@v0.3.0
```

CLI releases in [goblocks-cli](https://github.com/ymhhh/goblocks-cli); align version with framework when possible.

## Local dev (framework + CLI)

```bash
cd /path/to/github.com/ymhhh
go work init ./goblocks ./goblocks-cli
```

CLI integration tests need `GOBLOCKS_PATH` pointing at local goblocks:

```bash
export GOBLOCKS_PATH=/path/to/goblocks
cd goblocks-cli && make test-integration
```

## Code conventions

- Public APIs need godoc comments
- New features need unit tests
- Default logging: [`github.com/ymhhh/go-common/logger`](https://github.com/ymhhh/go-common)

## Known limitations

- No service discovery (Consul/Etcd)
- Redis rate limit needs a dedicated Redis; `memory` backend is single-process only
- gRPC and HTTP cannot share a port
- HTTP/3 requires extra UDP port and TLS
