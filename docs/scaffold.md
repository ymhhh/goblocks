# 脚手架 CLI

`goblocks` CLI 已拆至独立仓库 **[goblocks-cli](https://github.com/ymhhh/goblocks-cli)**，用于快速生成符合洋葱架构的服务工程。

## 安装

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
```

> 旧版 `go install github.com/ymhhh/goblocks/cmd/goblocks@...` 已废弃。

本地开发 CLI：

```bash
git clone https://github.com/ymhhh/goblocks-cli.git
cd goblocks-cli && make build
./bin/goblocks --help
```

## 命令

### goblocks new

```bash
goblocks new <output-dir> [flags]
```

| Flag | 说明 |
|------|------|
| `--module` | Go module 路径（必填） |
| `--goblocks-version` | 生成工程引用的框架版本（默认 `v0.2.0`） |
| `--demo` | 生成 User Demo |
| `--with-grpc` | 额外生成 `proto/user/v1/user.proto` 示例 |
| `--with-ai` | 生成 AI Chat handler |

### 示例

```bash
goblocks new my-service --module github.com/acme/my-service
goblocks new demo-svc --module github.com/acme/demo-svc --demo
goblocks new full-svc --module github.com/acme/full-svc --demo --with-grpc --with-ai \
  --goblocks-version v0.2.0
```

## 生成工程与框架的关系

生成工程的 `go.mod` 声明：

```
require github.com/ymhhh/goblocks v0.2.0
```

本地联调框架：

```bash
echo "replace github.com/ymhhh/goblocks => /path/to/goblocks" >> go.mod
go mod tidy
```

## 模板维护

模板源码、开发指南、集成测试均在 [goblocks-cli](https://github.com/ymhhh/goblocks-cli) 仓库维护。

## 与 ddd-onion-sample 对照

| ddd-onion-sample | goblocks 生成工程 |
|------------------|-------------------|
| `core/user.go` | 同（`--demo`） |
| `domain/user_repository.go` | 同 |
| `handlers/user_handler.go` | 同 |
| `infrastructure/app.go` | 同（组合根） |
| 无 HTTP 服务器 | 集成 goblocks HTTP/gRPC 启动 |

参考：[ymhhh/ddd-onion-sample](https://github.com/ymhhh/ddd-onion-sample)
