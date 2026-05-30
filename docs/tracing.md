# Distributed tracing (optional)

Goblocks does not bundle an OpenTelemetry SDK. Use **optional hooks** to plug in your own OTel middleware.

## HTTP

```go
import (
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

app.New(cfg).
    WithHTTP(registerRoutes).
    WithHTTPTracing(otelgin.Middleware("my-service"))
```

Middleware from `WithHTTPTracing` runs after Recovery / metrics and before resilience.

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

## Logging and trace IDs

Tracing and logging are **independent** in Goblocks. OTel middleware writes spans to `context` and your exporter; it does **not** automatically add `trace_id` / `span_id` to logrus output.

Framework lifecycle logs (`app.Run` startup/shutdown) have no request context and never include trace IDs.

To correlate logs with traces in handlers, use [`logger.LFromContext`](https://github.com/ymhhh/go-common) from go-common (requires OTel on the request context):

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/ymhhh/go-common/logger"
)

func chatHandler(c *gin.Context) {
    logger.LFromContext(c.Request.Context()).Info("handling chat request")
    // JSON log includes trace_id and span_id when otelgin middleware is enabled
}
```

gRPC handlers: pass the RPC `ctx` to `logger.LFromContext(ctx)`.

Without OTel middleware, `LFromContext` behaves like `L()` (no trace fields).

## Notes

- Initialize your own OTel TracerProvider and exporter (Jaeger, OTLP, etc.)
- Complements Prometheus metrics; does not duplicate collection
- Default `logger.L()` does not include trace IDs; use `logger.LFromContext(ctx)` in request-scoped code
