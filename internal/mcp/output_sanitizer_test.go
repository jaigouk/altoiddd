package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeOutput_StripsAbsolutePaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unix home path",
			input: "Error in /home/user/project/src/main.go",
			want:  "Error in main.go",
		},
		{
			name:  "macOS Users path",
			input: "File at /Users/jaigouk/alto-cli/internal/mcp/audit.go",
			want:  "File at audit.go",
		},
		{
			name:  "tmp path",
			input: "Created /tmp/alto-test-12345/output.txt",
			want:  "Created output.txt",
		},
		{
			name:  "etc path",
			input: "Reading /etc/passwd",
			want:  "Reading passwd",
		},
		{
			name:  "windows Users path",
			input: `Error in C:\Users\dev\project\src\main.go`,
			want:  "Error in main.go",
		},
		{
			name:  "windows deep path",
			input: `File at C:\Projects\alto\internal\mcp\audit.go`,
			want:  "File at audit.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, SanitizeOutput(tt.input))
		})
	}
}

func TestSanitizeOutput_PreservesRelativePaths(t *testing.T) {
	t.Parallel()
	input := "Error in src/main.go at line 42"
	assert.Equal(t, input, SanitizeOutput(input))
}

func TestSanitizeOutput_StripsSecretPatterns(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
	}{
		{"api key", "api_key=sk_live_abc123def456ghi789"},
		{"openai key", "Using key sk-abcdefghijklmnopqrstuvwxyz1234"},
		{"password field", "password: hunter2"},
		{"github pat", "token: ghp_abcdefghijklmnopqrstuvwxyz1234567890"},
		{"aws key", "AKIAIOSFODNN7EXAMPLE"},
		{"bearer token", "bearer = eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := SanitizeOutput(tt.input)
			assert.Contains(t, result, "[REDACTED]", "should redact %s", tt.name)
			assert.NotContains(t, result, "hunter2")
			assert.NotContains(t, result, "sk_live_abc")
		})
	}
}

func TestSanitizeOutput_CleanTextUnchanged(t *testing.T) {
	t.Parallel()
	input := "All tests passed. Coverage: 85.3%"
	assert.Equal(t, input, SanitizeOutput(input))
}

func TestSanitizeOutput_EmptyString(t *testing.T) {
	t.Parallel()
	assert.Empty(t, SanitizeOutput(""))
}

func TestSanitizeOutput_MultiplePatterns(t *testing.T) {
	t.Parallel()
	input := "Error at /home/user/project/main.go: api_key=secret123"
	result := SanitizeOutput(input)
	assert.NotContains(t, result, "/home/user")
	assert.Contains(t, result, "[REDACTED]")
}
