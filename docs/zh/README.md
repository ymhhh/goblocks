# Goblocks 中文文档

Goblocks 是 Go 语言服务端框架，提供 HTTP（H1/H2/H3）、gRPC、OpenAI 兼容 AI 客户端，以及统一的熔断、**分层限流**（L1/L2/L3）与 Prometheus 指标。

脚手架 CLI 位于 **[goblocks-cli](https://github.com/ymhhh/goblocks-cli)**，命令为 `goblocks new`。

英文文档见 [docs/README.md](../README.md)。

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

最小示例见 [快速入门](quickstart.md) 与 [包 API 参考 — app](packages.md#app)。

## 文档目录

| 文档 | 说明 |
|------|------|
| [快速入门](quickstart.md) | 安装、脚手架、最小接入 |
| [架构设计](architecture.md) | 分层结构、依赖、限流目录映射、请求流转 |
| [配置参考](configuration.md) | YAML 配置项、环境变量 |
| [分层限流指南](rate-limiting.md) | L1/L2/L3 原理、配置、挂载与排查 |
| [包 API 参考](packages.md) | `app`、`http`、`grpc`、`ai`、`resilience` |
| [观测指标](metrics.md) | Prometheus 指标与 PromQL |
| [分布式追踪](tracing.md) | OpenTelemetry 可选 hook |
| [grpc-gateway](grpc-gateway.md) | REST 网关集成（可选） |
| [开发指南](development.md) | 框架开发、测试与发布 |
| [API 稳定性](stability.md) | 版本策略与兼容承诺 |
| [更新日志摘要](changelog.md) | 版本变更中文摘要 |

## 推荐阅读顺序

1. [快速入门](quickstart.md)
2. [分层限流指南](rate-limiting.md)
3. [配置参考](configuration.md)
4. [架构设计](architecture.md)

## 相关链接

- 框架仓库：[goblocks](https://github.com/ymhhh/goblocks)
- CLI 仓库：[goblocks-cli](https://github.com/ymhhh/goblocks-cli)
- 参考示例：[ddd-onion-sample](https://github.com/ymhhh/ddd-onion-sample)
- 英文 Changelog：[CHANGELOG.md](../../CHANGELOG.md)

## 最低要求

- Go >= 1.22
- HTTP/3 需 TLS 证书与 UDP 端口可达
