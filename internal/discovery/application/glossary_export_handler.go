package application

import (
	"context"
	"fmt"

	discoverydomain "github.com/alto-cli/alto/internal/discovery/domain"
)

// GlossaryExportHandler orchestrates glossary extraction and persistence.
type GlossaryExportHandler struct {
	reader    StoryReader
	writer    GlossaryWriter
	extractor discoverydomain.GlossaryExtractor
}

// NewGlossaryExportHandler creates a GlossaryExportHandler with the given StoryReader and GlossaryWriter.
func NewGlossaryExportHandler(reader StoryReader, writer GlossaryWriter) *GlossaryExportHandler {
	return &GlossaryExportHandler{
		reader: reader,
		writer: writer,
	}
}

// Export reads stories from storyPaths, extracts glossary entries, and writes them to outputPath.
func (h *GlossaryExportHandler) Export(
	ctx context.Context,
	storyPaths []string,
	contextMap *discoverydomain.ContextMap,
	outputPath string,
) error {
	stories := make([]*discoverydomain.DomainStory, 0, len(storyPaths))

	for _, path := range storyPaths {
		story, err := h.reader.Read(ctx, path)
		if err != nil {
			return fmt.Errorf("reading story %q: %w", path, err)
		}

		stories = append(stories, story)
	}

	entries, err := h.extractor.Extract(stories, contextMap)
	if err != nil {
		return fmt.Errorf("extracting glossary: %w", err)
	}

	if err := h.writer.Write(ctx, outputPath, entries); err != nil {
		return fmt.Errorf("writing glossary: %w", err)
	}

	return nil
}
