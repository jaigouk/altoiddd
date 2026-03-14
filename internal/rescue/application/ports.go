// Package application defines ports for the Rescue bounded context.
package application

import (
	"context"

	rescuedomain "github.com/alty-cli/alty/internal/rescue/domain"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// Test framework constants.
const (
	TestFrameworkGo     = "go"
	TestFrameworkNPM    = "npm"
	TestFrameworkPytest = "pytest"
)

// ProjectScan scans an existing project's structure to report what
// documentation, configs, and structure already exist.
type ProjectScan interface {
	// Scan scans a project directory and returns a frozen snapshot.
	Scan(ctx context.Context, projectDir string, profile vo.StackProfile) (rescuedomain.ProjectScan, error)
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

// TestRunner detects and runs project tests for migration validation.
type TestRunner interface {
	// Detect identifies the test framework used in a project directory.
	// Returns one of TestFramework* constants, or empty string if none detected.
	Detect(ctx context.Context, projectDir string) (string, error)

	// Run executes tests using the specified framework.
	// The framework parameter should be one of TestFramework* constants.
	Run(ctx context.Context, projectDir string, framework string) error
}
