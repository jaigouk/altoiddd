package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkDiscoveryCompleted_FlipsFalseToTrue(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configContent := `[project]
name = "my-app"

[discovery]
completed = false

# [llm]
# provider = ""
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.toml"), []byte(configContent), 0o644))

	markDiscoveryCompleted(tmpDir)

	data, err := os.ReadFile(filepath.Join(tmpDir, "config.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "completed = true")
	assert.NotContains(t, string(data), "completed = false")
}

func TestMarkDiscoveryCompleted_PreservesOtherContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configContent := `[project]
name = "my-app"
description = "A cool project"

[detection]
detected = ["claude", "cursor"]

[discovery]
completed = false

# [llm]
# provider = ""
# model = ""
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.toml"), []byte(configContent), 0o644))

	markDiscoveryCompleted(tmpDir)

	data, err := os.ReadFile(filepath.Join(tmpDir, "config.toml"))
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, `name = "my-app"`)
	assert.Contains(t, content, `description = "A cool project"`)
	assert.Contains(t, content, `detected = ["claude", "cursor"]`)
	assert.Contains(t, content, "completed = true")
	assert.Contains(t, content, "# [llm]")
	assert.Contains(t, content, `# provider = ""`)
}

func TestMarkDiscoveryCompleted_IdempotentWhenAlreadyTrue(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configContent := `[discovery]
completed = true
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.toml"), []byte(configContent), 0o644))

	markDiscoveryCompleted(tmpDir)

	data, err := os.ReadFile(filepath.Join(tmpDir, "config.toml"))
	require.NoError(t, err)
	assert.Equal(t, configContent, string(data))
}

func TestMarkDiscoveryCompleted_MissingFile_NoPanic(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Should not panic when config.toml does not exist
	assert.NotPanics(t, func() {
		markDiscoveryCompleted(tmpDir)
	})
}

func TestMarkDiscoveryCompleted_NoDiscoverySection_Noop(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configContent := `[project]
name = "my-app"

# [llm]
# provider = ""
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.toml"), []byte(configContent), 0o644))

	markDiscoveryCompleted(tmpDir)

	data, err := os.ReadFile(filepath.Join(tmpDir, "config.toml"))
	require.NoError(t, err)
	assert.Equal(t, configContent, string(data))
}
