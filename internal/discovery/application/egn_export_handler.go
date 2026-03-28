package application

import (
	"context"
	"fmt"

	sharedapp "github.com/alto-cli/alto/internal/shared/application"
)

// EgnExportHandler orchestrates reading a story, rendering Egon JSON, and writing to file.
type EgnExportHandler struct {
	reader   StoryReader
	renderer EgnRenderer
	writer   sharedapp.FileWriter
}

// NewEgnExportHandler creates an EgnExportHandler with the given dependencies.
func NewEgnExportHandler(reader StoryReader, renderer EgnRenderer, writer sharedapp.FileWriter) *EgnExportHandler {
	return &EgnExportHandler{
		reader:   reader,
		renderer: renderer,
		writer:   writer,
	}
}

// Export reads a story from storyPath, renders it to Egon JSON, and writes to outputPath.
func (h *EgnExportHandler) Export(ctx context.Context, storyPath, outputPath string) error {
	story, err := h.reader.Read(ctx, storyPath)
	if err != nil {
		return fmt.Errorf("reading story: %w", err)
	}

	content, err := h.renderer.Render(ctx, story)
	if err != nil {
		return fmt.Errorf("rendering egn: %w", err)
	}

	if err := h.writer.WriteFile(ctx, outputPath, content); err != nil {
		return fmt.Errorf("writing egn: %w", err)
	}

	return nil
}
