package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/research/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// TrustLevel
// ---------------------------------------------------------------------------

func TestTrustLevel(t *testing.T) {
	t.Parallel()

	t.Run("has four members", func(t *testing.T) {
		t.Parallel()
		all := domain.AllTrustLevels()
		assert.Len(t, all, 4)
	})

	t.Run("ordering lower value is higher trust", func(t *testing.T) {
		t.Parallel()
		assert.Less(t, domain.TrustUserStated, domain.TrustUserConfirmed)
		assert.Less(t, domain.TrustUserConfirmed, domain.TrustAIResearched)
		assert.Less(t, domain.TrustAIResearched, domain.TrustAIInferred)
	})

	t.Run("values", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 1, int(domain.TrustUserStated))
		assert.Equal(t, 2, int(domain.TrustUserConfirmed))
		assert.Equal(t, 3, int(domain.TrustAIResearched))
		assert.Equal(t, 4, int(domain.TrustAIInferred))
	})
}

// ---------------------------------------------------------------------------
// Confidence
// ---------------------------------------------------------------------------

func TestConfidence(t *testing.T) {
	t.Parallel()

	t.Run("has three members", func(t *testing.T) {
		t.Parallel()
		all := domain.AllConfidenceLevels()
		assert.Len(t, all, 3)
	})

	t.Run("values", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "high", string(domain.ConfidenceHigh))
		assert.Equal(t, "medium", string(domain.ConfidenceMedium))
		assert.Equal(t, "low", string(domain.ConfidenceLow))
	})
}

// ---------------------------------------------------------------------------
// SourceAttribution
// ---------------------------------------------------------------------------

func TestSourceAttribution(t *testing.T) {
	t.Parallel()

	t.Run("valid construction", func(t *testing.T) {
		t.Parallel()
		sa, err := domain.NewSourceAttribution(
			"https://example.com", "Example", "2026-03-06", domain.ConfidenceMedium,
		)
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", sa.URL())
		assert.Equal(t, "Example", sa.Title())
		assert.Equal(t, "2026-03-06", sa.RetrievedDate())
		assert.Equal(t, domain.ConfidenceMedium, sa.Confidence())
	})

	t.Run("requires url", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewSourceAttribution("", "Example", "2026-03-06", domain.ConfidenceHigh)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "url")
	})

	t.Run("requires title", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewSourceAttribution("https://example.com", "", "2026-03-06", domain.ConfidenceHigh)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "title")
	})

	t.Run("whitespace only url rejected", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewSourceAttribution("   ", "Example", "2026-03-06", domain.ConfidenceHigh)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
		assert.Contains(t, err.Error(), "url")
	})
}

// ---------------------------------------------------------------------------
// WebSearchResult
// ---------------------------------------------------------------------------

func TestWebSearchResult(t *testing.T) {
	t.Parallel()

	t.Run("captures all fields", func(t *testing.T) {
		t.Parallel()
		wsr := domain.NewWebSearchResult("https://example.com", "Search Result", "Some snippet text")
		assert.Equal(t, "https://example.com", wsr.URL())
		assert.Equal(t, "Search Result", wsr.Title())
		assert.Equal(t, "Some snippet text", wsr.Snippet())
	})
}

// ---------------------------------------------------------------------------
// ResearchFinding
// ---------------------------------------------------------------------------

func TestResearchFinding(t *testing.T) {
	t.Parallel()

	makeSource := func(t *testing.T) domain.SourceAttribution {
		t.Helper()
		sa, err := domain.NewSourceAttribution(
			"https://example.com", "Source", "2026-03-06", domain.ConfidenceMedium,
		)
		require.NoError(t, err)
		return sa
	}

	t.Run("carries trust level and source", func(t *testing.T) {
		t.Parallel()
		source := makeSource(t)
		f := domain.NewResearchFinding("Industry pattern", source, domain.TrustAIResearched, "Sales", false)
		assert.Equal(t, domain.TrustAIResearched, f.TrustLevel())
		assert.Equal(t, source, f.Source())
		assert.Equal(t, "Sales", f.DomainArea())
	})

	t.Run("outdated defaults false", func(t *testing.T) {
		t.Parallel()
		f := domain.NewResearchFinding("Finding", makeSource(t), domain.TrustAIInferred, "Marketing", false)
		assert.False(t, f.Outdated())
	})
}

// ---------------------------------------------------------------------------
// ResearchBriefing
// ---------------------------------------------------------------------------

func TestResearchBriefing(t *testing.T) {
	t.Parallel()

	makeFinding := func(t *testing.T, area string) domain.ResearchFinding {
		t.Helper()
		sa, _ := domain.NewSourceAttribution(
			"https://example.com", "Source", "2026-03-06", domain.ConfidenceMedium,
		)
		return domain.NewResearchFinding("Some finding", sa, domain.TrustAIResearched, area, false)
	}

	t.Run("separates findings from no data", func(t *testing.T) {
		t.Parallel()
		finding := makeFinding(t, "Sales")
		briefing := domain.NewResearchBriefing(
			[]domain.ResearchFinding{finding},
			[]string{"Marketing"},
			"Partial research",
		)
		assert.Len(t, briefing.Findings(), 1)
		assert.Equal(t, "Sales", briefing.Findings()[0].DomainArea())
		assert.Equal(t, []string{"Marketing"}, briefing.NoDataAreas())
	})

	t.Run("empty briefing", func(t *testing.T) {
		t.Parallel()
		briefing := domain.NewResearchBriefing(nil, []string{"Sales", "Marketing"}, "")
		assert.Empty(t, briefing.Findings())
		assert.Len(t, briefing.NoDataAreas(), 2)
		assert.Empty(t, briefing.Summary())
	})
}
