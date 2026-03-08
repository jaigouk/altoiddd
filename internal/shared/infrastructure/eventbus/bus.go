package eventbus

import (
	"fmt"
	"reflect"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// Bus wraps a Watermill GoChannel pub/sub backend for in-process event routing.
type Bus struct {
	pubsub *gochannel.GoChannel
}

// NewBus creates a new Bus with a GoChannel backend configured for synchronous CLI use.
func NewBus() *Bus {
	pubsub := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer:            0,
			Persistent:                     false,
			BlockPublishUntilSubscriberAck: true,
		},
		watermill.NopLogger{},
	)
	return &Bus{pubsub: pubsub}
}

// PubSub returns the underlying GoChannel instance.
// Both publisher and subscriber must use the same instance.
func (b *Bus) PubSub() *gochannel.GoChannel {
	return b.pubsub
}

// Close shuts down the GoChannel backend.
func (b *Bus) Close() error {
	if err := b.pubsub.Close(); err != nil {
		return fmt.Errorf("closing event bus: %w", err)
	}
	return nil
}

// EventTypeName returns the fully qualified type name for an event value.
// Pointer types are dereferenced. Example: "events.DomainModelGenerated".
// Returns empty string for nil events.
func EventTypeName(event any) string {
	t := reflect.TypeOf(event)
	if t == nil {
		return ""
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath()[lastSlash(t.PkgPath())+1:] + "." + t.Name()
}

func lastSlash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}
