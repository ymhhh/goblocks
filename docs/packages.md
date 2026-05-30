# 包 API 参考

## config

加载 YAML 配置，支持环境变量覆盖。

```go
import "github.com/ymhhh/goblocks/config"

cfg, err := config.Load("config/config.yaml")
defaults := config.Default()
```

主要类型：`Config`、`ServerConfig`、`ResilienceConfig`、`AIConfig`。

详见 [配置参考](configuration.md)。

---

## resilience

熔断 + 限流统一抽象。

### Policy

```go
import "github.com/ymhhh/goblocks/resilience"

// 从配置创建
policy := resilience.NewPolicyFromConfig(cfg.Resilience)

// 手动创建
breaker := resilience.NewBreaker(resilience.BreakerSettings{
    Name:        "my-service",
    MaxRequests: 3,
    Interval:    60 * time.Second,
    Timeout:     30 * time.Second,
})
limiter := resilience.NewLimiter(100, 200)
policy := resilience.NewPolicy(breaker, limiter)

// 限流检查（入站）
if err := policy.Allow(); err != nil {
    // ErrRateLimited
}

// 熔断包裹（出站/业务执行）
result, err := policy.Execute(func() (any, error) {
    return doSomething()
})
// err 可能为 ErrCircuitOpen
```

### 错误

| 错误 | 含义 |
|------|------|
| `resilience.ErrRateLimited` | 令牌桶拒绝 |
| `resilience.ErrCircuitOpen` | 熔断器打开 |

---

## http

Gin 封装，支持 HTTP/1、TLS 下 HTTP/2、可选 HTTP/3。

```go
import (
    "github.com/gin-gonic/gin"
    gblockshttp "github.com/ymhhh/goblocks/http"
)

engine := gin.New()
srv := gblockshttp.NewServer(engine, gblockshttp.Config{
    Addr: ":8080",
    TLS: gblockshttp.TLSOptions{
        Enabled:  true,
        CertFile: "cert.pem",
        KeyFile:  "key.pem",
    },
    H3: gblockshttp.H3Options{
        Enabled: true,
        Addr:    ":8443",
    },
})
srv.Start()
defer srv.Shutdown()
```

### Middleware

```go
import httpmiddleware "github.com/ymhhh/goblocks/http/middleware"

engine.Use(httpmiddleware.Resilience(policy))           // 仅限流
engine.Use(httpmiddleware.ResilienceWithBreaker(policy)) // 限流 + 熔断状态检查
```

`app.App` 默认使用 `ResilienceWithBreaker`。

---

## grpc

### Server

```go
import (
    gblocksgrpc "github.com/ymhhh/goblocks/grpc"
    grpcinterceptors "github.com/ymhhh/goblocks/grpc/interceptors"
    "google.golang.org/grpc"
)

opts := []grpc.ServerOption{
    grpc.UnaryInterceptor(grpcinterceptors.UnaryServerInterceptor(policy)),
}
srv := gblocksgrpc.NewServer(gblocksgrpc.Config{Addr: ":9090"}, opts...)
// srv.GRPCServer() 注册 pb 服务
srv.Start()
defer srv.Shutdown()
```

### Client

```go
conn, err := gblocksgrpc.Dial(
    gblocksgrpc.ClientConfig{Addr: "localhost:9090"},
    grpc.WithUnaryInterceptor(grpcinterceptors.UnaryClientInterceptor(policy)),
)
```

---

## ai

OpenAI 兼容 Chat 接口。

```go
import "github.com/ymhhh/goblocks/ai"

client := ai.NewOpenAIClient(ai.OpenAIConfig{
    BaseURL: "https://api.openai.com/v1",
    APIKey:  os.Getenv("OPENAI_API_KEY"),
    Model:   "gpt-4o-mini",
    Policy:  policy,  // 可选，接入熔断限流
})

resp, err := client.Chat(ctx, ai.ChatRequest{
    Messages: []ai.Message{
        {Role: "user", Content: "Hello"},
    },
})
// resp.Content, resp.Model
```

`ai.Client` 接口便于测试时注入 Mock。

---

## app

应用生命周期编排，推荐在 `infrastructure/run.go` 中使用。

```go
import (
    "context"
    gblocksapp "github.com/ymhhh/goblocks/app"
    "github.com/ymhhh/goblocks/config"
    "github.com/ymhhh/goblocks/resilience"
    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
)

func main() {
    cfg, _ := config.Load("config/config.yaml")

    application := gblocksapp.New(cfg).
        WithHTTP(func(engine *gin.Engine, policy *resilience.Policy) {
            engine.GET("/health", healthHandler)
        }).
        WithGRPC(func(server *grpc.Server, policy *resilience.Policy) {
            // registerGRPCServices(server)
        })

    // AI Client（需 cfg.AI.Enabled = true）
    aiClient := application.AIClient()

    if err := application.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
    _ = aiClient
}
```

### 主要方法

| 方法 | 说明 |
|------|------|
| `New(cfg)` | 创建 App，自动构建 Policy |
| `WithHTTP(fn)` | 注册 HTTP 路由 |
| `WithGRPC(fn)` | 注册 gRPC 服务（`server.grpc.enabled: true` 时**必须**调用，否则启动失败） |
| `Policy()` | 获取共享 Policy |
| `AIClient()` | 获取 AI Client（lazy init） |
| `Config()` | 获取配置 |
| `Run(ctx)` | 启动并阻塞至信号关闭 |
| `Shutdown(ctx)` | 手动关闭 |

---

## 典型组合：infrastructure 组合根

生成工程的标准模式（Demo）：

```go
// infrastructure/run.go
func Run(configPath string) error {
    cfg, err := config.Load(configPath)
    app, err := NewApp(Config{ConfigPath: configPath})

    gblocks := gblocksapp.New(cfg).WithHTTP(app.registerHTTP)
    if cfg.Server.GRPC.Enabled {
        gblocks = gblocks.WithGRPC(app.registerGRPC) // 必须显式注册
    }
    return gblocks.Run(context.Background())
}

// infrastructure/grpc_server.go
func (a *App) registerGRPC(server *grpc.Server, _ *resilience.Policy) {
    healthServer := health.NewServer()
    grpc_health_v1.RegisterHealthServer(server, healthServer)
    healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
}
```

业务 Handler 在 `NewApp` 中构造，路由在 `registerHTTP` / `registerGRPC` 中绑定，保持 **handlers 不感知 HTTP 框架细节**（Demo 中 Gin handler 在 infrastructure 层做薄适配）。
