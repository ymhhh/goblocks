# 观测指标

Goblocks 内置 [Prometheus](https://prometheus.io/) 指标，默认开启，通过 HTTP 端点暴露。

## 配置

```yaml
metrics:
  enabled: true
  path: "/metrics"
```

| 字段 | 默认值 | 说明 |
|------|--------|------|
| `enabled` | `true` | 是否采集并暴露指标 |
| `path` | `/metrics` | Prometheus scrape 路径（挂载在 HTTP 服务上） |

环境变量：`GOBLOCKS_METRICS_ENABLED=true|false`

## 抓取示例

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

## 指标清单

### HTTP

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `goblocks_http_requests_total` | Counter | `method`, `path`, `status` | HTTP 请求总数 |
| `goblocks_http_request_duration_seconds` | Histogram | `method`, `path` | HTTP 请求延迟 |

`path` 使用 Gin 路由模板（如 `/users/:id`），避免高基数。

### gRPC

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `goblocks_grpc_server_handled_total` | Counter | `grpc_method`, `grpc_code` | gRPC 请求总数 |
| `goblocks_grpc_server_handling_seconds` | Histogram | `grpc_method` | gRPC 处理延迟 |

### Resilience

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `goblocks_resilience_rate_limit_rejected_total` | Counter | `protocol` | 限流拒绝次数（`http` / `grpc`） |
| `goblocks_resilience_circuit_breaker_rejected_total` | Counter | `protocol` | 熔断拒绝次数 |
| `goblocks_resilience_circuit_breaker_state` | Gauge | `name` | 熔断器状态：0=closed, 1=half-open, 2=open |

### AI

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `goblocks_ai_requests_total` | Counter | `model`, `status` | AI 请求总数 |
| `goblocks_ai_request_duration_seconds` | Histogram | `model` | AI 请求延迟 |

`status` 取值：`success`、`error`、`rate_limited`、`circuit_open`

## 常用 PromQL

```promql
# HTTP QPS
rate(goblocks_http_requests_total[1m])

# HTTP P99 延迟
histogram_quantile(0.99, rate(goblocks_http_request_duration_seconds_bucket[5m]))

# 限流拒绝率
rate(goblocks_resilience_rate_limit_rejected_total[1m])

# 熔断器是否打开
goblocks_resilience_circuit_breaker_state == 2

# AI 错误率
rate(goblocks_ai_requests_total{status="error"}[5m])
/ rate(goblocks_ai_requests_total[5m])
```

## 代码接入

指标由 `app.App` 自动装配。若需自定义采集：

```go
import "github.com/ymhhh/goblocks/metrics"

reg := metrics.NewRegistry()
reg.HTTPRequestsTotal.WithLabelValues("GET", "/custom", "200").Inc()
```

禁用指标：`metrics.enabled: false`，此时 `app.Metrics()` 返回 `nil`。
