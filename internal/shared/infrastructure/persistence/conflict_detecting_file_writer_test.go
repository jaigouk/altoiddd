package persistence_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/shared/domain/valueobjects"
	"github.com/alto-cli/alto/internal/shared/infrastructure/persistence"
)

func TestConflictDetectingFileWriter_WriteFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		strategy       valueobjects.ConflictStrategy
		setup          func(t *testing.T, dir string)
		writePath      string // relative to dir
		content        string
		wantWrittenAt  string // relative to dir — where content ends up
		wantConflicts  int
		wantOrigPath   string // original path in conflict record (relative)
		wantActualPath string // actual path in conflict record (relative)
	}{
		{
			name:          "no conflict writes to original path",
			strategy:      valueobjects.ConflictStrategyRename,
			writePath:     "file.md",
			content:       "hello",
			wantWrittenAt: "file.md",
			wantConflicts: 0,
		},
		{
			name:     "rename adds _alto suffix before extension",
			strategy: valueobjects.ConflictStrategyRename,
			setup: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "file.md"), []byte("existing"), 0o644))
			},
			writePath:      "file.md",
			content:        "new content",
			wantWrittenAt:  "file_alto.md",
			wantConflicts:  1,
			wantOrigPath:   "file.md",
			wantActualPath: "file_alto.md",
		},
		{
			name:     "rename increments when _alto also exists",
			strategy: valueobjects.ConflictStrategyRename,
			setup: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "file.md"), []byte("v1"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(dir, "file_alto.md"), []byte("v2"), 0o644))
			},
			writePath:      "file.md",
			content:        "v3",
			wantWrittenAt:  "file_alto_2.md",
			wantConflicts:  1,
			wantOrigPath:   "file.md",
			wantActualPath: "file_alto_2.md",
		},
		{
			name:     "rename handles no extension",
			strategy: valueobjects.ConflictStrategyRename,
			setup: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "README"), []byte("existing"), 0o644))
			},
			writePath:      "README",
			content:        "new readme",
			wantWrittenAt:  "README_alto",
			wantConflicts:  1,
			wantOrigPath:   "README",
			wantActualPath: "README_alto",
		},
		{
			name:     "rename handles hidden files",
			strategy: valueobjects.ConflictStrategyRename,
			setup: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("existing"), 0o644))
			},
			writePath:      ".gitignore",
			content:        "new ignore",
			wantWrittenAt:  ".gitignore_alto",
			wantConflicts:  1,
			wantOrigPath:   ".gitignore",
			wantActualPath: ".gitignore_alto",
		},
		{
			name:     "skip strategy does not write when conflict exists",
			strategy: valueobjects.ConflictStrategySkip,
			setup: func(t *testing.T, dir string) {
				t.Helper()
				require.NoError(t, os.WriteFile(filepath.Join(dir, "file.md"), []byte("keep me"), 0o644))
			},
			writePath:     "file.md",
			content:       "should not appear",
			wantConflicts: 1,
		},
		{
			name:     "rename handles nested paths",
			strategy: valueobjects.ConflictStrategyRename,
			setup: func(t *testing.T, dir string) {
				t.Helper()
				sub := filepath.Join(dir, "docs")
				require.NoError(t, os.MkdirAll(sub, 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(sub, "guide.md"), []byte("existing"), 0o644))
			},
			writePath:      filepath.Join("docs", "guide.md"),
			content:        "new guide",
			wantWrittenAt:  filepath.Join("docs", "guide_alto.md"),
			wantConflicts:  1,
			wantOrigPath:   filepath.Join("docs", "guide.md"),
			wantActualPath: filepath.Join("docs", "guide_alto.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()

			if tt.setup != nil {
				tt.setup(t, dir)
			}

			inner := persistence.NewFilesystemFileWriter()
			writer := persistence.NewConflictDetectingFileWriter(inner, tt.strategy)

			fullPath := filepath.Join(dir, tt.writePath)
			err := writer.WriteFile(context.Background(), fullPath, tt.content)
			require.NoError(t, err)

			conflicts := writer.Conflicts()
			assert.Len(t, conflicts, tt.wantConflicts)

			if tt.strategy == valueobjects.ConflictStrategySkip && tt.wantConflicts > 0 {
				// Original file should be untouched.
				got, readErr := os.ReadFile(fullPath)
				require.NoError(t, readErr)
				assert.NotEqual(t, tt.content, string(got))
				return
			}

			if tt.wantWrittenAt != "" {
				writtenPath := filepath.Join(dir, tt.wantWrittenAt)
				got, readErr := os.ReadFile(writtenPath)
				require.NoError(t, readErr)
				assert.Equal(t, tt.content, string(got))
			}

			if tt.wantConflicts > 0 {
				c := conflicts[0]
				assert.Equal(t, filepath.Join(dir, tt.wantOrigPath), c.OriginalPath())
				assert.Equal(t, filepath.Join(dir, tt.wantActualPath), c.ActualPath())
				assert.Equal(t, tt.strategy, c.Strategy())
			}
		})
	}
}

func TestConflictDetectingFileWriter_Reset(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.md"), []byte("existing"), 0o644))

	inner := persistence.NewFilesystemFileWriter()
	writer := persistence.NewConflictDetectingFileWriter(inner, valueobjects.ConflictStrategyRename)

	err := writer.WriteFile(context.Background(), filepath.Join(dir, "file.md"), "new")
	require.NoError(t, err)
	assert.Len(t, writer.Conflicts(), 1)

	writer.Reset()
	assert.Empty(t, writer.Conflicts())
}
