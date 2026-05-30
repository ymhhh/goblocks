# Goblocks documentation

Goblocks is a Go server framework: HTTP (H1/H2/H3), gRPC, OpenAI-compatible AI client, unified breaker and **layered rate limiting** (L1/L2/L3), and Prometheus metrics.

The scaffold CLI is in **[goblocks-cli](https://github.com/ymhhh/goblocks-cli)** (`goblocks new`).

## Feature overview

| Module | Description |
|--------|-------------|
| HTTP | Gin wrapper, HTTP/2 over TLS, optional HTTP/3 |
| gRPC | Server / Client with resilience interceptors |
| AI | OpenAI-compatible Chat (OpenAI, Azure, Ollama, etc.) |
| Resilience | Breaker + layered rate limits L1/L2/L3 (`memory` / `redis`) |
| Metrics | Prometheus (HTTP / gRPC / resilience / AI) |

## Quick start

**New project (recommended)**

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
goblocks new my-service --module github.com/acme/my-service
cd my-service && go mod tidy && go run .
```

CLI flags (`--demo`, `--with-grpc`, `--with-ai`): [goblocks-cli README](https://github.com/ymhhh/goblocks-cli).

**Existing project**

```bash
go get github.com/ymhhh/goblocks@latest
```

Minimal `main.go`: [Package API — app](packages.md#app).

Config example: [Configuration — full example](configuration.md#full-example). Rate limiting: [Chinese guide](zh/rate-limiting.md).

## Documentation index

| Doc | Description |
|-----|-------------|
| **[中文文档 / Chinese](zh/README.md)** | Full Chinese translations |
| [Architecture](architecture.md) · [架构](zh/architecture.md) | Layering, dependencies, rate-limit layout, request flow |
| [Configuration](configuration.md) · [配置](zh/configuration.md) | YAML, env vars, module examples |
| [Package API](packages.md) · [包 API](zh/packages.md) | `app`, `http`, `grpc`, `ai`, `resilience` |
| [Metrics](metrics.md) · [指标](zh/metrics.md) | Prometheus metrics and PromQL |
| [Tracing](tracing.md) · [追踪](zh/tracing.md) | Optional OpenTelemetry hooks |
| [grpc-gateway](grpc-gateway.md) · [网关](zh/grpc-gateway.md) | REST gateway integration (optional) |
| [Development](development.md) · [开发](zh/development.md) | Framework dev, test, release |
| [API stability](stability.md) · [稳定性](zh/stability.md) | Version policy and compatibility |
| [Rate limiting (中文)](zh/rate-limiting.md) | L1/L2/L3 guide (Chinese only) |
| [Quickstart (中文)](zh/quickstart.md) | Install and minimal example |
| [CHANGELOG summary (中文)](zh/changelog.md) | Version notes in Chinese |

## Links

- Framework: [goblocks](https://github.com/ymhhh/goblocks)
- CLI: [goblocks-cli](https://github.com/ymhhh/goblocks-cli)
- Reference: [ddd-onion-sample](https://github.com/ymhhh/ddd-onion-sample)
- Install framework: `go get github.com/ymhhh/goblocks@latest`
- Install CLI: `go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest`

## Requirements

- Go >= 1.22
- HTTP/3: TLS certificates and reachable UDP port
