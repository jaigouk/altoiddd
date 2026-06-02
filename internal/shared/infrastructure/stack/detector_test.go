package stack_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/infrastructure/stack"
)

func TestDetectProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		files         map[string]string
		expectedStack string
		fitnessAvail  bool
	}{
		{
			name:          "Go project",
			files:         map[string]string{"go.mod": "module test\n\ngo 1.26\n"},
			expectedStack: "go-mod",
			fitnessAvail:  true,
		},
		{
			name:          "Python project",
			files:         map[string]string{"pyproject.toml": "[project]\nname = \"test\"\n"},
			expectedStack: "python-uv",
			fitnessAvail:  true,
		},
		{
			name:          "Generic project",
			files:         map[string]string{},
			expectedStack: "generic",
			fitnessAvail:  false,
		},
		{
			name: "Go takes precedence over Python",
			files: map[string]string{
				"go.mod":         "module test\n\ngo 1.26\n",
				"pyproject.toml": "[project]\nname = \"test\"\n",
			},
			expectedStack: "go-mod",
			fitnessAvail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			for name, content := range tt.files {
				require.NoError(t, os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644))
			}

			profile := stack.DetectProfile(tmpDir)
			assert.Equal(t, tt.expectedStack, profile.StackID())
			assert.Equal(t, tt.fitnessAvail, profile.FitnessAvailable())
		})
	}
}

func TestDetectProfile_EmptyStringFallsBackToCwd(t *testing.T) {
	t.Parallel()
	// Empty string should trigger cwd fallback logic.
	// We can't reliably test cwd behavior without changing it (which breaks parallel tests),
	// but we can verify the function doesn't panic and returns a valid profile.
	profile := stack.DetectProfile("")
	assert.NotEmpty(t, profile.StackID())
}

// ExtractLanguageFromText — keyword-based language detection for the README fallback.
// Precedence: Go keywords (go.mod, golang, " go ", "package main") are checked
// before Python keywords (python, pyproject, pip install, .py). First match wins,
// returning "go" or "python" respectively, or "" for empty or unrecognized input.

func TestExtractLanguageFromText_WhenGoModMentioned_ExpectGo(t *testing.T) {
	t.Parallel()
	got := stack.ExtractLanguageFromText("This project uses go.mod for module management.")
	assert.Equal(t, "go", got)
}

func TestExtractLanguageFromText_WhenGolangMentioned_ExpectGo(t *testing.T) {
	t.Parallel()
	got := stack.ExtractLanguageFromText("A small CLI tool written in Golang.")
	assert.Equal(t, "go", got)
}

func TestExtractLanguageFromText_WhenPythonMentioned_ExpectPython(t *testing.T) {
	t.Parallel()
	got := stack.ExtractLanguageFromText("This is a Python service using FastAPI.")
	assert.Equal(t, "python", got)
}

func TestExtractLanguageFromText_WhenPyprojectMentioned_ExpectPython(t *testing.T) {
	t.Parallel()
	got := stack.ExtractLanguageFromText("Configured via pyproject.toml at the repo root.")
	assert.Equal(t, "python", got)
}

func TestExtractLanguageFromText_WhenEmptyText_ExpectEmptyString(t *testing.T) {
	t.Parallel()
	got := stack.ExtractLanguageFromText("")
	assert.Empty(t, got)
}

func TestExtractLanguageFromText_WhenUnrecognizedLanguage_ExpectEmptyString(t *testing.T) {
	t.Parallel()
	got := stack.ExtractLanguageFromText("A Rust workspace with Cargo and tokio.")
	assert.Empty(t, got)
}

func TestExtractLanguageFromText_WhenBothGoAndPython_ExpectGo(t *testing.T) {
	t.Parallel()
	// Go group is checked first; "golang" appears before any python keyword in
	// precedence, so a mixed mention must yield "go".
	got := stack.ExtractLanguageFromText("Polyglot repo: golang services and python scripts.")
	assert.Equal(t, "go", got)
}

func TestExtractLanguageFromText_WhenCaseInsensitiveGolang_ExpectGo(t *testing.T) {
	t.Parallel()
	got := stack.ExtractLanguageFromText("Built with GOLANG and gRPC.")
	assert.Equal(t, "go", got)
}
