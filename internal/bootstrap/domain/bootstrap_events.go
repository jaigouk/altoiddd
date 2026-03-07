// Package domain contains the Bootstrap bounded context domain model.
package domain

// BootstrapCompletedEvent is emitted when a bootstrap session completes execution.
type BootstrapCompletedEvent struct {
	sessionID  string
	projectDir string
}

// NewBootstrapCompletedEvent creates a BootstrapCompletedEvent.
func NewBootstrapCompletedEvent(sessionID, projectDir string) BootstrapCompletedEvent {
	return BootstrapCompletedEvent{sessionID: sessionID, projectDir: projectDir}
}

// SessionID returns the session identifier.
func (e BootstrapCompletedEvent) SessionID() string { return e.sessionID }

// ProjectDir returns the bootstrapped project directory.
func (e BootstrapCompletedEvent) ProjectDir() string { return e.projectDir }

// Equal returns true if two events have the same values.
func (e BootstrapCompletedEvent) Equal(other BootstrapCompletedEvent) bool {
	return e.sessionID == other.sessionID && e.projectDir == other.projectDir
}
