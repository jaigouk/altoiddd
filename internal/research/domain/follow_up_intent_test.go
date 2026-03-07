package domain_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/research/domain"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// FollowUpIntent
// ---------------------------------------------------------------------------

func TestFollowUpIntent(t *testing.T) {
	t.Parallel()

	t.Run("creates with title and description", func(t *testing.T) {
		t.Parallel()
		intent, err := domain.NewFollowUpIntent("Implement SessionStore", "Create in-memory store with TTL")
		require.NoError(t, err)
		assert.Equal(t, "Implement SessionStore", intent.Title())
		assert.Equal(t, "Create in-memory store with TTL", intent.Description())
	})

	t.Run("empty title raises", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewFollowUpIntent("", "Details")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("whitespace only title raises", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewFollowUpIntent("   ", "Details")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("empty description allowed", func(t *testing.T) {
		t.Parallel()
		intent, err := domain.NewFollowUpIntent("Task", "")
		require.NoError(t, err)
		assert.Empty(t, intent.Description())
	})

	t.Run("equality by value", func(t *testing.T) {
		t.Parallel()
		a, _ := domain.NewFollowUpIntent("Task A", "Desc")
		b, _ := domain.NewFollowUpIntent("Task A", "Desc")
		assert.Equal(t, a, b)
	})

	t.Run("different intents not equal", func(t *testing.T) {
		t.Parallel()
		a, _ := domain.NewFollowUpIntent("Task A", "Desc")
		b, _ := domain.NewFollowUpIntent("Task B", "Desc")
		assert.NotEqual(t, a, b)
	})
}

// ---------------------------------------------------------------------------
// FollowUpAuditResult
// ---------------------------------------------------------------------------

func TestFollowUpAuditResult(t *testing.T) {
	t.Parallel()

	t.Run("creates with all fields", func(t *testing.T) {
		t.Parallel()
		intent, _ := domain.NewFollowUpIntent("Task 1", "")
		result := domain.NewFollowUpAuditResult(
			"k7m.8",
			"docs/research/20260223_gap_analysis_design.md",
			[]domain.FollowUpIntent{intent},
			nil,
			[]domain.FollowUpIntent{intent},
		)
		assert.Equal(t, "k7m.8", result.SpikeID())
		assert.Equal(t, 1, result.OrphanedCount())
	})

	t.Run("orphaned count property", func(t *testing.T) {
		t.Parallel()
		var intents []domain.FollowUpIntent
		for i := range 5 {
			fi, _ := domain.NewFollowUpIntent(fmt.Sprintf("Task %d", i), "")
			intents = append(intents, fi)
		}
		result := domain.NewFollowUpAuditResult(
			"k7m.8", "report.md",
			intents,
			[]string{"alty-abc", "alty-def"},
			intents[:3],
		)
		assert.Equal(t, 3, result.OrphanedCount())
	})

	t.Run("has orphans true", func(t *testing.T) {
		t.Parallel()
		intent, _ := domain.NewFollowUpIntent("Lost task", "")
		result := domain.NewFollowUpAuditResult(
			"k7m.8", "report.md",
			[]domain.FollowUpIntent{intent},
			nil,
			[]domain.FollowUpIntent{intent},
		)
		assert.True(t, result.HasOrphans())
	})

	t.Run("has orphans false when all matched", func(t *testing.T) {
		t.Parallel()
		intent, _ := domain.NewFollowUpIntent("Created task", "")
		result := domain.NewFollowUpAuditResult(
			"k7m.8", "report.md",
			[]domain.FollowUpIntent{intent},
			[]string{"alty-abc"},
			nil,
		)
		assert.False(t, result.HasOrphans())
	})

	t.Run("no intents means no orphans", func(t *testing.T) {
		t.Parallel()
		result := domain.NewFollowUpAuditResult("k7m.8", "report.md", nil, nil, nil)
		assert.Equal(t, 0, result.OrphanedCount())
		assert.False(t, result.HasOrphans())
	})

	t.Run("defined count property", func(t *testing.T) {
		t.Parallel()
		var intents []domain.FollowUpIntent
		for i := range 17 {
			fi, _ := domain.NewFollowUpIntent(fmt.Sprintf("T%d", i), "")
			intents = append(intents, fi)
		}
		result := domain.NewFollowUpAuditResult("k7m.8", "report.md", intents, nil, intents)
		assert.Equal(t, 17, result.DefinedCount())
	})
}
