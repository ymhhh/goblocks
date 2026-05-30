# Goblocks 文档

Goblocks 是 Go 语言服务端框架，提供 HTTP（H1/H2/H3）、gRPC、OpenAI 兼容 AI 客户端，以及统一的熔断、限流与 Prometheus 指标。

脚手架 CLI 位于独立仓库 **[goblocks-cli](https://github.com/ymhhh/goblocks-cli)**，命令仍为 `goblocks new`。

## 文档目录

| 文档 | 说明 |
|------|------|
| [架构设计](architecture.md) | 分层结构、依赖方向、请求流转 |
| [配置参考](configuration.md) | YAML 配置项、环境变量、各模块配置示例 |
| [脚手架 CLI](scaffold.md) | `goblocks new` 用法（goblocks-cli） |
| [包 API 参考](packages.md) | `app`、`http`、`grpc`、`ai`、`resilience` 使用说明 |
| [观测指标](metrics.md) | Prometheus 指标说明与 PromQL 示例 |
| [开发指南](development.md) | 框架开发、测试与发布 |

## 快速链接

- 框架仓库：[goblocks](https://github.com/ymhhh/goblocks)
- CLI 仓库：[goblocks-cli](https://github.com/ymhhh/goblocks-cli)
- 参考示例：[ddd-onion-sample](https://github.com/ymhhh/ddd-onion-sample)
- 安装框架：`go get github.com/ymhhh/goblocks@latest`
- 安装 CLI：`go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest`

## 最低要求

- Go >= 1.22
- 生成 HTTP/3 时需 TLS 证书与 UDP 端口可达
