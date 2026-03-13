package application

import (
	"context"
	"fmt"
	"time"

	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	sharedapp "github.com/alty-cli/alty/internal/shared/application"
)

// DDDVersionParser parses and applies version metadata to DDD.md content.
// This is a port interface for infrastructure adapters that handle YAML parsing.
type DDDVersionParser interface {
	// ParseVersion extracts version metadata from DDD.md content.
	ParseVersion(content string) (challengedomain.DDDVersion, error)
	// ApplyVersion updates or adds version frontmatter to DDD.md content.
	ApplyVersion(content string, version challengedomain.DDDVersion) string
}

// VersionHandler handles versioning of DDD.md documents.
type VersionHandler struct {
	reader sharedapp.FileReader
	writer sharedapp.FileWriter
	parser DDDVersionParser
}

// NewVersionHandler creates a new VersionHandler with the given dependencies.
func NewVersionHandler(
	reader sharedapp.FileReader,
	writer sharedapp.FileWriter,
	parser DDDVersionParser,
) *VersionHandler {
	return &VersionHandler{
		reader: reader,
		writer: writer,
		parser: parser,
	}
}

// VersionDDDDocument reads the current DDD.md, increments its version,
// and writes it back with updated metadata.
func (h *VersionHandler) VersionDDDDocument(
	ctx context.Context,
	path string,
	round string,
	convergenceDelta int,
	updatedAt time.Time,
) error {
	// Read current content
	content, err := h.reader.ReadFile(ctx, path)
	if err != nil {
		return fmt.Errorf("reading DDD document: %w", err)
	}

	// Parse current version
	currentVersion, err := h.parser.ParseVersion(content)
	if err != nil {
		return fmt.Errorf("parsing version: %w", err)
	}

	// Increment version
	newVersion := currentVersion.Increment(round, convergenceDelta, updatedAt)

	// Apply new version to content
	updatedContent := h.parser.ApplyVersion(content, newVersion)

	// Write back
	if err := h.writer.WriteFile(ctx, path, updatedContent); err != nil {
		return fmt.Errorf("writing DDD document: %w", err)
	}

	return nil
}
