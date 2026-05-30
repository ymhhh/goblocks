# 开发指南

本文面向 goblocks **框架库**贡献者与维护者。脚手架 CLI 开发见 [goblocks-cli](https://github.com/ymhhh/goblocks-cli)。

## 环境要求

- Go >= 1.22（CI 在 1.22 / 1.23 上验证；`go.mod` 最低版本 1.22）
- Make（可选）

## 仓库结构

```
goblocks/
├── app/              应用生命周期（默认 L1 + L3）
├── config/           配置加载（分层 rate_limit schema）
├── resilience/       熔断 + RateLimiter（memory/redis）+ Policy
├── http/             HTTP 服务
│   └── middleware/   GlobalRateLimit、UserRateLimit、RouteRateLimit
├── grpc/             gRPC 服务
│   └── interceptors/ L1/L2/L3 unary interceptors
├── ai/               AI 客户端（出站 L1 + breaker）
├── metrics/          Prometheus（scope label）
├── docs/             说明文档
├── Makefile
└── README.md
```

## 常用命令

```bash
make test     # go test ./... -race -count=1
make lint     # go vet ./... && go fmt ./...
```

## 测试

```bash
go test ./config/... -v
go test ./resilience/... -v
go test ./metrics/... -v
go test ./... -race -count=1
```

## 添加新依赖

```bash
go get example.com/foo
go mod tidy
```

框架库应保持依赖精简，新增依赖需在 PR 中说明理由。

## 发布流程

1. 确保 `go test ./... -race` 通过
2. 更新 README / docs 如有 API 变更
3. 打 tag：`git tag v0.3.0`
4. 用户引用：

```bash
go get github.com/ymhhh/goblocks@v0.3.0
```

CLI 发布在 [goblocks-cli](https://github.com/ymhhh/goblocks-cli) 仓库，建议与框架版本号对齐。

## 本地联调（框架 + CLI）

```bash
cd /path/to/github.com/ymhhh
go work init ./goblocks ./goblocks-cli
```

CLI 集成测试需设置 `GOBLOCKS_PATH` 指向本地 goblocks 仓库：

```bash
export GOBLOCKS_PATH=/path/to/goblocks
cd goblocks-cli && make test-integration
```

## 代码规范

- 公开 API 需有 godoc 注释
- 新功能需附带单元测试
- 使用 [`github.com/ymhhh/go-common/logger`](https://github.com/ymhhh/go-common) 作为默认日志

## 已知限制

- 无服务注册发现（Consul/Etcd）
- 限流 Redis 后端需独立 Redis；`memory` 后端仅单进程有效
- gRPC 与 HTTP 不可同端口
- HTTP/3 需额外 UDP 端口与 TLS
