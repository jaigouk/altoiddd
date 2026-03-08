// Package application defines ports for the Rescue bounded context.
package application

import (
	"context"

	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ProjectScan scans an existing project's structure to report what
// documentation, configs, and structure already exist.
type ProjectScan interface {
	// Scan scans a project directory and returns a frozen snapshot.
	Scan(ctx context.Context, projectDir string, profile vo.StackProfile) (rescuedomain.ProjectScan, error)
}

// GitOps provides git operations needed by the rescue flow.
type GitOps interface {
	// HasGit checks whether the directory is inside a git repository.
	HasGit(ctx context.Context, projectDir string) (bool, error)

	// IsClean checks whether the git working tree is clean.
	IsClean(ctx context.Context, projectDir string) (bool, error)

	// BranchExists checks whether a git branch already exists.
	BranchExists(ctx context.Context, projectDir string, branchName string) (bool, error)

	// CreateBranch creates and checks out a new git branch.
	CreateBranch(ctx context.Context, projectDir string, branchName string) error
}

// Rescue handles analyzing an existing project, planning migration steps,
// and executing the rescue flow (alty init --existing).
type Rescue interface {
	// Analyze analyzes an existing project for structural gaps.
	Analyze(ctx context.Context, projectDir string) (*rescuedomain.GapAnalysis, error)

	// Plan creates a migration plan from the gap analysis.
	Plan(ctx context.Context, gapAnalysis *rescuedomain.GapAnalysis) (rescuedomain.MigrationPlan, error)

	// Execute executes the migration plan.
	Execute(ctx context.Context, plan rescuedomain.MigrationPlan) error
}
