# 配置参考

Goblocks 使用 YAML 配置文件，默认路径 `config/config.yaml`。通过 [`github.com/ymhhh/go-common/config`](https://github.com/ymhhh/go-common) 加载，支持 `#include`、 `${ENV}` 与 `${a.b.c}` 引用；`GOBLOCKS_*` 环境变量在加载后覆盖部分字段。

## 高级特性

### 拆分配置（#include）

```yaml
#include base.yaml
#include secrets.yaml
server:
  http:
    addr: ":8080"
```

### 占位符

- `${OPENAI_API_KEY}` — 环境变量
- `${server.http.addr}` — 同文件内交叉引用

## 完整示例

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
    rps: 100
    burst: 200
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

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `addr` | string | `:8080` | HTTP/HTTPS 监听地址 |
| `tls.enabled` | bool | `false` | 启用 TLS（同时开启 HTTP/2 ALPN） |
| `tls.cert_file` | string | — | TLS 证书路径 |
| `tls.key_file` | string | — | TLS 私钥路径 |
| `h3.enabled` | bool | `false` | 启用 HTTP/3（需 TLS） |
| `h3.addr` | string | `:8443` | HTTP/3 QUIC 监听地址 |
| `health.enabled` | bool | `true` | 注册 `/health` 与 `/ready` 探针 |
| `health.liveness_path` | string | `/health` | 存活探针路径 |
| `health.readiness_path` | string | `/ready` | 就绪探针路径（检查熔断器与 gRPC 状态） |

### server.grpc

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `enabled` | bool | `true` | 是否启动 gRPC 服务（需在 `infrastructure/run.go` 中调用 `WithGRPC(registerGRPC)`） |
| `addr` | string | `:9090` | gRPC 监听地址 |

> HTTP 与 gRPC 使用不同端口，不可合并为同一 listener。

## resilience

### breaker（熔断）

基于 [sony/gobreaker](https://github.com/sony/gobreaker)。连续失败 3 次触发打开（ReadyToTrip 逻辑见 `resilience/breaker.go`）。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `max_requests` | uint32 | `3` | 半开状态允许的最大请求数 |
| `interval` | duration | `60s` | 统计窗口（关闭状态下重置计数） |
| `timeout` | duration | `30s` | 打开状态持续时间，之后进入半开 |

### rate_limit（限流）

基于令牌桶 [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate)。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `rps` | float64 | `100` | 每秒令牌生成速率 |
| `burst` | int | `200` | 桶容量（突发上限） |

HTTP 超限返回 **429**；gRPC 超限返回 **ResourceExhausted**；熔断打开 HTTP 返回 **503**，gRPC 返回 **Unavailable**。

## ai

OpenAI 兼容 HTTP API，底层使用 [go-openai](https://github.com/sashabaranov/go-openai)。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `enabled` | bool | `false` | 是否初始化 AI Client |
| `base_url` | string | `https://api.openai.com/v1` | API Base URL |
| `api_key` | string | — | API Key，支持 `${ENV_VAR}` 占位 |
| `model` | string | `gpt-4o-mini` | 默认模型 |

### AI 接入示例

**OpenAI**

```yaml
ai:
  enabled: true
  base_url: "https://api.openai.com/v1"
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini"
```

**Ollama（本地）**

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

基于 [`github.com/ymhhh/go-common/logger`](https://github.com/ymhhh/go-common)（logrus）。

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `level` | string | `info` | 日志级别：debug/info/warn/error |
| `format` | string | `text` | `text` 或 `json` |
| `output` | string | `stderr` | `stdout`/`stderr`/`discard`/文件路径/`file:/path` |
| `reportCaller` | bool | `false` | 是否记录调用方 |
| `file.path` | string | — | 文件输出路径 |
| `file.rotate.enabled` | bool | `false` | 启用 lumberjack 轮转 |
| `file.rotate.maxSizeMB` | int | `100` | 单文件最大 MB |
| `file.rotate.maxBackups` | int | `7` | 保留备份数 |
| `file.rotate.maxAgeDays` | int | `7` | 保留天数 |
| `text.disableColors` | bool | `false` | text 格式禁用颜色 |
| `text.fullTimestamp` | bool | `false` | text 格式完整时间戳 |
| `json.prettyPrint` | bool | `false` | JSON 格式化输出 |

## metrics

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `enabled` | bool | `true` | 是否采集并暴露 Prometheus 指标 |
| `path` | string | `/metrics` | 指标 HTTP 路径 |
| `addr` | string | — | 独立 metrics 监听地址（设则不与业务 HTTP 共端口） |
| `auth_token` | string | — | Bearer token 保护 metrics 端点 |

详见 [观测指标](metrics.md)。

## 环境变量

| 变量 | 覆盖字段 |
|------|----------|
| `GOBLOCKS_HTTP_ADDR` | `server.http.addr` |
| `GOBLOCKS_GRPC_ADDR` | `server.grpc.addr` |
| `GOBLOCKS_AI_API_KEY` | `ai.api_key` |
| `GOBLOCKS_AI_BASE_URL` | `ai.base_url` |
| `GOBLOCKS_LOGGER_LEVEL` | `logger.level` |
| `GOBLOCKS_LOG_LEVEL` | `logger.level`（已废弃，仍兼容） |
| `GOBLOCKS_METRICS_ENABLED` | `metrics.enabled` |

### 占位符展开

配置中 `api_key: "${OPENAI_API_KEY}"` 由 go-common 在加载时解析为环境变量值。加载完成后，`GOBLOCKS_*` 变量仍可覆盖对应字段（优先级更高）。

## HTTP/2 与 HTTP/3 启用步骤

### HTTP/2

1. 准备 TLS 证书
2. 设置 `server.http.tls.enabled: true` 并填写 `cert_file`、`key_file`
3. 客户端使用 HTTPS 访问，ALPN 自动协商 h2

```bash
curl --http2 -k https://localhost:8080/health
```

### HTTP/3

1. 完成 HTTP/2 的 TLS 配置（H3 复用同一证书）
2. 设置 `server.http.h3.enabled: true`
3. 确保防火墙放行 UDP 端口（默认 `:8443`）

> 生产环境 HTTP/3 部署需考虑 UDP 负载均衡与 CDN 兼容性，默认建议关闭。

## 代码加载

```go
cfg, err := config.Load("config/config.yaml")
if err != nil {
    log.Fatal(err)
}

// 或使用默认值
cfg := config.Default()
```
