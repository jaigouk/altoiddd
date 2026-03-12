package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/rescue/application"
	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakeProjectScan struct {
	scan rescuedomain.ProjectScan
}

func newFakeScanner(scan *rescuedomain.ProjectScan) *fakeProjectScan {
	if scan != nil {
		return &fakeProjectScan{scan: *scan}
	}
	return &fakeProjectScan{
		scan: rescuedomain.NewProjectScan(
			"/tmp/proj",
			nil, nil, nil,
			false, false, true, false, false,
		),
	}
}

func (f *fakeProjectScan) Scan(_ context.Context, _ string, _ vo.StackProfile) (rescuedomain.ProjectScan, error) {
	return f.scan, nil
}

type fakeGitOps struct {
	hasGit             bool
	isClean            bool
	branchExists       bool
	createdBranches    []string
	checkoutPrevCalled bool
	deleteBranchCalled bool
	deletedBranches    []string
	checkoutPrevErr    error
	deleteBranchErr    error
}

func newFakeGitOps(hasGit, isClean, branchExists bool) *fakeGitOps {
	return &fakeGitOps{hasGit: hasGit, isClean: isClean, branchExists: branchExists}
}

func defaultFakeGitOps() *fakeGitOps {
	return newFakeGitOps(true, true, false)
}

func (f *fakeGitOps) HasGit(_ context.Context, _ string) (bool, error)  { return f.hasGit, nil }
func (f *fakeGitOps) IsClean(_ context.Context, _ string) (bool, error) { return f.isClean, nil }
func (f *fakeGitOps) BranchExists(_ context.Context, _ string, _ string) (bool, error) {
	return f.branchExists, nil
}

func (f *fakeGitOps) CreateBranch(_ context.Context, _ string, branchName string) error {
	f.createdBranches = append(f.createdBranches, branchName)
	return nil
}

func (f *fakeGitOps) CheckoutPrevious(_ context.Context, _ string) error {
	f.checkoutPrevCalled = true
	return f.checkoutPrevErr
}

func (f *fakeGitOps) DeleteBranch(_ context.Context, _ string, branchName string) error {
	f.deleteBranchCalled = true
	f.deletedBranches = append(f.deletedBranches, branchName)
	return f.deleteBranchErr
}

type fakeTestRunner struct {
	framework    string
	detectErr    error
	runErr       error
	detectCalled bool
	runCalled    bool
}

func newFakeTestRunner(framework string) *fakeTestRunner {
	return &fakeTestRunner{framework: framework}
}

func (f *fakeTestRunner) Detect(_ context.Context, _ string) (string, error) {
	f.detectCalled = true
	return f.framework, f.detectErr
}

func (f *fakeTestRunner) Run(_ context.Context, _ string, _ string) error {
	f.runCalled = true
	return f.runErr
}

type fakePublisherR struct {
	published []any
}

func (f *fakePublisherR) Publish(_ context.Context, event any) error {
	f.published = append(f.published, event)
	return nil
}

type fakeFileWriter struct {
	writtenFiles map[string]string
}

func newFakeFileWriter() *fakeFileWriter {
	return &fakeFileWriter{writtenFiles: make(map[string]string)}
}

func (f *fakeFileWriter) WriteFile(_ context.Context, path string, content string) error {
	f.writtenFiles[path] = content
	return nil
}

// ---------------------------------------------------------------------------
// Tests — Validate Preconditions
// ---------------------------------------------------------------------------

func TestRescueHandler_ValidatePreconditions(t *testing.T) {
	t.Parallel()

	t.Run("raises if not git repo", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), newFakeGitOps(false, true, false), nil, &fakePublisherR{}, nil)
		err := handler.ValidatePreconditions(context.Background(), "/tmp/proj")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a git repository")
	})

	t.Run("raises on dirty tree", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), newFakeGitOps(true, false, false), nil, &fakePublisherR{}, nil)
		err := handler.ValidatePreconditions(context.Background(), "/tmp/proj")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dirty")
	})

	t.Run("raises if branch exists", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), newFakeGitOps(true, true, true), nil, &fakePublisherR{}, nil)
		err := handler.ValidatePreconditions(context.Background(), "/tmp/proj")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "alty/init")
	})

	t.Run("passes for clean repo", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		err := handler.ValidatePreconditions(context.Background(), "/tmp/proj")
		require.NoError(t, err)
	})
}

// ---------------------------------------------------------------------------
// Tests — Happy Path
// ---------------------------------------------------------------------------

func TestRescueHandler_HappyPath(t *testing.T) {
	t.Parallel()

	t.Run("returns gap analysis in planned state", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, err := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		require.NoError(t, err)
		assert.Equal(t, rescuedomain.AnalysisStatusPlanned, analysis.Status())
		assert.NotNil(t, analysis.Scan())
		assert.NotNil(t, analysis.Plan())
		assert.NotEmpty(t, analysis.Gaps())
	})

	t.Run("creates branch before scanning", func(t *testing.T) {
		t.Parallel()
		gitOps := defaultFakeGitOps()
		handler := application.NewRescueHandler(newFakeScanner(nil), gitOps, nil, &fakePublisherR{}, nil)
		handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		assert.Contains(t, gitOps.createdBranches, "alty/init")
	})

	t.Run("detects missing docs", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		gapPaths := make([]string, 0)
		for _, g := range analysis.Gaps() {
			gapPaths = append(gapPaths, g.Path())
		}
		assert.Contains(t, gapPaths, "docs/PRD.md")
		assert.Contains(t, gapPaths, "docs/DDD.md")
		assert.Contains(t, gapPaths, "docs/ARCHITECTURE.md")
	})

	t.Run("detects missing knowledge dir", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		var knowledgeGaps []rescuedomain.Gap
		for _, g := range analysis.Gaps() {
			if g.GapType() == rescuedomain.GapTypeMissingKnowledge {
				knowledgeGaps = append(knowledgeGaps, g)
			}
		}
		assert.Len(t, knowledgeGaps, 1)
		assert.Equal(t, ".alty/knowledge/", knowledgeGaps[0].Path())
	})

	t.Run("all artifacts present returns analyzed with no gaps", func(t *testing.T) {
		t.Parallel()
		scan := rescuedomain.NewProjectScan(
			"/tmp/proj",
			[]string{"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"},
			[]string{".claude/CLAUDE.md", "pyproject.toml"},
			[]string{"src/domain/", "src/application/", "src/infrastructure/"},
			true, true, true, true, true,
		)
		handler := application.NewRescueHandler(newFakeScanner(&scan), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, err := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		require.NoError(t, err)
		assert.Equal(t, rescuedomain.AnalysisStatusAnalyzed, analysis.Status())
		assert.Empty(t, analysis.Gaps())
		assert.Nil(t, analysis.Plan())
	})

	t.Run("with profile detects missing config", func(t *testing.T) {
		t.Parallel()
		profile := vo.PythonUvProfile{}
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", profile, false)
		var configPaths []string
		for _, g := range analysis.Gaps() {
			if g.GapType() == rescuedomain.GapTypeMissingConfig {
				configPaths = append(configPaths, g.Path())
			}
		}
		assert.Contains(t, configPaths, ".claude/CLAUDE.md")
		assert.Contains(t, configPaths, "pyproject.toml")
	})

	t.Run("with profile detects missing structure", func(t *testing.T) {
		t.Parallel()
		profile := vo.PythonUvProfile{}
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", profile, false)
		var structurePaths []string
		for _, g := range analysis.Gaps() {
			if g.GapType() == rescuedomain.GapTypeMissingStructure {
				structurePaths = append(structurePaths, g.Path())
			}
		}
		assert.Contains(t, structurePaths, "src/domain/")
		assert.Contains(t, structurePaths, "src/application/")
		assert.Contains(t, structurePaths, "src/infrastructure/")
	})

	t.Run("none profile no structure gaps", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		for _, g := range analysis.Gaps() {
			if g.GapType() == rescuedomain.GapTypeMissingStructure {
				// Only .alty/ structure gaps allowed without profile
				assert.Contains(t, g.Path(), ".alty/")
			}
		}
	})

	t.Run("none profile no pyproject gap", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		for _, g := range analysis.Gaps() {
			if g.GapType() == rescuedomain.GapTypeMissingConfig {
				assert.NotEqual(t, "pyproject.toml", g.Path())
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Tests — Validated Parameter
// ---------------------------------------------------------------------------

func TestRescueHandler_ValidatedParameter(t *testing.T) {
	t.Parallel()

	t.Run("validated skips precondition check", func(t *testing.T) {
		t.Parallel()
		gitOps := newFakeGitOps(false, true, false) // not a git repo
		handler := application.NewRescueHandler(newFakeScanner(nil), gitOps, nil, &fakePublisherR{}, nil)
		analysis, err := handler.Rescue(context.Background(), "/tmp/proj", nil, true)
		require.NoError(t, err)
		assert.NotEmpty(t, analysis.Gaps())
	})

	t.Run("default validates preconditions", func(t *testing.T) {
		t.Parallel()
		gitOps := newFakeGitOps(false, true, false)
		handler := application.NewRescueHandler(newFakeScanner(nil), gitOps, nil, &fakePublisherR{}, nil)
		_, err := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a git repository")
	})
}

// ---------------------------------------------------------------------------
// Tests — Execute Plan
// ---------------------------------------------------------------------------

func TestRescueHandler_ExecutePlan(t *testing.T) {
	t.Parallel()

	t.Run("completes analysis", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), writer, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)
		require.NoError(t, err)
		assert.Equal(t, rescuedomain.AnalysisStatusCompleted, analysis.Status())
		assert.Len(t, analysis.Events(), 1)
	})

	t.Run("writes files", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), writer, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		handler.ExecutePlan(context.Background(), analysis)
		assert.NotEmpty(t, writer.writtenFiles)
	})

	t.Run("without file writer raises", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no file writer")
	})

	t.Run("wrong state raises", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		scan := rescuedomain.NewProjectScan(
			"/tmp/proj",
			[]string{"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"},
			[]string{".claude/CLAUDE.md", "pyproject.toml"},
			[]string{"src/domain/", "src/application/", "src/infrastructure/"},
			true, true, true, true, true,
		)
		handler := application.NewRescueHandler(newFakeScanner(&scan), defaultFakeGitOps(), writer, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot execute plan")
	})

	t.Run("skips agents md when flagged", func(t *testing.T) {
		t.Parallel()
		scan := rescuedomain.NewProjectScan(
			"/tmp/proj",
			[]string{"AGENTS.md"}, nil, nil,
			false, true, true, false, false,
		)
		writer := newFakeFileWriter()
		handler := application.NewRescueHandler(newFakeScanner(&scan), defaultFakeGitOps(), writer, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		handler.ExecutePlan(context.Background(), analysis)

		for path := range writer.writtenFiles {
			assert.NotContains(t, path, "AGENTS.md")
		}
	})
}

// ---------------------------------------------------------------------------
// Tests — Gap Severity
// ---------------------------------------------------------------------------

func TestRescueHandler_GapSeverity(t *testing.T) {
	t.Parallel()

	t.Run("required docs have required severity", func(t *testing.T) {
		t.Parallel()
		profile := vo.PythonUvProfile{}
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", profile, false)
		for _, g := range analysis.Gaps() {
			if g.GapType() == rescuedomain.GapTypeMissingDoc {
				if g.Path() == "docs/PRD.md" || g.Path() == "docs/DDD.md" || g.Path() == "docs/ARCHITECTURE.md" {
					assert.Equal(t, rescuedomain.GapSeverityRequired, g.Severity())
				}
			}
		}
	})

	t.Run("alty config gap has recommended severity", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		for _, g := range analysis.Gaps() {
			if g.Path() == ".alty/config.toml" {
				assert.Equal(t, rescuedomain.GapSeverityRecommended, g.Severity())
			}
		}
	})

	t.Run("knowledge gap has recommended severity", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		for _, g := range analysis.Gaps() {
			if g.Path() == ".alty/knowledge/" {
				assert.Equal(t, rescuedomain.GapSeverityRecommended, g.Severity())
			}
		}
	})

	t.Run("agents md gap has recommended severity", func(t *testing.T) {
		t.Parallel()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), nil, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
		for _, g := range analysis.Gaps() {
			if g.Path() == "AGENTS.md" {
				assert.Equal(t, rescuedomain.GapSeverityRecommended, g.Severity())
			}
		}
	})
}

func TestRescueHandler_ExecutePlan_PublishesEvent(t *testing.T) {
	t.Parallel()

	pub := &fakePublisherR{}
	writer := newFakeFileWriter()
	handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), writer, pub, nil)
	analysis, err := handler.Rescue(context.Background(), "/tmp/proj", nil, false)
	require.NoError(t, err)

	err = handler.ExecutePlan(context.Background(), analysis)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(pub.published), 1)
	_, ok := pub.published[0].(rescuedomain.GapAnalysisCompleted)
	assert.True(t, ok, "expected GapAnalysisCompleted, got %T", pub.published[0])
}

// ---------------------------------------------------------------------------
// Tests — Execute Plan with Test Runner
// ---------------------------------------------------------------------------

func TestRescueHandler_ExecutePlan_RunsTestsAfterScaffolding(t *testing.T) {
	t.Parallel()

	t.Run("runs tests when framework detected", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		testRunner := newFakeTestRunner(application.TestFrameworkGo)
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), writer, &fakePublisherR{}, testRunner)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)

		require.NoError(t, err)
		assert.True(t, testRunner.detectCalled, "Detect should be called")
		assert.True(t, testRunner.runCalled, "Run should be called")
		assert.Equal(t, rescuedomain.AnalysisStatusCompleted, analysis.Status())
	})

	t.Run("skips tests when no framework detected", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		testRunner := newFakeTestRunner("") // no framework
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), writer, &fakePublisherR{}, testRunner)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)

		require.NoError(t, err)
		assert.True(t, testRunner.detectCalled, "Detect should be called")
		assert.False(t, testRunner.runCalled, "Run should not be called when no framework")
		assert.Equal(t, rescuedomain.AnalysisStatusCompleted, analysis.Status())
	})

	t.Run("completes without test runner configured", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), writer, &fakePublisherR{}, nil)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)

		require.NoError(t, err)
		assert.Equal(t, rescuedomain.AnalysisStatusCompleted, analysis.Status())
	})
}

// ---------------------------------------------------------------------------
// Tests — Execute Plan Rollback
// ---------------------------------------------------------------------------

func TestRescueHandler_ExecutePlan_Rollback(t *testing.T) {
	t.Parallel()

	t.Run("rolls back on test detection failure", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		gitOps := defaultFakeGitOps()
		testRunner := newFakeTestRunner("")
		testRunner.detectErr = assert.AnError
		handler := application.NewRescueHandler(newFakeScanner(nil), gitOps, writer, &fakePublisherR{}, testRunner)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "rollback")
		assert.Contains(t, err.Error(), "detect")
		assert.True(t, gitOps.checkoutPrevCalled, "should checkout previous branch")
		assert.True(t, gitOps.deleteBranchCalled, "should delete branch")
		assert.Contains(t, gitOps.deletedBranches, "alty/init")
	})

	t.Run("rolls back on test run failure", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		gitOps := defaultFakeGitOps()
		testRunner := newFakeTestRunner(application.TestFrameworkGo)
		testRunner.runErr = assert.AnError
		handler := application.NewRescueHandler(newFakeScanner(nil), gitOps, writer, &fakePublisherR{}, testRunner)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "rollback")
		assert.Contains(t, err.Error(), "test")
		assert.True(t, gitOps.checkoutPrevCalled, "should checkout previous branch")
		assert.True(t, gitOps.deleteBranchCalled, "should delete branch")
	})

	t.Run("fails analysis state on rollback", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		testRunner := newFakeTestRunner(application.TestFrameworkGo)
		testRunner.runErr = assert.AnError
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), writer, &fakePublisherR{}, testRunner)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		_ = handler.ExecutePlan(context.Background(), analysis)

		assert.Equal(t, rescuedomain.AnalysisStatusFailed, analysis.Status())
	})

	t.Run("rollback continues despite checkout error", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		gitOps := defaultFakeGitOps()
		gitOps.checkoutPrevErr = assert.AnError
		testRunner := newFakeTestRunner(application.TestFrameworkGo)
		testRunner.runErr = assert.AnError
		handler := application.NewRescueHandler(newFakeScanner(nil), gitOps, writer, &fakePublisherR{}, testRunner)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)

		require.Error(t, err)
		// Rollback should still attempt to delete branch even if checkout fails
		assert.True(t, gitOps.deleteBranchCalled, "should still try to delete branch")
	})

	t.Run("rollback continues despite delete branch error", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		gitOps := defaultFakeGitOps()
		gitOps.deleteBranchErr = assert.AnError
		testRunner := newFakeTestRunner(application.TestFrameworkGo)
		testRunner.runErr = assert.AnError
		handler := application.NewRescueHandler(newFakeScanner(nil), gitOps, writer, &fakePublisherR{}, testRunner)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)

		// Should return original test error, not rollback error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rollback")
	})

	t.Run("error includes test failure details", func(t *testing.T) {
		t.Parallel()
		writer := newFakeFileWriter()
		testRunner := newFakeTestRunner(application.TestFrameworkGo)
		testRunner.runErr = assert.AnError
		handler := application.NewRescueHandler(newFakeScanner(nil), defaultFakeGitOps(), writer, &fakePublisherR{}, testRunner)
		analysis, _ := handler.Rescue(context.Background(), "/tmp/proj", nil, false)

		err := handler.ExecutePlan(context.Background(), analysis)

		require.Error(t, err)
		// Error should wrap the original test failure
		assert.ErrorIs(t, err, assert.AnError)
	})
}
