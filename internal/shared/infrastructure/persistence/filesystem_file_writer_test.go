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

func TestFilesystemFileWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		subPath    string
		content    string
		wantErr    bool
		errString  string
		setup      func(t *testing.T, dir string)
		teardown   func(t *testing.T, dir string)
		assertFunc func(t *testing.T, dir string)
	}{
		{
			name:    "creates file with content",
			subPath: "output.md",
			content: "hello world",
		},
		{
			name:    "creates parent directories",
			subPath: filepath.Join("a", "b", "c", "file.md"),
			content: "nested content",
		},
		{
			name:    "overwrites existing file",
			subPath: "file.md",
			content: "new content",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				err := os.WriteFile(filepath.Join(dir, "file.md"), []byte("old content"), 0o644)
				require.NoError(t, err)
			},
		},
		{
			name:    "writes empty content",
			subPath: "empty.md",
			content: "",
		},
		{
			name:    "writes unicode content",
			subPath: "unicode.md",
			content: "日本語テスト -- U o a n",
		},
		{
			name:    "permission error on readonly directory",
			subPath: filepath.Join("readonly", "file.md"),
			content: "content",
			wantErr: true,
			setup: func(t *testing.T, dir string) {
				t.Helper()
				readonlyDir := filepath.Join(dir, "readonly")
				require.NoError(t, os.Mkdir(readonlyDir, 0o500))
			},
			teardown: func(t *testing.T, dir string) {
				t.Helper()
				readonlyDir := filepath.Join(dir, "readonly")
				_ = os.Chmod(readonlyDir, 0o700)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()

			if tt.setup != nil {
				tt.setup(t, dir)
			}
			if tt.teardown != nil {
				defer tt.teardown(t, dir)
			}

			writer := persistence.NewFilesystemFileWriter()
			fullPath := filepath.Join(dir, tt.subPath)
			err := writer.WriteFile(context.Background(), fullPath, tt.content)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			got, err := os.ReadFile(fullPath)
			require.NoError(t, err)
			assert.Equal(t, tt.content, string(got))
		})
	}
}
