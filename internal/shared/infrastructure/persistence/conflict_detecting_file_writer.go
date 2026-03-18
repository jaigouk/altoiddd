package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sharedapp "github.com/alto-cli/alto/internal/shared/application"
	"github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// Compile-time interface satisfaction check.
var _ sharedapp.FileWriter = (*ConflictDetectingFileWriter)(nil)

// ConflictDetectingFileWriter is a decorator around FileWriter that detects
// existing files and renames alto output to avoid overwriting user content.
type ConflictDetectingFileWriter struct {
	inner     sharedapp.FileWriter
	strategy  valueobjects.ConflictStrategy
	mu        sync.Mutex
	conflicts []valueobjects.FileConflict
}

// NewConflictDetectingFileWriter creates a new ConflictDetectingFileWriter
// wrapping the given inner writer with the specified conflict strategy.
func NewConflictDetectingFileWriter(inner sharedapp.FileWriter, strategy valueobjects.ConflictStrategy) *ConflictDetectingFileWriter {
	return &ConflictDetectingFileWriter{
		inner:    inner,
		strategy: strategy,
	}
}

// WriteFile writes content to the given path. If the file already exists,
// the configured conflict strategy determines behavior:
//   - Rename: writes to an alternative path (e.g., file_alto.md)
//   - Skip: does not write, records the conflict
func (w *ConflictDetectingFileWriter) WriteFile(ctx context.Context, path string, content string) error {
	if !fileExists(path) {
		if err := w.inner.WriteFile(ctx, path, content); err != nil {
			return fmt.Errorf("writing file %s: %w", path, err)
		}
		return nil
	}

	switch w.strategy {
	case valueobjects.ConflictStrategySkip:
		conflict, err := valueobjects.NewFileConflict(path, path, w.strategy)
		if err != nil {
			return fmt.Errorf("creating file conflict record: %w", err)
		}
		w.addConflict(*conflict)
		return nil

	case valueobjects.ConflictStrategyRename:
		renamed := findAvailablePath(path)
		conflict, err := valueobjects.NewFileConflict(path, renamed, w.strategy)
		if err != nil {
			return fmt.Errorf("creating file conflict record: %w", err)
		}
		w.addConflict(*conflict)
		if err := w.inner.WriteFile(ctx, renamed, content); err != nil {
			return fmt.Errorf("writing renamed file %s: %w", renamed, err)
		}
		return nil

	default:
		if err := w.inner.WriteFile(ctx, path, content); err != nil {
			return fmt.Errorf("writing file %s: %w", path, err)
		}
		return nil
	}
}

// Conflicts returns a defensive copy of all recorded file conflicts.
func (w *ConflictDetectingFileWriter) Conflicts() []valueobjects.FileConflict {
	w.mu.Lock()
	defer w.mu.Unlock()

	out := make([]valueobjects.FileConflict, len(w.conflicts))
	copy(out, w.conflicts)
	return out
}

// Reset clears all recorded conflicts.
func (w *ConflictDetectingFileWriter) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conflicts = nil
}

func (w *ConflictDetectingFileWriter) addConflict(c valueobjects.FileConflict) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.conflicts = append(w.conflicts, c)
}

// findAvailablePath computes the first available renamed path:
//
//	file.md       -> file_alto.md -> file_alto_2.md -> ...
//	README        -> README_alto  -> README_alto_2  -> ...
//	.gitignore    -> .gitignore_alto -> .gitignore_alto_2 -> ...
func findAvailablePath(path string) string {
	dir := filepath.Dir(path)
	base, ext := splitNameExt(filepath.Base(path))

	candidate := filepath.Join(dir, base+"_alto"+ext)
	if !fileExists(candidate) {
		return candidate
	}

	for i := 2; ; i++ {
		candidate = filepath.Join(dir, fmt.Sprintf("%s_alto_%d%s", base, i, ext))
		if !fileExists(candidate) {
			return candidate
		}
	}
}

// splitNameExt splits a filename into name and extension, handling hidden files.
// For hidden files like ".gitignore", the entire name is the base with no extension.
func splitNameExt(filename string) (name, ext string) {
	// Hidden files: treat the whole thing as name, no extension.
	if strings.HasPrefix(filename, ".") && !strings.Contains(filename[1:], ".") {
		return filename, ""
	}

	ext = filepath.Ext(filename)
	name = strings.TrimSuffix(filename, ext)
	return name, ext
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
