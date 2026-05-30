# goblocks

Go 服务端框架：Gin HTTP（HTTP/1、HTTP/2、HTTP/3）、gRPC、OpenAI 兼容 AI 客户端，统一熔断与限流；CLI 脚手架生成 [洋葱架构 / DDD](https://github.com/ymhhh/ddd-onion-sample) 服务骨架。

**完整文档：[docs/](docs/README.md)**

## 特性

- **HTTP**：基于 Gin，TLS 启用时自动协商 HTTP/2，可选 HTTP/3（QUIC）
- **gRPC**：Server / Client 封装，Unary Interceptor 接入 resilience
- **AI**：OpenAI 兼容 API（OpenAI、Azure、Ollama 等），统一 Chat 接口
- **Resilience**：熔断（`sony/gobreaker`）+ 令牌桶限流（`golang.org/x/time/rate`）
- **脚手架**：`goblocks new` 生成洋葱架构工程（空骨架或 User Demo）

## 架构

```
handlers → domain → core
infrastructure（组合根）→ goblocks/app → http | grpc | ai
                              ↓
                         resilience
```

生成工程目录与 [ddd-onion-sample](https://github.com/ymhhh/ddd-onion-sample) 一致：

| 目录 | 职责 |
|------|------|
| `core` | 实体 / 值对象 |
| `domain` | 仓储接口（端口） |
| `handlers` | 应用层用例 |
| `infrastructure` | 依赖注入、路由注册、启动 |
| `config/` | YAML 配置 |

## 快速开始

### 安装 CLI

```bash
go install github.com/ymhhh/goblocks/cmd/goblocks@latest
```

本地开发（本仓库）：

```bash
make build
./bin/goblocks new my-service --module github.com/acme/my-service
```

### 生成空工程

```bash
goblocks new my-service --module github.com/acme/my-service
cd my-service
# 本地引用 goblocks 时追加 replace
echo "replace github.com/ymhhh/goblocks => /path/to/goblocks" >> go.mod
go mod tidy
go run .
```

### 生成 Demo 工程

```bash
goblocks new demo-svc --module github.com/acme/demo-svc --demo
go mod tidy && go run .
curl http://localhost:8080/users/1
```

### CLI 参数

```bash
goblocks new [output-dir] \
  --module github.com/acme/my-service \  # Go module 路径
  --demo \                               # 生成 User Demo
  --with-grpc \                          # 额外生成 proto 示例（gRPC 健康检查默认已内置）
  --with-ai                              # 含 AI Chat 端点
```

## 配置

`config/config.yaml` 示例：

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

环境变量覆盖（`GOBLOCKS_` 前缀）：

| 变量 | 说明 |
|------|------|
| `GOBLOCKS_HTTP_ADDR` | HTTP 监听地址 |
| `GOBLOCKS_GRPC_ADDR` | gRPC 监听地址 |
| `GOBLOCKS_AI_API_KEY` | AI API Key |
| `GOBLOCKS_AI_BASE_URL` | AI Base URL |
| `GOBLOCKS_LOG_LEVEL` | 日志级别 |

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

### HTTP/2 与 HTTP/3

- **HTTP/2**：设置 `server.http.tls.enabled: true` 并提供证书，ALPN 自动协商 h2
- **HTTP/3**：需 TLS，另设 `server.http.h3.enabled: true`（独立 QUIC 端口，默认 `:8443`）

## 框架 API 示例

```go
import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    gblocksapp "github.com/ymhhh/goblocks/app"
    "github.com/ymhhh/goblocks/config"
    "github.com/ymhhh/goblocks/resilience"
)

func main() {
    cfg, _ := config.Load("config/config.yaml")
    app := gblocksapp.New(cfg).WithHTTP(func(engine *gin.Engine, policy *resilience.Policy) {
        engine.GET("/health", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"status": "ok"})
        })
    })
    _ = app.Run(context.Background())
}
```

## 文档

| 文档 | 说明 |
|------|------|
| [文档首页](docs/README.md) | 目录与快速链接 |
| [架构设计](docs/architecture.md) | 分层、依赖方向、请求流转 |
| [配置参考](docs/configuration.md) | YAML 与环境变量 |
| [脚手架 CLI](docs/scaffold.md) | `goblocks new` 用法 |
| [包 API 参考](docs/packages.md) | 各包使用说明 |
| [开发指南](docs/development.md) | 测试、发布、贡献 |

## 包结构

```
goblocks/
├── app/           # 应用生命周期编排
├── config/        # YAML 配置加载
├── resilience/    # 熔断、限流、Policy
├── http/          # Gin HTTP/1/2/3
├── grpc/          # gRPC Server/Client + Interceptors
├── ai/            # OpenAI 兼容 Client
├── cmd/goblocks/  # CLI 脚手架
└── internal/scaffold/
```

## 开发

```bash
make test    # go test ./... -race
make build   # 构建 CLI
make lint    # go vet && go fmt
```

## License

GPL-3.0（与 ddd-onion-sample 保持一致）
