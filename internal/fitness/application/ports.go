// Package application defines ports for the Fitness bounded context.
package application

import (
	"context"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// FitnessGeneration generates architecture fitness functions (import-linter
// contracts, pytestarch tests) from a domain model.
type FitnessGeneration interface {
	// Generate generates fitness function tests from a domain model.
	Generate(ctx context.Context, model *ddd.DomainModel, rootPackage string, outputDir string) error
}

// --- Quality Gate Ports (ISP: single-gate runner vs multi-gate orchestrator) ---

// GateRunner runs a single quality gate command and returns structured results.
type GateRunner interface {
	// Run executes a single quality gate and returns its result.
	Run(ctx context.Context, gate vo.QualityGate) (vo.GateResult, error)
}

// QualityGateChecker runs multiple quality gate checks and returns an aggregated report.
type QualityGateChecker interface {
	// Check runs the specified quality gates (or all if nil).
	Check(ctx context.Context, gates []vo.QualityGate) (vo.QualityReport, error)
}
