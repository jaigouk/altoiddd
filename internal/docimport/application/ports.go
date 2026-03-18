// Package application contains use-case handlers and port interfaces for the DocImport BC.
package application

import (
	"context"

	"github.com/alto-cli/alto/internal/docimport/domain"
)

// DocImporter is the port interface for importing domain models from existing documentation.
type DocImporter interface {
	Import(ctx context.Context, docDir string) (*domain.ImportResult, error)
}
