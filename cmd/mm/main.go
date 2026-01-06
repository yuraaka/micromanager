package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"micromanager/internal/config"
	"micromanager/internal/lang"
	"micromanager/internal/runtime"
	"micromanager/internal/scaffold"
	mmtest "micromanager/internal/testing"
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
	rootCmd.AddCommand(packsCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initCommand() *cobra.Command {
	var lang string

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
				Lang: lang,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Defaults written to %s\n", filepath.Join(absTarget, ".mm", "defaults.toml"))
			fmt.Printf("Lang: %s\n", defaults.Lang)
			return nil
		},
	}

	cmd.Flags().StringVar(&lang, "lang", "", "service language (default: go)")
	return cmd
}

func newCommand() *cobra.Command {
	var empty bool

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
				Empty:    empty,
				Defaults: defaults,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Service '%s' created in services/%s\n", name, name)
			return nil
		},
	}

	cmd.Flags().BoolVar(&empty, "empty", false, "generate an external/empty service (Dockerfile + service.toml)")
	return cmd
}

func runCommand() *cobra.Command {
	var docker bool
	var minikube bool

	cmd := &cobra.Command{
		Use:   "run [service]",
		Short: "Run services",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "all"
			if len(args) == 1 {
				target = args[0]
			}

			mode := runtime.ModeLocal
			if docker {
				mode = runtime.ModeDocker
			}
			if minikube {
				mode = runtime.ModeMinikube
			}

			root, err := os.Getwd()
			if err != nil {
				return err
			}

			endpoint, err := runtime.Run(cmd.Context(), root, target, mode)
			if err != nil {
				return err
			}

			fmt.Printf("Service '%s' running in %s mode at %s\n", target, mode, endpoint)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&docker, "docker", "d", false, "run services in docker compose")
	cmd.Flags().BoolVarP(&minikube, "minikube", "m", false, "run services in minikube")
	return cmd
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
	cmd := &cobra.Command{
		Use:   "test [path]",
		Short: "Run tests",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := "all"
			if len(args) == 1 {
				target = args[0]
			}

			root, err := os.Getwd()
			if err != nil {
				return err
			}

			return mmtest.Run(cmd.Context(), root, target)
		},
	}
	return cmd
}

func packsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packs",
		Short: "Manage language packs",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available language packs",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			packs, err := lang.LoadAll(root)
			if err != nil {
				return err
			}
			if len(packs) == 0 {
				fmt.Println("No packs found in .mm/packs")
				return nil
			}
			for _, p := range packs {
				fmt.Printf("%s\t%s\t(lang=%s, v=%s)\n", p.Meta.ID, p.Meta.Name, p.Meta.Lang, p.Meta.Version)
			}
			return nil
		},
	}

	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate all language packs",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			packs, err := lang.LoadAll(root)
			if err != nil {
				return err
			}
			if len(packs) == 0 {
				fmt.Println("No packs found in .mm/packs")
				return nil
			}
			for _, p := range packs {
				if err := lang.Validate(p); err != nil {
					fmt.Printf("%s: INVALID (%v)\n", p.Meta.ID, err)
				} else {
					fmt.Printf("%s: OK\n", p.Meta.ID)
				}
			}
			return nil
		},
	}

	cmd.AddCommand(listCmd)
	cmd.AddCommand(validateCmd)
	return cmd
}
