# Watermill CQRS Deep Dive: Patterns, GoChannel, and DDD Integration

**Date:** 2026-03-07
**Type:** Spike Research (Deep Dive)
**Status:** Complete
**Builds on:** `docs/research/20260306_2_go_ddd_event_systems_messaging.md`

## Research Questions

1. How is Watermill's Message structured and how do messages flow through the system?
2. How is CQRS set up (CommandBus, EventBus, handlers, processors)?
3. What are the GoChannel backend specifics for local/CLI use?
4. What middleware patterns are available (retry, recovery, correlation)?
5. How to integrate Watermill with DDD domain events?
6. What does the migration path from GoChannel to NATS look like?
7. What best practices do ThreeDotsLabs recommend for DDD in Go?

---

## 1. Message Structure and Flow

### Message Anatomy

A Watermill `message.Message` has three components:

```go
type Message struct {
    UUID     string            // Unique ID for debugging (optional, can be empty)
    Metadata Metadata          // Key-value headers (like HTTP headers)
    Payload  []byte            // Actual content (JSON, Protobuf, etc.)
    // ... internal: ack/nack channels, context
}
```

Source: [Watermill Message Docs](https://watermill.io/docs/message/)

**UUID** -- Used for debugging and deduplication. Generated via `watermill.NewUUID()` (returns UUID v4 string). Not required to be globally unique across all systems.

**Metadata** -- Key-value map (`map[string]string`) persisted alongside the message in the pub/sub backend. Analogous to HTTP headers. Use cases:
- Correlation IDs for tracing
- Partition keys for Kafka
- Timestamps
- Content type indicators
- Any data that does not require unmarshaling the full payload

**Payload** -- Raw bytes. Watermill is format-agnostic: JSON, Protobuf, MessagePack, or any serialization. The CQRS component's `Marshaler` handles conversion between Go structs and `[]byte`.

### Creating Messages

```go
// Low-level (raw message)
msg := message.NewMessage(watermill.NewUUID(), []byte(`{"user_id": "123"}`))
msg.Metadata.Set("correlation_id", "abc-123")

// v1.5+ context-aware creation
msg := message.NewMessageWithContext(ctx, watermill.NewUUID(), payload)
```

Source: [Watermill Getting Started](https://watermill.io/learn/getting-started/)

### Acknowledgment Protocol

Messages use an explicit ack/nack system:

```go
// In a subscriber loop
for msg := range messages {
    err := processMessage(msg)
    if err != nil {
        msg.Nack()  // Signal failure -- message will be redelivered
        continue
    }
    msg.Ack()  // Signal success -- message is consumed
}

// In a publisher waiting for confirmation
select {
case <-msg.Acked():
    log.Info("Message acknowledged")
case <-msg.Nacked():
    log.Error("Message rejected")
}
```

Key properties:
- `Ack()` and `Nack()` are **non-blocking** and **idempotent**
- `Ack()` returns `false` if `Nack()` was already sent (and vice versa)
- First call wins -- subsequent calls to either are no-ops
- Default context is `context.Background()` if none set

Source: [Watermill Message Docs](https://watermill.io/docs/message/)

### Three-Layer Architecture

Watermill provides three progressively higher-level APIs:

```
Layer 3:  CQRS Component (CommandBus, EventBus, typed handlers)
             |
Layer 2:  Router + Middleware (routing, retry, correlation, recovery)
             |
Layer 1:  Publisher/Subscriber (raw message send/receive)
             |
Backend:  GoChannel | NATS | Kafka | SQL | SQLite | Redis | ...
```

Each layer builds on the one below. For alto, we primarily use Layer 3 (CQRS) which internally manages Layers 2 and 1.

Source: [Watermill Getting Started](https://watermill.io/learn/getting-started/)

---

## 2. CQRS Setup: CommandBus + EventBus + Handlers

### Overview

Watermill's CQRS component (`github.com/ThreeDotsLabs/watermill/components/cqrs`) provides typed command and event handling on top of the raw pub/sub layer. The key insight: **you work with Go structs, not raw messages.**

Source: [Watermill CQRS Docs](https://watermill.io/docs/cqrs/)

### Component Roles

| Component | Role | Cardinality |
|-----------|------|-------------|
| `CommandBus` | Sends commands to handlers | N publishers : 1 handler per command type |
| `CommandProcessor` | Routes received commands to handlers | 1 per application |
| `EventBus` | Publishes events | N publishers |
| `EventProcessor` | Routes events to handlers | 1 per application |
| `EventGroupProcessor` | Groups handlers sharing one subscriber | Optional, for ordering |
| `CommandEventMarshaler` | Serializes/deserializes Go structs | 1 per application |

### CommandBus Configuration

```go
commandBus, err := cqrs.NewCommandBusWithConfig(
    publisher,  // message.Publisher (GoChannel, NATS, etc.)
    cqrs.CommandBusConfig{
        GeneratePublishTopic: func(params cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
            // Generate topic name from command type
            return "commands." + params.CommandName, nil
        },
        OnSend: func(params cqrs.CommandBusOnSendParams) error {
            // Optional: enrich message metadata before sending
            params.Message.Metadata.Set("sent_at", time.Now().String())
            return nil
        },
        Marshaler: cqrs.JSONMarshaler{},
        Logger:    logger,
    },
)
```

**Sending a command:**

```go
err := commandBus.Send(ctx, &GeneratePRD{
    ProjectID: "proj-123",
    Answers:   answers,
})
```

Source: [pkg.go.dev cqrs package](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/cqrs)

### CommandProcessor Configuration

```go
commandProcessor, err := cqrs.NewCommandProcessorWithConfig(
    router,  // message.Router
    cqrs.CommandProcessorConfig{
        GenerateSubscribeTopic: func(params cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
            return "commands." + params.CommandName, nil
        },
        SubscriberConstructor: func(params cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
            return pubSub, nil  // Return the same GoChannel instance
        },
        Marshaler: cqrs.JSONMarshaler{},
        Logger:    logger,
    },
)
```

Source: [pkg.go.dev cqrs package](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/cqrs)

### EventBus Configuration

```go
eventBus, err := cqrs.NewEventBusWithConfig(
    publisher,
    cqrs.EventBusConfig{
        GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
            return "events." + params.EventName, nil
        },
        OnPublish: func(params cqrs.OnEventSendParams) error {
            params.Message.Metadata.Set("published_at", time.Now().String())
            return nil
        },
        Marshaler: cqrs.JSONMarshaler{},
        Logger:    logger,
    },
)
```

**Publishing an event:**

```go
err := eventBus.Publish(ctx, &PRDGenerated{
    ProjectID: "proj-123",
    PRDPath:   "/path/to/PRD.md",
    Timestamp: time.Now(),
})
```

Source: [Watermill CQRS Docs](https://watermill.io/docs/cqrs/)

### EventProcessor Configuration

```go
eventProcessor, err := cqrs.NewEventProcessorWithConfig(
    router,
    cqrs.EventProcessorConfig{
        GenerateSubscribeTopic: func(params cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
            return "events." + params.EventName, nil
        },
        SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
            return pubSub, nil
        },
        AckOnUnknownEvent: true,  // Don't fail on events with no handler
        Marshaler:         cqrs.JSONMarshaler{},
        Logger:            logger,
    },
)
```

Source: [pkg.go.dev cqrs package](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/cqrs)

### Handler Registration (Generic API, v1.3+)

**Command Handlers** -- exactly one handler per command type:

```go
err := commandProcessor.AddHandlers(
    cqrs.NewCommandHandler[GeneratePRD](
        "GeneratePRDHandler",
        func(ctx context.Context, cmd *GeneratePRD) error {
            prd, err := prdService.Generate(cmd.Answers)
            if err != nil {
                return err
            }
            return eventBus.Publish(ctx, &PRDGenerated{
                ProjectID: cmd.ProjectID,
                PRDPath:   prd.Path,
            })
        },
    ),
    cqrs.NewCommandHandler[GenerateTickets](
        "GenerateTicketsHandler",
        func(ctx context.Context, cmd *GenerateTickets) error {
            // ...
            return nil
        },
    ),
)
```

**Event Handlers** -- multiple handlers per event type:

```go
err := eventProcessor.AddHandlers(
    // Handler 1: Generate tickets when PRD is ready
    cqrs.NewEventHandler[PRDGenerated](
        "OnPRDGenerated_GenerateTickets",
        func(ctx context.Context, event *PRDGenerated) error {
            return commandBus.Send(ctx, &GenerateTickets{
                ProjectID: event.ProjectID,
            })
        },
    ),
    // Handler 2: Log PRD generation
    cqrs.NewEventHandler[PRDGenerated](
        "OnPRDGenerated_AuditLog",
        func(ctx context.Context, event *PRDGenerated) error {
            log.Info("PRD generated", "project", event.ProjectID)
            return nil
        },
    ),
)
```

Source: [Watermill CQRS Docs](https://watermill.io/docs/cqrs/),
[pkg.go.dev cqrs package](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/cqrs)

### EventGroupProcessor (Ordered Processing)

When multiple event handlers must process events in order (e.g., building a read model), use `EventGroupProcessor`:

```go
groupProcessor, err := cqrs.NewEventGroupProcessorWithConfig(
    router,
    cqrs.EventGroupProcessorConfig{
        GenerateSubscribeTopic: func(params cqrs.EventGroupProcessorGenerateSubscribeTopicParams) (string, error) {
            return "events." + params.EventName, nil
        },
        SubscriberConstructor: func(params cqrs.EventGroupProcessorSubscriberConstructorParams) (message.Subscriber, error) {
            return pubSub, nil
        },
        Marshaler: cqrs.JSONMarshaler{},
    },
)

err = groupProcessor.AddHandlersGroup(
    "bootstrap_progress",
    cqrs.NewGroupEventHandler[PRDGenerated](
        func(ctx context.Context, event *PRDGenerated) error { /* ... */ },
    ),
    cqrs.NewGroupEventHandler[DDDArtifactsGenerated](
        func(ctx context.Context, event *DDDArtifactsGenerated) error { /* ... */ },
    ),
)
```

All handlers in a group share a single subscriber instance, guaranteeing ordered processing across event types on one topic.

Source: [Watermill CQRS Docs](https://watermill.io/docs/cqrs/)

### Marshaler Options

| Marshaler | Format | Use Case | Status |
|-----------|--------|----------|--------|
| `JSONMarshaler` | JSON | Default, human-readable, debugging | Recommended |
| `ProtoMarshaler` | Protocol Buffers | Performance, schema evolution | v1.4.4+ |
| `ProtobufMarshaler` | Protocol Buffers (old) | Legacy | **Deprecated** (use `ProtoMarshaler`) |
| `CommandEventMarshalerDecorator` | Wraps any marshaler | Add metadata (partition keys, etc.) | v1.5.1+ |

For alto, `JSONMarshaler` is the clear choice -- human-readable payloads, simple debugging, no protobuf codegen needed.

Source: [pkg.go.dev cqrs package](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/cqrs)

### Utility Functions

```go
// Get struct name for topic generation
cqrs.StructName(PRDGenerated{})           // "PRDGenerated"
cqrs.FullyQualifiedStructName(PRDGenerated{}) // "events.PRDGenerated"

// Access original Watermill message from handler context
msg := cqrs.OriginalMessageFromCtx(ctx)   // v1.3.5+
correlationID := msg.Metadata.Get("correlation_id")
```

Source: [pkg.go.dev cqrs package](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/cqrs)

### Complete Wiring Example

```go
func setupCQRS(logger watermill.LoggerAdapter) (*cqrs.CommandBus, *cqrs.EventBus, *message.Router, error) {
    // 1. Create GoChannel pub/sub
    pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)

    // 2. Create router with middleware
    router, err := message.NewRouter(message.RouterConfig{}, logger)
    if err != nil {
        return nil, nil, nil, err
    }
    router.AddPlugin(plugin.SignalsHandler)
    router.AddMiddleware(
        middleware.Recoverer,
        middleware.CorrelationID,
    )

    // 3. Create marshaler
    marshaler := cqrs.JSONMarshaler{}

    // 4. Create command bus + processor
    commandBus, err := cqrs.NewCommandBusWithConfig(pubSub, cqrs.CommandBusConfig{
        GeneratePublishTopic: func(p cqrs.CommandBusGeneratePublishTopicParams) (string, error) {
            return "commands." + p.CommandName, nil
        },
        Marshaler: marshaler,
        Logger:    logger,
    })
    if err != nil {
        return nil, nil, nil, err
    }

    commandProcessor, err := cqrs.NewCommandProcessorWithConfig(router, cqrs.CommandProcessorConfig{
        GenerateSubscribeTopic: func(p cqrs.CommandProcessorGenerateSubscribeTopicParams) (string, error) {
            return "commands." + p.CommandName, nil
        },
        SubscriberConstructor: func(p cqrs.CommandProcessorSubscriberConstructorParams) (message.Subscriber, error) {
            return pubSub, nil
        },
        Marshaler: marshaler,
        Logger:    logger,
    })
    if err != nil {
        return nil, nil, nil, err
    }

    // 5. Create event bus + processor
    eventBus, err := cqrs.NewEventBusWithConfig(pubSub, cqrs.EventBusConfig{
        GeneratePublishTopic: func(p cqrs.GenerateEventPublishTopicParams) (string, error) {
            return "events." + p.EventName, nil
        },
        Marshaler: marshaler,
        Logger:    logger,
    })
    if err != nil {
        return nil, nil, nil, err
    }

    eventProcessor, err := cqrs.NewEventProcessorWithConfig(router, cqrs.EventProcessorConfig{
        GenerateSubscribeTopic: func(p cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
            return "events." + p.EventName, nil
        },
        SubscriberConstructor: func(p cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
            return pubSub, nil
        },
        AckOnUnknownEvent: true,
        Marshaler:         marshaler,
        Logger:            logger,
    })
    if err != nil {
        return nil, nil, nil, err
    }

    // 6. Register handlers (see examples in section 5)
    _ = commandProcessor  // Register handlers via commandProcessor.AddHandlers(...)
    _ = eventProcessor    // Register handlers via eventProcessor.AddHandlers(...)

    return commandBus, eventBus, router, nil
}
```

Source: Synthesized from [Watermill CQRS Docs](https://watermill.io/docs/cqrs/),
[Getting Started](https://watermill.io/learn/getting-started/),
[pkg.go.dev](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/cqrs)

---

## 3. GoChannel Backend Specifics

### Configuration

```go
type Config struct {
    // OutputChannelBuffer sets the buffer size for output channels.
    // Default: 0 (unbuffered)
    OutputChannelBuffer int64

    // Persistent: if true, subscribers receive all previously published messages.
    // Messages are stored in memory (a simple slice).
    // WARNING: large message volumes can cause OOM.
    // NOTE: When persistent is true, message ordering is NOT guaranteed.
    // Default: false
    Persistent bool

    // BlockPublishUntilSubscriberAck: if true, Publish blocks until subscriber
    // calls Ack(). If no subscribers exist, Publish does NOT block
    // (also when Persistent is true).
    // Default: false
    BlockPublishUntilSubscriberAck bool
}
```

Source: [GoChannel source](https://github.com/ThreeDotsLabs/watermill/blob/master/pubsub/gochannel/pubsub.go),
[pkg.go.dev gochannel](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill/pubsub/gochannel)

### Feature Matrix

| Feature | GoChannel | Notes |
|---------|-----------|-------|
| Consumer Groups | No | All subscribers get all messages |
| Exactly-Once Delivery | Yes | In-process guarantee |
| Guaranteed Order | Yes* | *Only when `Persistent` is `false` |
| Persistence | No* | *Optional in-memory persistence (not disk) |
| Marshaling Required | No | Messages stay in same process memory |
| Blocking Publish | No* | *Unless `BlockPublishUntilSubscriberAck` is `true` |

Source: [Watermill GoChannel Docs](https://watermill.io/pubsubs/gochannel/)

### Key Behaviors

1. **Same instance required.** GoChannel has no global state. You must use the same `*GoChannel` instance for publishing and subscribing. This is a common gotcha.

2. **Non-persistent by default.** If a message is published with no active subscribers, it is **discarded**. For alto's CLI, this is fine -- events are published and consumed in the same process during a single run.

3. **Non-blocking publish.** `Publish()` sends messages to subscribers in background goroutines. The call returns immediately.

4. **No consumer groups.** Every subscriber receives every message. This matches DDD event fan-out (multiple handlers per event type).

5. **v1.5 addition: PreserveContext.** GoChannel now supports `PreserveContext` configuration to propagate context from publisher to subscriber. Use with caution -- production pub/subs do NOT propagate context (it would cross process boundaries).

Source: [Watermill GoChannel Docs](https://watermill.io/pubsubs/gochannel/),
[Watermill 1.5 Release](https://threedots.tech/post/watermill-1-5/)

### Recommended Config for alto CLI

```go
pubSub := gochannel.NewGoChannel(
    gochannel.Config{
        // Unbuffered -- handler processes before next message
        OutputChannelBuffer: 0,
        // No persistence -- CLI runs are short-lived
        Persistent: false,
        // Non-blocking -- let handlers run concurrently
        BlockPublishUntilSubscriberAck: false,
    },
    watermill.NewSlogLogger(slog.Default()),
)
```

For testing, consider `BlockPublishUntilSubscriberAck: true` to make event flow deterministic.

### Performance

GoChannel achieves **315,776 msg/s publish** and **138,743 msg/s subscribe** in Watermill benchmarks. This is orders of magnitude beyond what alto needs (tens of events per CLI session).

Source: [Watermill Benchmarks](https://watermill.io/)

---

## 4. Middleware Patterns

### Adding Middleware

Middleware can be applied at two levels:

```go
// Router-level: applies to ALL handlers
router.AddMiddleware(
    middleware.Recoverer,
    middleware.CorrelationID,
    middleware.Retry{MaxRetries: 3}.Middleware,
)

// Handler-level: applies to ONE specific handler
handler, _ := router.AddHandler(/* ... */)
handler.AddMiddleware(mySpecificMiddleware)
```

Source: [Watermill Middleware Docs](https://watermill.io/docs/middlewares/)

### Available Middleware

| Middleware | Purpose | alto Relevance |
|-----------|---------|----------------|
| **Recoverer** | Captures panics, converts to errors with stack trace | HIGH -- always use |
| **CorrelationID** | Propagates correlation ID across messages | HIGH -- traces bootstrap session |
| **Retry** | Exponential backoff retry on handler error | MEDIUM -- useful for LLM calls |
| **Timeout** | Sets context deadline on handler execution | MEDIUM -- prevents hung handlers |
| **Circuit Breaker** | Stops calling failing handlers | LOW -- CLI is short-lived |
| **Throttle** | Rate limits message processing | LOW -- not needed for CLI |
| **Deduplicator** | Removes duplicate messages (Adler-32 or SHA-256) | LOW -- GoChannel has exactly-once |
| **Poison Queue** | Routes unprocessable messages to dead letter topic | MEDIUM -- useful for debugging |
| **Delay On Error** | Adds exponential backoff metadata | LOW -- requires compatible pub/sub |
| **Instant Ack** | Acks immediately, processes in background | LOW -- trades delivery guarantee |
| **Duplicator** | Processes messages twice (chaos testing) | LOW -- testing only |
| **Randomfail/Panic** | Random failures (chaos testing) | LOW -- testing only |

Source: [Watermill Middleware Docs](https://watermill.io/docs/middlewares/)

### Recommended Middleware Stack for alto

```go
router.AddMiddleware(
    // 1. Recover from panics -- always outermost
    middleware.Recoverer,

    // 2. Propagate correlation ID (bootstrap session ID)
    middleware.CorrelationID,

    // 3. Retry with backoff for LLM-dependent handlers
    middleware.Retry{
        MaxRetries:      3,
        InitialInterval: 500 * time.Millisecond,
        MaxInterval:     5 * time.Second,
        Multiplier:      2.0,
        Logger:          logger,
    }.Middleware,

    // 4. Timeout to prevent hung handlers (especially LLM calls)
    middleware.Timeout(30 * time.Second),
)
```

### Custom Middleware Pattern

Watermill middleware follows a simple function signature:

```go
func MyMiddleware(next message.HandlerFunc) message.HandlerFunc {
    return func(msg *message.Message) ([]*message.Message, error) {
        // Before handler
        startTime := time.Now()
        msg.Metadata.Set("processing_started", startTime.String())

        // Call next handler
        msgs, err := next(msg)

        // After handler
        duration := time.Since(startTime)
        log.Info("Handler completed", "duration", duration, "error", err)

        return msgs, err
    }
}
```

Source: [Watermill Middleware Docs](https://watermill.io/docs/middlewares/)

---

## 5. DDD Integration: Domain Events with Watermill

### Design Principle

ThreeDotsLabs' core principle for DDD in Go:

> "Always keep a valid state in the memory. All fields are private, validation occurs in constructors, and methods reflect business behaviors rather than data operations."

Source: [Combining DDD, CQRS, and Clean Architecture](https://threedots.tech/post/ddd-cqrs-clean-architecture-combined/)

### Domain Event Design for alto

Based on alto's DDD model (from `docs/DDD.md`), here are the domain events mapped to Watermill:

#### Bootstrap Context Events

```go
// events/bootstrap.go
package events

import "time"

// BootstrapStarted is published when a user runs `alto init`
type BootstrapStarted struct {
    SessionID string    `json:"session_id"`
    ProjectDir string   `json:"project_dir"`
    IsExisting bool     `json:"is_existing"`  // true for --existing
    Timestamp  time.Time `json:"timestamp"`
}

// BootstrapCompleted is published when the full bootstrap pipeline finishes
type BootstrapCompleted struct {
    SessionID    string    `json:"session_id"`
    ProjectDir   string    `json:"project_dir"`
    ArtifactPaths []string `json:"artifact_paths"`
    Timestamp    time.Time `json:"timestamp"`
}
```

#### Guided Discovery Context Events

```go
// events/discovery.go
package events

import "time"

// DiscoveryStarted is published when the question flow begins
type DiscoveryStarted struct {
    SessionID string    `json:"session_id"`
    Persona   string    `json:"persona"`    // "solo_developer", "team_lead", etc.
    Register  string    `json:"register"`   // "technical" or "non_technical"
    Timestamp time.Time `json:"timestamp"`
}

// QuestionPhaseCompleted is published after each of 5 question phases
type QuestionPhaseCompleted struct {
    SessionID string    `json:"session_id"`
    Phase     string    `json:"phase"`      // "seed", "actors", "story", "events", "boundaries"
    PhaseNum  int       `json:"phase_num"`  // 1-5
    Answers   int       `json:"answers"`    // Number of answers collected
    Timestamp time.Time `json:"timestamp"`
}

// DiscoveryCompleted is published when all questions are answered
type DiscoveryCompleted struct {
    SessionID    string    `json:"session_id"`
    TotalAnswers int       `json:"total_answers"`
    Timestamp    time.Time `json:"timestamp"`
}
```

#### Document Generation Context Events

```go
// events/generation.go
package events

import "time"

// PRDGenerated is published when the PRD document is written
type PRDGenerated struct {
    SessionID string    `json:"session_id"`
    PRDPath   string    `json:"prd_path"`
    Timestamp time.Time `json:"timestamp"`
}

// DDDArtifactsGenerated is published when DDD docs are written
type DDDArtifactsGenerated struct {
    SessionID         string    `json:"session_id"`
    BoundedContexts   []string  `json:"bounded_contexts"`
    SubdomainCount    int       `json:"subdomain_count"`
    Timestamp         time.Time `json:"timestamp"`
}

// ArchitectureGenerated is published when architecture doc is written
type ArchitectureGenerated struct {
    SessionID string    `json:"session_id"`
    DocPath   string    `json:"doc_path"`
    Timestamp time.Time `json:"timestamp"`
}
```

#### Ticket Pipeline Context Events

```go
// events/tickets.go
package events

import "time"

// TicketsGenerated is published when beads tickets are created
type TicketsGenerated struct {
    SessionID   string    `json:"session_id"`
    TicketCount int       `json:"ticket_count"`
    EpicCount   int       `json:"epic_count"`
    Timestamp   time.Time `json:"timestamp"`
}

// FitnessTestsGenerated is published when architecture test files are written
type FitnessTestsGenerated struct {
    SessionID string    `json:"session_id"`
    TestCount int       `json:"test_count"`
    TestPaths []string  `json:"test_paths"`
    Timestamp time.Time `json:"timestamp"`
}
```

#### Tool Translation Context Events

```go
// events/tools.go
package events

import "time"

// ToolConfigsGenerated is published when AI tool configs are written
type ToolConfigsGenerated struct {
    SessionID   string    `json:"session_id"`
    ToolNames   []string  `json:"tool_names"`  // ["claude_code", "cursor", ...]
    ConfigPaths []string  `json:"config_paths"`
    Timestamp   time.Time `json:"timestamp"`
}
```

### Command Design for alto

```go
// commands/bootstrap.go
package commands

// StartBootstrap initiates the bootstrap pipeline
type StartBootstrap struct {
    ProjectDir string `json:"project_dir"`
    IsExisting bool   `json:"is_existing"`
}

// RunDiscovery starts the guided DDD question flow
type RunDiscovery struct {
    SessionID string `json:"session_id"`
    Persona   string `json:"persona"`
    Register  string `json:"register"`
}

// GeneratePRD produces the PRD from discovery answers
type GeneratePRD struct {
    SessionID string `json:"session_id"`
}

// GenerateDDDArtifacts produces domain model documents
type GenerateDDDArtifacts struct {
    SessionID string `json:"session_id"`
}

// GenerateArchitecture produces architecture document
type GenerateArchitecture struct {
    SessionID string `json:"session_id"`
}

// GenerateTickets produces beads tickets from DDD artifacts
type GenerateTickets struct {
    SessionID string `json:"session_id"`
}

// GenerateFitnessTests produces architecture test files
type GenerateFitnessTests struct {
    SessionID string `json:"session_id"`
}

// GenerateToolConfigs produces AI tool-native configurations
type GenerateToolConfigs struct {
    SessionID string   `json:"session_id"`
    ToolNames []string `json:"tool_names"`
}
```

### Handler Wiring: The Bootstrap Pipeline

```go
// internal/application/bootstrap_handlers.go

func RegisterBootstrapHandlers(
    cp *cqrs.CommandProcessor,
    ep *cqrs.EventProcessor,
    commandBus *cqrs.CommandBus,
    eventBus *cqrs.EventBus,
    // ... domain service ports
) error {
    // Command handlers
    err := cp.AddHandlers(
        cqrs.NewCommandHandler[commands.StartBootstrap](
            "StartBootstrapHandler",
            func(ctx context.Context, cmd *commands.StartBootstrap) error {
                sessionID := watermill.NewUUID()
                return eventBus.Publish(ctx, &events.BootstrapStarted{
                    SessionID:  sessionID,
                    ProjectDir: cmd.ProjectDir,
                    IsExisting: cmd.IsExisting,
                    Timestamp:  time.Now(),
                })
            },
        ),
        // ... more command handlers
    )
    if err != nil {
        return err
    }

    // Event handlers (pipeline orchestration)
    // Policy: "When DiscoveryCompleted, then GeneratePRD"
    err = ep.AddHandlers(
        cqrs.NewEventHandler[events.DiscoveryCompleted](
            "OnDiscoveryCompleted_GeneratePRD",
            func(ctx context.Context, event *events.DiscoveryCompleted) error {
                return commandBus.Send(ctx, &commands.GeneratePRD{
                    SessionID: event.SessionID,
                })
            },
        ),
        // Policy: "When PRDGenerated, then GenerateDDDArtifacts"
        cqrs.NewEventHandler[events.PRDGenerated](
            "OnPRDGenerated_GenerateDDDArtifacts",
            func(ctx context.Context, event *events.PRDGenerated) error {
                return commandBus.Send(ctx, &commands.GenerateDDDArtifacts{
                    SessionID: event.SessionID,
                })
            },
        ),
        // Policy: "When DDDArtifactsGenerated, then GenerateArchitecture"
        cqrs.NewEventHandler[events.DDDArtifactsGenerated](
            "OnDDDArtifacts_GenerateArchitecture",
            func(ctx context.Context, event *events.DDDArtifactsGenerated) error {
                return commandBus.Send(ctx, &commands.GenerateArchitecture{
                    SessionID: event.SessionID,
                })
            },
        ),
        // Policy: "When ArchitectureGenerated, then GenerateTickets + GenerateFitnessTests"
        cqrs.NewEventHandler[events.ArchitectureGenerated](
            "OnArchitecture_GenerateTickets",
            func(ctx context.Context, event *events.ArchitectureGenerated) error {
                return commandBus.Send(ctx, &commands.GenerateTickets{
                    SessionID: event.SessionID,
                })
            },
        ),
        cqrs.NewEventHandler[events.ArchitectureGenerated](
            "OnArchitecture_GenerateFitnessTests",
            func(ctx context.Context, event *events.ArchitectureGenerated) error {
                return commandBus.Send(ctx, &commands.GenerateFitnessTests{
                    SessionID: event.SessionID,
                })
            },
        ),
    )

    return err
}
```

This maps directly to alto's DDD Policy pattern: "Whenever [event], then [command]."

---

## 6. Migration Path: GoChannel to Embedded NATS

### Why Migrate

| Phase | Trigger | Backend |
|-------|---------|---------|
| Phase 1 (CLI) | Initial release | GoChannel (in-process) |
| Phase 2 (MCP) | MCP server with concurrent agent requests | Embedded NATS |
| Phase 3 (Scale) | Multiple MCP servers (unlikely for CLI tool) | External NATS |

### What Changes

Only the pub/sub backend initialization changes. All command handlers, event handlers, domain events, and middleware remain identical.

**Phase 1 (GoChannel):**

```go
pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)
// Used for both publisher and subscriber
```

**Phase 2 (Embedded NATS):**

```go
import (
    natsserver "github.com/nats-io/nats-server/v2/server"
    natsclient "github.com/nats-io/nats.go"
    watermill_nats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
)

// Start embedded NATS (no TCP listener)
opts := &natsserver.Options{DontListen: true}
ns, _ := natsserver.NewServer(opts)
go ns.Start()
ns.ReadyForConnections(5 * time.Second)

// Connect in-process
nc, _ := natsclient.Connect(
    ns.ClientURL(),
    natsclient.InProcessServer(ns),
)

// Create Watermill publisher/subscriber
publisher, _ := watermill_nats.NewPublisher(
    watermill_nats.PublisherConfig{
        URL:       ns.ClientURL(),
        Marshaler: watermill_nats.GobMarshaler{},
        NatsOptions: []natsclient.Option{
            natsclient.InProcessServer(ns),
        },
    },
    logger,
)

subscriber, _ := watermill_nats.NewSubscriber(
    watermill_nats.SubscriberConfig{
        URL:       ns.ClientURL(),
        Marshaler: watermill_nats.GobMarshaler{},
        NatsOptions: []natsclient.Option{
            natsclient.InProcessServer(ns),
        },
    },
    logger,
)
```

**Everything else stays the same.** The `CommandBusConfig`, `EventBusConfig`, handlers, middleware -- all unchanged.

Source: [Watermill NATS Docs](https://watermill.io/pubsubs/nats/),
[Embedded NATS Pattern](https://gosuda.org/blog/posts/how-embedded-nats-communicate-with-go-application-z36089af0),
[watermill-nats GitHub](https://github.com/ThreeDotsLabs/watermill-nats)

### NATS Feature Comparison

| Feature | GoChannel | NATS JetStream |
|---------|-----------|----------------|
| Consumer Groups | No | Yes (`QueueGroupPrefix`) |
| Exactly-Once | Yes | Yes (`TrackMsgId` + sync ack) |
| Guaranteed Order | Yes* | No (redelivery) |
| Persistence | No | Yes (durable streams) |
| Performance (publish) | 315,776 msg/s | 50,668 msg/s |
| Performance (subscribe) | 138,743 msg/s | 34,713 msg/s |
| Memory overhead | ~0 | 5-20MB |
| External dependency | None | Embedded or external server |

Source: [Watermill Benchmarks](https://watermill.io/),
[NATS JetStream Docs](https://watermill.io/pubsubs/nats/)

### Alternative: SQLite Backend (v1.5+)

Watermill 1.5 introduced SQLite as a pub/sub backend. This is interesting for alto because:

- **Single binary** -- no external processes (like GoChannel)
- **Persistent** -- survives process restarts (unlike GoChannel)
- **Consumer groups** -- supported (unlike GoChannel)
- **Guaranteed order** -- supported
- **CGO-free** -- pure Go drivers (modernc.org/sqlite or zombiezen.com/go/sqlite)
- **File on disk** -- just a `.db` file, no server

This could be a Phase 1.5 option: more durable than GoChannel, simpler than NATS.

| Feature | GoChannel | SQLite | NATS |
|---------|-----------|--------|------|
| Persistence | No | Yes (file) | Yes (stream) |
| Consumer Groups | No | Yes | Yes |
| Exactly-Once | Yes | No | Yes |
| External Process | No | No | Yes* |
| Single Binary | Yes | Yes | Yes* |
| Complexity | Trivial | Low | Medium |

*NATS can be embedded, but adds ~20MB RAM.

Source: [Watermill SQLite Docs](https://watermill.io/pubsubs/sqlite/),
[Watermill 1.5 Release](https://threedots.tech/post/watermill-1-5/),
[Durable Execution with SQLite](https://threedots.tech/post/sqlite-durable-execution/)

### Recommended Migration Path (Updated)

```
Phase 1 (CLI MVP):      GoChannel (in-process, zero deps)
Phase 1.5 (Durability): SQLite (persistent, still single binary, zero external deps)
Phase 2 (MCP):          Embedded NATS (concurrent agents, in-process)
Phase 3 (Scale):        External NATS (if ever needed -- unlikely for CLI tool)
```

---

## 7. Watermill 1.5 Release Notes

Released as v1.5.1 on September 2, 2024. Key changes relevant to alto:

### New Features

1. **SQLite Pub/Sub (Beta)** -- File-based persistent messaging, no external infrastructure. Pure Go (CGO-free) drivers. Ideal for single-binary CLI tools.

2. **`message.NewMessageWithContext()`** -- Cleaner context propagation when creating messages.

3. **`AddConsumerHandler`** -- Renamed from `AddNoPublisherHandler` for semantic clarity.

4. **GoChannel `PreserveContext`** -- Optional context propagation from publisher to subscriber. Use cautiously (not how production pub/subs work).

5. **`CommandEventMarshalerDecorator`** -- Wraps any marshaler to add metadata (partition keys, routing info).

### Breaking Changes

1. **`ProtobufMarshaler` deprecated** -- Use `ProtoMarshaler` (switches from `gogo/protobuf` to `google.golang.org/protobuf`).

2. **watermill-sql v4.0.0** -- Publisher/Subscriber constructors now accept `Beginner` interface. Wrap with `BeginnerFromStdSQL(db)`.

3. **watermill-googlecloud v2** -- `SubscriptionConfig` replaced with `GenerateSubscription` function.

### Community

43 unique contributors across all Watermill repositories since the previous release.

Source: [Watermill 1.5 Release](https://threedots.tech/post/watermill-1-5/)

---

## 8. ThreeDotsLabs Best Practices for DDD in Go

### Key Articles

| Article | URL | Key Takeaway |
|---------|-----|--------------|
| DDD Lite Introduction | [threedots.tech](https://threedots.tech/post/ddd-lite-in-go-introduction/) | Adapt DDD to Go idioms; strategic DDD (event storming) provides 70% of value |
| Basic CQRS in Go | [threedots.tech](https://threedots.tech/post/basic-cqrs-in-go/) | CQRS is a simple pattern; separate read/write stores are optional |
| Clean Architecture | [threedots.tech](https://threedots.tech/post/introducing-clean-architecture/) | Hexagonal architecture with ports and adapters |
| DDD + CQRS + Clean Combined | [threedots.tech](https://threedots.tech/post/ddd-cqrs-clean-architecture-combined/) | The three patterns are synergistic: domain encapsulation + command orchestration + dependency inversion |
| Repository Pattern | [threedots.tech](https://threedots.tech/post/repository-pattern-in-go/) | Simplify to 3 methods (Add, Get, Update); domain logic in entities, not repos |
| Modern Business Software | [threedots.tech](https://threedots.tech/series/modern-business-software-in-go/) | Full series covering all DDD patterns in Go |
| Go With The Domain | [threedots.tech](https://threedots.tech/go-with-the-domain/) | Free e-book on DDD in Go |

Source: [ThreeDotsLabs Blog](https://threedots.tech/)

### DDD Lite Rules (adapted for Go)

1. **Reflect business logic directly in code.** Methods like `ScheduleTraining()` instead of `SetStatus("scheduled")`.

2. **Keep valid state in memory.** Private fields, constructor validation, no exported setters.

3. **Keep domain database-agnostic.** Repository interfaces in the domain layer; implementations in infrastructure.

4. **Commands orchestrate, domain objects decide.** Command handlers call domain methods but do not contain business logic.

5. **Simplify repositories.** With rich domain objects, repositories need only `Add()`, `Get()`, and `Update()`.

Source: [DDD Lite Introduction](https://threedots.tech/post/ddd-lite-in-go-introduction/)

### Wild Workouts Architecture

The [Wild Workouts](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example) project demonstrates:

- `internal/` directory for bounded contexts
- CQRS with Watermill for cross-context communication
- Clean Architecture layering (domain -> application -> ports -> adapters)
- Repository pattern with Firestore
- Integration testing with Docker Compose
- Progressive refactoring from monolith to DDD (15+ articles documenting each step)

The project was updated January 5, 2026, confirming active maintenance.

Source: [Wild Workouts GitHub](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)

### CQRS Best Practices

1. **Commands are imperative** (do something): `GeneratePRD`, `StartBootstrap`
2. **Events are past tense** (something happened): `PRDGenerated`, `BootstrapCompleted`
3. **Events are immutable** -- once published, never modified
4. **One command, one handler** -- commands have exactly one handler
5. **One event, many handlers** -- events fan out to multiple handlers
6. **Handlers must be thread-safe** -- they execute concurrently
7. **Handle duplicates** -- if using at-least-once delivery, design handlers to be idempotent
8. **Use `EventGroupProcessor`** for ordered processing across event types

Source: [Watermill CQRS Docs](https://watermill.io/docs/cqrs/)

### Durable Execution Pattern (from SQLite article)

ThreeDotsLabs defines durable execution as having three requirements:

1. **Persistent input storage** -- events written to disk before processing
2. **Idempotency** -- handlers can safely process the same event multiple times
3. **Atomicity** -- all state changes in a handler succeed or fail together

For alto, this matters when generating artifacts (PRD, DDD docs, tickets). If the CLI crashes mid-pipeline, idempotent handlers can safely re-run without corrupting output.

Source: [Durable Execution with SQLite](https://threedots.tech/post/sqlite-durable-execution/)

---

## 9. Event Sourcing Support

Watermill does **not** provide a built-in event store or event sourcing component. It is designed for event-driven architecture (publish/subscribe), not event sourcing (storing all events as the source of truth).

However, Watermill can be used as the **transport layer** for an event sourcing system:
- Events published to Watermill after being persisted to an event store
- Projections built via Watermill event handlers
- Sagas orchestrated via command/event patterns

For alto, event sourcing is not needed. The CQRS component (commands + events without event store) is sufficient.

Source: [Watermill README](https://github.com/ThreeDotsLabs/watermill),
[eventhorizon comparison](https://github.com/looplab/eventhorizon) (provides full ES, Watermill does not)

---

## 10. Deprecated APIs to Avoid

| Deprecated | Replacement | Since |
|-----------|-------------|-------|
| `cqrs.Facade` | `CommandProcessor` + `EventProcessor` | v1.3.0 |
| `cqrs.FacadeConfig` | Individual processor configs | v1.3.0 |
| `cqrs.ProtobufMarshaler` | `cqrs.ProtoMarshaler` | v1.5.0 |
| `router.AddNoPublisherHandler` | `router.AddConsumerHandler` | v1.5.0 |

Source: [pkg.go.dev cqrs package](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/cqrs)

---

## Summary: Watermill Integration Strategy for alto

### Architecture Mapping

```
alto Domain Layer (Go)
    |
    +-- domain/events/          Go structs (PRDGenerated, DiscoveryCompleted, etc.)
    +-- domain/commands/        Go structs (GeneratePRD, StartBootstrap, etc.)
    |
alto Application Layer
    |
    +-- application/ports/      EventPublisher, CommandSender interfaces (Protocols)
    +-- application/handlers/   Command handlers + event handlers (policies)
    |
alto Infrastructure Layer
    |
    +-- infrastructure/messaging/
        +-- watermill_adapter.go   Implements ports using Watermill
        +-- setup.go               CommandBus/EventBus/Router wiring
        +-- gochannel_backend.go   GoChannel configuration
        +-- nats_backend.go        NATS configuration (Phase 2)
```

### Key Design Decisions

1. **Domain events are plain Go structs** in `domain/events/`. They have zero dependency on Watermill. The infrastructure layer marshals them.

2. **Ports define the contract.** `application/ports/EventPublisher` is an interface; `infrastructure/messaging/watermill_adapter.go` implements it.

3. **JSONMarshaler for simplicity.** No protobuf codegen. Events are small, human-readable JSON.

4. **Middleware: Recoverer + CorrelationID + Retry.** Minimal but sufficient for CLI.

5. **GoChannel for Phase 1.** Zero external dependencies, single binary.

6. **SQLite as Phase 1.5 option.** If durability is needed before MCP, SQLite pub/sub adds persistence without external infrastructure.

---

## Sources

### Official Documentation
- [Watermill CQRS Component](https://watermill.io/docs/cqrs/)
- [Watermill Message Docs](https://watermill.io/docs/message/)
- [Watermill Middleware Docs](https://watermill.io/docs/middlewares/)
- [Watermill GoChannel Docs](https://watermill.io/pubsubs/gochannel/)
- [Watermill NATS JetStream Docs](https://watermill.io/pubsubs/nats/)
- [Watermill SQLite Docs](https://watermill.io/pubsubs/sqlite/)
- [Watermill Getting Started](https://watermill.io/learn/getting-started/)

### API Reference
- [pkg.go.dev cqrs v1.5.1](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill@v1.5.1/components/cqrs)
- [pkg.go.dev gochannel](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill/pubsub/gochannel)
- [GoChannel source code](https://github.com/ThreeDotsLabs/watermill/blob/master/pubsub/gochannel/pubsub.go)

### Release Notes
- [Watermill 1.5 Release](https://threedots.tech/post/watermill-1-5/)
- [Watermill GitHub Releases](https://github.com/ThreeDotsLabs/watermill/releases)

### ThreeDotsLabs DDD Articles
- [DDD Lite Introduction](https://threedots.tech/post/ddd-lite-in-go-introduction/)
- [Basic CQRS in Go](https://threedots.tech/post/basic-cqrs-in-go/)
- [Combining DDD, CQRS, and Clean Architecture](https://threedots.tech/post/ddd-cqrs-clean-architecture-combined/)
- [Clean Architecture in Go](https://threedots.tech/post/introducing-clean-architecture/)
- [Repository Pattern in Go](https://threedots.tech/post/repository-pattern-in-go/)
- [Modern Business Software in Go (series)](https://threedots.tech/series/modern-business-software-in-go/)
- [Go With The Domain (e-book)](https://threedots.tech/go-with-the-domain/)
- [Durable Execution with SQLite](https://threedots.tech/post/sqlite-durable-execution/)

### Reference Projects
- [Wild Workouts Go DDD Example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
- [Watermill CQRS Protobuf Example](https://github.com/ThreeDotsLabs/watermill/blob/master/_examples/basic/5-cqrs-protobuf/main.go)
- [watermill-nats](https://github.com/ThreeDotsLabs/watermill-nats)
- [watermill-sqlite](https://github.com/ThreeDotsLabs/watermill-sqlite)

### Prior Research
- [Go DDD Event Systems and Messaging](docs/research/20260306_2_go_ddd_event_systems_messaging.md) -- initial evaluation
