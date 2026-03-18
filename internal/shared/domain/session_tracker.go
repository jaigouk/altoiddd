// Package domain contains shared domain types for the alto application.
package domain

// ---------------------------------------------------------------------------
// WorkflowStep Value Object
// ---------------------------------------------------------------------------

// WorkflowStep represents a named step in the alto workflow.
// Steps transition through pending → ready → completed.
type WorkflowStep string

const (
	// StepArtifactGeneration becomes ready after DiscoveryCompleted.
	StepArtifactGeneration WorkflowStep = "artifact_generation"

	// StepFitness becomes ready after DomainModelGenerated.
	StepFitness WorkflowStep = "fitness"

	// StepTickets becomes ready after DomainModelGenerated.
	StepTickets WorkflowStep = "tickets"

	// StepConfigs becomes ready after DomainModelGenerated.
	StepConfigs WorkflowStep = "configs"

	// StepRippleReview becomes ready after TicketPlanApproved.
	StepRippleReview WorkflowStep = "ripple_review"
)

// String returns the string representation of the workflow step.
func (s WorkflowStep) String() string {
	return string(s)
}

// AllWorkflowSteps returns all defined workflow steps.
func AllWorkflowSteps() []WorkflowStep {
	return []WorkflowStep{
		StepArtifactGeneration,
		StepFitness,
		StepTickets,
		StepConfigs,
		StepRippleReview,
	}
}

// ---------------------------------------------------------------------------
// ReadyAction Value Object
// ---------------------------------------------------------------------------

// ReadyAction represents an action that is available for the user to take.
// Immutable value object.
type ReadyAction struct {
	name string
}

// NewReadyAction creates a ReadyAction with the given name.
func NewReadyAction(name string) ReadyAction {
	return ReadyAction{name: name}
}

// Name returns the action name.
func (a ReadyAction) Name() string {
	return a.name
}

// ---------------------------------------------------------------------------
// SessionTracker — Backward Compatibility Alias
// ---------------------------------------------------------------------------

// SessionTracker is an alias for WorkflowCoordinator.
//
// Deprecated: Use WorkflowCoordinator directly for new code.
type SessionTracker = WorkflowCoordinator

// NewSessionTracker creates a new SessionTracker (WorkflowCoordinator).
//
// Deprecated: Use NewWorkflowCoordinator directly for new code.
func NewSessionTracker() *SessionTracker {
	return NewWorkflowCoordinator()
}
