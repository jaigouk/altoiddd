package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/rescue/domain"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeScan() domain.ProjectScan {
	return domain.NewProjectScan("/tmp/proj", []string{"docs/PRD.md"}, nil, nil,
		false, false, true, false, false)
}

func makeGap() domain.Gap {
	return domain.NewGap("gap-1", domain.GapTypeMissingDoc, "docs/DDD.md",
		"Missing documentation: docs/DDD.md", domain.GapSeverityRequired)
}

func makePlan(gaps []domain.Gap) domain.MigrationPlan {
	if gaps == nil {
		gaps = []domain.Gap{makeGap()}
	}
	return domain.NewMigrationPlan("plan-1", gaps, "", false)
}

// ---------------------------------------------------------------------------
// Enums
// ---------------------------------------------------------------------------

func TestGapTypeValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "missing_doc", string(domain.GapTypeMissingDoc))
	assert.Equal(t, "missing_config", string(domain.GapTypeMissingConfig))
	assert.Equal(t, "missing_structure", string(domain.GapTypeMissingStructure))
	assert.Equal(t, "missing_tooling", string(domain.GapTypeMissingTooling))
	assert.Equal(t, "missing_knowledge", string(domain.GapTypeMissingKnowledge))
	assert.Equal(t, "conflict", string(domain.GapTypeConflict))
}

func TestAnalysisStatusValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "scanning", string(domain.AnalysisStatusScanning))
	assert.Equal(t, "analyzed", string(domain.AnalysisStatusAnalyzed))
	assert.Equal(t, "planned", string(domain.AnalysisStatusPlanned))
	assert.Equal(t, "executing", string(domain.AnalysisStatusExecuting))
	assert.Equal(t, "completed", string(domain.AnalysisStatusCompleted))
	assert.Equal(t, "failed", string(domain.AnalysisStatusFailed))
}

func TestGapSeverityValues(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "required", string(domain.GapSeverityRequired))
	assert.Equal(t, "recommended", string(domain.GapSeverityRecommended))
	assert.Equal(t, "optional", string(domain.GapSeverityOptional))
}

func TestAllGapSeveritiesReturnsAllConstants(t *testing.T) {
	t.Parallel()
	all := domain.AllGapSeverities()
	assert.Len(t, all, 3)
	assert.Contains(t, all, domain.GapSeverityRequired)
	assert.Contains(t, all, domain.GapSeverityRecommended)
	assert.Contains(t, all, domain.GapSeverityOptional)
}

// ---------------------------------------------------------------------------
// Creation
// ---------------------------------------------------------------------------

func TestGapAnalysisCreation(t *testing.T) {
	t.Parallel()

	t.Run("initial state is scanning", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		assert.Equal(t, domain.AnalysisStatusScanning, a.Status())
	})

	t.Run("has unique id", func(t *testing.T) {
		t.Parallel()
		a1 := domain.NewGapAnalysis("/tmp/a")
		a2 := domain.NewGapAnalysis("/tmp/b")
		assert.NotEqual(t, a1.AnalysisID(), a2.AnalysisID())
	})

	t.Run("stores project dir", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		assert.Equal(t, "/tmp/proj", a.ProjectDir())
	})

	t.Run("initial gaps empty", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		assert.Empty(t, a.Gaps())
	})

	t.Run("initial scan nil", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		assert.Nil(t, a.Scan())
	})

	t.Run("initial plan nil", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		assert.Nil(t, a.Plan())
	})

	t.Run("initial events empty", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		assert.Empty(t, a.Events())
	})
}

// ---------------------------------------------------------------------------
// Set Scan
// ---------------------------------------------------------------------------

func TestGapAnalysisSetScan(t *testing.T) {
	t.Parallel()

	t.Run("set scan from scanning", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		scan := makeScan()
		err := a.SetScan(scan)
		require.NoError(t, err)
		require.NotNil(t, a.Scan())
		assert.Equal(t, scan, *a.Scan())
		assert.Equal(t, domain.AnalysisStatusScanning, a.Status())
	})

	t.Run("set scan wrong state raises", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		_ = a.Analyze([]domain.Gap{makeGap()})
		err := a.SetScan(makeScan())
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// Analyze
// ---------------------------------------------------------------------------

func TestGapAnalysisAnalyze(t *testing.T) {
	t.Parallel()

	t.Run("analyze sets gaps", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		gap := makeGap()
		err := a.Analyze([]domain.Gap{gap})
		require.NoError(t, err)
		assert.Len(t, a.Gaps(), 1)
		assert.Equal(t, domain.AnalysisStatusAnalyzed, a.Status())
	})

	t.Run("analyze without scan raises", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		err := a.Analyze([]domain.Gap{makeGap()})
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("analyze wrong state raises", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		_ = a.Analyze([]domain.Gap{makeGap()})
		err := a.Analyze([]domain.Gap{makeGap()})
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("empty gaps transitions to analyzed", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		err := a.Analyze(nil)
		require.NoError(t, err)
		assert.Equal(t, domain.AnalysisStatusAnalyzed, a.Status())
		assert.Empty(t, a.Gaps())
	})
}

// ---------------------------------------------------------------------------
// Create Plan
// ---------------------------------------------------------------------------

func TestGapAnalysisCreatePlan(t *testing.T) {
	t.Parallel()

	t.Run("create plan from analyzed", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		_ = a.Analyze([]domain.Gap{makeGap()})
		plan := makePlan(nil)
		err := a.CreatePlan(plan)
		require.NoError(t, err)
		require.NotNil(t, a.Plan())
		assert.Equal(t, domain.AnalysisStatusPlanned, a.Status())
	})

	t.Run("create plan wrong state raises", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		err := a.CreatePlan(makePlan(nil))
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// Begin Execution
// ---------------------------------------------------------------------------

func TestGapAnalysisBeginExecution(t *testing.T) {
	t.Parallel()

	t.Run("begin execution from planned", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		_ = a.Analyze([]domain.Gap{makeGap()})
		_ = a.CreatePlan(makePlan(nil))
		err := a.BeginExecution()
		require.NoError(t, err)
		assert.Equal(t, domain.AnalysisStatusExecuting, a.Status())
	})

	t.Run("begin execution wrong state raises", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		err := a.BeginExecution()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// Complete
// ---------------------------------------------------------------------------

func TestGapAnalysisComplete(t *testing.T) {
	t.Parallel()

	t.Run("complete emits event", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		gap := makeGap()
		_ = a.Analyze([]domain.Gap{gap})
		_ = a.CreatePlan(makePlan([]domain.Gap{gap}))
		_ = a.BeginExecution()
		err := a.Complete()
		require.NoError(t, err)
		assert.Equal(t, domain.AnalysisStatusCompleted, a.Status())
		assert.Len(t, a.Events(), 1)
		event := a.Events()[0]
		assert.Equal(t, a.AnalysisID(), event.AnalysisID())
		assert.Equal(t, "/tmp/proj", event.ProjectDir())
		assert.Equal(t, 1, event.GapsFound())
		assert.Equal(t, 1, event.GapsResolved())
	})

	t.Run("complete wrong state raises", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		err := a.Complete()
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})

	t.Run("events returns defensive copy", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		_ = a.Analyze([]domain.Gap{makeGap()})
		_ = a.CreatePlan(makePlan(nil))
		_ = a.BeginExecution()
		_ = a.Complete()
		events := a.Events()
		assert.Len(t, events, 1)
		// Modifying returned slice shouldn't affect aggregate
		_ = events[:0]
		assert.Len(t, a.Events(), 1)
	})
}

// ---------------------------------------------------------------------------
// Fail
// ---------------------------------------------------------------------------

func TestGapAnalysisFail(t *testing.T) {
	t.Parallel()

	t.Run("fail from executing", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		_ = a.Analyze([]domain.Gap{makeGap()})
		_ = a.CreatePlan(makePlan(nil))
		_ = a.BeginExecution()
		err := a.Fail("Something went wrong")
		require.NoError(t, err)
		assert.Equal(t, domain.AnalysisStatusFailed, a.Status())
	})

	t.Run("fail stores reason", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		_ = a.SetScan(makeScan())
		_ = a.Analyze([]domain.Gap{makeGap()})
		_ = a.CreatePlan(makePlan(nil))
		_ = a.BeginExecution()
		err := a.Fail("disk full")
		require.NoError(t, err)
		assert.Equal(t, "disk full", a.FailureReason())
	})

	t.Run("failure reason empty before fail", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		assert.Empty(t, a.FailureReason())
	})

	t.Run("fail wrong state raises", func(t *testing.T) {
		t.Parallel()
		a := domain.NewGapAnalysis("/tmp/proj")
		err := a.Fail("reason")
		require.ErrorIs(t, err, domainerrors.ErrInvariantViolation)
	})
}

// ---------------------------------------------------------------------------
// Value Objects
// ---------------------------------------------------------------------------

func TestProjectScanFields(t *testing.T) {
	t.Parallel()

	t.Run("has knowledge dir", func(t *testing.T) {
		t.Parallel()
		scan := domain.NewProjectScan("/tmp/proj", nil, nil, nil, false, false, true, false, false)
		assert.False(t, scan.HasKnowledgeDir())
	})

	t.Run("has agents md", func(t *testing.T) {
		t.Parallel()
		scan := domain.NewProjectScan("/tmp/proj", nil, nil, nil, false, true, true, false, false)
		assert.True(t, scan.HasAgentsMD())
	})
}

func TestMigrationPlanFields(t *testing.T) {
	t.Parallel()

	t.Run("skip agents md", func(t *testing.T) {
		t.Parallel()
		plan := domain.NewMigrationPlan("plan-1", []domain.Gap{makeGap()}, "", true)
		assert.True(t, plan.SkipAgentsMD())
	})

	t.Run("default branch name", func(t *testing.T) {
		t.Parallel()
		plan := domain.NewMigrationPlan("plan-1", []domain.Gap{makeGap()}, "", false)
		assert.Equal(t, "alty/init", plan.BranchName())
	})

	t.Run("custom branch name", func(t *testing.T) {
		t.Parallel()
		plan := domain.NewMigrationPlan("plan-1", []domain.Gap{makeGap()}, "custom/branch", false)
		assert.Equal(t, "custom/branch", plan.BranchName())
	})
}

func TestGapConflictType(t *testing.T) {
	t.Parallel()
	gap := domain.NewGap("gap-conflict", domain.GapTypeConflict, ".claude/CLAUDE.md",
		"Conflicting config", domain.GapSeverityRequired)
	assert.Equal(t, domain.GapTypeConflict, gap.GapType())
}

// ---------------------------------------------------------------------------
// IsDirectory
// ---------------------------------------------------------------------------

func TestGap_IsDirectory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		wantDir bool
	}{
		{"trailing slash is directory", ".alty/knowledge/", true},
		{"no trailing slash is file", "docs/PRD.md", false},
		{"root slash is directory", "/", true},
		{"nested trailing slash", "src/domain/", true},
		{"empty path is not directory", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gap := domain.NewGap("gap-1", domain.GapTypeMissingStructure, tt.path,
				"test gap", domain.GapSeverityRequired)
			assert.Equal(t, tt.wantDir, gap.IsDirectory())
		})
	}
}

// ---------------------------------------------------------------------------
// Uses shared GapAnalysisCompleted event
// ---------------------------------------------------------------------------

func TestGapAnalysisUsesSharedEvent(t *testing.T) {
	t.Parallel()
	a := domain.NewGapAnalysis("/tmp/proj")
	_ = a.SetScan(makeScan())
	_ = a.Analyze([]domain.Gap{makeGap()})
	_ = a.CreatePlan(makePlan(nil))
	_ = a.BeginExecution()
	_ = a.Complete()
	event := a.Events()[0]
	assert.Equal(t, a.AnalysisID(), event.AnalysisID())
}
