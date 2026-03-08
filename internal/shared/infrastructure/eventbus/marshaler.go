// Package eventbus provides an in-process event bus backed by Watermill GoChannel.
package eventbus

import (
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

// MetadataEventType is the metadata key used to store the event type name.
const MetadataEventType = "event_type"

// JSONMarshaler handles JSON serialization of domain events into Watermill messages.
type JSONMarshaler struct{}

// NewJSONMarshaler creates a new JSONMarshaler.
func NewJSONMarshaler() *JSONMarshaler {
	return &JSONMarshaler{}
}

// Marshal converts a domain event into a Watermill message with JSON payload.
func (m *JSONMarshaler) Marshal(eventType string, event any) (*message.Message, error) {
	if event == nil {
		return nil, fmt.Errorf("cannot marshal nil event")
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("marshaling event %q: %w", eventType, err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set(MetadataEventType, eventType)

	return msg, nil
}

// Unmarshal deserializes a Watermill message payload into the given target.
func (m *JSONMarshaler) Unmarshal(msg *message.Message, target any) error {
	if err := json.Unmarshal(msg.Payload, target); err != nil {
		return fmt.Errorf("unmarshaling event: %w", err)
	}
	return nil
}
