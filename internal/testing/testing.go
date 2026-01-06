package testing

import (
	"context"
	"fmt"
)

// Run executes tests for a specific service path or all services.
// Currently a stub that reports intent.
func Run(ctx context.Context, root, target string) error {
	_ = ctx
	if target == "" {
		target = "all"
	}
	fmt.Printf("[dry-run] would run tests for '%s' under %s\n", target, root)
	return nil
}
