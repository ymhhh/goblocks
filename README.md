# goblocks

Go 服务端框架：HTTP（H1/H2/H3）、gRPC、OpenAI 兼容 AI 客户端，统一熔断与分层限流。

配合 [goblocks-cli](https://github.com/ymhhh/goblocks-cli) 生成符合 [洋葱架构 / DDD](https://github.com/ymhhh/ddd-onion-sample) 的服务工程。

## 快速开始

```bash
go install github.com/ymhhh/goblocks-cli/cmd/goblocks@latest
goblocks new my-service --module github.com/acme/my-service
cd my-service && go mod tidy && go run .
```

已有工程引入框架：`go get github.com/ymhhh/goblocks@latest`

## 文档

详细说明见 **[docs/](docs/README.md)**：

| 文档 | 说明 |
|------|------|
| [架构设计](docs/architecture.md) | 分层、依赖、限流挂载 |
| [配置参考](docs/configuration.md) | YAML 与环境变量 |
| [包 API 参考](docs/packages.md) | 代码接入与示例 |
| [观测指标](docs/metrics.md) | Prometheus |
| [开发指南](docs/development.md) | 贡献、测试、发布 |

## License

GPL-3.0。API 稳定性见 [docs/stability.md](docs/stability.md)。
