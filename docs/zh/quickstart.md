# 快速入门

## 环境要求

- Go >= 1.22
- （可选）Redis — 分布式限流 `backend: redis` 时使用

## 新建服务（推荐）

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
goblocks new my-service --module github.com/acme/my-service
cd my-service
go mod tidy
go run .
```

验证：

```bash
curl http://localhost:8080/health
```

带 Demo（用户 API + 可选 AI/gRPC）：

```bash
goblocks new demo-svc --module github.com/acme/demo-svc --demo --with-ai
```

## 已有项目接入

```bash
go get github.com/ymhhh/goblocks@v0.3.0
```

`main.go` 或 `infrastructure/run.go` 最小示例：

```go
package main

import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    gblocksapp "github.com/ymhhh/goblocks/app"
    "github.com/ymhhh/goblocks/config"
    "github.com/ymhhh/goblocks/resilience"
)

func main() {
    cfg, err := config.Load("config/config.yaml")
    if err != nil {
        panic(err)
    }

    app := gblocksapp.New(cfg).WithHTTP(func(engine *gin.Engine, _ *resilience.Policy) {
        engine.GET("/health", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"status": "ok"})
        })
    })

    if err := app.Run(context.Background()); err != nil {
        panic(err)
    }
}
```

> `app.Run` 已自动挂载 L1 全局限流与熔断检查；配置了 `resilience.rate_limit.routes` 时自动挂载 L3。

## 最小配置

`config/config.yaml`：

```yaml
server:
  http:
    addr: ":8080"
  grpc:
    enabled: false
resilience:
  rate_limit:
    backend: memory
    global:
      rps: 100
      burst: 200
logger:
  level: info
metrics:
  enabled: true
  path: /metrics
```

## 下一步

- [分层限流指南](rate-limiting.md) — 启用 L2 用户限流与 L3 路由限流
- [配置参考](../configuration.md) — 完整配置项
- [包 API 参考](../packages.md) — HTTP/gRPC/AI 代码示例
