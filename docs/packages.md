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

熔断 + **分层限流**（L1 全局 / L2 用户 / L3 路由）统一抽象。

### RateLimiter 与后端

```go
// 接口：按 key + rule 检查（memory 或 redis）
type RateLimiter interface {
    Allow(ctx context.Context, key string, rule LimitRule) (bool, error)
}

backend, err := resilience.NewRateLimiterBackend(cfg.Resilience.RateLimit)
// memory: 进程内令牌桶；redis: GCRA，key 前缀见 config redis.key_prefix
```

Key 约定（`resilience/keyed.go`）：

| 层 | Key 示例 |
|----|----------|
| L1 | `global` |
| L2 | `user:alice`（无 userId 时为 `user:anonymous`） |
| L3 | `route:POST:/ai/chat` |

### Policy

```go
import "github.com/ymhhh/goblocks/resilience"

// 从配置创建（需处理 error；redis backend 需有效 addr）
policy, err := resilience.NewPolicyFromConfig(cfg.Resilience)

// 分层限流检查
policy.AllowGlobal(ctx)                       // L1
policy.AllowUser(ctx, "")                     // L2（context 或显式 key）
policy.AllowRoute(ctx, "POST", "/ai/chat")    // L3（config routes 有规则时生效）

// 用户身份（HTTP）
ctx = resilience.ContextWithUserID(ctx, userID)
key := resilience.UserKeyFromContext(ctx)

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

engine.Use(httpmiddleware.GlobalRateLimit(policy, metrics)) // L1（app 默认已挂载）
engine.Use(httpmiddleware.BreakerCheck(policy, metrics))
engine.Use(httpmiddleware.UserRateLimit(policy, metrics))  // L2（infrastructure 挂载）
engine.Use(httpmiddleware.RouteRateLimit(policy, metrics)) // L3
// 或自定义 key：
engine.Use(httpmiddleware.RateLimitByKey(keyFn, rule, scope, policy, metrics))
```

`app.App` 默认挂载 `GlobalRateLimit` + `BreakerCheck`。

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
    grpc.ChainUnaryInterceptor(
        grpcinterceptors.UnaryServerInterceptor(policy, metrics),       // L1 + breaker
        grpcinterceptors.UserUnaryServerInterceptor(policy, metrics),   // L2（infrastructure）
        grpcinterceptors.RouteUnaryServerInterceptor(policy, metrics),  // L3
    ),
}
```

`app.Run` 在 gRPC 启用时默认只链接 **L1** `UnaryServerInterceptor`。L2/L3 在 `registerGRPC` 或自定义 `ServerOption` 中追加。

```go
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

生成工程的标准模式（Demo，`--demo` 含 L2/L3 示例）：

```go
// infrastructure/run.go
func (a *App) registerHTTP(engine *gin.Engine, policy *resilience.Policy) {
    // L2：路由链内注入 userId 后限流
    users := engine.Group("/users")
    users.GET("/:id",
        func(c *gin.Context) {
            httpmiddleware.GinContextWithUserID(c, c.Param("id"))
        },
        httpmiddleware.UserRateLimit(policy, nil),
        a.getUserHandler,
    )

    // L3：昂贵 API 路由组
    ai := engine.Group("/ai")
    ai.Use(httpmiddleware.RouteRateLimit(policy, nil))
    ai.POST("/chat", a.AIHandler.Chat)
}

func Run(configPath string) error {
    cfg, _ := config.Load(configPath)
    app, _ := NewApp(Config{ConfigPath: configPath})

    gblocks := gblocksapp.New(cfg).WithHTTP(app.registerHTTP)
    if cfg.Server.GRPC.Enabled {
        gblocks = gblocks.WithGRPC(app.registerGRPC)
    }
    return gblocks.Run(context.Background())
}
```

```go
// infrastructure/grpc_server.go — L2 需 metadata x-user-id
func (a *App) registerGRPC(server *grpc.Server, policy *resilience.Policy) {
    _ = policy
    // app 已挂 L1；追加 L2/L3 见 grpc ChainUnaryInterceptor 示例
}
```

业务 Handler 在 `NewApp` 中构造，路由在 `registerHTTP` / `registerGRPC` 中绑定，保持 **handlers / domain / core 不感知限流基础设施**。
