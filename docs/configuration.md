# 配置参考

Goblocks 使用 YAML 配置文件，默认路径 `config/config.yaml`。可通过 `config.Load(path)` 加载，环境变量以 `GOBLOCKS_` 前缀覆盖部分字段。

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
log:
  level: info
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

## log

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `level` | string | `info` | 日志级别（框架使用 `log/slog`） |

## 环境变量

| 变量 | 覆盖字段 |
|------|----------|
| `GOBLOCKS_HTTP_ADDR` | `server.http.addr` |
| `GOBLOCKS_GRPC_ADDR` | `server.grpc.addr` |
| `GOBLOCKS_AI_API_KEY` | `ai.api_key` |
| `GOBLOCKS_AI_BASE_URL` | `ai.base_url` |
| `GOBLOCKS_LOG_LEVEL` | `log.level` |

### 占位符展开

配置中 `api_key: "${OPENAI_API_KEY}"` 会在加载时替换为对应环境变量值。

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
