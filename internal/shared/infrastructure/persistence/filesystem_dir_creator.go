package persistence

import (
	"context"
	"fmt"
	"os"

	sharedapp "github.com/alty-cli/alty/internal/shared/application"
)

// Compile-time check that FilesystemDirCreator satisfies DirCreator.
var _ sharedapp.DirCreator = (*FilesystemDirCreator)(nil)

// FilesystemDirCreator creates directories on the local filesystem.
type FilesystemDirCreator struct{}

// NewFilesystemDirCreator creates a new FilesystemDirCreator.
func NewFilesystemDirCreator() *FilesystemDirCreator {
	return &FilesystemDirCreator{}
}

// EnsureDir creates the directory at path (including parents) if it does not exist.
func (c *FilesystemDirCreator) EnsureDir(_ context.Context, path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", path, err)
	}
	return nil
}
