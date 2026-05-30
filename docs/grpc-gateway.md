# grpc-gateway（可选）

对外 REST、对内 gRPC 时，可结合 [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) 与 goblocks 使用。

## 推荐方式

1. 在业务仓库定义 `.proto` 并添加 `google.api.http` 注解
2. 使用 `protoc-gen-grpc-gateway` 生成 reverse-proxy 代码
3. 在 `WithHTTP` 中注册 gateway mux，与现有 Gin 路由并存

```go
func registerHTTP(engine *gin.Engine, _ *resilience.Policy) {
    mux := runtime.NewServeMux()
    _ = pb.RegisterUserServiceHandlerFromEndpoint(
        context.Background(), mux, "localhost:9090",
        []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
    )
    engine.Any("/api/*path", gin.WrapH(mux))
}
```

## CLI 支持（规划中）

goblocks-cli 可提供 `--with-grpc-gateway` 生成 proto 注解与 Makefile 目标。当前版本请手动集成。

## 注意事项

- gateway 与 gRPC 服务需同时启用（`server.grpc.enabled: true` + `WithGRPC`）
- 生产环境应对 gateway 路径单独做鉴权与限流
