# 分层限流指南

Goblocks 采用 **三层限流模型**，由粗到细保护服务：

| 层级 | 名称 | 目的 | 默认挂载 |
|------|------|------|----------|
| **L1** | 全局限流 | 保护整个服务/集群不被打垮 | ✅ `app.Run` |
| **L2** | 用户限流 | 公平配额，防止单用户滥用 | ❌ 业务 `infrastructure` |
| **L3** | 路由限流 | 昂贵 API 单独控流 | ✅ `app.Run`（config 有 `routes` 时） |

## 配置

```yaml
resilience:
  rate_limit:
    backend: memory          # memory | redis
    global:                  # L1
      rps: 1000
      burst: 2000
    redis:                   # backend=redis 时必填
      addr: "redis://localhost:6379/0"
      key_prefix: "goblocks:rl:"
    user:                    # L2
      enabled: true
      default_rps: 20
      burst: 40
    routes:                  # L3
      - method: POST
        path: /ai/chat
        rps: 5
        burst: 10
      - method: GRPC
        path: /my.v1.UserService/GetUser
        rps: 10
        burst: 20
```

环境变量：

| 变量 | 作用 |
|------|------|
| `GOBLOCKS_RATE_LIMIT_BACKEND` | 覆盖 `backend` |
| `GOBLOCKS_REDIS_ADDR` | 覆盖 `redis.addr` |

## L1 全局限流

- **Key**：`global`（Redis 中为 `{prefix}global`）
- **算法**：令牌桶（memory）或 GCRA（redis）
- **挂载**：`app.Run` → `GlobalRateLimit` + `BreakerCheck`
- **超限**：HTTP 429 / gRPC `ResourceExhausted`

无需额外代码，配置 `global.rps` / `global.burst` 即可。

## L2 用户限流

### 启用条件

1. 配置 `user.enabled: true`
2. 在 `infrastructure/registerHTTP` 挂载 `UserRateLimit`
3. **鉴权中间件必须先于 L2**，将 userId 写入 context

### HTTP 示例

```go
import httpmiddleware "github.com/ymhhh/goblocks/http/middleware"

// 方式一：全站挂载（适合 JWT 已解析到 context 的场景）
engine.Use(authMiddleware)
engine.Use(httpmiddleware.UserRateLimit(policy, metrics))

// 方式二：单路由链（Demo 模板用法）
users.GET("/:id",
    func(c *gin.Context) {
        httpmiddleware.GinContextWithUserID(c, c.Param("id"))
    },
    httpmiddleware.UserRateLimit(policy, nil),
    handler,
)
```

### 用户 Key 规则

| 场景 | Key |
|------|-----|
| context 有 userId | `user:{userId}` |
| 未登录 / 未注入 | `user:anonymous`（共用一个桶） |

### gRPC

链式挂载 `UserUnaryServerInterceptor`，客户端 metadata 携带：

```
x-user-id: alice
```

## L3 路由限流

### 配置驱动（推荐）

在 `routes` 中声明规则后，**`app.Run` 自动挂载**，无需在代码中重复 `RouteRateLimit`。

**HTTP path** 须与 Gin **路由模板**一致（`c.FullPath()`），例如注册了 `POST /ai/chat`，则：

```yaml
routes:
  - method: POST
    path: /ai/chat
    rps: 5
    burst: 10
```

**gRPC** 使用 `method: GRPC`，`path` 为 FullMethod：

```yaml
routes:
  - method: GRPC
    path: /my.v1.UserService/GetUser
    rps: 10
    burst: 10
```

### 后端 Key

限流计数 key 为 `route:{METHOD}:{path}`，例如 `route:POST:/ai/chat`。

### 代码手动挂载（可选）

对未写入 config 的路由，可用 `RateLimitByKey`：

```go
engine.Use(httpmiddleware.RateLimitByKey(
    func(c *gin.Context) string {
        return resilience.RouteKey(c.Request.Method, c.FullPath())
    },
    resilience.LimitRule{RPS: 5, Burst: 10},
    resilience.ScopeRoute,
    policy, metrics,
))
```

## 后端选择

| backend | 适用场景 | 说明 |
|---------|----------|------|
| `memory` | 开发、单 Pod、单测 | 进程内 map，重启丢失计数 |
| `redis` | 生产多副本 | GCRA，跨 Pod 一致；启动时 Ping Redis |

## 中间件顺序（HTTP）

```
Recovery → Metrics → Tracing
  → L1 GlobalRateLimit      ← app
  → BreakerCheck            ← app
  → L3 RouteRateLimit       ← app（有 routes 时）
  → [Auth]                  ← infrastructure
  → L2 UserRateLimit        ← infrastructure
  → Handler
```

## 观测

Prometheus 指标 `goblocks_resilience_rate_limit_rejected_total` 带标签：

- `protocol`：`http` / `grpc`
- `scope`：`global` / `user` / `route`

```promql
rate(goblocks_resilience_rate_limit_rejected_total{scope="user"}[1m])
```

详见 [观测指标](../metrics.md)。

## 常见问题

**Q：配置了 L3 但未生效？**

- 检查 `path` 是否与 Gin `FullPath` 一致（非实际 URL 中带参数的值，而是模板如 `/users/:id`）
- 确认 `routes` 非空且 method 大小写无关（内部会 `ToUpper`）

**Q：L2 未生效？**

- 确认 `user.enabled: true`
- 确认已挂载 `UserRateLimit` 且 context 中已有 userId

**Q：L1 和 L2 关系？**

- 先过 L1 全站桶，再过 L2 用户桶；两层独立计数。

## 相关文档

- [架构设计 — 限流目录映射](../architecture.md#分层限流目录与职责)
- [配置参考 — rate_limit](../configuration.md#rate_limit分层限流)
- [包 API — resilience](../packages.md#resilience)
