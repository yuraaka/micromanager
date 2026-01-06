package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
