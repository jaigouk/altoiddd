// Package persistence provides filesystem-based infrastructure adapters
// for the shared bounded context.
package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	sharedapp "github.com/alto-cli/alto/internal/shared/application"
)

// Compile-time interface satisfaction check.
var _ sharedapp.FileWriter = (*FilesystemFileWriter)(nil)

// FilesystemFileWriter writes files to the local filesystem.
// Implements FileWriter by writing content to the given path,
// creating parent directories as needed.
type FilesystemFileWriter struct{}

// NewFilesystemFileWriter creates a new FilesystemFileWriter.
func NewFilesystemFileWriter() *FilesystemFileWriter {
	return &FilesystemFileWriter{}
}

// WriteFile writes content to a file at the given path.
// Creates parent directories if they don't exist. Overwrites existing files.
func (w *FilesystemFileWriter) WriteFile(_ context.Context, path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing file %s: %w", path, err)
	}
	return nil
}
