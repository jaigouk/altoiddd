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

// FileWriter writes files to the filesystem. Shared kernel port used by
// multiple bounded contexts for writing generated artifacts to disk.
type FileWriter interface {
	// WriteFile writes content to a file at the given path.
	WriteFile(ctx context.Context, path string, content string) error
}
