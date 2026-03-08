package eventbus_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alty-cli/alty/internal/shared/application"
	"github.com/alty-cli/alty/internal/shared/infrastructure/eventbus"
)

// testEvent is a simple event used in tests.
type testEvent struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// anotherEvent is a different event type for routing tests.
type anotherEvent struct {
	Code int    `json:"code"`
	Info string `json:"info"`
}

// --- Compile-time interface checks ---

var (
	_ application.EventPublisher  = (*eventbus.Publisher)(nil)
	_ application.EventSubscriber = (*eventbus.Subscriber)(nil)
)

// --- Marshaler tests ---

func TestMarshal_Roundtrip(t *testing.T) {
	t.Parallel()

	m := eventbus.NewJSONMarshaler()
	original := testEvent{ID: "evt-1", Message: "hello world"}

	msg, err := m.Marshal("testEvent", original)
	require.NoError(t, err)
	assert.NotEmpty(t, msg.UUID)
	assert.Equal(t, "testEvent", msg.Metadata.Get(eventbus.MetadataEventType))

	var restored testEvent
	err = m.Unmarshal(msg, &restored)
	require.NoError(t, err)
	assert.Equal(t, original.ID, restored.ID)
	assert.Equal(t, original.Message, restored.Message)
}

func TestMarshal_PreservesAllFields(t *testing.T) {
	t.Parallel()

	m := eventbus.NewJSONMarshaler()
	original := anotherEvent{Code: 42, Info: "test info"}

	msg, err := m.Marshal("anotherEvent", original)
	require.NoError(t, err)

	// Verify payload is valid JSON with correct fields
	var raw map[string]any
	err = json.Unmarshal(msg.Payload, &raw)
	require.NoError(t, err)
	assert.InDelta(t, float64(42), raw["code"].(float64), 0)
	assert.Equal(t, "test info", raw["info"])
}

func TestMarshal_NilEvent_ReturnsError(t *testing.T) {
	t.Parallel()

	m := eventbus.NewJSONMarshaler()
	_, err := m.Marshal("nilEvent", nil)
	require.Error(t, err)
}

// --- Pub/Sub roundtrip tests ---

func TestPubSub_Roundtrip(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBus()
	defer bus.Close()

	pub := eventbus.NewPublisher(bus)
	sub := eventbus.NewSubscriber(bus)

	received := make(chan testEvent, 1)
	err := eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *testEvent) error {
		received <- *evt
		return nil
	})
	require.NoError(t, err)

	err = sub.Start(context.Background())
	require.NoError(t, err)

	ctx := context.Background()
	err = pub.Publish(ctx, testEvent{ID: "roundtrip-1", Message: "hello"})
	require.NoError(t, err)

	select {
	case evt := <-received:
		assert.Equal(t, "roundtrip-1", evt.ID)
		assert.Equal(t, "hello", evt.Message)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestPubSub_MultipleSubscribers_FanOut(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBus()
	defer bus.Close()

	pub := eventbus.NewPublisher(bus)
	sub1 := eventbus.NewSubscriber(bus)
	sub2 := eventbus.NewSubscriber(bus)

	var mu sync.Mutex
	var received []string

	makeHandler := func(id string) func(context.Context, *testEvent) error {
		return func(ctx context.Context, evt *testEvent) error {
			mu.Lock()
			defer mu.Unlock()
			received = append(received, id)
			return nil
		}
	}

	err := eventbus.SubscribeTyped(sub1, makeHandler("sub1"))
	require.NoError(t, err)
	err = eventbus.SubscribeTyped(sub2, makeHandler("sub2"))
	require.NoError(t, err)

	err = sub1.Start(context.Background())
	require.NoError(t, err)
	err = sub2.Start(context.Background())
	require.NoError(t, err)

	err = pub.Publish(context.Background(), testEvent{ID: "fanout-1", Message: "broadcast"})
	require.NoError(t, err)

	// Wait for both subscribers to process
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(received) == 2
	}, 2*time.Second, 10*time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Contains(t, received, "sub1")
	assert.Contains(t, received, "sub2")
}

func TestPubSub_EventTypeRouting(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBus()
	defer bus.Close()

	pub := eventbus.NewPublisher(bus)
	sub := eventbus.NewSubscriber(bus)

	testReceived := make(chan struct{}, 1)
	anotherReceived := make(chan struct{}, 1)

	err := eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *testEvent) error {
		testReceived <- struct{}{}
		return nil
	})
	require.NoError(t, err)

	err = eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *anotherEvent) error {
		anotherReceived <- struct{}{}
		return nil
	})
	require.NoError(t, err)

	err = sub.Start(context.Background())
	require.NoError(t, err)

	// Publish only a testEvent
	err = pub.Publish(context.Background(), testEvent{ID: "route-1", Message: "test"})
	require.NoError(t, err)

	select {
	case <-testReceived:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("testEvent handler did not receive event")
	}

	// anotherEvent handler should NOT have been called
	select {
	case <-anotherReceived:
		t.Fatal("anotherEvent handler should not have received testEvent")
	case <-time.After(200 * time.Millisecond):
		// expected: no event for this handler
	}
}

func TestPubSub_SubscriberError_DoesNotCrashBus(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBus()
	defer bus.Close()

	pub := eventbus.NewPublisher(bus)
	sub := eventbus.NewSubscriber(bus)

	callCount := 0
	var mu sync.Mutex
	received := make(chan string, 2)

	err := eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *testEvent) error {
		mu.Lock()
		callCount++
		count := callCount
		mu.Unlock()

		if count == 1 {
			received <- evt.ID
			return errors.New("handler error")
		}
		received <- evt.ID
		return nil
	})
	require.NoError(t, err)

	err = sub.Start(context.Background())
	require.NoError(t, err)

	// First publish -- handler returns error (but message is Ack'd to avoid redelivery)
	err = pub.Publish(context.Background(), testEvent{ID: "err-1", Message: "fail"})
	require.NoError(t, err)

	// Second publish -- should still work (bus not crashed)
	err = pub.Publish(context.Background(), testEvent{ID: "err-2", Message: "succeed"})
	require.NoError(t, err)

	// Both events should be received (handler error doesn't crash the bus)
	for i := 0; i < 2; i++ {
		select {
		case <-received:
			// OK
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for event %d", i+1)
		}
	}
}

func TestPubSub_UntypedSubscribe_ReceivesMap(t *testing.T) {
	t.Parallel()

	bus := eventbus.NewBus()
	defer bus.Close()

	pub := eventbus.NewPublisher(bus)
	sub := eventbus.NewSubscriber(bus)

	received := make(chan map[string]any, 1)
	eventType := eventbus.EventTypeName(testEvent{})
	err := sub.Subscribe(eventType, func(ctx context.Context, event any) error {
		m, ok := event.(map[string]any)
		require.True(t, ok, "expected map[string]any, got %T", event)
		received <- m
		return nil
	})
	require.NoError(t, err)

	err = sub.Start(context.Background())
	require.NoError(t, err)

	err = pub.Publish(context.Background(), testEvent{ID: "untyped-1", Message: "raw"})
	require.NoError(t, err)

	select {
	case m := <-received:
		assert.Equal(t, "untyped-1", m["id"])
		assert.Equal(t, "raw", m["message"])
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for untyped event")
	}
}

func TestEventTypeName_ReturnsFullyQualifiedName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    any
		expected string
	}{
		{"struct value", testEvent{}, "eventbus_test.testEvent"},
		{"struct pointer", &testEvent{}, "eventbus_test.testEvent"},
		{"another type", anotherEvent{}, "eventbus_test.anotherEvent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := eventbus.EventTypeName(tt.event)
			assert.Equal(t, tt.expected, got)
		})
	}
}
