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
	validator discoverydomain.CoherenceValidator
}

// NewGlossaryExportHandler creates a GlossaryExportHandler with the given StoryReader and GlossaryWriter.
func NewGlossaryExportHandler(reader StoryReader, writer GlossaryWriter) *GlossaryExportHandler {
	return &GlossaryExportHandler{
		reader:    reader,
		writer:    writer,
		validator: discoverydomain.CoherenceValidator{},
	}
}

// Export reads stories from storyPaths, validates cross-story coherence,
// extracts glossary entries, and writes them to outputPath.
// Coherence findings are advisory — the glossary is always written regardless of findings.
func (h *GlossaryExportHandler) Export(
	ctx context.Context,
	storyPaths []string,
	contextMap *discoverydomain.ContextMap,
	outputPath string,
) ([]discoverydomain.CoherenceFinding, error) {
	stories := make([]*discoverydomain.DomainStory, 0, len(storyPaths))

	for _, path := range storyPaths {
		story, err := h.reader.Read(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("reading story %q: %w", path, err)
		}

		stories = append(stories, story)
	}

	report := h.validator.Validate(stories)

	entries, err := h.extractor.Extract(stories, contextMap)
	if err != nil {
		return nil, fmt.Errorf("extracting glossary: %w", err)
	}

	if err := h.writer.Write(ctx, outputPath, entries); err != nil {
		return nil, fmt.Errorf("writing glossary: %w", err)
	}

	return report.Findings(), nil
}
