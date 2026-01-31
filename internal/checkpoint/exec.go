package checkpoint

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// safeCommand returns a CommandContext for the given tool with sanitized arguments.
// It resolves the tool's path and ensures only whitelisted tools are executed.
func safeCommand(ctx context.Context, tool string, args ...string) (*exec.Cmd, error) {
	var allowedTools = map[string]bool{
		"go":          true,
		"gosec":       true,
		"govulncheck": true,
	}

	if !allowedTools[tool] {
		return nil, fmt.Errorf("tool %s is not allowed", tool)
	}

	// Resolve the absolute path of the tool
	path, err := exec.LookPath(tool)
	if err != nil {
		return nil, fmt.Errorf("failed to find tool %s: %w", tool, err)
	}

	// Sanitize arguments - basic check for command injection characters
	for _, arg := range args {
		if strings.ContainsAny(arg, ";&|`$()") {
			return nil, fmt.Errorf("invalid argument: %s", arg)
		}
	}

	// Reviewed: SECURITY-PROCESS-EXEC - Arguments are sanitized above and tools are whitelisted.
	// We use exec.CommandContext to ensure timeouts are respected.
	return exec.CommandContext(ctx, path, args...), nil
}
