package application

import (
	"context"
	"fmt"

	sharedapp "github.com/alto-cli/alto/internal/shared/application"
)

// PlantUMLExportHandler orchestrates reading a story, rendering PlantUML, and writing to file.
type PlantUMLExportHandler struct {
	reader   StoryReader
	renderer PlantUMLRenderer
	writer   sharedapp.FileWriter
}

// NewPlantUMLExportHandler creates a PlantUMLExportHandler with the given dependencies.
func NewPlantUMLExportHandler(reader StoryReader, renderer PlantUMLRenderer, writer sharedapp.FileWriter) *PlantUMLExportHandler {
	return &PlantUMLExportHandler{
		reader:   reader,
		renderer: renderer,
		writer:   writer,
	}
}

// Export reads a story from storyPath, renders it to PlantUML, and writes to outputPath.
func (h *PlantUMLExportHandler) Export(ctx context.Context, storyPath, outputPath string) error {
	story, err := h.reader.Read(ctx, storyPath)
	if err != nil {
		return fmt.Errorf("reading story: %w", err)
	}

	content, err := h.renderer.Render(ctx, story)
	if err != nil {
		return fmt.Errorf("rendering plantuml: %w", err)
	}

	if err := h.writer.WriteFile(ctx, outputPath, content); err != nil {
		return fmt.Errorf("writing plantuml: %w", err)
	}

	return nil
}
