package application_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/knowledge/application"
	"github.com/alty-cli/alty/internal/knowledge/domain"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// Mock reader
// ---------------------------------------------------------------------------

type mockKnowledgeReader struct {
	entries map[string]domain.KnowledgeEntry
	topics  map[string][]string
}

func newMockKnowledgeReader() *mockKnowledgeReader {
	tacticalPath, _ := domain.NewKnowledgePath("ddd/tactical-patterns")
	toolsPath, _ := domain.NewKnowledgePath("tools/claude-code/agent-format")

	meta := domain.NewEntryMetadata("", "", "high", false, "", "", nil)

	return &mockKnowledgeReader{
		entries: map[string]domain.KnowledgeEntry{
			"ddd/tactical-patterns": domain.NewKnowledgeEntry(
				tacticalPath,
				"Tactical Patterns",
				"# Tactical Patterns\nAggregates, Entities, VOs.",
				&meta,
				"",
			),
			"tools/claude-code/agent-format": domain.NewKnowledgeEntry(
				toolsPath,
				"Agent Format",
				"# Agent Format",
				&meta,
				"toml",
			),
		},
		topics: map[string][]string{
			"ddd":               {"tactical-patterns", "strategic-patterns"},
			"tools:claude-code": {"agent-format", "config-structure"},
		},
	}
}

func (m *mockKnowledgeReader) ReadEntry(_ context.Context, path domain.KnowledgePath, _ string) (domain.KnowledgeEntry, error) {
	entry, ok := m.entries[path.Raw()]
	if !ok {
		return domain.KnowledgeEntry{}, fmt.Errorf("entry not found: %s: %w", path.Raw(), domainerrors.ErrNotFound)
	}
	return entry, nil
}

func (m *mockKnowledgeReader) ListTopics(_ context.Context, category domain.KnowledgeCategory, tool *string) ([]string, error) {
	key := string(category)
	if tool != nil {
		key = key + ":" + *tool
	}
	topics, ok := m.topics[key]
	if !ok {
		return nil, nil
	}
	return topics, nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestKnowledgeLookupHandler(t *testing.T) {
	t.Parallel()

	t.Run("lookup delegates to reader", func(t *testing.T) {
		t.Parallel()
		reader := newMockKnowledgeReader()
		handler := application.NewKnowledgeLookupHandler(reader)

		entry, err := handler.Lookup(context.Background(), "ddd/tactical-patterns", "current")

		require.NoError(t, err)
		assert.Equal(t, "Tactical Patterns", entry.Title())
		assert.Contains(t, entry.Content(), "Aggregates")
	})

	t.Run("lookup tools path", func(t *testing.T) {
		t.Parallel()
		reader := newMockKnowledgeReader()
		handler := application.NewKnowledgeLookupHandler(reader)

		entry, err := handler.Lookup(context.Background(), "tools/claude-code/agent-format", "current")

		require.NoError(t, err)
		assert.Equal(t, "Agent Format", entry.Title())
	})

	t.Run("list categories", func(t *testing.T) {
		t.Parallel()
		reader := newMockKnowledgeReader()
		handler := application.NewKnowledgeLookupHandler(reader)

		categories := handler.ListCategories()

		assert.Contains(t, categories, "ddd")
		assert.Contains(t, categories, "tools")
		assert.Contains(t, categories, "conventions")
		assert.Contains(t, categories, "cross-tool")
	})

	t.Run("list topics delegates to reader", func(t *testing.T) {
		t.Parallel()
		reader := newMockKnowledgeReader()
		handler := application.NewKnowledgeLookupHandler(reader)

		topics, err := handler.ListTopics(context.Background(), "ddd", nil)

		require.NoError(t, err)
		assert.Contains(t, topics, "tactical-patterns")
		assert.Contains(t, topics, "strategic-patterns")
	})

	t.Run("list topics with tool filter", func(t *testing.T) {
		t.Parallel()
		reader := newMockKnowledgeReader()
		handler := application.NewKnowledgeLookupHandler(reader)

		tool := "claude-code"
		topics, err := handler.ListTopics(context.Background(), "tools", &tool)

		require.NoError(t, err)
		assert.Contains(t, topics, "agent-format")
		assert.Contains(t, topics, "config-structure")
	})

	t.Run("lookup invalid path raises error", func(t *testing.T) {
		t.Parallel()
		reader := newMockKnowledgeReader()
		handler := application.NewKnowledgeLookupHandler(reader)

		_, err := handler.Lookup(context.Background(), "", "current")

		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("lookup nonexistent entry raises error", func(t *testing.T) {
		t.Parallel()
		reader := newMockKnowledgeReader()
		handler := application.NewKnowledgeLookupHandler(reader)

		_, err := handler.Lookup(context.Background(), "ddd/nonexistent-topic", "current")

		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrNotFound)
	})

	t.Run("list topics invalid category raises error", func(t *testing.T) {
		t.Parallel()
		reader := newMockKnowledgeReader()
		handler := application.NewKnowledgeLookupHandler(reader)

		_, err := handler.ListTopics(context.Background(), "invalid-category", nil)

		require.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}
