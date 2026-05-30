package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerateAndBuildDemo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build integration in short mode")
	}

	dir := t.TempDir()
	out := filepath.Join(dir, "demo-svc")
	goblocksRoot := findGoblocksRoot(t)

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

	goModPath := filepath.Join(out, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data) + "\nreplace github.com/ymhhh/goblocks => " + goblocksRoot + "\n"
	if err := os.WriteFile(goModPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = out
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, outBytes)
	}

	cmd = exec.Command("go", "build", "-o", filepath.Join(dir, "demo-svc-bin"), ".")
	cmd.Dir = out
	if outBytes, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, outBytes)
	}
}

func findGoblocksRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/scaffold/scaffold_test.go -> repo root
	root := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	if !strings.HasSuffix(root, "goblocks") {
		// fallback: walk up from cwd
		cwd, _ := os.Getwd()
		for d := cwd; d != filepath.Dir(d); d = filepath.Dir(d) {
			if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
				return d
			}
		}
	}
	return root
}
