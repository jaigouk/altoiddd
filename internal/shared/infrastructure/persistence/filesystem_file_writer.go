// Package persistence provides filesystem-based infrastructure adapters
// for the shared bounded context.
package persistence

import (
	"context"
	"os"
	"path/filepath"

	sharedapp "github.com/alty-cli/alty/internal/shared/application"
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
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
