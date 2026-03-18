package persistence_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/infrastructure/persistence"
)

func TestFilesystemFileReader_ReadFile(t *testing.T) {
	t.Parallel()

	t.Run("reads existing file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.md")
		content := "# Test Content\n\nSome text here."
		require.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

		reader := persistence.NewFilesystemFileReader()
		result, err := reader.ReadFile(context.Background(), testFile)

		require.NoError(t, err)
		assert.Equal(t, content, result)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		t.Parallel()

		reader := persistence.NewFilesystemFileReader()
		_, err := reader.ReadFile(context.Background(), "/non/existent/file.md")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading file")
	})

	t.Run("reads empty file", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "empty.md")
		require.NoError(t, os.WriteFile(testFile, []byte(""), 0o644))

		reader := persistence.NewFilesystemFileReader()
		result, err := reader.ReadFile(context.Background(), testFile)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("reads file with UTF-8 content", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "utf8.md")
		content := "# Überschrift\n\n日本語テスト\n\nEmoji: 🎉"
		require.NoError(t, os.WriteFile(testFile, []byte(content), 0o644))

		reader := persistence.NewFilesystemFileReader()
		result, err := reader.ReadFile(context.Background(), testFile)

		require.NoError(t, err)
		assert.Equal(t, content, result)
	})
}
