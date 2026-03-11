package valueobjects

import (
	"fmt"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// FileConflict records that an alty output file was redirected because
// the original path already existed on disk.
type FileConflict struct {
	originalPath string
	actualPath   string
	strategy     ConflictStrategy
}

// NewFileConflict creates a FileConflict value object.
func NewFileConflict(originalPath, actualPath string, strategy ConflictStrategy) (*FileConflict, error) {
	if originalPath == "" {
		return nil, fmt.Errorf("original path required: %w", domainerrors.ErrInvariantViolation)
	}
	if actualPath == "" {
		return nil, fmt.Errorf("actual path required: %w", domainerrors.ErrInvariantViolation)
	}
	return &FileConflict{
		originalPath: originalPath,
		actualPath:   actualPath,
		strategy:     strategy,
	}, nil
}

// OriginalPath returns the originally requested file path.
func (fc *FileConflict) OriginalPath() string { return fc.originalPath }

// ActualPath returns the path where the file was actually written.
func (fc *FileConflict) ActualPath() string { return fc.actualPath }

// Strategy returns the conflict resolution strategy that was applied.
func (fc *FileConflict) Strategy() ConflictStrategy { return fc.strategy }
