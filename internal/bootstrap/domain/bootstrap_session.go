package domain

import (
	"fmt"

	"github.com/google/uuid"

	domainerrors "github.com/alty-cli/alty/internal/shared/domain/errors"
	vo "github.com/alty-cli/alty/internal/shared/domain/valueobjects"
)

// SessionStatus represents states in the bootstrap session lifecycle.
type SessionStatus string

// SessionStatus constants.
const (
	SessionStatusCreated   SessionStatus = "created"
	SessionStatusPreviewed SessionStatus = "previewed"
	SessionStatusConfirmed SessionStatus = "confirmed"
	SessionStatusExecuting SessionStatus = "executing"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusCancelled SessionStatus = "cancelled"
)

// BootstrapSession is the aggregate root for the bootstrap flow.
// Enforces the preview-before-action invariant.
type BootstrapSession struct {
	sessionID     string
	projectDir    string
	status        SessionStatus
	preview       *vo.Preview
	detectedTools []string
	events        []BootstrapCompletedEvent
}

// NewBootstrapSession creates a new session in CREATED state.
func NewBootstrapSession(projectDir string) *BootstrapSession {
	return &BootstrapSession{
		sessionID:  uuid.New().String(),
		projectDir: projectDir,
		status:     SessionStatusCreated,
	}
}

// SessionID returns the unique session identifier.
func (s *BootstrapSession) SessionID() string { return s.sessionID }

// ProjectDir returns the project directory being bootstrapped.
func (s *BootstrapSession) ProjectDir() string { return s.projectDir }

// Status returns the current session state.
func (s *BootstrapSession) Status() SessionStatus { return s.status }

// Preview returns the current preview, or nil if not set.
func (s *BootstrapSession) Preview() *vo.Preview { return s.preview }

// DetectedTools returns a defensive copy of detected tool names.
func (s *BootstrapSession) DetectedTools() []string {
	out := make([]string, len(s.detectedTools))
	copy(out, s.detectedTools)
	return out
}

// SetDetectedTools records which AI coding tools were found.
func (s *BootstrapSession) SetDetectedTools(tools []string) {
	s.detectedTools = make([]string, len(tools))
	copy(s.detectedTools, tools)
}

// Events returns a defensive copy of domain events.
func (s *BootstrapSession) Events() []BootstrapCompletedEvent {
	out := make([]BootstrapCompletedEvent, len(s.events))
	copy(out, s.events)
	return out
}

// SetPreview sets or replaces the preview.
func (s *BootstrapSession) SetPreview(preview *vo.Preview) error {
	if s.status != SessionStatusCreated && s.status != SessionStatusPreviewed {
		return fmt.Errorf("cannot preview in %s state: %w", s.status, domainerrors.ErrInvariantViolation)
	}
	s.preview = preview
	s.status = SessionStatusPreviewed
	return nil
}

// Confirm confirms the preview, allowing execution.
func (s *BootstrapSession) Confirm() error {
	if s.status != SessionStatusPreviewed {
		return fmt.Errorf("cannot confirm without preview: %w", domainerrors.ErrInvariantViolation)
	}
	s.status = SessionStatusConfirmed
	return nil
}

// Cancel cancels the session after preview.
func (s *BootstrapSession) Cancel() error {
	if s.status != SessionStatusPreviewed {
		return fmt.Errorf("can only cancel from previewed state: %w", domainerrors.ErrInvariantViolation)
	}
	s.status = SessionStatusCancelled
	return nil
}

// BeginExecution transitions to EXECUTING state.
func (s *BootstrapSession) BeginExecution() error {
	if s.status != SessionStatusConfirmed {
		return fmt.Errorf("cannot execute without confirmation: %w", domainerrors.ErrInvariantViolation)
	}
	s.status = SessionStatusExecuting
	return nil
}

// Complete marks execution as completed and emits BootstrapCompleted event.
func (s *BootstrapSession) Complete() error {
	if s.status != SessionStatusExecuting {
		return fmt.Errorf("cannot complete unless executing: %w", domainerrors.ErrInvariantViolation)
	}
	s.status = SessionStatusCompleted
	s.events = append(s.events, NewBootstrapCompletedEvent(s.sessionID, s.projectDir))
	return nil
}
