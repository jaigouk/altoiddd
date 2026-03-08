package eventbus

import (
	"context"
	"fmt"
)

// Publisher implements the application.EventPublisher port using a Watermill GoChannel backend.
type Publisher struct {
	bus       *Bus
	marshaler *JSONMarshaler
}

// NewPublisher creates a Publisher that writes to the given Bus.
func NewPublisher(bus *Bus) *Publisher {
	return &Publisher{
		bus:       bus,
		marshaler: NewJSONMarshaler(),
	}
}

// Publish marshals the event to JSON and publishes it to the topic derived from the event type.
func (p *Publisher) Publish(ctx context.Context, event any) error {
	if event == nil {
		return fmt.Errorf("cannot publish nil event")
	}

	eventType := EventTypeName(event)
	msg, err := p.marshaler.Marshal(eventType, event)
	if err != nil {
		return fmt.Errorf("publishing %s: %w", eventType, err)
	}
	msg.SetContext(ctx)

	if err := p.bus.PubSub().Publish(eventType, msg); err != nil {
		return fmt.Errorf("publishing %s to bus: %w", eventType, err)
	}

	return nil
}
