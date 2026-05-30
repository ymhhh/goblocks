# 脚手架 CLI

`goblocks` CLI 用于快速生成符合洋葱架构的服务工程。模板通过 `embed.FS` 内嵌，无需联网。

## 安装

```bash
# 从远程安装（需已发布 tag）
go install github.com/ymhhh/goblocks/cmd/goblocks@latest

# 本地开发
git clone https://github.com/ymhhh/goblocks.git
cd goblocks && make build
./bin/goblocks --help
```

## 命令

### goblocks new

```bash
goblocks new <output-dir> [flags]
```

| Flag | 说明 |
|------|------|
| `--module` | Go module 路径（必填，如 `github.com/acme/my-service`） |
| `--demo` | 生成 User Demo（含 Mock 仓储、HTTP `GET /users/:id`） |
| `--with-grpc` | 生成 gRPC 健康检查注册 + `proto/user/v1/user.proto` 示例 |
| `--with-ai` | 生成 AI Chat handler（`POST /ai/chat`）并启用 ai 配置 |

### 示例

**空工程（最小骨架）**

```bash
goblocks new my-service --module github.com/acme/my-service
cd my-service
echo "replace github.com/ymhhh/goblocks => /path/to/goblocks" >> go.mod  # 本地开发时
go mod tidy
go run .
curl http://localhost:8080/health
```

**User Demo**

```bash
goblocks new demo-svc --module github.com/acme/demo-svc --demo
go mod tidy && go run .
curl http://localhost:8080/users/1
# {"id":"1","name":"Alice"}
```

**完整功能（Demo + gRPC + AI）**

```bash
goblocks new full-svc --module github.com/acme/full-svc \
  --demo --with-grpc --with-ai

export OPENAI_API_KEY=sk-...
go mod tidy && go run .

curl http://localhost:8080/users/1
curl -X POST http://localhost:8080/ai/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"hello"}'
```

## 生成目录结构

### 空工程（empty 模板）

```
my-service/
├── main.go                 # 仅调用 infrastructure.Run
├── go.mod
├── Makefile
├── README.md
├── config/
│   └── config.yaml
├── core/
│   └── doc.go              # 实体包占位
├── domain/
│   └── domain.go           # 端口包占位
├── handlers/
│   └── handlers.go         # 用例包占位
└── infrastructure/
    ├── app.go              # 组合根
    └── run.go              # 启动与路由注册
```

`--with-grpc` 额外生成：

```
infrastructure/grpc_server.go   # gRPC 健康检查
```

### Demo 工程（demo 模板）

在 empty 基础上增加：

```
core/user.go
domain/user_repository.go
domain/repo_user_mock.go
domain/errors.go
handlers/user_handler.go
handlers/context.go
```

`--with-grpc` 额外生成：

```
proto/user/v1/user.proto
infrastructure/grpc_server.go
```

`--with-ai` 额外生成：

```
handlers/ai_handler.go
```

## 生成工程与框架的关系

生成工程的 `go.mod` 声明：

```
require github.com/ymhhh/goblocks v0.0.0
```

- **已发布版本**：`go get github.com/ymhhh/goblocks@v0.1.0`
- **本地开发**：在 `go.mod` 追加 `replace github.com/ymhhh/goblocks => /path/to/goblocks`

业务代码位于 `core/domain/handlers/infrastructure`，**不修改** goblocks 源码。

## 扩展生成工程

### 添加 HTTP 路由

在 `infrastructure/run.go` 的 `registerHTTP` 中注册：

```go
func (a *App) registerHTTP(engine *gin.Engine, _ *resilience.Policy) {
    engine.GET("/health", ...)
    engine.GET("/users/:id", ...)
    // 在此添加新路由
}
```

### 替换 Mock 仓储为真实 DB

1. 在 `domain/` 保持 `UserRepository` 接口不变
2. 在 `infrastructure/` 或单独 `domain/` 实现文件中新写 `PostgresUserRepo`
3. 在 `NewApp` 中切换构造：

```go
repo := domain.NewPostgresUserRepo(db)  // 替换 NewUserMockRepo()
```

### 添加 gRPC 服务

1. 在 `proto/` 定义 `.proto` 文件
2. 使用 `buf generate` 或 `protoc` 生成 Go 代码
3. 在 `infrastructure/grpc_server.go` 的 `registerGRPC` 中注册实现

`--with-grpc` 生成的 `proto/user/v1/user.proto` 可作为起点。

## 与 ddd-onion-sample 对照

| ddd-onion-sample | goblocks 生成工程 |
|------------------|-------------------|
| `core/user.go` | 同（`--demo`） |
| `domain/user_repository.go` | 同 |
| `handlers/user_handler.go` | 同 |
| `infrastructure/app.go` | 同（组合根） |
| `infrastructure/entry.go` | 合并为 `run.go` + goblocks `app.Run` |
| 无 HTTP 服务器 | 集成 goblocks HTTP/gRPC 启动 |

参考仓库：[ymhhh/ddd-onion-sample](https://github.com/ymhhh/ddd-onion-sample)
