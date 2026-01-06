package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"micromanager/internal/config"
	"micromanager/internal/scaffold"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mm",
		Short: "Micromanager CLI",
	}

	rootCmd.AddCommand(initCommand())
	rootCmd.AddCommand(newCommand())
	rootCmd.AddCommand(runCommand())
	rootCmd.AddCommand(updateCommand())
	rootCmd.AddCommand(testCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initCommand() *cobra.Command {
	var backendLang string
	var frontendLang string
	var frontendServer string
	var frontendClient string
	var databaseEngine string
	var packageManager string

	cmd := &cobra.Command{
		Use:   "init <path>",
		Short: "Initialize repository defaults and structure",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			if target == "" {
				return fmt.Errorf("path is required")
			}
			absTarget, err := filepath.Abs(target)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(absTarget, 0o755); err != nil {
				return err
			}

			defaults, err := scaffold.InitRepo(cmd.Context(), absTarget, scaffold.InitOptions{
				BackendLang:    backendLang,
				FrontendLang:   frontendLang,
				FrontendServer: frontendServer,
				FrontendClient: frontendClient,
				DatabaseEngine: databaseEngine,
				PackageManager: packageManager,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Defaults written to %s\n", filepath.Join(absTarget, ".mm", "defaults.toml"))
			fmt.Printf("Backend: %s | Frontend: %s/%s (%s, %s) | Database: %s\n",
				defaults.Backend.Lang,
				defaults.Frontend.Server,
				defaults.Frontend.Client,
				defaults.Frontend.Lang,
				defaults.Frontend.PackageManager,
				defaults.Database.Engine,
			)
			return nil
		},
	}

	cmd.Flags().StringVar(&backendLang, "back", "", "backend language (default: go)")
	cmd.Flags().StringVar(&frontendLang, "front-lang", "", "frontend language (default: ts)")
	cmd.Flags().StringVar(&frontendServer, "front", "", "frontend server framework (default: next.js)")
	cmd.Flags().StringVar(&frontendClient, "front-client", "", "frontend client library (default: react)")
	cmd.Flags().StringVar(&databaseEngine, "db", "", "database engine (default: postgres)")
	cmd.Flags().StringVar(&packageManager, "pkg", "", "frontend package manager (default: pnpm)")
	return cmd
}

func newCommand() *cobra.Command {
	var frontendOnly bool
	var empty bool
	var deps []string

	cmd := &cobra.Command{
		Use:   "new <service-name>",
		Short: "Create a new service skeleton",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.TrimSpace(args[0])
			if name == "" {
				return fmt.Errorf("service name is required")
			}

			root, err := os.Getwd()
			if err != nil {
				return err
			}

			defaults, err := config.LoadDefaults(root)
			if err != nil {
				return fmt.Errorf("load defaults: %w", err)
			}

			_, err = scaffold.NewService(root, name, scaffold.NewServiceOptions{
				FrontendOnly: frontendOnly,
				Empty:        empty,
				Defaults:     defaults,
				Dependencies: deps,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Service '%s' created in services/%s\n", name, name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&frontendOnly, "front", false, "generate a frontend-only service")
	cmd.Flags().BoolVar(&empty, "empty", false, "generate an external/empty service (Dockerfile + service.toml)")
	cmd.Flags().StringSliceVar(&deps, "dep", nil, "declare service dependencies (names)")
	return cmd
}

func runCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func updateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Rescan services and apply structural updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}

func testCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Run tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("not implemented yet")
		},
	}
}
