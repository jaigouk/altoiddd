package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/knowledge/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// KnowledgeCategory
// ---------------------------------------------------------------------------

func TestKnowledgeCategory(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "ddd", string(domain.CategoryDDD))
	assert.Equal(t, "tools", string(domain.CategoryTools))
	assert.Equal(t, "conventions", string(domain.CategoryConventions))
	assert.Equal(t, "cross-tool", string(domain.CategoryCrossTool))
}

// ---------------------------------------------------------------------------
// KnowledgePath
// ---------------------------------------------------------------------------

func TestKnowledgePath(t *testing.T) {
	t.Parallel()

	t.Run("valid ddd path", func(t *testing.T) {
		t.Parallel()
		p, err := domain.NewKnowledgePath("ddd/tactical-patterns")
		require.NoError(t, err)
		assert.Equal(t, "ddd/tactical-patterns", p.Raw())
	})

	t.Run("valid tools path", func(t *testing.T) {
		t.Parallel()
		p, err := domain.NewKnowledgePath("tools/claude-code/agent-format")
		require.NoError(t, err)
		assert.Equal(t, "tools/claude-code/agent-format", p.Raw())
	})

	t.Run("valid conventions path", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewKnowledgePath("conventions/tdd")
		require.NoError(t, err)
	})

	t.Run("valid cross-tool path", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewKnowledgePath("cross-tool/agents-md")
		require.NoError(t, err)
	})

	t.Run("rejects empty", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewKnowledgePath("")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("rejects traversal", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewKnowledgePath("ddd/../secrets")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "traversal")
	})

	t.Run("rejects invalid category", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewKnowledgePath("unknown/topic")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "category")
	})

	t.Run("rejects single segment tools", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewKnowledgePath("tools")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "category/topic")
	})

	t.Run("rejects single segment ddd", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewKnowledgePath("ddd")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "category/topic")
	})

	t.Run("rejects empty topic", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewKnowledgePath("tools/")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "category/topic")
	})

	t.Run("category extraction", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			raw      string
			expected domain.KnowledgeCategory
		}{
			{"ddd/tactical-patterns", domain.CategoryDDD},
			{"tools/cursor/rules-format", domain.CategoryTools},
			{"conventions/solid", domain.CategoryConventions},
			{"cross-tool/agents-md", domain.CategoryCrossTool},
		}
		for _, tt := range tests {
			p, err := domain.NewKnowledgePath(tt.raw)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, p.Category())
		}
	})

	t.Run("tool extraction", func(t *testing.T) {
		t.Parallel()
		p, _ := domain.NewKnowledgePath("tools/claude-code/agent-format")
		tool := p.Tool()
		require.NotNil(t, tool)
		assert.Equal(t, "claude-code", *tool)
	})

	t.Run("tool is nil for non-tools", func(t *testing.T) {
		t.Parallel()
		p, _ := domain.NewKnowledgePath("ddd/tactical-patterns")
		assert.Nil(t, p.Tool())
	})

	t.Run("subtopic extraction", func(t *testing.T) {
		t.Parallel()
		p, _ := domain.NewKnowledgePath("tools/claude-code/agent-format")
		sub := p.Subtopic()
		require.NotNil(t, sub)
		assert.Equal(t, "agent-format", *sub)
	})

	t.Run("subtopic is nil for non-tools", func(t *testing.T) {
		t.Parallel()
		p, _ := domain.NewKnowledgePath("ddd/tactical-patterns")
		assert.Nil(t, p.Subtopic())
	})

	t.Run("topic ddd", func(t *testing.T) {
		t.Parallel()
		p, _ := domain.NewKnowledgePath("ddd/tactical-patterns")
		assert.Equal(t, "tactical-patterns", p.Topic())
	})

	t.Run("topic tools combined", func(t *testing.T) {
		t.Parallel()
		p, _ := domain.NewKnowledgePath("tools/claude-code/agent-format")
		assert.Equal(t, "claude-code/agent-format", p.Topic())
	})
}

// ---------------------------------------------------------------------------
// EntryMetadata
// ---------------------------------------------------------------------------

func TestEntryMetadata(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()
		m := domain.NewEntryMetadata("", "", "high", false, "", "", nil)
		assert.Empty(t, m.LastVerified())
		assert.Empty(t, m.VerifiedAgainst())
		assert.Equal(t, "high", m.Confidence())
		assert.False(t, m.Deprecated())
		assert.Empty(t, m.NextReviewDate())
		assert.Empty(t, m.SchemaVersion())
		assert.Empty(t, m.SourceURLs())
	})

	t.Run("all fields", func(t *testing.T) {
		t.Parallel()
		m := domain.NewEntryMetadata(
			"2026-01-01", "v2.0", "medium", true, "2026-06-01", "1",
			[]string{"https://example.com"},
		)
		assert.Equal(t, "2026-01-01", m.LastVerified())
		assert.Equal(t, "v2.0", m.VerifiedAgainst())
		assert.Equal(t, "medium", m.Confidence())
		assert.True(t, m.Deprecated())
		assert.Equal(t, "2026-06-01", m.NextReviewDate())
		assert.Equal(t, "1", m.SchemaVersion())
		assert.Equal(t, []string{"https://example.com"}, m.SourceURLs())
	})
}

// ---------------------------------------------------------------------------
// KnowledgeEntry
// ---------------------------------------------------------------------------

func TestKnowledgeEntry(t *testing.T) {
	t.Parallel()

	t.Run("equality by path", func(t *testing.T) {
		t.Parallel()
		pathA, _ := domain.NewKnowledgePath("ddd/tactical-patterns")
		pathB, _ := domain.NewKnowledgePath("ddd/tactical-patterns")
		a := domain.NewKnowledgeEntry(pathA, "Tactical Patterns", "# Patterns", nil, "markdown")
		b := domain.NewKnowledgeEntry(pathB, "Different Title", "Different content", nil, "markdown")
		assert.True(t, a.EqualByPath(b))
	})

	t.Run("inequality different paths", func(t *testing.T) {
		t.Parallel()
		pathA, _ := domain.NewKnowledgePath("ddd/tactical-patterns")
		pathB, _ := domain.NewKnowledgePath("ddd/strategic-patterns")
		a := domain.NewKnowledgeEntry(pathA, "Tactical Patterns", "# Patterns", nil, "markdown")
		b := domain.NewKnowledgeEntry(pathB, "Tactical Patterns", "# Patterns", nil, "markdown")
		assert.False(t, a.EqualByPath(b))
	})

	t.Run("default format", func(t *testing.T) {
		t.Parallel()
		path, _ := domain.NewKnowledgePath("ddd/tactical-patterns")
		e := domain.NewKnowledgeEntry(path, "Tactical Patterns", "# Patterns", nil, "markdown")
		assert.Equal(t, "markdown", e.Format())
		assert.Nil(t, e.Metadata())
	})
}
