package stack_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/shared/infrastructure/stack"
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
