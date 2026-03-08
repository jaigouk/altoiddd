package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/alty-cli/alty/internal/shared/application"
)

// registration holds a handler, the event type name, and the Go type for unmarshaling.
type registration struct {
	handler   application.EventHandler
	eventType string
	goType    reflect.Type
}

// Subscriber implements the application.EventSubscriber port.
// It subscribes to GoChannel topics and dispatches deserialized events to handlers.
type Subscriber struct {
	bus           *Bus
	mu            sync.Mutex
	wg            sync.WaitGroup
	registrations []registration
}

// NewSubscriber creates a Subscriber backed by the given Bus.
func NewSubscriber(bus *Bus) *Subscriber {
	return &Subscriber{
		bus: bus,
	}
}

// Subscribe registers a handler for the given event type name.
// The event type name should match EventTypeName output (e.g., "events.DomainModelGenerated").
// Must be called before Start.
func (s *Subscriber) Subscribe(eventType string, handler application.EventHandler) error {
	if handler == nil {
		return fmt.Errorf("handler must not be nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registrations = append(s.registrations, registration{
		handler:   handler,
		eventType: eventType,
	})
	return nil
}

// SubscribeTyped registers a handler for a specific Go type. The event type name is
// derived from the type, and the payload is unmarshaled into a pointer of that type
// before being passed to the handler.
func SubscribeTyped[T any](s *Subscriber, handler func(ctx context.Context, event *T) error) error {
	var zero T
	eventType := EventTypeName(zero)
	goType := reflect.TypeOf(zero)

	return s.subscribeWithType(eventType, goType, func(ctx context.Context, event any) error {
		typed, ok := event.(*T)
		if !ok {
			return fmt.Errorf("unexpected event type: got %T, want *%s", event, eventType)
		}
		return handler(ctx, typed)
	})
}

func (s *Subscriber) subscribeWithType(
	eventType string,
	goType reflect.Type,
	handler application.EventHandler,
) error {
	if handler == nil {
		return fmt.Errorf("handler must not be nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registrations = append(s.registrations, registration{
		handler:   handler,
		eventType: eventType,
		goType:    goType,
	})
	return nil
}

// Start begins listening on all registered topics. Each registration gets its own
// goroutine that reads from the GoChannel subscription and invokes the handler.
// Errors from handlers are logged but do not stop the subscriber.
func (s *Subscriber) Start(ctx context.Context) error {
	s.mu.Lock()
	regs := make([]registration, len(s.registrations))
	copy(regs, s.registrations)
	s.mu.Unlock()

	for _, reg := range regs {
		msgs, err := s.bus.PubSub().Subscribe(ctx, reg.eventType)
		if err != nil {
			return fmt.Errorf("subscribing to %s: %w", reg.eventType, err)
		}
		s.wg.Add(1)
		go func(r registration) {
			defer s.wg.Done()
			s.listen(ctx, msgs, r)
		}(reg)
	}
	return nil
}

// Wait blocks until all listener goroutines have exited.
func (s *Subscriber) Wait() {
	s.wg.Wait()
}

// listen processes messages from a subscription channel and dispatches them to the handler.
func (s *Subscriber) listen(
	ctx context.Context,
	msgs <-chan *message.Message,
	reg registration,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			s.dispatch(ctx, msg, reg)
		}
	}
}

// dispatch unmarshals a message payload into the registered Go type and calls the handler.
// Messages are always Ack'd after processing (even on error) because with
// BlockPublishUntilSubscriberAck=true, Nack causes infinite redelivery which
// deadlocks in a synchronous CLI process. Handler errors are swallowed here;
// callers that need error visibility should use typed channels or return values.
func (s *Subscriber) dispatch(ctx context.Context, msg *message.Message, reg registration) {
	defer msg.Ack()

	var event any

	if reg.goType != nil {
		// Typed registration: unmarshal into the concrete type.
		ptr := reflect.New(reg.goType).Interface()
		if err := json.Unmarshal(msg.Payload, ptr); err != nil {
			return
		}
		event = ptr
	} else {
		// Untyped registration: pass raw JSON as a map.
		var m map[string]any
		if err := json.Unmarshal(msg.Payload, &m); err != nil {
			return
		}
		event = m
	}

	_ = reg.handler(ctx, event)
}
