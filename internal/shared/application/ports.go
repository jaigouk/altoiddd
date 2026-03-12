// Package application defines ports (interfaces) for the shared application layer.
package application

import "context"

// EventPublisher publishes domain events to the event bus.
type EventPublisher interface {
	Publish(ctx context.Context, event any) error
}

// EventHandler handles a domain event of a specific type.
type EventHandler func(ctx context.Context, event any) error

// EventSubscriber subscribes to domain events by type name.
type EventSubscriber interface {
	Subscribe(eventType string, handler EventHandler) error
}

// FileReader reads files from the filesystem. Shared kernel port used by
// multiple bounded contexts for reading existing artifacts.
type FileReader interface {
	// ReadFile reads content from a file at the given path.
	ReadFile(ctx context.Context, path string) (string, error)
}

// FileWriter writes files to the filesystem. Shared kernel port used by
// multiple bounded contexts for writing generated artifacts to disk.
type FileWriter interface {
	// WriteFile writes content to a file at the given path.
	WriteFile(ctx context.Context, path string, content string) error
}

// ReadinessQuerier queries workflow readiness state for a session.
// Used by MCP/CLI to determine what actions are available next.
type ReadinessQuerier interface {
	// ReadyActions returns the list of actions currently available for the session.
	// Returns an empty slice for unknown sessions.
	ReadyActions(sessionID string) []ReadyAction
}

// ReadyAction represents an action that is available for the user to take.
// This is a simplified interface matching the domain ReadyAction value object.
type ReadyAction interface {
	Name() string
}
