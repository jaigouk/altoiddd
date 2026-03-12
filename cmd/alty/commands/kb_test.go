package commands_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/cmd/alty/commands"
	"github.com/alty-cli/alty/internal/composition"
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

func mustSignal(entryPath string, signalType domain.DriftSignalType, desc string, severity domain.DriftSeverity) domain.DriftSignal {
	s, err := domain.NewDriftSignal(entryPath, signalType, desc, severity)
	if err != nil {
		panic(err)
	}
	return s
}

// ---------------------------------------------------------------------------
// Tests — kb drift
// ---------------------------------------------------------------------------

func TestKBDriftCmd_OutputsReport(t *testing.T) {
	t.Parallel()

	signals := []domain.DriftSignal{
		mustSignal("tools/claude-code/v2_1", domain.DriftStale, "Knowledge entry stale: last verified 2025-01-01 (>90 days ago)", domain.SeverityInfo),
	}
	detector := &fakeDriftDetector{report: domain.NewDriftReport(signals)}
	handler := application.NewDriftDetectionHandler(detector)

	app := &composition.App{
		DriftDetectionHandler: handler,
	}

	cmd := commands.NewKBCmd(app)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"drift"})

	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Drift Report")
	assert.Contains(t, output, "claude-code")
	assert.Contains(t, output, "stale")
}

func TestKBDriftCmd_FiltersByTool(t *testing.T) {
	t.Parallel()

	signals := []domain.DriftSignal{
		mustSignal("tools/claude-code/v2_1", domain.DriftStale, "Stale", domain.SeverityInfo),
		mustSignal("tools/cursor/v1_0", domain.DriftStale, "Stale", domain.SeverityWarning),
	}
	detector := &fakeDriftDetector{report: domain.NewDriftReport(signals)}
	handler := application.NewDriftDetectionHandler(detector)

	app := &composition.App{
		DriftDetectionHandler: handler,
	}

	cmd := commands.NewKBCmd(app)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"drift", "claude-code"})

	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "claude-code")
	assert.NotContains(t, output, "cursor")
}

func TestKBDriftCmd_EmptyReport(t *testing.T) {
	t.Parallel()

	detector := &fakeDriftDetector{report: domain.NewDriftReport(nil)}
	handler := application.NewDriftDetectionHandler(detector)

	app := &composition.App{
		DriftDetectionHandler: handler,
	}

	cmd := commands.NewKBCmd(app)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"drift"})

	err := cmd.Execute()

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "No drift detected")
}

func TestKBDriftCmd_ExitCodeOnError(t *testing.T) {
	t.Parallel()

	signals := []domain.DriftSignal{
		mustSignal("tools/test/v1", domain.DriftStale, "Stale", domain.SeverityError),
	}
	detector := &fakeDriftDetector{report: domain.NewDriftReport(signals)}
	handler := application.NewDriftDetectionHandler(detector)

	app := &composition.App{
		DriftDetectionHandler: handler,
	}

	cmd := commands.NewKBCmd(app)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"drift"})

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "drift detected")
}
