package application

import (
	"context"
	"fmt"

	vo "github.com/alto-cli/alto/internal/shared/domain/valueobjects"
)

// allGates is the default set of quality gates to run.
var allGates = []vo.QualityGate{
	vo.QualityGateLint,
	vo.QualityGateTypes,
	vo.QualityGateTests,
	vo.QualityGateFitness,
}

// QualityGateHandler orchestrates quality gate execution.
// Runs one or more quality gates via a GateRunner adapter,
// collecting results into a QualityReport. Continues running
// remaining gates even when earlier gates fail.
type QualityGateHandler struct {
	runner GateRunner
}

// NewQualityGateHandler creates a new QualityGateHandler with an injected GateRunner.
func NewQualityGateHandler(runner GateRunner) *QualityGateHandler {
	return &QualityGateHandler{runner: runner}
}

// Check runs quality gates and returns an aggregated report.
// If gates is nil, all four gates are run.
func (h *QualityGateHandler) Check(
	ctx context.Context,
	gates []vo.QualityGate,
) (vo.QualityReport, error) {
	gatesToRun := gates
	if gatesToRun == nil {
		gatesToRun = allGates
	}

	var results []vo.GateResult
	for _, gate := range gatesToRun {
		result, err := h.runner.Run(ctx, gate)
		if err != nil {
			return vo.QualityReport{}, fmt.Errorf("run gate %s: %w", gate, err)
		}
		results = append(results, result)
	}

	return vo.NewQualityReport(results), nil
}
