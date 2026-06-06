# 更新日志（中文摘要）

完整英文记录见 [CHANGELOG.md](../../CHANGELOG.md)。

## v0.3.2 — 2026-06-06

### 破坏性变更

- `app.New(cfg)` 改为 `(*App, error)`，不再 panic
- 移除废弃类型：`resilience.Limiter`、`NewPolicy`、`legacyLimiterAdapter`、`Policy.Limiter`

### 新增

- gRPC stream 拦截器：`StreamServerInterceptor`、`StreamClientInterceptor` 等，streaming 调用现在经过限流和熔断
- `app.Run` 在 L2 启用时自动挂载 gRPC user 拦截器
- 可配置的优雅关闭超时 `server.shutdown_timeout`（默认 30s）
- 可选熔断 `resilience.breaker.enabled`（默认 true）
- `metrics.Registry.PromRegistry()` 暴露底层 `*prometheus.Registry`

### 修复

- **内存泄漏：** `MemoryRateLimiter` 增加后台清理，可配置 TTL（默认 5m）和清理间隔（默认 1m）
- **连接泄漏：** `RedisRateLimiter` 增加 `Close()`，接入 `App.Shutdown`
- **竞态条件：** `AIClient()` 懒初始化用 `sync.Once` 保护
- `AllowUser` / `AllowRoute` 增加 nil receiver 保护
- AI 客户端改用 `AllowGlobal(ctx)`，保留追踪上下文
- gRPC 客户端拦截器指标标签改为 `grpc_client`
- Redis 限流器防止 `int(0.5) == 0` 截断
- burst 默认值统一为 `RPS * 2`
- gRPC 路由限流 key 直接用方法路径，不再硬编码 `"GRPC"`
- L3 路由中间件跳过未注册路由

### 变更

- 移除 `Metrics.Path` 重复默认值
- `Policy.allow()` 移除未使用的 `Scope` 参数
- `config.Normalized()` burst 默认值改为动态计算

## v0.3.1 — 2026-05-30

### 新增

- `app.Run` 在 config 配置 `resilience.rate_limit.routes` 时**自动挂载 L3**
- 完整中文文档 `docs/zh/`
- Tracing 文档：`logger.LFromContext` 关联 trace_id / span_id
- 扩展单测：L3 集成、tracing、HTTP/gRPC 限流中间件

### 变更

- 文档：L3 挂载语义、中英文交叉链接
- 依赖：升级 go-common（`LFromContext`）

## v0.3.0 — 2026-05-30

### 新增

- **分层限流** L1 全局 / L2 用户 / L3 路由
- `RateLimiter` 接口，`MemoryRateLimiter` 与 `RedisRateLimiter`（GCRA）
- HTTP：`GlobalRateLimit`、`UserRateLimit`、`RouteRateLimit`、`RateLimitByKey`、`BreakerCheck`
- gRPC：`UserUnaryServerInterceptor`、`RouteUnaryServerInterceptor`
- 配置扩展：`global` / `user` / `routes` / `backend`
- 指标：`rate_limit_rejected_total` 增加 `scope` 标签

### 破坏性变更

- `NewPolicyFromConfig` 返回 `(*Policy, error)`
- 旧版 `rate_limit.rps`/`burst` 迁移至 `global.*`
- 移除 `http/middleware/resilience.go`，改用 `ratelimit.go`

### 移除

- `examples/minimal`、`docs/scaffold.md`

## v0.2.2 — 2026-05-30

- `/health`、`/ready` 探针
- 独立 metrics 端口与 Bearer 鉴权
- 可配置熔断 `consecutive_failures`
- 可选 OTel tracing hooks

## v0.2.1 — 2026-05-30

- Prometheus 指标
- go-common 配置与日志
- CI；gRPC 需显式 `WithGRPC`

## v0.2.0 — 2026-05-30

- 初始版本：HTTP/gRPC/AI、熔断限流、应用生命周期
