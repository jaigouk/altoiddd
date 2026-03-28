package application

import (
	"context"
	"fmt"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// ContextMapExportHandler orchestrates context map building and persistence.
type ContextMapExportHandler struct {
	writer ContextMapWriter
}

// NewContextMapExportHandler creates a ContextMapExportHandler with the given writer.
func NewContextMapExportHandler(writer ContextMapWriter) *ContextMapExportHandler {
	return &ContextMapExportHandler{writer: writer}
}

// Export builds a context map from sketches and writes it to outputPath.
func (h *ContextMapExportHandler) Export(
	ctx context.Context,
	projectName string,
	sketches []discoverydomain.BoundedContextSketch,
	outputPath string,
) error {
	cm, err := discoverydomain.BuildContextMap(projectName, sketches)
	if err != nil {
		return fmt.Errorf("building context map: %w", err)
	}

	if err := h.writer.Write(ctx, outputPath, cm); err != nil {
		return fmt.Errorf("writing context map: %w", err)
	}

	return nil
}
