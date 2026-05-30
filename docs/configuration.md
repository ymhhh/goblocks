# Configuration reference

Goblocks uses YAML configuration, default path `config/config.yaml`. Loaded via [`github.com/ymhhh/go-common/config`](https://github.com/ymhhh/go-common) with `#include`, `${ENV}`, and `${a.b.c}` references; `GOBLOCKS_*` environment variables override fields after load.

## Advanced features

### Split config (#include)

```yaml
#include base.yaml
#include secrets.yaml
server:
  http:
    addr: ":8080"
```

### Placeholders

- `${OPENAI_API_KEY}` — environment variable
- `${server.http.addr}` — cross-reference within the same file

## Full example

```yaml
server:
  http:
    addr: ":8080"
    tls:
      enabled: false
      cert_file: "cert.pem"
      key_file: "key.pem"
    h3:
      enabled: false
      addr: ":8443"
  grpc:
    enabled: true
    addr: ":9090"
resilience:
  breaker:
    max_requests: 3
    interval: 60s
    timeout: 30s
  rate_limit:
    backend: memory          # memory | redis
    global:
      rps: 100
      burst: 200
    redis:
      addr: "redis://localhost:6379/0"
      key_prefix: "goblocks:rl:"
    user:
      enabled: true
      default_rps: 20
      burst: 40
    routes:
      - method: POST
        path: /api/v1/ai/chat
        rps: 5
        burst: 10
    # Legacy (equivalent to global):
    # rps: 100
    # burst: 200
ai:
  enabled: false
  base_url: "https://api.openai.com/v1"
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini"
logger:
  level: info
  format: text
  output: stderr
```

## server

### server.http

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `addr` | string | `:8080` | HTTP/HTTPS listen address |
| `tls.enabled` | bool | `false` | Enable TLS (HTTP/2 ALPN) |
| `tls.cert_file` | string | — | TLS certificate path |
| `tls.key_file` | string | — | TLS private key path |
| `h3.enabled` | bool | `false` | Enable HTTP/3 (requires TLS) |
| `h3.addr` | string | `:8443` | HTTP/3 QUIC listen address |
| `health.enabled` | bool | `true` | Register `/health` and `/ready` probes |
| `health.liveness_path` | string | `/health` | Liveness probe path |
| `health.readiness_path` | string | `/ready` | Readiness probe path (breaker and gRPC state) |

### server.grpc

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Start gRPC server (call `WithGRPC(registerGRPC)` in `infrastructure/run.go`) |
| `addr` | string | `:9090` | gRPC listen address |

> HTTP and gRPC use different ports; they cannot share one listener.

## resilience

### breaker

Based on [sony/gobreaker](https://github.com/sony/gobreaker). Opens after `consecutive_failures` consecutive failures (default 3).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `max_requests` | uint32 | `3` | Max requests allowed in half-open state |
| `consecutive_failures` | uint32 | `3` | Consecutive failures before opening |
| `interval` | duration | `60s` | Stats window (resets in closed state) |
| `timeout` | duration | `30s` | Open state duration before half-open |

### rate_limit (layered)

Three layers: **L1 global (service/cluster) → L2 user → L3 route/API**. The framework `app` mounts **L1** and **L3 (when `routes` is configured)** by default; L2 is mounted in business `infrastructure/registerHTTP`.

Backend: `memory` (single process / dev) or `redis` (multi-Pod, GCRA).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `backend` | string | `memory` | `memory` or `redis` |
| `global.rps` | float64 | `100` | L1 global RPS |
| `global.burst` | int | `200` | L1 burst capacity |
| `redis.addr` | string | — | Redis address (required when `backend: redis`) |
| `redis.key_prefix` | string | `goblocks:rl:` | Redis key prefix |
| `user.enabled` | bool | `false` | Enable L2 (mount middleware in infrastructure) |
| `user.default_rps` | float64 | `20` | L2 default RPS per user |
| `user.burst` | int | `40` | L2 burst per user |
| `routes` | list | — | L3 route rules (method + path + rps + burst) |
| `rps` | float64 | — | **Deprecated**, maps to `global.rps` |
| `burst` | int | — | **Deprecated**, maps to `global.burst` |

HTTP rate limit exceeded → **429**; gRPC → **ResourceExhausted**; breaker open → HTTP **503**, gRPC **Unavailable**.

See [Environment variables](#environment-variables) below.

#### routes rules (L3)

| Field | Description |
|-------|-------------|
| `method` | HTTP verb (`GET` / `POST` …) or `GRPC` for gRPC |
| `path` | HTTP: Gin route template (e.g. `/ai/chat`); gRPC: FullMethod (e.g. `/my.v1.Service/Method`) |
| `rps` / `burst` | Token bucket for this route |

When `routes` is configured, `app.Run` mounts L3 automatically; no duplicate registration in code. See the [rate-limiting guide (中文)](zh/rate-limiting.md).

## ai

OpenAI-compatible HTTP API via [go-openai](https://github.com/sashabaranov/go-openai).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `false` | Initialize AI client |
| `base_url` | string | `https://api.openai.com/v1` | API base URL |
| `api_key` | string | — | API key; supports `${ENV_VAR}` |
| `model` | string | `gpt-4o-mini` | Default model |

### AI provider examples

**OpenAI**

```yaml
ai:
  enabled: true
  base_url: "https://api.openai.com/v1"
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini"
```

**Ollama (local)**

```yaml
ai:
  enabled: true
  base_url: "http://localhost:11434/v1"
  api_key: "ollama"
  model: "llama3"
```

**Azure OpenAI**

```yaml
ai:
  enabled: true
  base_url: "https://YOUR_RESOURCE.openai.azure.com/openai/deployments/YOUR_DEPLOYMENT"
  api_key: "${AZURE_OPENAI_API_KEY}"
  model: "gpt-4o"
```

## logger

Via [`github.com/ymhhh/go-common/logger`](https://github.com/ymhhh/go-common) (logrus).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `level` | string | `info` | Log level: debug/info/warn/error |
| `format` | string | `text` | `text` or `json` |
| `output` | string | `stderr` | `stdout`/`stderr`/`discard`/file path/`file:/path` |
| `reportCaller` | bool | `false` | Log caller |
| `file.path` | string | — | File output path |
| `file.rotate.enabled` | bool | `false` | Enable lumberjack rotation |
| `file.rotate.maxSizeMB` | int | `100` | Max size per file (MB) |
| `file.rotate.maxBackups` | int | `7` | Backup count |
| `file.rotate.maxAgeDays` | int | `7` | Retention days |
| `text.disableColors` | bool | `false` | Disable colors in text format |
| `text.fullTimestamp` | bool | `false` | Full timestamp in text format |
| `json.prettyPrint` | bool | `false` | Pretty-print JSON |

## metrics

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Collect and expose Prometheus metrics |
| `path` | string | `/metrics` | Metrics HTTP path |
| `addr` | string | — | Dedicated metrics listen address (separate from app HTTP) |
| `auth_token` | string | — | Bearer token for metrics endpoint |

See [Metrics](metrics.md).

## Environment variables

| Variable | Overrides |
|----------|-----------|
| `GOBLOCKS_HTTP_ADDR` | `server.http.addr` |
| `GOBLOCKS_GRPC_ADDR` | `server.grpc.addr` |
| `GOBLOCKS_AI_API_KEY` | `ai.api_key` |
| `GOBLOCKS_AI_BASE_URL` | `ai.base_url` |
| `GOBLOCKS_LOGGER_LEVEL` | `logger.level` |
| `GOBLOCKS_LOG_LEVEL` | `logger.level` (deprecated, still supported) |
| `GOBLOCKS_METRICS_ENABLED` | `metrics.enabled` |
| `GOBLOCKS_REDIS_ADDR` | `resilience.rate_limit.redis.addr` |
| `GOBLOCKS_RATE_LIMIT_BACKEND` | `resilience.rate_limit.backend` |

### Placeholder expansion

`api_key: "${OPENAI_API_KEY}"` is resolved by go-common at load time. After load, `GOBLOCKS_*` variables still override (higher priority).

## Enabling HTTP/2 and HTTP/3

### HTTP/2

1. Prepare TLS certificates
2. Set `server.http.tls.enabled: true` and set `cert_file`, `key_file`
3. Clients use HTTPS; ALPN negotiates h2

```bash
curl --http2 -k https://localhost:8080/health
```

### HTTP/3

1. Complete HTTP/2 TLS setup (H3 reuses the same certs)
2. Set `server.http.h3.enabled: true`
3. Allow UDP port in firewall (default `:8443`)

> Production HTTP/3 needs UDP load balancing and CDN compatibility; disabled by default is recommended.

## Loading in code

```go
cfg, err := config.Load("config/config.yaml")
if err != nil {
    log.Fatal(err)
}

// Or use defaults
cfg := config.Default()
```
