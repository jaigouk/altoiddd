package persistence_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/shared/infrastructure/persistence"
)

func TestFilesystemDirCreator_EnsureDir_WhenPathNotExists_ExpectCreated(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "sub", "nested")

	creator := persistence.NewFilesystemDirCreator()
	err := creator.EnsureDir(context.Background(), target)
	require.NoError(t, err)

	info, err := os.Stat(target)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFilesystemDirCreator_EnsureDir_WhenPathAlreadyExists_ExpectIdempotent(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "existing")

	require.NoError(t, os.MkdirAll(target, 0o755))

	creator := persistence.NewFilesystemDirCreator()
	err := creator.EnsureDir(context.Background(), target)
	require.NoError(t, err)

	info, err := os.Stat(target)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
