package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

// Options holds scaffold generation options.
type Options struct {
	OutputDir  string
	ModulePath string
	ServiceName string
	Demo       bool
	WithGRPC   bool
	WithAI     bool
}

// Generate creates a new service project from templates.
func Generate(opts Options) error {
	if opts.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	if opts.ModulePath == "" {
		return fmt.Errorf("module path is required")
	}
	if opts.ServiceName == "" {
		opts.ServiceName = filepath.Base(opts.OutputDir)
	}

	templateSet := "empty"
	if opts.Demo {
		templateSet = "demo"
	}

	root := filepath.Join("templates", templateSet)
	return fs.WalkDir(templateFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if d.IsDir() {
			if shouldSkipDir(rel, opts) {
				return fs.SkipDir
			}
			outDir := filepath.Join(opts.OutputDir, rel)
			return os.MkdirAll(outDir, 0o755)
		}

		if shouldSkipFile(rel, opts) {
			return nil
		}

		return renderFile(path, filepath.Join(opts.OutputDir, stripTmpl(rel)), opts)
	})
}

func shouldSkipDir(rel string, opts Options) bool {
	if !opts.WithGRPC && strings.HasPrefix(rel, "proto") {
		return true
	}
	if !opts.WithGRPC && strings.HasPrefix(rel, "api") {
		return true
	}
	return false
}

func shouldSkipFile(rel string, opts Options) bool {
	rel = filepath.ToSlash(rel)
	if !opts.WithGRPC {
		if strings.HasPrefix(rel, "proto/") ||
			strings.HasPrefix(rel, "api/") {
			return true
		}
	}
	if !opts.WithAI && strings.Contains(rel, "ai_handler") {
		return true
	}
	return false
}

func stripTmpl(path string) string {
	if strings.HasSuffix(path, ".tmpl") {
		return strings.TrimSuffix(path, ".tmpl")
	}
	return path
}

func renderFile(srcPath, dstPath string, opts Options) error {
	data, err := templateFS.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read template %s: %w", srcPath, err)
	}

	tmpl, err := template.New(filepath.Base(srcPath)).Parse(string(data))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", srcPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	f, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, opts)
}
