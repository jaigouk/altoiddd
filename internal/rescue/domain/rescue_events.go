// Package domain provides the Rescue bounded context's core domain model.
// It contains domain events for the rescue/gap-analysis workflow.
package domain

// GapAnalysisCompleted is emitted when a rescue gap analysis completes execution.
type GapAnalysisCompleted struct {
	analysisID   string
	projectDir   string
	gapsFound    int
	gapsResolved int
}

// NewGapAnalysisCompleted creates a GapAnalysisCompleted event.
func NewGapAnalysisCompleted(analysisID, projectDir string, gapsFound, gapsResolved int) GapAnalysisCompleted {
	return GapAnalysisCompleted{
		analysisID:   analysisID,
		projectDir:   projectDir,
		gapsFound:    gapsFound,
		gapsResolved: gapsResolved,
	}
}

// AnalysisID returns the analysis identifier.
func (e GapAnalysisCompleted) AnalysisID() string { return e.analysisID }

// ProjectDir returns the project directory.
func (e GapAnalysisCompleted) ProjectDir() string { return e.projectDir }

// GapsFound returns the number of gaps found.
func (e GapAnalysisCompleted) GapsFound() int { return e.gapsFound }

// GapsResolved returns the number of gaps resolved.
func (e GapAnalysisCompleted) GapsResolved() int { return e.gapsResolved }
