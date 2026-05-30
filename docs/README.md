# Goblocks 文档

Goblocks 是 Go 语言服务端框架，提供 HTTP（H1/H2/H3）、gRPC、OpenAI 兼容 AI 客户端，以及统一的熔断与限流能力；CLI 脚手架可生成符合 [洋葱架构 / DDD](https://github.com/ymhhh/ddd-onion-sample) 的服务工程。

## 文档目录

| 文档 | 说明 |
|------|------|
| [架构设计](architecture.md) | 分层结构、依赖方向、请求流转 |
| [配置参考](configuration.md) | YAML 配置项、环境变量、各模块配置示例 |
| [脚手架 CLI](scaffold.md) | `goblocks new` 用法、模板说明、生成工程结构 |
| [包 API 参考](packages.md) | `app`、`http`、`grpc`、`ai`、`resilience` 使用说明 |
| [开发指南](development.md) | 本地开发、测试、发布与贡献 |

## 快速链接

- 仓库 README：[../README.md](../README.md)
- 参考示例：[ddd-onion-sample](https://github.com/ymhhh/ddd-onion-sample)
- 安装 CLI：`go install github.com/ymhhh/goblocks/cmd/goblocks@latest`

## 最低要求

- Go >= 1.22
- 生成 HTTP/3 时需 TLS 证书与 UDP 端口可达
