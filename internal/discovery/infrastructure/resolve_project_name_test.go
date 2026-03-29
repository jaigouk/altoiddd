package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED phase tests for alty-cli-vhs: CLIDiscoveryAdapter.resolveProjectName.
//
// The current implementation of resolveProjectName always returns filepath.Base(dir).
// These tests verify the NEW behavior: reading README.md/README to extract the H1 heading.
// They will compile (the method exists) but FAIL because the method doesn't yet read README files.

func TestCLIDiscoveryAdapter_resolveProjectName_WhenReadmeMdHasH1_ReturnsHeading(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# kura\nSome description"), 0o644))

	adapter := &CLIDiscoveryAdapter{projectDir: dir}

	got, err := adapter.resolveProjectName()
	require.NoError(t, err)
	assert.Equal(t, "kura", got)
}

func TestCLIDiscoveryAdapter_resolveProjectName_WhenNoReadme_FallsBackToBaseName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	adapter := &CLIDiscoveryAdapter{projectDir: dir}

	got, err := adapter.resolveProjectName()
	require.NoError(t, err)
	assert.Equal(t, filepath.Base(dir), got)
}
