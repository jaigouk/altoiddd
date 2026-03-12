package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/knowledge/application"
	"github.com/alty-cli/alty/internal/knowledge/domain"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakeDriftDetector struct {
	report domain.DriftReport
	err    error
}

func (f *fakeDriftDetector) Detect(_ context.Context) (domain.DriftReport, error) {
	return f.report, f.err
}

func newFakeDriftDetector(signals []domain.DriftSignal) *fakeDriftDetector {
	return &fakeDriftDetector{report: domain.NewDriftReport(signals)}
}

func mustSignal(entryPath string, signalType domain.DriftSignalType, desc string, severity domain.DriftSeverity) domain.DriftSignal {
	s, err := domain.NewDriftSignal(entryPath, signalType, desc, severity)
	if err != nil {
		panic(err)
	}
	return s
}

// ---------------------------------------------------------------------------
// Tests — DetectDrift
// ---------------------------------------------------------------------------

func TestDriftDetectionHandler_DetectDrift(t *testing.T) {
	t.Parallel()

	t.Run("returns full report when no filter", func(t *testing.T) {
		t.Parallel()
		signals := []domain.DriftSignal{
			mustSignal("tools/claude-code/config", domain.DriftVersionChange, "Changed", domain.SeverityWarning),
			mustSignal("tools/cursor/rules", domain.DriftStale, "Stale", domain.SeverityInfo),
		}
		detector := newFakeDriftDetector(signals)
		handler := application.NewDriftDetectionHandler(detector)

		report, err := handler.DetectDrift(context.Background(), nil)

		require.NoError(t, err)
		assert.Equal(t, 2, report.TotalCount())
	})

	t.Run("returns full report when empty string filter", func(t *testing.T) {
		t.Parallel()
		signals := []domain.DriftSignal{
			mustSignal("tools/claude-code/config", domain.DriftVersionChange, "Changed", domain.SeverityWarning),
			mustSignal("tools/cursor/rules", domain.DriftStale, "Stale", domain.SeverityInfo),
		}
		detector := newFakeDriftDetector(signals)
		handler := application.NewDriftDetectionHandler(detector)

		emptyFilter := ""
		report, err := handler.DetectDrift(context.Background(), &emptyFilter)

		require.NoError(t, err)
		assert.Equal(t, 2, report.TotalCount())
	})

	t.Run("filters by tool when filter provided", func(t *testing.T) {
		t.Parallel()
		signals := []domain.DriftSignal{
			mustSignal("tools/claude-code/config", domain.DriftVersionChange, "Changed", domain.SeverityWarning),
			mustSignal("tools/cursor/rules", domain.DriftStale, "Stale", domain.SeverityInfo),
			mustSignal("tools/claude-code/commands", domain.DriftDocCodeMismatch, "Mismatch", domain.SeverityError),
		}
		detector := newFakeDriftDetector(signals)
		handler := application.NewDriftDetectionHandler(detector)

		filter := "claude-code"
		report, err := handler.DetectDrift(context.Background(), &filter)

		require.NoError(t, err)
		assert.Equal(t, 2, report.TotalCount())
		for _, sig := range report.Signals() {
			assert.Contains(t, sig.EntryPath(), "claude-code")
		}
	})

	t.Run("returns empty report when tool not found", func(t *testing.T) {
		t.Parallel()
		signals := []domain.DriftSignal{
			mustSignal("tools/claude-code/config", domain.DriftVersionChange, "Changed", domain.SeverityWarning),
		}
		detector := newFakeDriftDetector(signals)
		handler := application.NewDriftDetectionHandler(detector)

		filter := "nonexistent-tool"
		report, err := handler.DetectDrift(context.Background(), &filter)

		require.NoError(t, err)
		assert.Equal(t, 0, report.TotalCount())
		assert.False(t, report.HasDrift())
	})

	t.Run("case insensitive tool matching", func(t *testing.T) {
		t.Parallel()
		signals := []domain.DriftSignal{
			mustSignal("tools/claude-code/config", domain.DriftVersionChange, "Changed", domain.SeverityWarning),
		}
		detector := newFakeDriftDetector(signals)
		handler := application.NewDriftDetectionHandler(detector)

		filter := "Claude-Code" // Different case
		report, err := handler.DetectDrift(context.Background(), &filter)

		require.NoError(t, err)
		assert.Equal(t, 1, report.TotalCount())
	})

	t.Run("non-tool entries excluded when filtering", func(t *testing.T) {
		t.Parallel()
		signals := []domain.DriftSignal{
			mustSignal("tools/claude-code/config", domain.DriftVersionChange, "Changed", domain.SeverityWarning),
			mustSignal("ddd/aggregate", domain.DriftStale, "Stale", domain.SeverityInfo),
			mustSignal("conventions/naming", domain.DriftDocCodeMismatch, "Mismatch", domain.SeverityWarning),
		}
		detector := newFakeDriftDetector(signals)
		handler := application.NewDriftDetectionHandler(detector)

		filter := "claude-code"
		report, err := handler.DetectDrift(context.Background(), &filter)

		require.NoError(t, err)
		assert.Equal(t, 1, report.TotalCount())
		assert.Equal(t, "tools/claude-code/config", report.Signals()[0].EntryPath())
	})

	t.Run("propagates port error", func(t *testing.T) {
		t.Parallel()
		detector := &fakeDriftDetector{err: assert.AnError}
		handler := application.NewDriftDetectionHandler(detector)

		_, err := handler.DetectDrift(context.Background(), nil)

		require.Error(t, err)
		require.ErrorIs(t, err, assert.AnError)
		assert.Contains(t, err.Error(), "detect drift")
	})

	t.Run("empty report from port returns empty report", func(t *testing.T) {
		t.Parallel()
		detector := newFakeDriftDetector(nil) // empty
		handler := application.NewDriftDetectionHandler(detector)

		report, err := handler.DetectDrift(context.Background(), nil)

		require.NoError(t, err)
		assert.Equal(t, 0, report.TotalCount())
		assert.False(t, report.HasDrift())
	})
}
