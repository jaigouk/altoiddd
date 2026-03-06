# Go DDD Event Systems and Messaging for Local CLI/MCP Application

**Date:** 2026-03-06
**Type:** Spike Research
**Status:** Complete

## Research Question

What is the best event/messaging architecture for a Go-based local-first DDD CLI tool
that may grow into handling concurrent MCP sub-agent requests? Evaluate in-process
event buses, NATS (standalone and embedded), Watermill, and mediator patterns.

## Context

alty is a project bootstrapper that enforces DDD + TDD + SOLID. A Go rewrite is under
evaluation. The application needs:

- Domain events within bounded contexts (e.g., "PRD generated" triggers ticket pipeline)
- Command/query separation (CQRS-lite, not full event sourcing)
- Future MCP server handling concurrent sub-agent requests
- Single local machine (developer workstation), not distributed
- Small footprint -- CLI tool, not a server farm

---

## Option 1: Pure Go Channels (DIY Event Bus)

### Description

Implement a custom event bus using Go's built-in channels with fan-out pattern.
No external dependencies.

### Implementation Pattern

```go
type EventBus struct {
    mu       sync.RWMutex
    handlers map[string][]chan Event
}

func (b *EventBus) Publish(topic string, event Event) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    subs := b.handlers[topic]
    for _, ch := range subs {
        go func(c chan Event) { c <- event }(ch)
    }
}
```

Source: [Go Event Bus pattern](https://leapcell.medium.com/how-to-create-a-event-bus-in-go-d7919b59a584),
[Go Pipelines and Fan-out](https://go.dev/blog/pipelines)

### Evaluation

| Criterion | Assessment |
|-----------|------------|
| License | N/A (stdlib) |
| Memory footprint | Negligible -- just goroutines and channels |
| Setup complexity | Low initial, HIGH long-term (must build marshaling, error handling, retry, middleware) |
| DDD event support | Manual -- no typed event handlers, no command/event separation built in |
| CQRS support | None -- must build from scratch |
| Upgrade path | Poor -- no abstraction over transport; locked to in-process |
| Single binary | Yes |
| Maintenance burden | High -- you own all the code |

### Verdict

Suitable only if event needs are trivially simple (2-3 event types, no retry, no
middleware). For a DDD application with multiple bounded contexts, the boilerplate
cost is prohibitive. You end up reimplementing what Watermill already provides.

---

## Option 2: asaskevich/EventBus

### Facts

| Property | Value | Source |
|----------|-------|--------|
| Stars | ~1.9k | [GitHub](https://github.com/asaskevich/EventBus) |
| License | MIT | [GitHub](https://github.com/asaskevich/EventBus) |
| Last push | June 2024 | [GitHub Contributors](https://github.com/asaskevich/EventBus/graphs/contributors) |
| Go module | `github.com/asaskevich/EventBus` | [pkg.go.dev](https://pkg.go.dev/github.com/asaskevich/eventbus) |

### Features

- Subscribe/Publish with string topic keys
- Async publish support
- Cross-process via network (client/server mode)
- Uses reflection for handler dispatch

### Evaluation

| Criterion | Assessment |
|-----------|------------|
| DDD event support | Weak -- string topics, reflection-based, no typed handlers |
| CQRS support | None |
| Upgrade path | None -- no transport abstraction |
| Maintenance | Low activity; open issues dating back years |
| Type safety | Poor -- uses `interface{}` and reflection |

### Verdict

Too simplistic for DDD use. Reflection-based dispatch loses Go's type safety.
No command/event separation, no middleware pipeline. Not recommended.

---

## Option 3: mustafaturan/bus v3

### Facts

| Property | Value | Source |
|----------|-------|--------|
| Version | v3.0.3 (stable) | [pkg.go.dev](https://pkg.go.dev/github.com/mustafaturan/bus/v3) |
| License | Apache 2.0 | [GitHub](https://github.com/mustafaturan/bus) |
| Stars | ~340 | [GitHub](https://github.com/mustafaturan/bus) |
| Key feature | Zero-allocation on Emit | [GitHub README](https://github.com/mustafaturan/bus) |

### Features

- Topic-based pub/sub with handler registration
- Zero-allocation emit (performance-optimized)
- Requires external ID generator (e.g., mustafaturan/monoton)
- Stable API (v3.x.x frozen except bugfixes)

### Evaluation

| Criterion | Assessment |
|-----------|------------|
| DDD event support | Basic -- topic+handler, but no typed event structs |
| CQRS support | None |
| Upgrade path | None -- in-process only |
| Unique value | Zero-allocation -- matters for hot paths, not for CLI tool |

### Verdict

Better than asaskevich/EventBus for performance, but still lacks DDD/CQRS
abstractions. The zero-allocation optimization is irrelevant for a CLI tool that
processes tens of events per session, not millions per second.

---

## Option 4: mehdihadeli/Go-MediatR

### Facts

| Property | Value | Source |
|----------|-------|--------|
| Stars | ~276 | [GitHub](https://github.com/mehdihadeli/Go-MediatR) |
| License | MIT | [GitHub](https://github.com/mehdihadeli/Go-MediatR/blob/main/LICENCE) |
| Go module | `github.com/mehdihadeli/go-mediatr` | [pkg.go.dev](https://pkg.go.dev/github.com/mehdihadeli/go-mediatr) |
| Pattern | Mediator (inspired by .NET MediatR) | [GitHub README](https://github.com/mehdihadeli/Go-MediatR) |

### Features

- Request/Response messages dispatched to single handler (commands + queries)
- Notification messages dispatched to multiple handlers (events)
- Pipeline behaviors (middleware) for cross-cutting concerns
- Generics support (Go 1.18+)

### Evaluation

| Criterion | Assessment |
|-----------|------------|
| DDD event support | Good -- notifications = domain events, multiple handlers |
| CQRS support | Good -- request/response for commands, notifications for events |
| Upgrade path | Poor -- purely in-process, no transport layer |
| Type safety | Good -- uses Go generics |
| Middleware | Yes -- pipeline behaviors |
| Community | Small (276 stars) but concept is well-proven from .NET |

### Verdict

Best pure mediator option for CQRS-lite. However, it has no transport abstraction --
if you later need events to cross process boundaries (MCP server handling multiple
agents), you must add a separate messaging layer. Could work as the application-layer
CQRS dispatcher paired with Watermill or NATS for infrastructure events.

---

## Option 5: looplab/eventhorizon

### Facts

| Property | Value | Source |
|----------|-------|--------|
| Stars | ~1.7k | [GitHub](https://github.com/looplab/eventhorizon) |
| License | Apache 2.0 | [GitHub](https://github.com/looplab/eventhorizon) |
| Latest release | v0.16.0 (Dec 2022) | [GitHub Releases](https://github.com/looplab/eventhorizon/releases) |
| Last significant update | 2022 | [GitHub](https://github.com/looplab/eventhorizon) |

### Features

- Full CQRS/ES (Event Sourcing) toolkit
- Aggregate store, event store, command handlers, event handlers
- Multiple backends: Memory, MongoDB, Kafka, NATS, Redis, GCP Pub/Sub
- Saga support

### Evaluation

| Criterion | Assessment |
|-----------|------------|
| DDD event support | Excellent -- purpose-built for DDD aggregates |
| CQRS support | Full CQRS + Event Sourcing |
| Upgrade path | Good -- multiple backends |
| Maintenance | CONCERN: last release Dec 2022, API marked "not final" |
| Complexity | High -- full ES is overkill for alty's needs |

### Verdict

Powerful but overengineered for this use case. alty does not need event sourcing --
it needs domain events and command/query separation. The stale maintenance (3+ years
since last release) is a risk. The "API not final" warning after 3 years of no
releases suggests abandonment risk.

---

## Option 6: NATS (Standalone or Embedded)

### Facts

| Property | Value | Source |
|----------|-------|--------|
| Server version | v2.12.4 (Jan 27, 2026) | [GitHub Releases](https://github.com/nats-io/nats-server/releases) |
| Client version | nats.go (actively maintained) | [GitHub](https://github.com/nats-io/nats.go) |
| License | Apache 2.0 (server + client) | [nats.io](https://nats.io/about/) |
| Server RAM | <20MB typical | [nats.io](https://nats.io/about/) |
| Binary size | ~6.3MB (deb/rpm packages for amd64) | [GitHub Releases](https://github.com/nats-io/nats-server/releases) |
| Stars (server) | 16k+ | [GitHub](https://github.com/nats-io/nats-server) |
| Stars (client) | 5k+ | [GitHub](https://github.com/nats-io/nats.go) |

### Embedded Mode

NATS server can be imported as a Go library and run in-process:

```go
import "github.com/nats-io/nats-server/v2/server"
import "github.com/nats-io/nats.go"

opts := &server.Options{DontListen: true}  // No TCP listener
ns, _ := server.NewServer(opts)
go ns.Start()
ns.ReadyForConnections(5 * time.Second)

// Connect in-process (net.Pipe, no network)
nc, _ := nats.Connect(ns.ClientURL(), nats.InProcessServer(ns))
```

Source: [Embedding NATS in Go](https://gosuda.org/blog/posts/how-embedded-nats-communicate-with-go-application-z36089af0),
[Synadia Screencast](https://www.synadia.com/screencasts/how-to-embed-nats-server-in-your-app),
[Dev.to Guide](https://dev.to/karanpratapsingh/embedding-nats-in-go-19o)

Key insight: With `DontListen: true` + `InProcessServer`, communication uses
`net.Pipe()` -- pure in-memory, no TCP socket overhead.

### JetStream (Durable Events)

- Core NATS: fire-and-forget pub/sub (sufficient for domain events)
- JetStream: durable streams, replay, at-least-once delivery
- JetStream adds storage overhead but enables event replay

### Evaluation

| Criterion | Assessment |
|-----------|------------|
| DDD event support | Transport only -- no typed handlers, no command/event distinction |
| CQRS support | None -- raw pub/sub, must build CQRS layer on top |
| Memory footprint | <20MB standalone; likely 5-10MB embedded with DontListen |
| Setup complexity | Medium (embedded) to Low (Docker sidecar) |
| Upgrade path | Excellent -- same code works local or distributed |
| Single binary | Yes (embedded mode) |
| Community | Massive (16k stars, Synadia-backed, CNCF) |
| Concurrency | Battle-tested for millions of msgs/sec |

### Verdict

Excellent infrastructure layer. Overkill as standalone for a CLI tool, but the
embedded mode makes it a zero-cost option that provides a real upgrade path.
However, NATS is raw transport -- it does not provide DDD/CQRS abstractions.
You still need an application-level framework on top.

---

## Option 7: Watermill (ThreeDotsLabs) -- RECOMMENDED

### Facts

| Property | Value | Source |
|----------|-------|--------|
| Version | v1.5.1 (Sep 2, 2024) | [GitHub Releases](https://github.com/ThreeDotsLabs/watermill/releases) |
| Stars | ~9.4k | [GitHub](https://github.com/ThreeDotsLabs/watermill) |
| License | MIT | [GitHub](https://github.com/ThreeDotsLabs/watermill) |
| Go module | `github.com/ThreeDotsLabs/watermill` | [pkg.go.dev](https://pkg.go.dev/github.com/ThreeDotsLabs/watermill) |
| API stability | Stable since v1.0.0 | [watermill.io](https://watermill.io/) |
| Contributors | 17+ on v1.5.0 alone | [GitHub Releases](https://github.com/ThreeDotsLabs/watermill/releases) |
| DDD example | Wild Workouts Go DDD | [GitHub](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example) |

### Architecture

Watermill provides a layered architecture that maps cleanly to DDD:

```
Application Layer (your code)
    |
    v
CQRS Component (CommandBus, EventBus, CommandProcessor, EventProcessor)
    |
    v
Router + Middleware (retry, correlation ID, recovery, metrics)
    |
    v
Pub/Sub Interface (pluggable backend)
    |
    v
GoChannel | NATS | Kafka | Redis | SQL | ...
```

### CQRS Component (First-Class)

From Context7 documentation:

**Command Bus:**
- Commands dispatched to exactly one handler
- Typed command structs (Go generics)
- JSON marshaling built in
- OnSend hooks for metadata enrichment

**Event Bus:**
- Events dispatched to multiple handlers
- Multiple handlers per event type
- OnPublish hooks
- Fan-out to independent processors

**Event Processor Groups:**
- Multiple handlers share a single subscriber
- Ordered processing within a group
- Essential for read model projections

Source: [Watermill CQRS Docs](https://watermill.io/docs/cqrs/),
[Context7 Watermill](https://context7.com/threedotslabs/watermill)

### Middleware

Built-in middleware pipeline:
- `Recoverer` -- panic recovery
- `Retry` -- configurable retry with backoff
- `CorrelationID` -- trace propagation
- `Throttle` -- rate limiting
- `Timeout` -- handler timeout
- Custom middleware via simple function signature

### Pub/Sub Backends

| Backend | Package | Use Case |
|---------|---------|----------|
| GoChannel | `watermill/pubsub/gochannel` | In-process, testing, local dev |
| NATS JetStream | `watermill-nats/v2` | Durable events, distributed |
| Kafka | `watermill-kafka` | High-throughput streaming |
| SQL (Postgres) | `watermill-sql` | Transactional outbox pattern |
| Redis | `watermill-redisstream` | Redis Streams |

**Critical feature: start with GoChannel, upgrade to NATS later with zero application code changes.**

### GoChannel Backend (In-Memory)

```go
pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)
```

- Non-blocking publish (messages sent in background goroutines)
- No persistence (messages lost if no subscriber)
- No global state (same instance for pub and sub)
- Sufficient for CLI tool domain events

Source: [Watermill GoChannel docs](https://watermill.io/pubsubs/gochannel/)

### NATS Backend (Upgrade Path)

The `watermill-nats/v2` package supports JetStream:

```go
// Same application code, different backend
publisher, _ := jetstream.NewPublisher(jetstream.PublisherConfig{
    URL: nats.DefaultURL,  // or embedded NATS URL
}, logger)
subscriber, _ := jetstream.NewSubscriber(jetstream.SubscriberConfig{
    URL: nats.DefaultURL,
}, logger)
```

Source: [watermill-nats](https://github.com/ThreeDotsLabs/watermill-nats),
[Watermill NATS docs](https://watermill.io/pubsubs/nats/)

### Wild Workouts DDD Reference

ThreeDotsLabs maintains a complete DDD+CQRS+Clean Architecture example in Go
using Watermill: [wild-workouts-go-ddd-example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example).
This provides a battle-tested reference architecture for alty's Go rewrite.

### Evaluation

| Criterion | Assessment |
|-----------|------------|
| DDD event support | Excellent -- typed event handlers, event bus, multiple handlers per event |
| CQRS support | Excellent -- first-class CommandBus, EventBus, CommandProcessor, EventProcessor |
| Memory footprint | Negligible with GoChannel; ~20MB additional with embedded NATS |
| Setup complexity | Low -- CQRS component is ~50 lines of config |
| Upgrade path | Excellent -- swap GoChannel for NATS/Kafka/SQL with no app code changes |
| Single binary | Yes (with GoChannel or embedded NATS) |
| Community | Strong (9.4k stars, active development, ThreeDotsLabs blog) |
| Middleware | Rich built-in set (retry, recovery, correlation, throttle) |
| Type safety | Good -- generics-based handlers since v1.3+ |
| Saga support | Yes -- via event-driven command patterns |

---

## Comparison Table

| Criterion | Go Channels | asaskevich/EventBus | mustafaturan/bus | Go-MediatR | eventhorizon | NATS Embedded | Watermill |
|-----------|:-----------:|:-------------------:|:----------------:|:----------:|:------------:|:-------------:|:---------:|
| **License** | N/A | MIT | Apache 2.0 | MIT | Apache 2.0 | Apache 2.0 | MIT |
| **Stars** | N/A | 1.9k | 340 | 276 | 1.7k | 16k (server) | 9.4k |
| **Last release** | N/A | Jun 2024 | Stable v3 | Active | Dec 2022 | Jan 2026 | Sep 2024 |
| **Memory** | ~0 | ~0 | ~0 | ~0 | ~0 | 5-20MB | ~0 (GoChannel) |
| **DDD events** | Manual | Weak | Basic | Good | Excellent | Transport only | Excellent |
| **CQRS** | None | None | None | Good | Full ES | None | Excellent |
| **Typed handlers** | No | No (reflection) | No | Yes (generics) | Yes | No | Yes |
| **Middleware** | None | None | None | Pipeline behaviors | Middleware | None | Rich set |
| **Upgrade path** | None | None | None | None | Multi-backend | Excellent | Excellent |
| **Single binary** | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| **Setup effort** | High (build everything) | Low | Low | Medium | High | Medium | Low-Medium |
| **Maintenance risk** | Self-maintained | Low activity | Frozen | Small community | Stale (3yr) | Synadia-backed | Active |

---

## Recommendation

### Primary: Watermill with GoChannel backend

**Rationale:**

1. **DDD-native CQRS.** Watermill's CQRS component provides exactly what alty needs:
   typed command handlers (one handler per command), typed event handlers (multiple
   handlers per event), and a clean separation between command bus and event bus.
   This maps directly to DDD bounded context communication patterns.

2. **Zero-cost start, smooth upgrade.** Begin with `gochannel.NewGoChannel()` for
   the CLI -- pure in-process, zero external dependencies, single binary. When the
   MCP server needs concurrent agent handling, swap to embedded NATS or external NATS
   by changing only the pub/sub backend configuration. Application code (handlers,
   commands, events) remains untouched.

3. **Battle-tested DDD reference.** ThreeDotsLabs' Wild Workouts project provides a
   complete Go DDD+CQRS+Clean Architecture example using Watermill, directly
   applicable to alty's architecture.

4. **Middleware pipeline.** Built-in retry, recovery, correlation ID propagation, and
   throttling -- essential for an MCP server handling concurrent sub-agent requests.

5. **Community and maintenance.** 9.4k stars, MIT license, stable API since v1.0,
   active development (v1.5.1 Sep 2024, 17 contributors on v1.5.0).

### Migration Path

```
Phase 1 (CLI):     Watermill + GoChannel (in-process, single binary)
Phase 2 (MCP):     Watermill + Embedded NATS (in-process, single binary, concurrent)
Phase 3 (Scale):   Watermill + External NATS (if ever needed)
```

Each phase requires ZERO changes to command handlers, event handlers, or domain code.
Only the pub/sub backend initialization changes.

### Why NOT the others?

| Option | Rejection Reason |
|--------|-----------------|
| Pure Go channels | Too much boilerplate; no CQRS, no middleware, no upgrade path |
| asaskevich/EventBus | Reflection-based, no type safety, no CQRS, low maintenance |
| mustafaturan/bus | Zero-allocation optimization irrelevant for CLI; no CQRS |
| Go-MediatR | Good mediator but no transport abstraction; could complement Watermill |
| eventhorizon | Overkill (full ES), stale maintenance (last release 2022), API "not final" |
| NATS alone | Raw transport -- still need CQRS layer on top; Watermill provides this |

### Complementary Option: Go-MediatR for Synchronous Dispatch

For purely synchronous command/query dispatch within a single bounded context
(e.g., CLI command -> use case handler -> response), Go-MediatR could complement
Watermill. Watermill handles async domain events between bounded contexts;
Go-MediatR handles sync request/response within a context. However, this adds
complexity -- evaluate whether Watermill's command bus alone is sufficient first.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Watermill development slows | Low | Medium | MIT license, fork-friendly; GoChannel backend is simple enough to maintain internally |
| GoChannel insufficient for MCP concurrency | Medium | Low | Upgrade to embedded NATS is well-documented; Watermill abstracts the switch |
| Watermill NATS v2 JetStream still "beta" | Medium | Low | Core NATS (non-JetStream) via watermill-nats v1 is stable; JetStream only needed for durability |
| Embedded NATS memory overhead too high | Low | Low | 5-20MB is acceptable on dev workstations; DontListen mode minimizes footprint |

---

## Follow-Up Actions

1. **Implementation ticket:** Set up Watermill + GoChannel as the event infrastructure
   in the Go rewrite. Define the `EventBus` and `CommandBus` ports in the application
   layer, with Watermill as the infrastructure adapter.

2. **Prototype ticket:** Build a minimal proof-of-concept with 2 bounded contexts
   communicating via Watermill domain events (e.g., "PRD Generated" event triggers
   ticket pipeline).

3. **MCP concurrency spike:** When MCP server design begins, evaluate embedded NATS
   as Watermill backend for handling concurrent sub-agent requests.

---

## Sources

- [Watermill GitHub](https://github.com/ThreeDotsLabs/watermill) -- 9.4k stars, MIT, v1.5.1
- [Watermill CQRS Docs](https://watermill.io/docs/cqrs/)
- [Watermill GoChannel Docs](https://watermill.io/pubsubs/gochannel/)
- [Watermill NATS Integration](https://github.com/ThreeDotsLabs/watermill-nats)
- [Watermill Releases](https://github.com/ThreeDotsLabs/watermill/releases)
- [Wild Workouts Go DDD Example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
- [Watermill History Blog Post](https://threedots.tech/episode/history-of-watermill/)
- [NATS Server Releases](https://github.com/nats-io/nats-server/releases) -- v2.12.4, Apache 2.0
- [NATS About Page](https://nats.io/about/) -- <20MB RAM
- [Embedding NATS in Go (gosuda)](https://gosuda.org/blog/posts/how-embedded-nats-communicate-with-go-application-z36089af0)
- [Embedding NATS in Go (dev.to)](https://dev.to/karanpratapsingh/embedding-nats-in-go-19o)
- [Synadia Embedded NATS Screencast](https://www.synadia.com/screencasts/how-to-embed-nats-server-in-your-app)
- [asaskevich/EventBus GitHub](https://github.com/asaskevich/EventBus) -- 1.9k stars, MIT
- [mustafaturan/bus GitHub](https://github.com/mustafaturan/bus) -- 340 stars, Apache 2.0
- [Go-MediatR GitHub](https://github.com/mehdihadeli/Go-MediatR) -- 276 stars, MIT
- [looplab/eventhorizon GitHub](https://github.com/looplab/eventhorizon) -- 1.7k stars, Apache 2.0
- [Practical DDD Domain Events in Go](https://www.ompluscator.com/article/golang/practical-ddd-domain-event/)
- [Go Pipelines and Fan-out](https://go.dev/blog/pipelines)
- [Go Event Bus Tutorial](https://leapcell.medium.com/how-to-create-a-event-bus-in-go-d7919b59a584)
