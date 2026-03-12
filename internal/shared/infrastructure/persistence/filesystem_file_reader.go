package persistence

import (
	"context"
	"fmt"
	"os"

	sharedapp "github.com/alty-cli/alty/internal/shared/application"
)

// Compile-time interface satisfaction check.
var _ sharedapp.FileReader = (*FilesystemFileReader)(nil)

// FilesystemFileReader reads files from the local filesystem.
// Implements FileReader by reading content from the given path.
type FilesystemFileReader struct{}

// NewFilesystemFileReader creates a new FilesystemFileReader.
func NewFilesystemFileReader() *FilesystemFileReader {
	return &FilesystemFileReader{}
}

// ReadFile reads content from a file at the given path.
func (r *FilesystemFileReader) ReadFile(_ context.Context, path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}
	return string(content), nil
}
