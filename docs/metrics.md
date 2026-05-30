# Metrics

Goblocks includes [Prometheus](https://prometheus.io/) metrics, enabled by default, exposed over HTTP.

## Configuration

```yaml
metrics:
  enabled: true
  path: "/metrics"
  addr: ":9091"          # optional: dedicated admin port, not on app HTTP
  auth_token: ""         # optional: Bearer token for /metrics
```

| Field | Default | Description |
|-------|---------|-------------|
| `enabled` | `true` | Collect and expose metrics |
| `path` | `/metrics` | Prometheus scrape path |
| `addr` | — | When set, expose metrics on a separate port (recommended in production) |
| `auth_token` | — | When set, requires `Authorization: Bearer <token>` |

Environment variable: `GOBLOCKS_METRICS_ENABLED=true|false`

## Scrape example

```yaml
# prometheus.yml
scrape_configs:
  - job_name: goblocks
    static_configs:
      - targets: ["localhost:8080"]
    metrics_path: /metrics
```

```bash
curl http://localhost:8080/metrics
```

## Metric catalog

### HTTP

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `goblocks_http_requests_total` | Counter | `method`, `path`, `status` | Total HTTP requests |
| `goblocks_http_request_duration_seconds` | Histogram | `method`, `path` | HTTP request latency |

`path` uses Gin route templates (e.g. `/users/:id`) to avoid high cardinality.

### gRPC

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `goblocks_grpc_server_handled_total` | Counter | `grpc_method`, `grpc_code` | Total gRPC requests |
| `goblocks_grpc_server_handling_seconds` | Histogram | `grpc_method` | gRPC handling latency |

### Resilience

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `goblocks_resilience_rate_limit_rejected_total` | Counter | `protocol`, `scope` | Rate limit rejections; `scope` is `global` / `user` / `route` |
| `goblocks_resilience_circuit_breaker_rejected_total` | Counter | `protocol` | Circuit breaker rejections |
| `goblocks_resilience_circuit_breaker_state` | Gauge | `name` | Breaker state: 0=closed, 1=half-open, 2=open |

### AI

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `goblocks_ai_requests_total` | Counter | `model`, `status` | Total AI requests |
| `goblocks_ai_request_duration_seconds` | Histogram | `model` | AI request latency |

`status` values: `success`, `error`, `rate_limited`, `circuit_open`

## Common PromQL

```promql
# HTTP QPS
rate(goblocks_http_requests_total[1m])

# HTTP P99 latency
histogram_quantile(0.99, rate(goblocks_http_request_duration_seconds_bucket[5m]))

# Rate limit rejections by layer
rate(goblocks_resilience_rate_limit_rejected_total[1m])
rate(goblocks_resilience_rate_limit_rejected_total{scope="user"}[1m])

# Circuit breaker open
goblocks_resilience_circuit_breaker_state == 2

# AI error rate
rate(goblocks_ai_requests_total{status="error"}[5m])
/ rate(goblocks_ai_requests_total[5m])
```

## Programmatic access

Metrics are wired by `app.App`. For custom instrumentation:

```go
import "github.com/ymhhh/goblocks/metrics"

reg := metrics.NewRegistry()
reg.HTTPRequestsTotal.WithLabelValues("GET", "/custom", "200").Inc()
```

Disable metrics: `metrics.enabled: false`; then `app.Metrics()` returns `nil`.
