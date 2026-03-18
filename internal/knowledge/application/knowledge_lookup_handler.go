package application

import (
	"context"
	"fmt"

	"github.com/alto-cli/alto/internal/knowledge/domain"
)

// KnowledgeReader is a handler-local interface for reading knowledge entries.
// Defined where consumed per Go convention.
type KnowledgeReader interface {
	// ReadEntry reads a single knowledge entry by path.
	ReadEntry(ctx context.Context, path domain.KnowledgePath, version string) (domain.KnowledgeEntry, error)

	// ListTopics lists available topics within a category.
	ListTopics(ctx context.Context, category domain.KnowledgeCategory, tool *string) ([]string, error)
}

// KnowledgeLookupHandler orchestrates knowledge base lookup operations.
// It parses user-facing path strings into KnowledgePath value objects
// and delegates to a KnowledgeReader for actual retrieval.
type KnowledgeLookupHandler struct {
	reader KnowledgeReader
}

// NewKnowledgeLookupHandler creates a new KnowledgeLookupHandler.
func NewKnowledgeLookupHandler(reader KnowledgeReader) *KnowledgeLookupHandler {
	return &KnowledgeLookupHandler{reader: reader}
}

// Lookup looks up a knowledge entry by its RLM path string.
func (h *KnowledgeLookupHandler) Lookup(ctx context.Context, pathStr string, version string) (domain.KnowledgeEntry, error) {
	path, err := domain.NewKnowledgePath(pathStr)
	if err != nil {
		return domain.KnowledgeEntry{}, fmt.Errorf("parse knowledge path: %w", err)
	}
	entry, err := h.reader.ReadEntry(ctx, path, version)
	if err != nil {
		return domain.KnowledgeEntry{}, fmt.Errorf("read entry: %w", err)
	}
	return entry, nil
}

// ListCategories returns all available knowledge category values.
func (h *KnowledgeLookupHandler) ListCategories() []string {
	return domain.AllCategoryValues()
}

// ListTopics lists topics within a category.
func (h *KnowledgeLookupHandler) ListTopics(ctx context.Context, category string, tool *string) ([]string, error) {
	cat, err := domain.ParseCategory(category)
	if err != nil {
		return nil, fmt.Errorf("parse category: %w", err)
	}
	topics, err := h.reader.ListTopics(ctx, cat, tool)
	if err != nil {
		return nil, fmt.Errorf("list topics: %w", err)
	}
	return topics, nil
}
