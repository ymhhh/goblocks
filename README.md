# goblocks

Go server framework: HTTP (H1/H2/H3), gRPC, OpenAI-compatible AI client, unified breaker and layered rate limiting.

Use [goblocks-cli](https://github.com/ymhhh/goblocks-cli) to scaffold [onion architecture / DDD](https://github.com/ymhhh/ddd-onion-sample) services.

## Quick start

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
goblocks new my-service --module github.com/acme/my-service
cd my-service && go mod tidy && go run .
```

Existing project: `go get github.com/ymhhh/goblocks@latest`

## Documentation

See **[docs/](docs/README.md)** (English) | **[中文文档](docs/zh/README.md)**:

| Doc | Description |
|-----|-------------|
| [Architecture](docs/architecture.md) · [架构](docs/zh/architecture.md) | Layering, rate-limit mounting |
| [Configuration](docs/configuration.md) · [配置](docs/zh/configuration.md) | YAML and env vars |
| [Rate limiting](docs/zh/rate-limiting.md) | L1/L2/L3 (中文) |
| [Quickstart](docs/zh/quickstart.md) | Install and minimal example (中文) |
| [Package API](docs/packages.md) · [包 API](docs/zh/packages.md) | Code integration |
| [Metrics](docs/metrics.md) · [指标](docs/zh/metrics.md) | Prometheus |
| [Development](docs/development.md) · [开发](docs/zh/development.md) | Contribute, test, release |

## License

GPL-3.0. API stability: [docs/stability.md](docs/stability.md).
