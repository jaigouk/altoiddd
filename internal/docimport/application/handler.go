package application

import (
	"context"
	"fmt"

	"github.com/alty-cli/alty/internal/docimport/domain"
)

// DocImportHandler orchestrates the import of domain models from existing documentation.
type DocImportHandler struct {
	importer DocImporter
}

// NewDocImportHandler creates a new DocImportHandler with the given importer port.
func NewDocImportHandler(importer DocImporter) *DocImportHandler {
	return &DocImportHandler{importer: importer}
}

// Import delegates to the DocImporter port to parse documentation into an ImportResult.
func (h *DocImportHandler) Import(ctx context.Context, docDir string) (*domain.ImportResult, error) {
	result, err := h.importer.Import(ctx, docDir)
	if err != nil {
		return nil, fmt.Errorf("importing docs from %q: %w", docDir, err)
	}
	return result, nil
}
