package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/dochealth/domain"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// 1. DocHealthStatus enum
// ---------------------------------------------------------------------------

func TestDocHealthStatusEnum(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		status   domain.DocHealthStatus
		expected string
	}{
		{"ok", domain.DocHealthOK, "ok"},
		{"stale", domain.DocHealthStale, "stale"},
		{"missing", domain.DocHealthMissing, "missing"},
		{"no_frontmatter", domain.DocHealthNoFrontmatter, "no_frontmatter"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

// ---------------------------------------------------------------------------
// 2. DocRegistryEntry
// ---------------------------------------------------------------------------

func TestDocRegistryEntry(t *testing.T) {
	t.Parallel()

	t.Run("defaults", func(t *testing.T) {
		t.Parallel()
		entry, err := domain.NewDocRegistryEntry("docs/PRD.md", "", 30)
		require.NoError(t, err)
		assert.Equal(t, "docs/PRD.md", entry.Path())
		assert.Empty(t, entry.Owner())
		assert.Equal(t, 30, entry.ReviewIntervalDays())
	})

	t.Run("custom values", func(t *testing.T) {
		t.Parallel()
		entry, err := domain.NewDocRegistryEntry("docs/DDD.md", "team-lead", 14)
		require.NoError(t, err)
		assert.Equal(t, "docs/DDD.md", entry.Path())
		assert.Equal(t, "team-lead", entry.Owner())
		assert.Equal(t, 14, entry.ReviewIntervalDays())
	})

	t.Run("rejects zero interval", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewDocRegistryEntry("docs/PRD.md", "", 0)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("rejects negative interval", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewDocRegistryEntry("docs/PRD.md", "", -5)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// 3. BrokenLink
// ---------------------------------------------------------------------------

func TestBrokenLink(t *testing.T) {
	t.Parallel()

	t.Run("valid creation", func(t *testing.T) {
		t.Parallel()
		bl, err := domain.NewBrokenLink(10, "ref", "gone.md", "not found")
		require.NoError(t, err)
		assert.Equal(t, 10, bl.LineNumber())
		assert.Equal(t, "ref", bl.LinkText())
		assert.Equal(t, "gone.md", bl.Target())
		assert.Equal(t, "not found", bl.Reason())
	})

	t.Run("rejects zero line number", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewBrokenLink(0, "x", "x.md", "test")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "line_number must be >= 1")
	})

	t.Run("rejects negative line number", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewBrokenLink(-1, "x", "x.md", "test")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "line_number must be >= 1")
	})
}

// ---------------------------------------------------------------------------
// 4. CreateDocStatus factory
// ---------------------------------------------------------------------------

func TestCreateDocStatus(t *testing.T) {
	t.Parallel()
	today := time.Now().Truncate(24 * time.Hour)

	t.Run("ok within interval", func(t *testing.T) {
		t.Parallel()
		reviewed := today.AddDate(0, 0, -10)
		s := domain.CreateDocStatus("docs/PRD.md", true, &reviewed, 30, "", &today, nil)
		assert.Equal(t, domain.DocHealthOK, s.Status())
		assert.NotNil(t, s.LastReviewed())
		assert.Equal(t, 10, *s.DaysSince())
		assert.Equal(t, "docs/PRD.md", s.Path())
	})

	t.Run("stale beyond interval", func(t *testing.T) {
		t.Parallel()
		reviewed := today.AddDate(0, 0, -45)
		s := domain.CreateDocStatus("docs/PRD.md", true, &reviewed, 30, "", &today, nil)
		assert.Equal(t, domain.DocHealthStale, s.Status())
		assert.Equal(t, 45, *s.DaysSince())
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		s := domain.CreateDocStatus("docs/MISSING.md", false, nil, 30, "", &today, nil)
		assert.Equal(t, domain.DocHealthMissing, s.Status())
		assert.Nil(t, s.LastReviewed())
		assert.Nil(t, s.DaysSince())
	})

	t.Run("no frontmatter", func(t *testing.T) {
		t.Parallel()
		s := domain.CreateDocStatus("docs/NO_FM.md", true, nil, 30, "", &today, nil)
		assert.Equal(t, domain.DocHealthNoFrontmatter, s.Status())
		assert.Nil(t, s.LastReviewed())
		assert.Nil(t, s.DaysSince())
	})

	t.Run("exactly at interval is ok", func(t *testing.T) {
		t.Parallel()
		reviewed := today.AddDate(0, 0, -30)
		s := domain.CreateDocStatus("docs/PRD.md", true, &reviewed, 30, "", &today, nil)
		assert.Equal(t, domain.DocHealthOK, s.Status())
	})

	t.Run("one day beyond interval is stale", func(t *testing.T) {
		t.Parallel()
		reviewed := today.AddDate(0, 0, -31)
		s := domain.CreateDocStatus("docs/PRD.md", true, &reviewed, 30, "", &today, nil)
		assert.Equal(t, domain.DocHealthStale, s.Status())
	})

	t.Run("preserves owner", func(t *testing.T) {
		t.Parallel()
		s := domain.CreateDocStatus("docs/PRD.md", true, &today, 30, "team-lead", &today, nil)
		assert.Equal(t, "team-lead", s.Owner())
	})
}

// ---------------------------------------------------------------------------
// 5. DocHealthReport
// ---------------------------------------------------------------------------

func TestDocHealthReport(t *testing.T) {
	t.Parallel()

	t.Run("issue count", func(t *testing.T) {
		t.Parallel()
		statuses := []domain.DocStatus{
			domain.NewDocStatus("a.md", domain.DocHealthOK, nil, nil, 30, "", nil),
			domain.NewDocStatus("b.md", domain.DocHealthStale, nil, nil, 30, "", nil),
			domain.NewDocStatus("c.md", domain.DocHealthMissing, nil, nil, 30, "", nil),
		}
		report := domain.NewDocHealthReport(statuses)
		assert.Equal(t, 2, report.IssueCount())
	})

	t.Run("total checked", func(t *testing.T) {
		t.Parallel()
		statuses := []domain.DocStatus{
			domain.NewDocStatus("a.md", domain.DocHealthOK, nil, nil, 30, "", nil),
			domain.NewDocStatus("b.md", domain.DocHealthStale, nil, nil, 30, "", nil),
		}
		report := domain.NewDocHealthReport(statuses)
		assert.Equal(t, 2, report.TotalChecked())
	})

	t.Run("has issues", func(t *testing.T) {
		t.Parallel()
		statuses := []domain.DocStatus{
			domain.NewDocStatus("a.md", domain.DocHealthOK, nil, nil, 30, "", nil),
			domain.NewDocStatus("b.md", domain.DocHealthStale, nil, nil, 30, "", nil),
		}
		report := domain.NewDocHealthReport(statuses)
		assert.True(t, report.HasIssues())
	})

	t.Run("no issues", func(t *testing.T) {
		t.Parallel()
		statuses := []domain.DocStatus{
			domain.NewDocStatus("a.md", domain.DocHealthOK, nil, nil, 30, "", nil),
			domain.NewDocStatus("b.md", domain.DocHealthOK, nil, nil, 30, "", nil),
		}
		report := domain.NewDocHealthReport(statuses)
		assert.False(t, report.HasIssues())
	})

	t.Run("empty report", func(t *testing.T) {
		t.Parallel()
		report := domain.NewDocHealthReport(nil)
		assert.Equal(t, 0, report.IssueCount())
		assert.Equal(t, 0, report.TotalChecked())
		assert.False(t, report.HasIssues())
	})

	t.Run("has issues with broken links on OK doc", func(t *testing.T) {
		t.Parallel()
		bl, _ := domain.NewBrokenLink(5, "ref", "gone.md", "not found")
		statuses := []domain.DocStatus{
			domain.NewDocStatus("a.md", domain.DocHealthOK, nil, nil, 30, "", []domain.BrokenLink{bl}),
		}
		report := domain.NewDocHealthReport(statuses)
		assert.True(t, report.HasIssues())
		assert.Equal(t, 1, report.IssueCount())
	})
}

// ---------------------------------------------------------------------------
// 6. DocReviewResult
// ---------------------------------------------------------------------------

func TestDocReviewResult(t *testing.T) {
	t.Parallel()
	now := time.Now().Truncate(24 * time.Hour)
	r := domain.NewDocReviewResult("docs/PRD.md", now)
	assert.Equal(t, "docs/PRD.md", r.Path())
	assert.Equal(t, now, r.NewDate())
}
