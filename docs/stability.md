# API 稳定性

## v0.x（当前）

- 允许 breaking change，但必须在 [CHANGELOG](../CHANGELOG.md) 中明确标注
- 配置 YAML 键名、环境变量变更视为 breaking change
- 小版本（v0.2.x）优先 bugfix 与向后兼容增强

## 已冻结（尽量保持兼容）

| API | 说明 |
|-----|------|
| `config.Load(path)` | 返回 `*Config, error` |
| `app.New(cfg)` | 构造应用 |
| `app.(*App).WithHTTP` / `WithGRPC` | 注册路由与服务 |
| `app.(*App).Run(ctx)` | 启动与优雅关闭 |
| `resilience.NewPolicyFromConfig` | 从配置创建策略 |

## v1.0 目标

- 冻结上述公开 API 签名
- semver 严格遵循 MAJOR.MINOR.PATCH
- 配置 schema 提供迁移指南与兼容期

## 实验性 API

以下 API 可能在 minor 版本中调整：

- `WithHTTPTracing` / `WithGRPCTracing`
- `server.http.health.*` 探针行为
- `metrics.addr` / `metrics.auth_token`
