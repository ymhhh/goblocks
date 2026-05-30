# 开发指南

本文面向 goblocks **框架库**贡献者与维护者。

## 环境要求

- Go >= 1.22（推荐与 `go.mod` 中版本一致）
- Make（可选，使用 Makefile 快捷命令）

## 仓库结构

```
goblocks/
├── app/              应用生命周期
├── config/           配置加载
├── resilience/       熔断限流
├── http/             HTTP 服务 + middleware
├── grpc/             gRPC 服务 + interceptors
├── ai/               AI 客户端
├── cmd/goblocks/     CLI 入口
├── internal/scaffold/  脚手架模板（embed）
├── docs/             说明文档
├── Makefile
└── README.md
```

## 常用命令

```bash
make test     # go test ./... -race -count=1
make build    # 构建 bin/goblocks
make install  # go install ./cmd/goblocks
make lint     # go vet ./... && go fmt ./...
make clean    # 删除 bin/
```

## 测试

### 单元测试

各包测试文件与源码同目录：

```bash
go test ./config/... -v
go test ./resilience/... -v
go test ./ai/... -v
```

### 脚手架集成测试

`internal/scaffold/scaffold_integration_test.go` 会：

1. 生成 Demo 工程到临时目录
2. 追加 `replace` 指向本仓库
3. 执行 `go mod tidy` 与 `go build`

```bash
go test ./internal/scaffold/... -v
```

跳过耗时集成测试：

```bash
go test ./internal/scaffold/... -short
```

### 手动验证脚手架

```bash
make build
OUT=/tmp/goblocks-test
./bin/goblocks new $OUT/demo --module github.com/acme/demo --demo --with-grpc

cd $OUT/demo
echo "replace github.com/ymhhh/goblocks => $(pwd)/../../.." >> go.mod
go mod tidy
go run .

# 另一终端
curl http://localhost:8080/users/1
curl http://localhost:8080/health
```

## 修改脚手架模板

模板位于 `internal/scaffold/templates/`：

| 目录 | 用途 |
|------|------|
| `empty/` | 空洋葱骨架 |
| `demo/` | User Demo |

文件以 `.tmpl` 结尾，使用 Go `text/template` 语法，可用变量：

| 变量 | 说明 |
|------|------|
| `{{.ModulePath}}` | Go module 路径 |
| `{{.ServiceName}}` | 服务名（输出目录 basename） |
| `{{.Demo}}` | 是否 demo 模式 |
| `{{.WithGRPC}}` | 是否含 gRPC |
| `{{.WithAI}}` | 是否含 AI |

修改模板后务必运行：

```bash
go test ./internal/scaffold/... -v -count=1
```

## 添加新依赖

```bash
go get example.com/foo
go mod tidy
```

框架库应保持依赖精简，新增依赖需在 PR 中说明理由。

## 发布流程（建议）

1. 确保 `go test ./... -race` 通过
2. 更新 README / docs 如有 API 变更
3. 打 tag：`git tag v0.1.0`
4. 推送 tag 后用户可：

```bash
go get github.com/ymhhh/goblocks@v0.1.0
go install github.com/ymhhh/goblocks/cmd/goblocks@v0.1.0
```

## 版本策略

| 版本 | 范围 |
|------|------|
| v0.1.x | 框架核心 + empty/demo 脚手架 |
| v0.2.x | grpc-gateway、Docker 模板等增强（规划中） |

Breaking Change 需在 README 与 docs 中注明迁移步骤。

## 代码规范

- 遵循现有包边界，不在 `core/domain` 层引入框架依赖（针对生成工程）
- 公开 API 需有 godoc 注释
- 新功能需附带单元测试
- 使用 `log/slog` 作为默认日志（`app` 包已使用）

## 已知限制（MVP）

- 无服务注册发现（Consul/Etcd）
- 限流为进程内令牌桶，非分布式
- Tracing 仅预留扩展点，未内置 OpenTelemetry
- gRPC 与 HTTP 不可同端口
- HTTP/3 需额外 UDP 端口与 TLS

详见计划文档中的 Risks 章节。
