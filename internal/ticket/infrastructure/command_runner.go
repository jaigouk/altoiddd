// Package infrastructure provides adapters for the Ticket bounded context.
package infrastructure

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// DefaultCommandTimeout is the default timeout for command execution.
const DefaultCommandTimeout = 30 * time.Second

// ShellCommandRunner executes verification commands with security controls.
type ShellCommandRunner struct {
	allowlist []*regexp.Regexp
	timeout   time.Duration
}

// NewShellCommandRunner creates a CommandRunner with default security settings.
func NewShellCommandRunner() *ShellCommandRunner {
	return &ShellCommandRunner{
		allowlist: defaultAllowlist(),
		timeout:   DefaultCommandTimeout,
	}
}

// NewShellCommandRunnerWithOptions creates a CommandRunner with custom settings.
func NewShellCommandRunnerWithOptions(allowlist []*regexp.Regexp, timeout time.Duration) *ShellCommandRunner {
	return &ShellCommandRunner{
		allowlist: allowlist,
		timeout:   timeout,
	}
}

// Run executes a command and returns its stdout output.
// Security controls:
// - Command must match allowlist patterns
// - No shell expansion (exec.Command, not sh -c)
// - Enforced timeout via context
func (r *ShellCommandRunner) Run(ctx context.Context, command string) (string, error) {
	if command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	// Security: Check allowlist
	if !r.isAllowed(command) {
		return "", fmt.Errorf("command not in allowlist: %s", truncate(command, 50))
	}

	// Parse command into executable and args (no shell expansion)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command after parsing")
	}

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// Execute without shell (no sh -c)
	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out after %v", r.timeout)
		}
		return "", fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// isAllowed checks if the command matches any allowlist pattern.
func (r *ShellCommandRunner) isAllowed(command string) bool {
	for _, pattern := range r.allowlist {
		if pattern.MatchString(command) {
			return true
		}
	}
	return false
}

// defaultAllowlist returns safe command patterns for verification.
func defaultAllowlist() []*regexp.Regexp {
	patterns := []string{
		`^deadcode\s+`,           // deadcode analysis
		`^go\s+(build|test|vet)`, // go toolchain
		`^grep\s+`,               // grep searches
		`^wc\s+`,                 // word count
		`^find\s+`,               // find files
		`^ls\s+`,                 // list files
		`^cat\s+`,                // read files (simple)
	}

	allowlist := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		allowlist = append(allowlist, regexp.MustCompile(p))
	}
	return allowlist
}

// truncate shortens a string for safe logging.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
