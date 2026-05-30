# grpc-gateway (optional)

When exposing REST externally and gRPC internally, combine [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) with goblocks.

## Recommended approach

1. Define `.proto` in your business repo with `google.api.http` annotations
2. Generate reverse-proxy code with `protoc-gen-grpc-gateway`
3. Register the gateway mux in `WithHTTP` alongside existing Gin routes

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

## CLI support (planned)

goblocks-cli may add `--with-grpc-gateway` to generate proto annotations and Makefile targets. Integrate manually for now.

## Notes

- Gateway and gRPC must both be enabled (`server.grpc.enabled: true` + `WithGRPC`)
- In production, apply auth and rate limits separately for gateway paths
