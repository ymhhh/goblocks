# 分布式追踪（可选）

Goblocks 不内置 OpenTelemetry SDK，通过 **可选 hook** 接入用户自选的 OTel 中间件。

## HTTP

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

app.New(cfg).
    WithHTTP(registerRoutes).
    WithHTTPTracing(otelgin.Middleware("my-service"))
```

`WithHTTPTracing` 注入的中间件在 Recovery / metrics 之后、resilience 之前执行。

## gRPC

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

app.New(cfg).
    WithGRPC(registerGRPC).
    WithGRPCTracing(
        grpc.StatsHandler(otelgrpc.NewServerHandler()),
    )
```

## 日志与 trace ID

Goblocks 中 **tracing 与 logging 相互独立**。OTel 中间件将 span 写入 `context` 并导出到 tracing 后端，**不会**自动在 logrus 日志中加入 `trace_id` / `span_id`。

框架生命周期日志（`app.Run` 启动/关闭）无请求 context，不会包含 trace ID。

在 handler 中关联日志与 trace，请使用 go-common 的 [`logger.LFromContext`](https://github.com/ymhhh/go-common)（需 OTel 中间件已写入 context）：

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/ymhhh/go-common/logger"
)

func chatHandler(c *gin.Context) {
    logger.LFromContext(c.Request.Context()).Info("handling chat request")
    // 启用 otelgin 后，JSON 日志会包含 trace_id、span_id
}
```

gRPC handler：将 RPC 的 `ctx` 传入 `logger.LFromContext(ctx)`。

未启用 OTel 中间件时，`LFromContext` 与 `L()` 行为相同（无 trace 字段）。

## 说明

- 需自行初始化 OTel TracerProvider 与 exporter（Jaeger、OTLP 等）
- 与 Prometheus metrics 互补，不重复采集
- 默认 `logger.L()` 不含 trace ID；请求内日志请用 `logger.LFromContext(ctx)`
