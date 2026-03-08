package application_test

import (
	"context"

	"github.com/alty-cli/alty/internal/fitness/application"
	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// Compile-time interface satisfaction checks.
var (
	_ application.FitnessGeneration  = (*mockFitnessGeneration)(nil)
	_ application.GateRunner         = (*mockGateRunner)(nil)
	_ application.QualityGateChecker = (*mockQualityGateChecker)(nil)
)

// --- mockFitnessGeneration ---

type mockFitnessGeneration struct{}

func (m *mockFitnessGeneration) Generate(_ context.Context, _ *ddd.DomainModel, _ string, _ string) error {
	return nil
}

// --- mockGateRunner ---

type mockGateRunner struct{}

func (m *mockGateRunner) Run(_ context.Context, _ vo.QualityGate) (vo.GateResult, error) {
	return vo.GateResult{}, nil
}

// --- mockQualityGateChecker ---

type mockQualityGateChecker struct{}

func (m *mockQualityGateChecker) Check(_ context.Context, _ []vo.QualityGate) (vo.QualityReport, error) {
	return vo.QualityReport{}, nil
}
