package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateEmpty(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "testsvc")

	opts := Options{
		OutputDir:  out,
		ModulePath: "github.com/acme/testsvc",
		ServiceName: "testsvc",
		Demo:       false,
	}

	if err := Generate(opts); err != nil {
		t.Fatal(err)
	}

	checkFile(t, out, "main.go")
	checkFile(t, out, "go.mod")
	checkFile(t, out, "config/config.yaml")
	checkFile(t, out, "infrastructure/run.go")
	checkFile(t, out, "infrastructure/grpc_server.go")
	checkNoFile(t, out, "core/user.go")
}

func TestGenerateDemoWithGRPCProto(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "demo-svc")

	opts := Options{
		OutputDir:   out,
		ModulePath:  "github.com/acme/demo-svc",
		ServiceName: "demo-svc",
		Demo:        true,
		WithGRPC:    true,
	}

	if err := Generate(opts); err != nil {
		t.Fatal(err)
	}

	checkFile(t, out, "proto/user/v1/user.proto")
}

func TestGenerateDemo(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "demo-svc")

	opts := Options{
		OutputDir:  out,
		ModulePath: "github.com/acme/demo-svc",
		ServiceName: "demo-svc",
		Demo:       true,
		WithGRPC:   true,
		WithAI:     true,
	}

	if err := Generate(opts); err != nil {
		t.Fatal(err)
	}

	checkFile(t, out, "core/user.go")
	checkFile(t, out, "handlers/user_handler.go")
	checkFile(t, out, "handlers/ai_handler.go")
	checkFile(t, out, "proto/user/v1/user.proto")
	checkFile(t, out, "infrastructure/grpc_server.go")
}

func checkFile(t *testing.T, root, rel string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s: %v", rel, err)
	}
}

func checkNoFile(t *testing.T, root, rel string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected no file %s", rel)
	}
}
