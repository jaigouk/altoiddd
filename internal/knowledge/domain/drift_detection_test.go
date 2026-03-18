package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alto-cli/alto/internal/knowledge/domain"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// DriftSignalType
// ---------------------------------------------------------------------------

func TestDriftSignalType(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "version_change", string(domain.DriftVersionChange))
	assert.Equal(t, "doc_code_mismatch", string(domain.DriftDocCodeMismatch))
	assert.Equal(t, "stale", string(domain.DriftStale))
}

// ---------------------------------------------------------------------------
// DriftSeverity
// ---------------------------------------------------------------------------

func TestDriftSeverity(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "info", string(domain.SeverityInfo))
	assert.Equal(t, "warning", string(domain.SeverityWarning))
	assert.Equal(t, "error", string(domain.SeverityError))
}

// ---------------------------------------------------------------------------
// DriftSignal
// ---------------------------------------------------------------------------

func TestDriftSignal(t *testing.T) {
	t.Parallel()

	t.Run("creates with all fields", func(t *testing.T) {
		t.Parallel()
		s, err := domain.NewDriftSignal(
			"tools/claude-code/config-structure",
			domain.DriftVersionChange,
			"Key 'rules/*.md' added in current but missing in v2.0",
			domain.SeverityWarning,
		)
		require.NoError(t, err)
		assert.Equal(t, "tools/claude-code/config-structure", s.EntryPath())
		assert.Equal(t, domain.DriftVersionChange, s.SignalType())
		assert.Equal(t, domain.SeverityWarning, s.Severity())
	})

	t.Run("empty entry path raises", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewDriftSignal("", domain.DriftVersionChange, "Something changed", domain.SeverityWarning)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("whitespace entry path raises", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewDriftSignal("   ", domain.DriftVersionChange, "Something changed", domain.SeverityWarning)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("empty description raises", func(t *testing.T) {
		t.Parallel()
		_, err := domain.NewDriftSignal("tools/cursor/rules-format", domain.DriftStale, "", domain.SeverityInfo)
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("equality by value", func(t *testing.T) {
		t.Parallel()
		a, _ := domain.NewDriftSignal("tools/cursor/rules-format", domain.DriftStale, "Stale entry", domain.SeverityInfo)
		b, _ := domain.NewDriftSignal("tools/cursor/rules-format", domain.DriftStale, "Stale entry", domain.SeverityInfo)
		assert.Equal(t, a, b)
	})

	t.Run("different signals not equal", func(t *testing.T) {
		t.Parallel()
		a, _ := domain.NewDriftSignal("tools/cursor/rules-format", domain.DriftStale, "Stale entry", domain.SeverityInfo)
		b, _ := domain.NewDriftSignal("tools/cursor/rules-format", domain.DriftVersionChange, "Changed entry", domain.SeverityWarning)
		assert.NotEqual(t, a, b)
	})
}

// ---------------------------------------------------------------------------
// DriftReport
// ---------------------------------------------------------------------------

func TestDriftReport(t *testing.T) {
	t.Parallel()

	t.Run("creates with signals", func(t *testing.T) {
		t.Parallel()
		s, _ := domain.NewDriftSignal(
			"tools/claude-code/config-structure",
			domain.DriftVersionChange,
			"Key added",
			domain.SeverityWarning,
		)
		report := domain.NewDriftReport([]domain.DriftSignal{s})
		assert.Equal(t, 1, report.TotalCount())
	})

	t.Run("empty report", func(t *testing.T) {
		t.Parallel()
		report := domain.NewDriftReport(nil)
		assert.Equal(t, 0, report.TotalCount())
		assert.False(t, report.HasDrift())
	})

	t.Run("has drift true", func(t *testing.T) {
		t.Parallel()
		s, _ := domain.NewDriftSignal("tools/cursor/rules-format", domain.DriftStale, "Stale", domain.SeverityInfo)
		report := domain.NewDriftReport([]domain.DriftSignal{s})
		assert.True(t, report.HasDrift())
	})

	t.Run("count by severity", func(t *testing.T) {
		t.Parallel()
		s1, _ := domain.NewDriftSignal("a", domain.DriftStale, "Stale 1", domain.SeverityInfo)
		s2, _ := domain.NewDriftSignal("b", domain.DriftVersionChange, "Changed", domain.SeverityWarning)
		s3, _ := domain.NewDriftSignal("c", domain.DriftDocCodeMismatch, "Mismatch", domain.SeverityError)
		s4, _ := domain.NewDriftSignal("d", domain.DriftStale, "Stale 2", domain.SeverityInfo)
		report := domain.NewDriftReport([]domain.DriftSignal{s1, s2, s3, s4})
		assert.Equal(t, 2, report.CountBySeverity(domain.SeverityInfo))
		assert.Equal(t, 1, report.CountBySeverity(domain.SeverityWarning))
		assert.Equal(t, 1, report.CountBySeverity(domain.SeverityError))
	})

	t.Run("count by type", func(t *testing.T) {
		t.Parallel()
		s1, _ := domain.NewDriftSignal("a", domain.DriftStale, "Stale", domain.SeverityInfo)
		s2, _ := domain.NewDriftSignal("b", domain.DriftVersionChange, "Changed 1", domain.SeverityWarning)
		s3, _ := domain.NewDriftSignal("c", domain.DriftVersionChange, "Changed 2", domain.SeverityWarning)
		report := domain.NewDriftReport([]domain.DriftSignal{s1, s2, s3})
		assert.Equal(t, 2, report.CountByType(domain.DriftVersionChange))
		assert.Equal(t, 1, report.CountByType(domain.DriftStale))
		assert.Equal(t, 0, report.CountByType(domain.DriftDocCodeMismatch))
	})
}
