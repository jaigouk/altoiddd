// Package domain contains shared domain types for the alty application.
package domain

import "sync"

// ---------------------------------------------------------------------------
// WorkflowStep Value Object
// ---------------------------------------------------------------------------

// WorkflowStep represents a named step in the alty workflow.
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
// SessionTracker Aggregate
// ---------------------------------------------------------------------------

// sessionState holds the workflow state for a single session.
type sessionState struct {
	ready     map[WorkflowStep]struct{}
	completed map[WorkflowStep]struct{}
}

// SessionTracker tracks workflow progress across multiple sessions.
// It maintains which steps are ready (available for user action) and which
// are completed for each session. Thread-safe for concurrent MCP sessions.
type SessionTracker struct {
	mu       sync.RWMutex
	sessions map[string]*sessionState
}

// NewSessionTracker creates a new SessionTracker.
func NewSessionTracker() *SessionTracker {
	return &SessionTracker{
		sessions: make(map[string]*sessionState),
	}
}

// getOrCreate returns the session state for the given ID, creating it if needed.
// Caller must hold the write lock.
func (t *SessionTracker) getOrCreate(sessionID string) *sessionState {
	state, ok := t.sessions[sessionID]
	if !ok {
		state = &sessionState{
			ready:     make(map[WorkflowStep]struct{}),
			completed: make(map[WorkflowStep]struct{}),
		}
		t.sessions[sessionID] = state
	}
	return state
}

// MarkReady marks one or more workflow steps as ready for the given session.
// Idempotent — marking an already-ready step has no effect.
func (t *SessionTracker) MarkReady(sessionID string, steps ...WorkflowStep) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.getOrCreate(sessionID)
	for _, step := range steps {
		// Only mark ready if not already completed
		if _, done := state.completed[step]; !done {
			state.ready[step] = struct{}{}
		}
	}
}

// MarkCompleted marks a workflow step as completed for the given session.
// Removes the step from ready actions. Idempotent.
func (t *SessionTracker) MarkCompleted(sessionID string, step WorkflowStep) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.getOrCreate(sessionID)
	delete(state.ready, step)
	state.completed[step] = struct{}{}
}

// ReadyActions returns the list of actions currently available for the session.
// Returns an empty slice for unknown sessions.
func (t *SessionTracker) ReadyActions(sessionID string) []ReadyAction {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, ok := t.sessions[sessionID]
	if !ok || len(state.ready) == 0 {
		return []ReadyAction{}
	}

	actions := make([]ReadyAction, 0, len(state.ready))
	for step := range state.ready {
		actions = append(actions, NewReadyAction(step.String()))
	}
	return actions
}

// IsReady returns true if the given step is ready for the session.
func (t *SessionTracker) IsReady(sessionID string, step WorkflowStep) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, ok := t.sessions[sessionID]
	if !ok {
		return false
	}
	_, ready := state.ready[step]
	return ready
}

// IsCompleted returns true if the given step is completed for the session.
func (t *SessionTracker) IsCompleted(sessionID string, step WorkflowStep) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, ok := t.sessions[sessionID]
	if !ok {
		return false
	}
	_, done := state.completed[step]
	return done
}
