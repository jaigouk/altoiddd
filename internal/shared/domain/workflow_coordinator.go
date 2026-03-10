// Package domain contains shared domain types for the alty application.
package domain

import (
	"fmt"
	"sync"
	"time"

	"github.com/alty-cli/alty/internal/shared/domain/ddd"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// ---------------------------------------------------------------------------
// StepStatus Value Object
// ---------------------------------------------------------------------------

// StepStatus represents the current status of a workflow step.
type StepStatus int

const (
	// StepPending means the step is not yet available (preconditions not met).
	StepPending StepStatus = iota
	// StepReady means the step is available for execution.
	StepReady
	// StepInProgress means the step is currently executing.
	StepInProgress
	// StepCompleted means the step has finished successfully.
	StepCompleted
	// StepSkipped means the user chose to skip this step.
	StepSkipped
)

// String returns the string representation of the step status.
func (s StepStatus) String() string {
	switch s {
	case StepPending:
		return "pending"
	case StepReady:
		return "ready"
	case StepInProgress:
		return "in_progress"
	case StepCompleted:
		return "completed"
	case StepSkipped:
		return "skipped"
	default:
		return "unknown"
	}
}

// AllStepStatuses returns all defined step statuses.
func AllStepStatuses() []StepStatus {
	return []StepStatus{
		StepPending,
		StepReady,
		StepInProgress,
		StepCompleted,
		StepSkipped,
	}
}

// ---------------------------------------------------------------------------
// SessionContext Value Object
// ---------------------------------------------------------------------------

// SessionContext holds the context for a workflow session including
// the domain model, stack profile, and project directory.
type SessionContext struct {
	SessionID    string
	DomainModel  *ddd.DomainModel
	StackProfile vo.StackProfile
	ProjectDir   string
	CreatedAt    time.Time
}

// ---------------------------------------------------------------------------
// WorkflowCoordinator Aggregate
// ---------------------------------------------------------------------------

// workflowSession holds the workflow state for a single session.
type workflowSession struct {
	steps   map[WorkflowStep]StepStatus
	context *SessionContext
}

// WorkflowCoordinator manages workflow progress across multiple sessions.
// It tracks step statuses, enforces state transitions, and holds session context.
// Thread-safe for concurrent MCP sessions.
//
// State transitions:
//
//	PENDING -> READY (via MarkReady)
//	READY -> IN_PROGRESS (via BeginStep)
//	READY -> SKIPPED (via SkipStep)
//	IN_PROGRESS -> COMPLETED (via CompleteStep)
type WorkflowCoordinator struct {
	mu       sync.RWMutex
	sessions map[string]*workflowSession
}

// NewWorkflowCoordinator creates a new WorkflowCoordinator.
func NewWorkflowCoordinator() *WorkflowCoordinator {
	return &WorkflowCoordinator{
		sessions: make(map[string]*workflowSession),
	}
}

// getOrCreate returns the session for the given ID, creating it if needed.
// Caller must hold the write lock.
func (c *WorkflowCoordinator) getOrCreate(sessionID string) *workflowSession {
	session, ok := c.sessions[sessionID]
	if !ok {
		session = &workflowSession{
			steps: make(map[WorkflowStep]StepStatus),
		}
		c.sessions[sessionID] = session
	}
	return session
}

// ---------------------------------------------------------------------------
// Status Queries
// ---------------------------------------------------------------------------

// StepStatus returns the current status of a step for the given session.
// Returns StepPending if the session or step is unknown.
func (c *WorkflowCoordinator) StepStatus(sessionID string, step WorkflowStep) StepStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, ok := c.sessions[sessionID]
	if !ok {
		return StepPending
	}
	status, ok := session.steps[step]
	if !ok {
		return StepPending
	}
	return status
}

// CanExecute returns true if the step is ready to be executed.
// A step can only be executed if its status is StepReady.
func (c *WorkflowCoordinator) CanExecute(sessionID string, step WorkflowStep) bool {
	return c.StepStatus(sessionID, step) == StepReady
}

// IsReady returns true if the given step is ready for the session.
// Provided for backward compatibility with SessionTracker.
func (c *WorkflowCoordinator) IsReady(sessionID string, step WorkflowStep) bool {
	return c.StepStatus(sessionID, step) == StepReady
}

// IsCompleted returns true if the given step is completed for the session.
// Provided for backward compatibility with SessionTracker.
func (c *WorkflowCoordinator) IsCompleted(sessionID string, step WorkflowStep) bool {
	return c.StepStatus(sessionID, step) == StepCompleted
}

// AvailableActions returns the list of actions currently available for the session.
// Returns steps that are in Ready status. Returns an empty slice for unknown sessions.
func (c *WorkflowCoordinator) AvailableActions(sessionID string) []ReadyAction {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, ok := c.sessions[sessionID]
	if !ok {
		return []ReadyAction{}
	}

	actions := make([]ReadyAction, 0)
	for step, status := range session.steps {
		if status == StepReady {
			actions = append(actions, NewReadyAction(step.String()))
		}
	}
	return actions
}

// ---------------------------------------------------------------------------
// State Transitions
// ---------------------------------------------------------------------------

// MarkReady marks one or more workflow steps as ready for the given session.
// Transitions: PENDING -> READY
// Idempotent — marking an already-ready step has no effect.
// Does not affect steps that are already in progress, completed, or skipped.
func (c *WorkflowCoordinator) MarkReady(sessionID string, steps ...WorkflowStep) {
	c.mu.Lock()
	defer c.mu.Unlock()

	session := c.getOrCreate(sessionID)
	for _, step := range steps {
		current := session.steps[step]
		// Only mark ready if pending (default) or not set
		if current == StepPending {
			session.steps[step] = StepReady
		}
	}
}

// BeginStep transitions a step from Ready to InProgress.
// Returns an error if the step is not in Ready status.
func (c *WorkflowCoordinator) BeginStep(sessionID string, step WorkflowStep) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session := c.getOrCreate(sessionID)
	current := session.steps[step]

	switch current {
	case StepReady:
		session.steps[step] = StepInProgress
		return nil
	case StepInProgress:
		return fmt.Errorf("step %q already in progress", step)
	case StepPending:
		return fmt.Errorf("step %q not ready: preconditions not met", step)
	case StepCompleted:
		return fmt.Errorf("step %q already completed", step)
	case StepSkipped:
		return fmt.Errorf("step %q was skipped", step)
	}
	return nil // unreachable, but required for exhaustive switch
}

// CompleteStep transitions a step from InProgress to Completed.
// Returns an error if the step is not in InProgress status.
func (c *WorkflowCoordinator) CompleteStep(sessionID string, step WorkflowStep) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session := c.getOrCreate(sessionID)
	current := session.steps[step]

	if current != StepInProgress {
		return fmt.Errorf("step %q not in progress: status is %s", step, current)
	}

	session.steps[step] = StepCompleted
	return nil
}

// SkipStep transitions a step from Ready to Skipped.
// Returns an error if the step is not in Ready status.
func (c *WorkflowCoordinator) SkipStep(sessionID string, step WorkflowStep) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	session := c.getOrCreate(sessionID)
	current := session.steps[step]

	switch current {
	case StepReady:
		session.steps[step] = StepSkipped
		return nil
	case StepInProgress:
		return fmt.Errorf("step %q in progress: cannot skip", step)
	case StepPending:
		return fmt.Errorf("step %q not ready: cannot skip", step)
	case StepCompleted:
		return fmt.Errorf("step %q already completed: cannot skip", step)
	case StepSkipped:
		return fmt.Errorf("step %q already skipped", step)
	}
	return nil // unreachable, but required for exhaustive switch
}

// ---------------------------------------------------------------------------
// Session Context Management
// ---------------------------------------------------------------------------

// SetSessionContext stores the session context (DomainModel, StackProfile, ProjectDir).
// Overwrites any existing context for the session.
func (c *WorkflowCoordinator) SetSessionContext(sessionID string, ctx *SessionContext) error {
	if ctx == nil {
		return fmt.Errorf("session context cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	session := c.getOrCreate(sessionID)
	session.context = ctx
	return nil
}

// SessionContext retrieves the session context.
// Returns an error if no context has been set for the session.
func (c *WorkflowCoordinator) SessionContext(sessionID string) (*SessionContext, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, ok := c.sessions[sessionID]
	if !ok || session.context == nil {
		return nil, fmt.Errorf("session context not found for %q", sessionID)
	}

	return session.context, nil
}

// ---------------------------------------------------------------------------
// Backward Compatibility — ReadyActions alias
// ---------------------------------------------------------------------------

// ReadyActions is an alias for AvailableActions for backward compatibility.
func (c *WorkflowCoordinator) ReadyActions(sessionID string) []ReadyAction {
	return c.AvailableActions(sessionID)
}

// MarkCompleted is provided for backward compatibility.
// It transitions a step directly to Completed status regardless of current state.
// Prefer using BeginStep + CompleteStep for proper state tracking.
func (c *WorkflowCoordinator) MarkCompleted(sessionID string, step WorkflowStep) {
	c.mu.Lock()
	defer c.mu.Unlock()

	session := c.getOrCreate(sessionID)
	session.steps[step] = StepCompleted
}
