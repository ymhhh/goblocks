# 更新日志（中文摘要）

完整英文记录见 [CHANGELOG.md](../../CHANGELOG.md)。

## v0.3.1（未发布）

### 变更

- `app.Run` 在 config 配置 `resilience.rate_limit.routes` 时**自动挂载 L3**（HTTP `RouteRateLimit`、gRPC `RouteUnaryServerInterceptor`）
- 补充 L3 单元测试与中文文档

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
