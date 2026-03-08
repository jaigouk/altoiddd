package application_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/fitness/application"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Fake gate runner
// ---------------------------------------------------------------------------

type fakeGateRunner struct {
	called    []vo.QualityGate
	failGates map[vo.QualityGate]bool
}

func newFakeGateRunner(failGates ...vo.QualityGate) *fakeGateRunner {
	fg := make(map[vo.QualityGate]bool)
	for _, g := range failGates {
		fg[g] = true
	}
	return &fakeGateRunner{failGates: fg}
}

func (f *fakeGateRunner) Run(_ context.Context, gate vo.QualityGate) (vo.GateResult, error) {
	f.called = append(f.called, gate)
	passed := !f.failGates[gate]
	output := fmt.Sprintf("ok: %s", string(gate))
	if !passed {
		output = fmt.Sprintf("fail: %s", string(gate))
	}
	return vo.NewGateResult(gate, passed, output, 10), nil
}

// ---------------------------------------------------------------------------
// Tests — Run All Gates
// ---------------------------------------------------------------------------

func TestQualityGateHandler_RunAllGates(t *testing.T) {
	t.Parallel()

	t.Run("runs all four gates when none specified", func(t *testing.T) {
		t.Parallel()
		runner := newFakeGateRunner()
		handler := application.NewQualityGateHandler(runner)

		report, err := handler.Check(context.Background(), nil)

		require.NoError(t, err)
		assert.Equal(t, 4, len(runner.called))
		assert.True(t, report.Passed())
	})

	t.Run("returns report with all results", func(t *testing.T) {
		t.Parallel()
		runner := newFakeGateRunner()
		handler := application.NewQualityGateHandler(runner)

		report, err := handler.Check(context.Background(), nil)

		require.NoError(t, err)
		results := report.Results()
		assert.Equal(t, 4, len(results))
		gatesInReport := make(map[vo.QualityGate]bool)
		for _, r := range results {
			gatesInReport[r.Gate()] = true
		}
		assert.True(t, gatesInReport[vo.QualityGateLint])
		assert.True(t, gatesInReport[vo.QualityGateTypes])
		assert.True(t, gatesInReport[vo.QualityGateTests])
		assert.True(t, gatesInReport[vo.QualityGateFitness])
	})
}

// ---------------------------------------------------------------------------
// Tests — Run Specific Gates
// ---------------------------------------------------------------------------

func TestQualityGateHandler_RunSpecificGates(t *testing.T) {
	t.Parallel()

	t.Run("runs only requested gates", func(t *testing.T) {
		t.Parallel()
		runner := newFakeGateRunner()
		handler := application.NewQualityGateHandler(runner)

		report, err := handler.Check(context.Background(), []vo.QualityGate{
			vo.QualityGateLint, vo.QualityGateTypes,
		})

		require.NoError(t, err)
		assert.Equal(t, 2, len(runner.called))
		assert.Equal(t, []vo.QualityGate{vo.QualityGateLint, vo.QualityGateTypes}, runner.called)
		assert.Equal(t, 2, len(report.Results()))
	})

	t.Run("single gate", func(t *testing.T) {
		t.Parallel()
		runner := newFakeGateRunner()
		handler := application.NewQualityGateHandler(runner)

		report, err := handler.Check(context.Background(), []vo.QualityGate{vo.QualityGateTests})

		require.NoError(t, err)
		assert.Equal(t, 1, len(runner.called))
		assert.Equal(t, vo.QualityGateTests, runner.called[0])
		assert.True(t, report.Passed())
	})
}

// ---------------------------------------------------------------------------
// Tests — Continues After Failure
// ---------------------------------------------------------------------------

func TestQualityGateHandler_ContinuesAfterFailure(t *testing.T) {
	t.Parallel()

	t.Run("continues after first gate fails", func(t *testing.T) {
		t.Parallel()
		runner := newFakeGateRunner(vo.QualityGateLint)
		handler := application.NewQualityGateHandler(runner)

		report, err := handler.Check(context.Background(), nil)

		require.NoError(t, err)
		assert.Equal(t, 4, len(runner.called))
		assert.False(t, report.Passed())

		results := report.Results()
		for _, r := range results {
			if r.Gate() == vo.QualityGateLint {
				assert.False(t, r.Passed())
			}
			if r.Gate() == vo.QualityGateTests {
				assert.True(t, r.Passed())
			}
		}
	})

	t.Run("multiple failures still runs all", func(t *testing.T) {
		t.Parallel()
		runner := newFakeGateRunner(vo.QualityGateLint, vo.QualityGateTypes)
		handler := application.NewQualityGateHandler(runner)

		report, err := handler.Check(context.Background(), nil)

		require.NoError(t, err)
		assert.Equal(t, 4, len(runner.called))
		assert.False(t, report.Passed())

		failedGates := make(map[vo.QualityGate]bool)
		for _, r := range report.Results() {
			if !r.Passed() {
				failedGates[r.Gate()] = true
			}
		}
		assert.True(t, failedGates[vo.QualityGateLint])
		assert.True(t, failedGates[vo.QualityGateTypes])
	})
}

// ---------------------------------------------------------------------------
// Tests — Report Correctness
// ---------------------------------------------------------------------------

func TestQualityGateHandler_ReportCorrectness(t *testing.T) {
	t.Parallel()

	t.Run("results match runner output", func(t *testing.T) {
		t.Parallel()
		runner := newFakeGateRunner(vo.QualityGateFitness)
		handler := application.NewQualityGateHandler(runner)

		report, err := handler.Check(context.Background(), []vo.QualityGate{
			vo.QualityGateLint, vo.QualityGateFitness,
		})

		require.NoError(t, err)
		results := report.Results()
		assert.Equal(t, 2, len(results))
		assert.Equal(t, vo.QualityGateLint, results[0].Gate())
		assert.True(t, results[0].Passed())
		assert.Equal(t, vo.QualityGateFitness, results[1].Gate())
		assert.False(t, results[1].Passed())
	})

	t.Run("empty gates returns empty report", func(t *testing.T) {
		t.Parallel()
		runner := newFakeGateRunner()
		handler := application.NewQualityGateHandler(runner)

		report, err := handler.Check(context.Background(), []vo.QualityGate{})

		require.NoError(t, err)
		assert.Equal(t, 0, len(report.Results()))
		assert.True(t, report.Passed())
		assert.Equal(t, 0, len(runner.called))
	})
}
