package mcp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeError_NilPassesThrough(t *testing.T) {
	t.Parallel()
	assert.NoError(t, SanitizeError(nil))
}

func TestSanitizeError_ConvertsInternalError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
	}{
		{"stack trace", fmt.Errorf("failed: stack trace: goroutine 1 [running]")},
		{"panic", fmt.Errorf("panic: runtime error: index out of range")},
		{"runtime error", fmt.Errorf("runtime error: invalid memory address")},
		{"internal server error", fmt.Errorf("internal server error in handler")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeError(tt.err)
			require.Error(t, result)
			assert.Equal(t, "internal error occurred", result.Error())
		})
	}
}

func TestSanitizeError_StripsAbsolutePaths(t *testing.T) {
	t.Parallel()
	err := fmt.Errorf("file not found: /home/user/project/src/main.go")
	result := SanitizeError(err)
	require.Error(t, result)
	assert.NotContains(t, result.Error(), "/home/user")
	assert.Contains(t, result.Error(), "main.go")
}

func TestSanitizeError_StripsSecrets(t *testing.T) {
	t.Parallel()
	err := fmt.Errorf("auth failed: api_key=sk_live_abc123def456")
	result := SanitizeError(err)
	require.Error(t, result)
	assert.Contains(t, result.Error(), "[REDACTED]")
	assert.NotContains(t, result.Error(), "sk_live_abc")
}

func TestSanitizeError_PreservesSimpleError(t *testing.T) {
	t.Parallel()
	err := fmt.Errorf("tool %q not found", "echo")
	result := SanitizeError(err)
	require.Error(t, result)
	assert.Equal(t, `tool "echo" not found`, result.Error())
}

func TestSanitizeError_PreservesValidationError(t *testing.T) {
	t.Parallel()
	err := fmt.Errorf("invalid ticket ID format: %q", "../etc")
	result := SanitizeError(err)
	require.Error(t, result)
	assert.Contains(t, result.Error(), "invalid ticket ID format")
}
