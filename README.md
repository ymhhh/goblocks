# goblocks

Go 服务端框架：Gin HTTP（HTTP/1、HTTP/2、HTTP/3）、gRPC、OpenAI 兼容 AI 客户端，统一熔断与限流；配合 [goblocks-cli](https://github.com/ymhhh/goblocks-cli) 脚手架生成 [洋葱架构 / DDD](https://github.com/ymhhh/ddd-onion-sample) 服务工程。

**完整文档：[docs/](docs/README.md)**

## 特性

- **HTTP**：基于 Gin，TLS 启用时自动协商 HTTP/2，可选 HTTP/3（QUIC）
- **gRPC**：Server / Client 封装，Unary Interceptor 接入 resilience
- **AI**：OpenAI 兼容 API（OpenAI、Azure、Ollama 等），统一 Chat 接口
- **Resilience**：熔断（`sony/gobreaker`）+ 令牌桶限流（`golang.org/x/time/rate`）
- **Metrics**：Prometheus 观测指标（HTTP/gRPC/resilience/AI）
- **脚手架**：由独立仓库 [goblocks-cli](https://github.com/ymhhh/goblocks-cli) 提供 `goblocks new` 命令

## 架构

```
handlers → domain → core
infrastructure（组合根）→ goblocks/app → http | grpc | ai
                              ↓
                         resilience | metrics
```

## 快速开始

### 安装框架

```bash
go get github.com/ymhhh/goblocks@latest
```

### 安装脚手架 CLI

脚手架已拆至独立仓库 [goblocks-cli](https://github.com/ymhhh/goblocks-cli)：

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
```

> **迁移说明**：旧版 `go install github.com/ymhhh/goblocks/cmd/goblocks@...` 已废弃，请改用上述命令。CLI 命令仍为 `goblocks new`。

### 生成并运行服务

```bash
goblocks new my-service --module github.com/acme/my-service
cd my-service
go mod tidy && go run .
curl http://localhost:8080/health
```

详见 [脚手架 CLI 文档](docs/scaffold.md) 与 [goblocks-cli README](https://github.com/ymhhh/goblocks-cli)。

## 配置

`config/config.yaml` 示例：

```yaml
server:
  http:
    addr: ":8080"
  grpc:
    enabled: true
    addr: ":9090"
resilience:
  rate_limit:
    rps: 100
    burst: 200
metrics:
  enabled: true
  path: "/metrics"
```

完整配置见 [配置参考](docs/configuration.md)。

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
| [脚手架 CLI](docs/scaffold.md) | `goblocks new` 用法（goblocks-cli） |
| [包 API 参考](docs/packages.md) | 各包使用说明 |
| [观测指标](docs/metrics.md) | Prometheus 指标与 PromQL |
| [开发指南](docs/development.md) | 框架开发与发布 |

## 包结构

```
goblocks/
├── app/           # 应用生命周期编排
├── config/        # YAML 配置加载
├── resilience/    # 熔断、限流、Policy
├── http/          # Gin HTTP/1/2/3
├── grpc/          # gRPC Server/Client + Interceptors
├── ai/            # OpenAI 兼容 Client
├── metrics/       # Prometheus 观测指标
└── docs/          # 说明文档
```

脚手架 CLI 见 [goblocks-cli](https://github.com/ymhhh/goblocks-cli)。

## 开发

```bash
make test    # go test ./... -race
make lint    # go vet && go fmt
```

## License

GPL-3.0（与 ddd-onion-sample 保持一致）
