// Package domain contains the Bootstrap bounded context domain model.
package domain

import (
	"encoding/json"
	"fmt"
)

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

// MarshalJSON implements json.Marshaler for event bus serialization.
func (e BootstrapCompletedEvent) MarshalJSON() ([]byte, error) {
	type proxy struct {
		SessionID  string `json:"session_id"`
		ProjectDir string `json:"project_dir"`
	}
	data, err := json.Marshal(proxy{
		SessionID:  e.sessionID,
		ProjectDir: e.projectDir,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling BootstrapCompletedEvent: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler for event bus deserialization.
func (e *BootstrapCompletedEvent) UnmarshalJSON(data []byte) error {
	type proxy struct {
		SessionID  string `json:"session_id"`
		ProjectDir string `json:"project_dir"`
	}
	var p proxy
	if err := json.Unmarshal(data, &p); err != nil {
		return fmt.Errorf("unmarshaling BootstrapCompletedEvent: %w", err)
	}
	e.sessionID = p.SessionID
	e.projectDir = p.ProjectDir
	return nil
}
