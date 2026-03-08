package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	sharedapp "github.com/alty-cli/alty/internal/shared/application"
	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	"github.com/alty-cli/alty/internal/shared/domain/identity"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// Required docs that any alty project should have.
var requiredDocs = []string{
	"docs/PRD.md",
	"docs/DDD.md",
	"docs/ARCHITECTURE.md",
}

// Required configs that any alty project should have.
var requiredConfigs = []string{".claude/CLAUDE.md"}

const branchName = "alty/init"

// RescueHandler orchestrates the rescue flow: scan -> validate -> analyze -> plan.
type RescueHandler struct {
	projectScan ProjectScan
	gitOps      GitOps
	fileWriter  sharedapp.FileWriter
}

// NewRescueHandler creates a new RescueHandler with injected dependencies.
func NewRescueHandler(
	projectScan ProjectScan,
	gitOps GitOps,
	fileWriter sharedapp.FileWriter,
) *RescueHandler {
	return &RescueHandler{
		projectScan: projectScan,
		gitOps:      gitOps,
		fileWriter:  fileWriter,
	}
}

// ValidatePreconditions validates git preconditions before rescue.
func (h *RescueHandler) ValidatePreconditions(ctx context.Context, projectDir string) error {
	hasGit, err := h.gitOps.HasGit(ctx, projectDir)
	if err != nil {
		return err
	}
	if !hasGit {
		return fmt.Errorf("Not a git repository: %w", domainerrors.ErrInvariantViolation)
	}

	isClean, err := h.gitOps.IsClean(ctx, projectDir)
	if err != nil {
		return err
	}
	if !isClean {
		return fmt.Errorf("Working tree is dirty: %w", domainerrors.ErrInvariantViolation)
	}

	exists, err := h.gitOps.BranchExists(ctx, projectDir, branchName)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Branch %s already exists. Delete it first or use --force-branch to override.: %w",
			branchName, domainerrors.ErrInvariantViolation)
	}

	return nil
}

// Rescue analyzes an existing project and produces a gap analysis with plan.
func (h *RescueHandler) Rescue(
	ctx context.Context,
	projectDir string,
	profile vo.StackProfile,
	validated bool,
) (*rescuedomain.GapAnalysis, error) {
	if !validated {
		if err := h.ValidatePreconditions(ctx, projectDir); err != nil {
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
	gaps := h.identifyGaps(scan, profile)

	// Build aggregate
	analysis := rescuedomain.NewGapAnalysis(projectDir)
	if err := analysis.SetScan(scan); err != nil {
		return nil, err
	}
	if err := analysis.Analyze(gaps); err != nil {
		return nil, err
	}

	if len(gaps) > 0 {
		plan := rescuedomain.NewMigrationPlan(
			identity.NewID(),
			gaps,
			branchName,
			scan.HasAgentsMD(),
		)
		if err := analysis.CreatePlan(plan); err != nil {
			return nil, err
		}
	}

	return analysis, nil
}

// ExecutePlan executes a planned migration.
func (h *RescueHandler) ExecutePlan(ctx context.Context, analysis *rescuedomain.GapAnalysis) error {
	if analysis.Status() != rescuedomain.AnalysisStatusPlanned {
		return fmt.Errorf("Cannot execute plan in %s state: %w",
			string(analysis.Status()), domainerrors.ErrInvariantViolation)
	}

	if h.fileWriter == nil {
		return fmt.Errorf("No file writer configured for plan execution: %w",
			domainerrors.ErrInvariantViolation)
	}

	if err := analysis.BeginExecution(); err != nil {
		return err
	}

	plan := analysis.Plan()
	if plan == nil {
		return analysis.Fail("No plan available")
	}

	for _, gap := range plan.Gaps() {
		if gap.GapType() == rescuedomain.GapTypeConflict {
			continue
		}
		if plan.SkipAgentsMD() && gap.Path() == "AGENTS.md" {
			continue
		}

		target := filepath.Join(analysis.ProjectDir(), gap.Path())
		stem := strings.TrimSuffix(filepath.Base(gap.Path()), filepath.Ext(gap.Path()))
		content := fmt.Sprintf("# %s\n\n> TODO: Fill in content.\n", stem)
		if err := h.fileWriter.WriteFile(ctx, target, content); err != nil {
			return err
		}
	}

	return analysis.Complete()
}

func (h *RescueHandler) identifyGaps(
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

	// Check required configs (alty-universal)
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
			".alty/knowledge/",
			"Missing knowledge base directory",
			rescuedomain.GapSeverityRecommended,
		))
	}

	// Check .alty/config.toml
	if !scan.HasAltyConfig() {
		gaps = append(gaps, rescuedomain.NewGap(
			identity.NewID(),
			rescuedomain.GapTypeMissingConfig,
			".alty/config.toml",
			"Missing alty project configuration",
			rescuedomain.GapSeverityRecommended,
		))
	}

	// Check .alty/maintenance/
	if !scan.HasMaintenanceDir() {
		gaps = append(gaps, rescuedomain.NewGap(
			identity.NewID(),
			rescuedomain.GapTypeMissingStructure,
			".alty/maintenance/",
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
