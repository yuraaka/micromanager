package runtime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const (
	ModeLocal    = "local"
	ModeDocker   = "docker"
	ModeMinikube = "minikube"
)

// Run is a placeholder that would orchestrate starting services in different modes.
// For now, it simulates build output placement and returns a reachable endpoint.
func Run(ctx context.Context, root, target, mode string) (string, error) {
	_ = ctx
	if target == "" {
		target = "all"
	}

	buildPath := filepath.Join(root, "build", target)
	if err := os.MkdirAll(buildPath, 0o755); err != nil {
		return "", err
	}

	port := defaultPortFor(mode)
	endpoint := fmt.Sprintf("http://localhost:%d", port)
	return endpoint, nil
}

// RunService builds and executes a service with environment variables from service.toml.
func RunService(ctx context.Context, servicePath, mode string) error {
	// Normalize path
	if !filepath.IsAbs(servicePath) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		servicePath = filepath.Join(cwd, servicePath)
	}

	// Check if service directory exists
	info, err := os.Stat(servicePath)
	if err != nil {
		return fmt.Errorf("service path error: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", servicePath)
	}

	// Load service configuration
	configPath := filepath.Join(servicePath, "service.toml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read service.toml: %w", err)
	}

	var cfg struct {
		Environment map[string]map[string]interface{} `toml:"environment"`
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse service.toml: %w", err)
	}

	// Find repo root by looking for .mm directory
	repoRoot := servicePath
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, ".mm")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			return fmt.Errorf("could not find repo root (.mm directory)")
		}
		repoRoot = parent
	}

	// Build service binary in <repo-root>/build/<service-name>
	serviceName := filepath.Base(servicePath)
	buildDir := filepath.Join(repoRoot, "build", serviceName)
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	binaryPath := filepath.Join(buildDir, serviceName)

	// Try building from ./server if it exists, otherwise use ./
	buildTarget := "./server"
	serverPath := filepath.Join(servicePath, "server")
	if _, err := os.Stat(serverPath); err != nil {
		buildTarget = "./"
	}

	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, buildTarget)
	buildCmd.Dir = servicePath
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %w\n%s", err, string(output))
	}

	// Prepare environment variables from service.toml
	env := os.Environ()
	for varName, modeValues := range cfg.Environment {
		if modeValue, ok := modeValues[mode]; ok {
			env = append(env, fmt.Sprintf("%s=%v", varName, modeValue))
		}
	}

	// Run the service
	runCmd := exec.CommandContext(ctx, binaryPath)
	runCmd.Dir = servicePath
	runCmd.Env = env
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	runCmd.Stdin = os.Stdin

	return runCmd.Run()
}

func defaultPortFor(mode string) int {
	switch mode {
	case ModeDocker:
		return 10000
	case ModeMinikube:
		return 2000
	default:
		return 8000
	}
}
