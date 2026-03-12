package application

import (
	"context"
	"fmt"
	"time"

	challengedomain "github.com/alty-cli/alty/internal/challenge/domain"
	sharedapp "github.com/alty-cli/alty/internal/shared/application"
)

// VersionHandler handles versioning of DDD.md documents.
type VersionHandler struct {
	reader sharedapp.FileReader
	writer sharedapp.FileWriter
}

// NewVersionHandler creates a new VersionHandler with the given dependencies.
func NewVersionHandler(reader sharedapp.FileReader, writer sharedapp.FileWriter) *VersionHandler {
	return &VersionHandler{
		reader: reader,
		writer: writer,
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
	currentVersion, err := challengedomain.ParseDDDVersion(content)
	if err != nil {
		return fmt.Errorf("parsing version: %w", err)
	}

	// Increment version
	newVersion := currentVersion.Increment(round, convergenceDelta, updatedAt)

	// Apply new version to content
	updatedContent := challengedomain.ApplyVersion(content, newVersion)

	// Write back
	if err := h.writer.WriteFile(ctx, path, updatedContent); err != nil {
		return fmt.Errorf("writing DDD document: %w", err)
	}

	return nil
}
