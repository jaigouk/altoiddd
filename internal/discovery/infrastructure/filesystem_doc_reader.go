package infrastructure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	discoveryapp "github.com/alto-cli/alto/internal/discovery/application"
)

// Compile-time interface check.
var _ discoveryapp.DocReader = (*FilesystemDocReader)(nil)

// FilesystemDocReader reads documentation files from the local filesystem.
type FilesystemDocReader struct{}

// NewFilesystemDocReader creates a FilesystemDocReader.
func NewFilesystemDocReader() *FilesystemDocReader {
	return &FilesystemDocReader{}
}

// ReadDocs reads markdown and text files from docsDir and returns a map of filename->content.
func (r *FilesystemDocReader) ReadDocs(_ context.Context, docsDir string) (map[string]string, error) {
	entries, err := os.ReadDir(docsDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	docs := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".md" && ext != ".txt" {
			continue
		}

		data, readErr := os.ReadFile(filepath.Join(docsDir, name))
		if readErr != nil {
			return nil, fmt.Errorf("reading %q: %w", name, readErr)
		}
		docs[name] = string(data)
	}
	return docs, nil
}
