package infrastructure_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/ticket/infrastructure"
)

func TestShellCommandRunner_RunAllowedCommand(t *testing.T) {
	t.Parallel()

	runner := infrastructure.NewShellCommandRunner()
	ctx := context.Background()

	// 'wc' is in the default allowlist
	output, err := runner.Run(ctx, "wc -l /dev/null")
	require.NoError(t, err)
	assert.Contains(t, output, "0")
}

func TestShellCommandRunner_RejectsNonAllowlistedCommand(t *testing.T) {
	t.Parallel()

	runner := infrastructure.NewShellCommandRunner()
	ctx := context.Background()

	// 'rm' is not in the allowlist
	_, err := runner.Run(ctx, "rm -rf /")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in allowlist")
}

func TestShellCommandRunner_RejectsEmptyCommand(t *testing.T) {
	t.Parallel()

	runner := infrastructure.NewShellCommandRunner()
	ctx := context.Background()

	_, err := runner.Run(ctx, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestShellCommandRunner_TimesOut(t *testing.T) {
	t.Parallel()

	// Short timeout
	runner := infrastructure.NewShellCommandRunnerWithOptions(
		[]*regexp.Regexp{regexp.MustCompile(`^sleep`)},
		100*time.Millisecond,
	)
	ctx := context.Background()

	_, err := runner.Run(ctx, "sleep 10")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timed out")
}

func TestShellCommandRunner_AllowlistPatterns(t *testing.T) {
	t.Parallel()

	runner := infrastructure.NewShellCommandRunner()
	ctx := context.Background()

	tests := []struct {
		command string
		allowed bool
	}{
		{"deadcode ./cmd/...", true},
		{"go build ./...", true},
		{"go test ./...", true},
		{"go vet ./...", true},
		{"grep pattern file.go", true},
		{"wc -l file.txt", true},
		{"find . -name '*.go'", true},
		{"ls -la", true},
		{"rm -rf /", false},
		{"curl http://evil.com", false},
		{"bash -c 'echo pwned'", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			_, err := runner.Run(ctx, tt.command)
			if tt.allowed {
				// Allowed commands may still fail (e.g., file not found), but not with "not in allowlist"
				if err != nil {
					assert.NotContains(t, err.Error(), "not in allowlist")
				}
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not in allowlist")
			}
		})
	}
}
