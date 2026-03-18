// Package domain provides the Rescue bounded context's core domain model.
// It contains domain events for the rescue/gap-analysis workflow.
//
// GapAnalysisCompleted lives in the shared kernel (internal/shared/domain/events)
// because it is consumed across context boundaries (fitness, integration tests).
// This file re-exports the type for convenience within the rescue context.
package domain

import (
	sharedevents "github.com/alto-cli/alto/internal/shared/domain/events"
)

// GapAnalysisCompleted is a type alias for the shared kernel event.
type GapAnalysisCompleted = sharedevents.GapAnalysisCompleted

// NewGapAnalysisCompleted delegates to the shared kernel constructor.
func NewGapAnalysisCompleted(analysisID, projectDir string, gapsFound, gapsResolved int) GapAnalysisCompleted {
	return sharedevents.NewGapAnalysisCompleted(analysisID, projectDir, gapsFound, gapsResolved)
}
