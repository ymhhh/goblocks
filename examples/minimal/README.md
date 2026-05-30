# Minimal Example

Smallest goblocks service without the CLI scaffold.

## Run

```bash
cd examples/minimal
go mod tidy
go run .
```

```bash
curl http://localhost:8080/hello
curl http://localhost:8080/health
```

Uses `replace` in `go.mod` to point at the parent framework repo.
