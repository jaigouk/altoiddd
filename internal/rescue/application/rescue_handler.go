package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	rescuedomain "github.com/alto-cli/alto/internal/rescue/domain"
	sharedapp "github.com/alto-cli/alto/internal/shared/application"
	domainerrors "github.com/alto-cli/alto/internal/shared/domain/errors"
	"github.com/alto-cli/alto/internal/shared/domain/identity"
	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// Required docs that any alto project should have.
var requiredDocs = []string{
	"docs/PRD.md",
	"docs/DDD.md",
	"docs/ARCHITECTURE.md",
}

// Required configs that any alto project should have.
var requiredConfigs = []string{".claude/CLAUDE.md"}

const branchName = "alto/init"

// RescueHandler orchestrates the rescue flow: scan -> validate -> analyze -> plan.
type RescueHandler struct {
	projectScan ProjectScan
	gitOps      sharedapp.GitOps
	fileWriter  sharedapp.FileWriter
	publisher   sharedapp.EventPublisher
	testRunner  TestRunner
	dirCreator  sharedapp.DirCreator
}

// NewRescueHandler creates a new RescueHandler with injected dependencies.
func NewRescueHandler(
	projectScan ProjectScan,
	gitOps sharedapp.GitOps,
	fileWriter sharedapp.FileWriter,
	publisher sharedapp.EventPublisher,
	testRunner TestRunner,
	dirCreator sharedapp.DirCreator,
) *RescueHandler {
	return &RescueHandler{
		projectScan: projectScan,
		gitOps:      gitOps,
		fileWriter:  fileWriter,
		publisher:   publisher,
		testRunner:  testRunner,
		dirCreator:  dirCreator,
	}
}

// ValidatePreconditions validates git preconditions before rescue.
// When forceBranch is true and the branch already exists, it deletes the
// existing branch instead of returning an error.
func (h *RescueHandler) ValidatePreconditions(ctx context.Context, projectDir string, forceBranch bool) error {
	hasGit, err := h.gitOps.HasGit(ctx, projectDir)
	if err != nil {
		return fmt.Errorf("check git repository: %w", err)
	}
	if !hasGit {
		return fmt.Errorf("not a git repository: %w", domainerrors.ErrInvariantViolation)
	}

	isClean, err := h.gitOps.IsClean(ctx, projectDir)
	if err != nil {
		return fmt.Errorf("check working tree: %w", err)
	}
	if !isClean {
		return fmt.Errorf("working tree is dirty: %w", domainerrors.ErrInvariantViolation)
	}

	exists, err := h.gitOps.BranchExists(ctx, projectDir, branchName)
	if err != nil {
		return fmt.Errorf("check branch existence: %w", err)
	}
	if exists {
		if forceBranch {
			if err := h.gitOps.DeleteBranch(ctx, projectDir, branchName); err != nil {
				return fmt.Errorf("delete existing branch: %w", err)
			}
		} else {
			return fmt.Errorf("branch %s already exists, delete it first or use --force-branch to override: %w",
				branchName, domainerrors.ErrInvariantViolation)
		}
	}

	return nil
}

// Rescue analyzes an existing project and produces a gap analysis with plan.
func (h *RescueHandler) Rescue(
	ctx context.Context,
	projectDir string,
	profile vo.StackProfile,
	validated bool,
	forceBranch bool,
) (*rescuedomain.GapAnalysis, error) {
	if !validated {
		if err := h.ValidatePreconditions(ctx, projectDir, forceBranch); err != nil {
			return nil, err
		}
	}

	// Create branch before scanning
	if err := h.gitOps.CreateBranch(ctx, projectDir, branchName); err != nil {
		return nil, fmt.Errorf("creating branch: %w", err)
	}

	// Scan project
	scan, err := h.projectScan.Scan(ctx, projectDir, profile)
	if err != nil {
		return nil, fmt.Errorf("scanning project: %w", err)
	}

	// Analyze gaps
	gaps := IdentifyGaps(scan, profile)

	// Build aggregate
	analysis := rescuedomain.NewGapAnalysis(projectDir)
	if err := analysis.SetScan(scan); err != nil {
		return nil, fmt.Errorf("set scan: %w", err)
	}
	if err := analysis.Analyze(gaps); err != nil {
		return nil, fmt.Errorf("analyze gaps: %w", err)
	}

	if len(gaps) > 0 {
		plan := rescuedomain.NewMigrationPlan(
			identity.NewID(),
			gaps,
			branchName,
			scan.HasAgentsMD(),
		)
		if err := analysis.CreatePlan(plan); err != nil {
			return nil, fmt.Errorf("create plan: %w", err)
		}
	}

	return analysis, nil
}

// ExecutePlan executes a planned migration.
func (h *RescueHandler) ExecutePlan(ctx context.Context, analysis *rescuedomain.GapAnalysis) error {
	if analysis.Status() != rescuedomain.AnalysisStatusPlanned {
		return fmt.Errorf("cannot execute plan in %s state: %w",
			string(analysis.Status()), domainerrors.ErrInvariantViolation)
	}

	if h.fileWriter == nil {
		return fmt.Errorf("no file writer configured for plan execution: %w",
			domainerrors.ErrInvariantViolation)
	}

	if err := analysis.BeginExecution(); err != nil {
		return fmt.Errorf("begin execution: %w", err)
	}

	plan := analysis.Plan()
	if plan == nil {
		if err := analysis.Fail("no plan available"); err != nil {
			return fmt.Errorf("fail analysis: %w", err)
		}
		return fmt.Errorf("no plan available: %w", domainerrors.ErrInvariantViolation)
	}

	for _, gap := range plan.Gaps() {
		if gap.GapType() == rescuedomain.GapTypeConflict {
			continue
		}
		if plan.SkipAgentsMD() && gap.Path() == "AGENTS.md" {
			continue
		}

		target := filepath.Join(analysis.ProjectDir(), gap.Path())

		if gap.IsDirectory() {
			if h.dirCreator != nil {
				if err := h.dirCreator.EnsureDir(ctx, target); err != nil {
					return fmt.Errorf("create directory %s: %w", gap.Path(), err)
				}
			}
			continue
		}

		stem := strings.TrimSuffix(filepath.Base(gap.Path()), filepath.Ext(gap.Path()))
		content := fmt.Sprintf("# %s\n\n> TODO: Fill in content.\n", stem)
		if err := h.fileWriter.WriteFile(ctx, target, content); err != nil {
			return fmt.Errorf("write file %s: %w", gap.Path(), err)
		}
	}

	// Run tests if test runner is configured
	if h.testRunner != nil {
		framework, err := h.testRunner.Detect(ctx, analysis.ProjectDir())
		if err != nil {
			_ = analysis.Fail("test detection failed")
			rollbackErr := h.rollback(ctx, analysis.ProjectDir(), plan.BranchName())
			if rollbackErr != nil {
				return fmt.Errorf("test detection failed, rollback also failed: %w", err)
			}
			return fmt.Errorf("test detection failed, rollback completed: %w", err)
		}

		if framework != "" {
			if err := h.testRunner.Run(ctx, analysis.ProjectDir(), framework); err != nil {
				_ = analysis.Fail("tests failed")
				rollbackErr := h.rollback(ctx, analysis.ProjectDir(), plan.BranchName())
				if rollbackErr != nil {
					return fmt.Errorf("tests failed, rollback also failed: %w", err)
				}
				return fmt.Errorf("tests failed, rollback completed: %w", err)
			}
		}
	}

	if err := analysis.Complete(); err != nil {
		return fmt.Errorf("complete analysis: %w", err)
	}
	for _, event := range analysis.Events() {
		_ = h.publisher.Publish(ctx, event)
	}
	return nil
}

// rollback undoes the migration by switching to the previous branch and deleting the migration branch.
// It attempts both operations even if one fails, returning the first error encountered.
func (h *RescueHandler) rollback(ctx context.Context, projectDir, branchName string) error {
	var firstErr error
	if err := h.gitOps.CheckoutPrevious(ctx, projectDir); err != nil {
		firstErr = fmt.Errorf("checkout previous: %w", err)
	}
	if err := h.gitOps.DeleteBranch(ctx, projectDir, branchName); err != nil {
		if firstErr == nil {
			firstErr = fmt.Errorf("delete branch: %w", err)
		}
	}
	return firstErr
}

// IdentifyGaps compares a project scan against required structure and returns gaps.
// It is a pure function with no handler dependencies, usable by both RescueHandler
// and GapQueryHandler.
func IdentifyGaps(
	scan rescuedomain.ProjectScan,
	profile vo.StackProfile,
) []rescuedomain.Gap {
	existingDocs := toSet(scan.ExistingDocs())
	existingConfigs := toSet(scan.ExistingConfigs())
	existingStructure := toSet(scan.ExistingStructure())

	var gaps []rescuedomain.Gap

	// Check required docs
	for _, docPath := range requiredDocs {
		if !existingDocs[docPath] {
			gaps = append(gaps, rescuedomain.NewGap(
				identity.NewID(),
				rescuedomain.GapTypeMissingDoc,
				docPath,
				fmt.Sprintf("Missing documentation: %s", docPath),
				rescuedomain.GapSeverityRequired,
			))
		}
	}

	// Check required configs (alto-universal)
	for _, configPath := range requiredConfigs {
		if !existingConfigs[configPath] {
			gaps = append(gaps, rescuedomain.NewGap(
				identity.NewID(),
				rescuedomain.GapTypeMissingConfig,
				configPath,
				fmt.Sprintf("Missing configuration: %s", configPath),
				rescuedomain.GapSeverityRequired,
			))
		}
	}

	// Check project manifest from profile
	if profile != nil {
		manifest := profile.ProjectManifest()
		if manifest != "" && !existingConfigs[manifest] {
			gaps = append(gaps, rescuedomain.NewGap(
				identity.NewID(),
				rescuedomain.GapTypeMissingConfig,
				manifest,
				fmt.Sprintf("Missing configuration: %s", manifest),
				rescuedomain.GapSeverityRequired,
			))
		}
	}

	// Check structure from profile
	if profile != nil {
		for _, structurePath := range profile.SourceLayout() {
			if !existingStructure[structurePath] {
				gaps = append(gaps, rescuedomain.NewGap(
					identity.NewID(),
					rescuedomain.GapTypeMissingStructure,
					structurePath,
					fmt.Sprintf("Missing directory: %s", structurePath),
					rescuedomain.GapSeverityRequired,
				))
			}
		}
	}

	// Check knowledge directory
	if !scan.HasKnowledgeDir() {
		gaps = append(gaps, rescuedomain.NewGap(
			identity.NewID(),
			rescuedomain.GapTypeMissingKnowledge,
			".alto/knowledge/",
			"Missing knowledge base directory",
			rescuedomain.GapSeverityRecommended,
		))
	}

	// Check .alto/config.toml
	if !scan.HasAltoConfig() {
		gaps = append(gaps, rescuedomain.NewGap(
			identity.NewID(),
			rescuedomain.GapTypeMissingConfig,
			".alto/config.toml",
			"Missing alto project configuration",
			rescuedomain.GapSeverityRecommended,
		))
	}

	// Check .alto/maintenance/
	if !scan.HasMaintenanceDir() {
		gaps = append(gaps, rescuedomain.NewGap(
			identity.NewID(),
			rescuedomain.GapTypeMissingStructure,
			".alto/maintenance/",
			"Missing doc maintenance directory",
			rescuedomain.GapSeverityRecommended,
		))
	}

	// Check AGENTS.md
	if !scan.HasAgentsMD() {
		gaps = append(gaps, rescuedomain.NewGap(
			identity.NewID(),
			rescuedomain.GapTypeMissingDoc,
			"AGENTS.md",
			"Missing AGENTS.md",
			rescuedomain.GapSeverityRecommended,
		))
	}

	return gaps
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
