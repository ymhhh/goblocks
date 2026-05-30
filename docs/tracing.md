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

## 说明

- 需自行初始化 OTel TracerProvider 与 exporter（Jaeger、OTLP 等）
- 与 Prometheus metrics 互补，不重复采集
