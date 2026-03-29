package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED phase tests for alty-cli-vhs: resolveProjectName from README.
//
// These tests call resolveProjectName(dir), which does NOT exist yet in
// the commands package. They will fail to compile until the function is
// implemented, at which point they become proper RED tests (compile but
// fail on assertions).

func TestResolveProjectName_WhenReadmeMdHasH1_ReturnsHeading(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# kura\nSome description"), 0o644))

	got := resolveProjectName(dir)
	assert.Equal(t, "kura", got)
}

func TestResolveProjectName_WhenNoReadme_FallsBackToBaseName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	got := resolveProjectName(dir)
	assert.Equal(t, filepath.Base(dir), got)
}

func TestResolveProjectName_WhenReadmeHasNoH1_FallsBackToBaseName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("## Not an H1\nJust text"), 0o644))

	got := resolveProjectName(dir)
	assert.Equal(t, filepath.Base(dir), got)
}

func TestResolveProjectName_WhenReadmeExistsButNoMd_ReturnsHeading(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README"), []byte("# kura\nSome description"), 0o644))

	got := resolveProjectName(dir)
	assert.Equal(t, "kura", got)
}

func TestResolveProjectName_WhenReadmeMdEmpty_FallsBackToBaseName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte(""), 0o644))

	got := resolveProjectName(dir)
	assert.Equal(t, filepath.Base(dir), got)
}

func TestResolveProjectName_WhenH1HasExtraSpaces_ReturnsTrimmed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("#  My App  \nDescription"), 0o644))

	got := resolveProjectName(dir)
	assert.Equal(t, "My App", got)
}
