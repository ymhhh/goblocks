package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ymhhh/goblocks/internal/scaffold"
)

func main() {
	root := &cobra.Command{
		Use:   "goblocks",
		Short: "Goblocks service framework CLI",
		Long:  "Scaffold Go services using the Goblocks framework with DDD onion architecture.",
	}

	root.AddCommand(newCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newCmd() *cobra.Command {
	var (
		modulePath string
		demo       bool
		withGRPC   bool
		withAI     bool
	)

	cmd := &cobra.Command{
		Use:   "new [output-dir]",
		Short: "Create a new service project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir := args[0]
			if modulePath == "" {
				modulePath = outputDir
			}

			opts := scaffold.Options{
				OutputDir:   outputDir,
				ModulePath:  modulePath,
				ServiceName: filepath.Base(outputDir),
				Demo:        demo,
				WithGRPC:    withGRPC,
				WithAI:      withAI,
			}

			if err := scaffold.Generate(opts); err != nil {
				return err
			}

			fmt.Printf("Created service at %s\n", outputDir)
			fmt.Printf("  module: %s\n", modulePath)
			if demo {
				fmt.Println("  template: demo (User service)")
			} else {
				fmt.Println("  template: empty")
			}
			fmt.Println("\nNext steps:")
			fmt.Printf("  cd %s\n", outputDir)
			fmt.Println("  go mod tidy")
			fmt.Println("  go run .")
			return nil
		},
	}

	cmd.Flags().StringVar(&modulePath, "module", "", "Go module path (default: output dir name)")
	cmd.Flags().BoolVar(&demo, "demo", false, "Generate demo User service")
	cmd.Flags().BoolVar(&withGRPC, "with-grpc", false, "Include gRPC proto and server")
	cmd.Flags().BoolVar(&withAI, "with-ai", false, "Include AI chat handler example")

	return cmd
}
