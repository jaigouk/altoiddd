package domain

import (
	"fmt"
	"strings"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	"github.com/alty-cli/alty/internal/shared/domain/identity"
)

// GapType classifies a structural gap found during project analysis.
type GapType string

// Gap type constants.
const (
	GapTypeMissingDoc       GapType = "missing_doc"
	GapTypeMissingConfig    GapType = "missing_config"
	GapTypeMissingStructure GapType = "missing_structure"
	GapTypeMissingTooling   GapType = "missing_tooling"
	GapTypeMissingKnowledge GapType = "missing_knowledge"
	GapTypeConflict         GapType = "conflict"
)

// GapSeverity indicates how critical a gap is.
type GapSeverity string

// Gap severity constants.
const (
	GapSeverityRequired    GapSeverity = "required"
	GapSeverityRecommended GapSeverity = "recommended"
	GapSeverityOptional    GapSeverity = "optional"
)

// AllGapSeverities returns all gap severity values.
func AllGapSeverities() []GapSeverity {
	return []GapSeverity{GapSeverityRequired, GapSeverityRecommended, GapSeverityOptional}
}

// AnalysisStatus represents states in the gap analysis lifecycle.
type AnalysisStatus string

// Analysis status constants.
const (
	AnalysisStatusScanning  AnalysisStatus = "scanning"
	AnalysisStatusAnalyzed  AnalysisStatus = "analyzed"
	AnalysisStatusPlanned   AnalysisStatus = "planned"
	AnalysisStatusExecuting AnalysisStatus = "executing"
	AnalysisStatusCompleted AnalysisStatus = "completed"
	AnalysisStatusFailed    AnalysisStatus = "failed"
)

// ProjectScan is an immutable snapshot of an existing project's current state.
type ProjectScan struct {
	projectDir        string
	existingDocs      []string
	existingConfigs   []string
	existingStructure []string
	hasKnowledgeDir   bool
	hasAgentsMD       bool
	hasGit            bool
	hasAltyConfig     bool
	hasMaintenanceDir bool
}

// NewProjectScan creates a ProjectScan value object.
func NewProjectScan(
	projectDir string,
	existingDocs, existingConfigs, existingStructure []string,
	hasKnowledgeDir, hasAgentsMD, hasGit, hasAltyConfig, hasMaintenanceDir bool,
) ProjectScan {
	ed := make([]string, len(existingDocs))
	copy(ed, existingDocs)
	ec := make([]string, len(existingConfigs))
	copy(ec, existingConfigs)
	es := make([]string, len(existingStructure))
	copy(es, existingStructure)
	return ProjectScan{
		projectDir:        projectDir,
		existingDocs:      ed,
		existingConfigs:   ec,
		existingStructure: es,
		hasKnowledgeDir:   hasKnowledgeDir,
		hasAgentsMD:       hasAgentsMD,
		hasGit:            hasGit,
		hasAltyConfig:     hasAltyConfig,
		hasMaintenanceDir: hasMaintenanceDir,
	}
}

// ProjectDir returns the project directory path.
func (s ProjectScan) ProjectDir() string { return s.projectDir }

// HasKnowledgeDir returns whether the project has a knowledge directory.
func (s ProjectScan) HasKnowledgeDir() bool { return s.hasKnowledgeDir }

// HasAgentsMD returns whether the project has an agents.md file.
func (s ProjectScan) HasAgentsMD() bool { return s.hasAgentsMD }

// HasGit returns whether the project has a git repository.
func (s ProjectScan) HasGit() bool { return s.hasGit }

// ExistingDocs returns a defensive copy.
func (s ProjectScan) ExistingDocs() []string {
	out := make([]string, len(s.existingDocs))
	copy(out, s.existingDocs)
	return out
}

// ExistingConfigs returns a defensive copy of existing config file paths.
func (s ProjectScan) ExistingConfigs() []string {
	out := make([]string, len(s.existingConfigs))
	copy(out, s.existingConfigs)
	return out
}

// ExistingStructure returns a defensive copy of existing directory paths.
func (s ProjectScan) ExistingStructure() []string {
	out := make([]string, len(s.existingStructure))
	copy(out, s.existingStructure)
	return out
}

// HasAltyConfig returns whether the project has .alty/config.toml.
func (s ProjectScan) HasAltyConfig() bool { return s.hasAltyConfig }

// HasMaintenanceDir returns whether the project has .alty/maintenance/.
func (s ProjectScan) HasMaintenanceDir() bool { return s.hasMaintenanceDir }

// Gap is a single structural gap found during project analysis.
type Gap struct {
	gapID       string
	gapType     GapType
	path        string
	description string
	severity    GapSeverity
}

// NewGap creates a Gap value object.
func NewGap(gapID string, gapType GapType, path, description string, severity GapSeverity) Gap {
	return Gap{
		gapID:       gapID,
		gapType:     gapType,
		path:        path,
		description: description,
		severity:    severity,
	}
}

// GapID returns the gap identifier.
func (g Gap) GapID() string { return g.gapID }

// GapType returns the gap type classification.
func (g Gap) GapType() GapType { return g.gapType }

// Path returns the path of the missing or conflicting item.
func (g Gap) Path() string { return g.path }

// Description returns a human-readable gap description.
func (g Gap) Description() string { return g.description }

// Severity returns the gap severity.
func (g Gap) Severity() GapSeverity { return g.severity }

// IsDirectory returns true if this gap represents a directory (path ends with "/").
func (g Gap) IsDirectory() bool { return strings.HasSuffix(g.path, "/") }

// MigrationPlan is an immutable plan for resolving gaps.
type MigrationPlan struct {
	planID       string
	branchName   string
	gaps         []Gap
	skipAgentsMD bool
}

// NewMigrationPlan creates a MigrationPlan value object.
func NewMigrationPlan(planID string, gaps []Gap, branchName string, skipAgentsMD bool) MigrationPlan {
	g := make([]Gap, len(gaps))
	copy(g, gaps)
	if branchName == "" {
		branchName = "alty/init"
	}
	return MigrationPlan{
		planID:       planID,
		gaps:         g,
		branchName:   branchName,
		skipAgentsMD: skipAgentsMD,
	}
}

// PlanID returns the plan identifier.
func (p MigrationPlan) PlanID() string { return p.planID }

// BranchName returns the git branch name for the migration.
func (p MigrationPlan) BranchName() string { return p.branchName }

// SkipAgentsMD returns whether to skip generating agents.md.
func (p MigrationPlan) SkipAgentsMD() bool { return p.skipAgentsMD }

// Gaps returns a defensive copy.
func (p MigrationPlan) Gaps() []Gap {
	out := make([]Gap, len(p.gaps))
	copy(out, p.gaps)
	return out
}

// GapAnalysis is the aggregate root for the rescue flow.
type GapAnalysis struct {
	analysisID    string
	projectDir    string
	status        AnalysisStatus
	scan          *ProjectScan
	gaps          []Gap
	plan          *MigrationPlan
	failureReason string
	events        []GapAnalysisCompleted
}

// NewGapAnalysis creates a new GapAnalysis aggregate root.
func NewGapAnalysis(projectDir string) *GapAnalysis {
	return &GapAnalysis{
		analysisID: identity.NewID(),
		projectDir: projectDir,
		status:     AnalysisStatusScanning,
	}
}

// AnalysisID returns the analysis identifier.
func (a *GapAnalysis) AnalysisID() string { return a.analysisID }

// ProjectDir returns the project directory.
func (a *GapAnalysis) ProjectDir() string { return a.projectDir }

// Status returns the current analysis status.
func (a *GapAnalysis) Status() AnalysisStatus { return a.status }

// Scan returns the project scan result, or nil.
func (a *GapAnalysis) Scan() *ProjectScan { return a.scan }

// Gaps returns a defensive copy.
func (a *GapAnalysis) Gaps() []Gap {
	out := make([]Gap, len(a.gaps))
	copy(out, a.gaps)
	return out
}

// Plan returns the migration plan, or nil.
func (a *GapAnalysis) Plan() *MigrationPlan { return a.plan }

// Events returns a defensive copy of domain events.
func (a *GapAnalysis) Events() []GapAnalysisCompleted {
	out := make([]GapAnalysisCompleted, len(a.events))
	copy(out, a.events)
	return out
}

// SetScan records scan results. Only from SCANNING state.
func (a *GapAnalysis) SetScan(scan ProjectScan) error {
	if a.status != AnalysisStatusScanning {
		return fmt.Errorf("cannot set scan in %s state: %w", string(a.status), domainerrors.ErrInvariantViolation)
	}
	a.scan = &scan
	return nil
}

// Analyze sets analysis results. Requires scan to exist.
func (a *GapAnalysis) Analyze(gaps []Gap) error {
	if a.status != AnalysisStatusScanning {
		return fmt.Errorf("cannot analyze in %s state: %w", string(a.status), domainerrors.ErrInvariantViolation)
	}
	if a.scan == nil {
		return fmt.Errorf("cannot analyze without scan: %w", domainerrors.ErrInvariantViolation)
	}
	g := make([]Gap, len(gaps))
	copy(g, gaps)
	a.gaps = g
	a.status = AnalysisStatusAnalyzed
	return nil
}

// CreatePlan creates a migration plan. Only from ANALYZED.
func (a *GapAnalysis) CreatePlan(plan MigrationPlan) error {
	if a.status != AnalysisStatusAnalyzed {
		return fmt.Errorf("cannot create plan in %s state: %w", string(a.status), domainerrors.ErrInvariantViolation)
	}
	a.plan = &plan
	a.status = AnalysisStatusPlanned
	return nil
}

// BeginExecution starts executing the plan. Only from PLANNED.
func (a *GapAnalysis) BeginExecution() error {
	if a.status != AnalysisStatusPlanned {
		return fmt.Errorf("cannot begin execution in %s state: %w", string(a.status), domainerrors.ErrInvariantViolation)
	}
	a.status = AnalysisStatusExecuting
	return nil
}

// Complete marks as completed and emits event.
func (a *GapAnalysis) Complete() error {
	if a.status != AnalysisStatusExecuting {
		return fmt.Errorf("cannot complete in %s state: %w", string(a.status), domainerrors.ErrInvariantViolation)
	}
	a.status = AnalysisStatusCompleted
	a.events = append(a.events, NewGapAnalysisCompleted(
		a.analysisID, a.projectDir, len(a.gaps), len(a.gaps),
	))
	return nil
}

// Fail marks as failed and records the reason. Only from EXECUTING.
func (a *GapAnalysis) Fail(reason string) error {
	if a.status != AnalysisStatusExecuting {
		return fmt.Errorf("cannot fail in %s state: %w", string(a.status), domainerrors.ErrInvariantViolation)
	}
	a.status = AnalysisStatusFailed
	a.failureReason = reason
	return nil
}

// FailureReason returns the reason for failure, or empty string if not failed.
func (a *GapAnalysis) FailureReason() string { return a.failureReason }
