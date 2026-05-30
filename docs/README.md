# Goblocks 文档

Goblocks 是 Go 语言服务端框架，提供 HTTP（H1/H2/H3）、gRPC、OpenAI 兼容 AI 客户端，以及统一的熔断、**分层限流**（L1/L2/L3）与 Prometheus 指标。

脚手架 CLI 位于独立仓库 **[goblocks-cli](https://github.com/ymhhh/goblocks-cli)**，命令为 `goblocks new`。

## 特性概览

| 模块 | 说明 |
|------|------|
| HTTP | Gin 封装，TLS 下 HTTP/2，可选 HTTP/3 |
| gRPC | Server / Client，Unary Interceptor 接入 resilience |
| AI | OpenAI 兼容 Chat（OpenAI、Azure、Ollama 等） |
| Resilience | 熔断 + 分层限流 L1/L2/L3（`memory` / `redis`） |
| Metrics | Prometheus（HTTP / gRPC / resilience / AI） |

## 快速开始

**新建工程（推荐）**

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
goblocks new my-service --module github.com/acme/my-service
cd my-service && go mod tidy && go run .
```

CLI 选项（`--demo`、`--with-grpc`、`--with-ai`）见 [goblocks-cli README](https://github.com/ymhhh/goblocks-cli)。

**已有工程接入**

```bash
go get github.com/ymhhh/goblocks@latest
```

最小 `main.go` 示例见 [包 API 参考 — app](packages.md#app)。

配置示例见 [配置参考](configuration.md#完整示例)。

## 文档目录

| 文档 | 说明 |
|------|------|
| [架构设计](architecture.md) | 分层结构、依赖方向、**限流目录映射**、请求流转 |
| [配置参考](configuration.md) | YAML 配置项、环境变量、各模块配置示例 |
| [包 API 参考](packages.md) | `app`、`http`、`grpc`、`ai`、`resilience` 使用说明 |
| [观测指标](metrics.md) | Prometheus 指标说明与 PromQL 示例 |
| [分布式追踪](tracing.md) | OpenTelemetry 可选 hook |
| [grpc-gateway](grpc-gateway.md) | REST 网关集成指南（可选） |
| [开发指南](development.md) | 框架开发、测试与发布 |
| [API 稳定性](stability.md) | 版本策略与兼容承诺 |

## 快速链接

- 框架仓库：[goblocks](https://github.com/ymhhh/goblocks)
- CLI 仓库：[goblocks-cli](https://github.com/ymhhh/goblocks-cli)
- 参考示例：[ddd-onion-sample](https://github.com/ymhhh/ddd-onion-sample)
- 安装框架：`go get github.com/ymhhh/goblocks@latest`
- 安装 CLI：`go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest`

## 最低要求

- Go >= 1.22
- 生成 HTTP/3 时需 TLS 证书与 UDP 端口可达
