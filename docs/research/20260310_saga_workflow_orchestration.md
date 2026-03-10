# Research: Saga/Workflow Orchestration for Human-in-the-Loop Event Chains

**Date:** 2026-03-10
**Spike Ticket:** alty-4pp.13
**Status:** Complete
**Dependencies:** alty-4pp.10 (SessionTracker), alty-4pp.11 (Tier 3 autonomous effects)

## Summary

This spike investigates whether alty should adopt saga/process manager patterns for orchestrating its multi-step bootstrap workflow (discovery -> artifacts -> fitness -> tickets -> configs), given that every step requires human approval ("preview before action").

**Recommendation:** Do NOT adopt a formal saga pattern. Instead, evolve the existing SessionTracker into a **WorkflowCoordinator** -- a lightweight state machine that tracks step completion, enforces ordering constraints, and provides "what's next?" queries. This gives us the benefits of saga observability without the complexity of event-driven saga orchestration, which is architecturally mismatched with alty's human-in-the-loop constraint.

---

## 1. Research Questions Answered

### Q1: How should alty orchestrate multi-step workflows?

**Answer: Imperative orchestration with state tracking is the correct pattern for alty.**

The key insight is that alty's workflow is **not a saga** in the distributed systems sense. Sagas solve the problem of coordinating transactions across services that may fail independently and need compensating actions. Alty's workflow has none of these characteristics:

| Saga Property | Alty's Workflow | Match? |
|---------------|-----------------|--------|
| Distributed transactions | Single process, single machine | No |
| Compensating actions needed | No rollback (preview prevents errors) | No |
| Services fail independently | All handlers in same process | No |
| Long-running (hours/days) | Minutes per session | No |
| Multiple actors | Single human + alty CLI | No |

What alty actually needs is **workflow state tracking**: knowing which steps are done, which are available, and what the user should do next. This is a state machine problem, not a saga problem.

**Sources:**
- [Saga Pattern Demystified](https://blog.bytebytego.com/p/saga-pattern-demystified-orchestration) -- Sagas coordinate distributed transactions
- [Saga and Process Manager](https://event-driven.io/en/saga_process_manager_distributed_transactions/) -- Process managers track state across events
- `docs/ARCHITECTURE.md:51-53` -- "Human-in-the-loop: The system flags, suggests, and previews. Humans decide."

### Q2: What saga/workflow patterns work for human-in-the-loop CLI + MCP?

**Answer: The "Checkpoint + Resume" pattern, not reactive event chains.**

The fundamental mismatch between sagas and alty's workflow is the **control inversion**:

- **Saga pattern:** Event A fires -> subscriber automatically triggers step B -> event B fires -> subscriber triggers step C. The system drives progress.
- **Alty's reality:** Step A completes -> user sees preview for step B -> user decides IF and WHEN to proceed -> user explicitly calls step B. The human drives progress.

Temporal.io handles this elegantly with **signals** -- a workflow blocks on `workflow.GetSignalChannel()` waiting for external input. But Temporal requires a server infrastructure (Temporal Server + database) that is completely mismatched with alty's "local-first, zero network" constraint.

The correct pattern for alty is:

```
Step completes -> emit event -> SessionTracker marks next steps "ready"
                                -> CLI/MCP returns "next steps available" list
                                -> User explicitly invokes next step
                                -> Handler checks "is this step ready?" before executing
```

This is what alty already does (SessionTracker from alty-4pp.10), but informally. The improvement is formalizing it.

**Sources:**
- [Temporal Human-in-the-Loop](https://docs.temporal.io/ai-cookbook/human-in-the-loop-python) -- Signal-based wait for human input
- [Temporal Long-Running Workflows](https://temporal.io/blog/very-long-running-workflows) -- Workflow.GetSignalChannel for human approval
- `internal/shared/domain/session_tracker.go:108-119` -- Current MarkReady/MarkCompleted pattern

### Q3: Should alty use Watermill's Router/CQRS EventProcessor or a custom saga coordinator?

**Answer: Neither. Watermill does not have a built-in saga/process manager.**

Key findings:

1. **Watermill has no built-in saga support.** GitHub issue [#7](https://github.com/ThreeDotsLabs/watermill/issues/7) (opened 2018) remains open. Maintainer Robert Laszczak confirmed it's a "long term plan" but not implemented. Users must build their own.

2. **Watermill's CQRS example shows a "saga" that is just event chaining** -- event handler reacts to event A by publishing command B. This is choreography, not orchestration. It works when steps are fully autonomous but breaks when human approval gates exist between steps.

3. **Watermill's Router is designed for autonomous message processing** -- handlers consume from a topic and produce to another topic. There is no built-in "pause and wait for external signal" capability.

4. **Current eventbus abstraction works well for what it does** (Tier 1 logging, Tier 2 readiness, Tier 3 autonomous effects). The problem is not with the event bus; it's that saga orchestration is the wrong pattern for alty's workflow.

**Sources:**
- [Watermill Issue #7: Saga Support](https://github.com/ThreeDotsLabs/watermill/issues/7) -- Still open, no built-in support
- [Watermill CQRS Docs](https://watermill.io/docs/cqrs/) -- Event chaining example (not true saga)
- `internal/shared/infrastructure/eventbus/bus.go:18-26` -- GoChannel with BlockPublishUntilSubscriberAck=true
- `docs/research/20260307_6_watermill_cqrs_deep_dive.md` -- Prior research on Watermill CQRS patterns

### Q4: How do saga patterns interact with GoChannel's synchronous semantics?

**Answer: Poorly. GoChannel + BlockPublishUntilSubscriberAck creates deadlock risk for saga patterns.**

With `BlockPublishUntilSubscriberAck: true` (current config, `bus.go:22`), `Publish()` blocks until all subscribers have called `msg.Ack()`. In a saga pattern, if step A's event handler tries to publish step B's event, and step B's subscriber tries to publish step C's event, the call chain becomes:

```
Publish(A) -> blocks until subscriber Acks
  -> subscriber handles A, Publishes(B) -> blocks until subscriber Acks
    -> subscriber handles B, Publishes(C) -> blocks
```

This creates a deep synchronous call chain on a single goroutine. If any step in the chain blocks waiting for human input, the entire chain deadlocks. This is fundamentally incompatible with human-in-the-loop approval gates.

GoChannel limitations relevant to saga patterns:
- No persistence: saga state lost on crash
- No consumer groups: can't have multiple saga instances processing events independently
- Synchronous delivery: `BlockPublishUntilSubscriberAck` means events are processed inline

These limitations confirm that GoChannel is designed for notification-style events (fire-and-observe), not for saga coordination.

**Sources:**
- `internal/shared/infrastructure/eventbus/subscriber.go:139-142` -- "Messages are always Ack'd after processing... Nack causes infinite redelivery which deadlocks"
- `internal/shared/infrastructure/eventbus/bus.go:20-22` -- BlockPublishUntilSubscriberAck: true
- `docs/research/20260307_6_watermill_cqrs_deep_dive.md` -- GoChannel limitations section

### Q5: What does the migration path look like from imperative to event-driven orchestration?

**Answer: There is no migration needed. Imperative orchestration with state tracking IS the right architecture for alty.**

The current architecture is correct for the problem domain:

1. **Imperative tool handlers** (`tools_bootstrap.go`) -- Each MCP tool calls the appropriate handler directly. Simple, debuggable, no hidden event chains.
2. **Event-based notifications** (Tier 1/2/3) -- Events fire AFTER operations complete to update readiness state and perform bookkeeping. Events don't drive the workflow.
3. **SessionTracker** (`session_tracker.go`) -- Tracks which steps are ready/completed. Answers "what can the user do next?"

The improvement path is not "imperative -> event-driven saga" but rather "informal state tracking -> formal state machine":

| Current | Evolved |
|---------|---------|
| SessionTracker with ready/completed maps | WorkflowCoordinator with defined state machine |
| MCP tools check handler availability | MCP tools check WorkflowCoordinator.CanExecute(step) |
| ModelStore caches DomainModel by sessionID | WorkflowCoordinator holds session context (model, profile, projectDir) |
| No crash recovery | Optional: SQLite persistence for session state |

---

## 2. Comparison Table: Orchestration Approaches

| Criterion | Imperative (Current) | Event-Driven Saga | State Machine Coordinator (Recommended) |
|-----------|---------------------|-------------------|----------------------------------------|
| **Complexity** | Low | High | Medium |
| **Human-in-the-loop** | Natural (user calls tools) | Unnatural (events pause, resume) | Natural (user queries, then acts) |
| **Debuggability** | Excellent (stack traces) | Poor (event chains) | Good (state transitions logged) |
| **Ordering enforcement** | Implicit (tool docs) | Event dependencies | Explicit (state machine guards) |
| **Crash recovery** | None (in-memory) | Requires persistent event store | Optional (persist state to file/SQLite) |
| **Concurrent sessions** | Via SessionTracker | Complex (saga instance per session) | Via WorkflowCoordinator (map of sessions) |
| **New step addition** | Add tool handler | Add event + subscriber + saga step | Add state + transition |
| **Testing** | Unit test handlers | Integration test event chains | Unit test state machine |
| **Dependencies** | None new | Watermill Router + middleware | qmuntal/stateless (optional) |
| **Fits alty's domain** | Yes | No (alty is not distributed) | Yes |

---

## 3. GoChannel vs SQLite Decision Matrix for Session State

| Factor | GoChannel (Current) | SQLite Persistence | Recommendation |
|--------|--------------------|--------------------|----------------|
| **Session survival** | Lost on crash/exit | Survives restart | GoChannel is fine for CLI (sessions are minutes long) |
| **MCP server sessions** | Lost if MCP server restarts | Survives restart | SQLite valuable when MCP server runs as daemon |
| **Storage overhead** | Zero | ~4KB per session | Negligible |
| **Implementation effort** | Already done | New adapter, migration | Defer until MCP daemon mode |
| **Complexity** | Zero | Schema management, file locking | Avoid premature |
| **When to upgrade** | -- | When MCP runs as long-lived daemon | Phase 2 |

**Decision: Stay with in-memory GoChannel for now.** The CLI is short-lived (minutes). MCP via stdio is also short-lived (one conversation). SQLite persistence becomes valuable only when alty runs as a long-lived MCP server (daemon mode), which is not currently on the roadmap.

If crash recovery becomes important, the simplest path is persisting `WorkflowCoordinator` state to a JSON file in `.alty/sessions/`, not adopting SQLite pub/sub.

---

## 4. Minimal Viable Design: WorkflowCoordinator

### Concept

Evolve `SessionTracker` into `WorkflowCoordinator` -- a formal state machine per session that:
1. Defines the allowed step transitions
2. Enforces preconditions before execution
3. Provides "what's next?" queries for CLI/MCP
4. Holds session context (DomainModel, StackProfile, ProjectDir) -- absorbing ModelStore

### State Machine Definition

```
States per step: PENDING -> READY -> IN_PROGRESS -> COMPLETED | SKIPPED

Transitions:
  DiscoveryCompleted event -> artifact_generation becomes READY
  ArtifactGeneration completed -> fitness, tickets, configs become READY
  Fitness completed/skipped -> (no downstream)
  Tickets completed -> ripple_review becomes READY
  Configs completed/skipped -> (no downstream)

Guards:
  CanExecute(step) returns true only if step is READY
  User can Skip(step) for optional steps (fitness, configs)
```

### Interface Design

```go
// WorkflowCoordinator manages bootstrap workflow state per session.
// Lives in internal/shared/domain/ (pure domain, no external deps).
type WorkflowCoordinator struct {
    mu       sync.RWMutex
    sessions map[string]*WorkflowSession
}

type WorkflowSession struct {
    sessionID  string
    steps      map[WorkflowStep]StepStatus
    context    SessionContext // DomainModel, StackProfile, ProjectDir
    createdAt  time.Time
}

type StepStatus int
const (
    StepPending StepStatus = iota
    StepReady
    StepInProgress
    StepCompleted
    StepSkipped
)

// Key methods:
func (c *WorkflowCoordinator) MarkReady(sessionID string, steps ...WorkflowStep)
func (c *WorkflowCoordinator) BeginStep(sessionID string, step WorkflowStep) error  // READY -> IN_PROGRESS
func (c *WorkflowCoordinator) CompleteStep(sessionID string, step WorkflowStep) error // IN_PROGRESS -> COMPLETED
func (c *WorkflowCoordinator) SkipStep(sessionID string, step WorkflowStep) error    // READY -> SKIPPED
func (c *WorkflowCoordinator) CanExecute(sessionID string, step WorkflowStep) bool
func (c *WorkflowCoordinator) AvailableActions(sessionID string) []ReadyAction
func (c *WorkflowCoordinator) SessionContext(sessionID string) (*SessionContext, error)
```

### How It Integrates

1. **Event subscribers (Tier 2)** call `coordinator.MarkReady()` -- same as current SessionTracker
2. **MCP tool handlers** call `coordinator.CanExecute()` before running -- replaces nil checks
3. **MCP tool handlers** call `coordinator.BeginStep()` / `coordinator.CompleteStep()` -- provides observability
4. **MCP "what's next?" resource** calls `coordinator.AvailableActions()` -- user sees what they can do
5. **ModelStore** is absorbed into `SessionContext` -- single source of session state

### Relationship to External Libraries

**qmuntal/stateless** was evaluated as a potential FSM library:
- License: BSD-2-Clause (permissive)
- Version: v1.8.0 (February 2026)
- Features: Guard clauses, entry/exit callbacks, hierarchical states, DOT graph export
- Context support: Yes
- Thread safety: Yes
- CGO: No (pure Go)

**Decision: Do NOT adopt stateless.** Alty's workflow state machine is simple enough (5 steps, 4 states, linear transitions) that a hand-rolled implementation is clearer, has zero dependencies, and avoids learning curve. The `stateless` library shines for complex FSMs with many states, guards, and nested hierarchies -- alty's workflow is none of these.

If the workflow grows significantly more complex (10+ steps, conditional branching, nested sub-workflows), revisit `stateless`.

**Sources:**
- [qmuntal/stateless](https://github.com/qmuntal/stateless) -- BSD-2, v1.8.0, pure Go, thread-safe
- `internal/shared/domain/session_tracker.go` -- Current implementation to evolve
- `internal/mcp/model_store.go` -- Session context to absorb

---

## 5. Migration Path from Current State

### Phase 1: Formalize WorkflowCoordinator (immediate, 1 ticket)

1. Rename `SessionTracker` to `WorkflowCoordinator`
2. Add `StepStatus` enum (Pending, Ready, InProgress, Completed, Skipped)
3. Add `BeginStep()`, `CompleteStep()`, `SkipStep()` methods
4. Add `CanExecute()` guard method
5. Add `SessionContext` struct to hold DomainModel + StackProfile + ProjectDir
6. Update event subscribers (Tier 2) to use new method names
7. All domain-only changes, no external deps

### Phase 2: Integrate with MCP Tools (next sprint, 1 ticket)

1. MCP tool handlers call `CanExecute()` before processing
2. MCP tool handlers call `BeginStep()` / `CompleteStep()` around execution
3. Remove `ModelStore` -- session context lives in `WorkflowCoordinator`
4. Add MCP resource `session_status` that returns `AvailableActions()`
5. Tool descriptions updated to show preconditions ("requires generate_artifacts first")

### Phase 3: Optional Persistence (deferred until MCP daemon mode)

1. Persist `WorkflowSession` state to `.alty/sessions/<session-id>.json`
2. On startup, load existing sessions
3. TTL-based cleanup of old sessions
4. Only pursue when MCP server runs as a long-lived daemon

### What We Explicitly Do NOT Do

- **No Watermill Router for workflow orchestration** -- Router is for autonomous message processing
- **No saga coordinator** -- alty's workflow is not a distributed transaction
- **No SQLite event store** -- in-memory state is fine for short-lived CLI sessions
- **No Temporal/workflow engine** -- massive overkill for 5-step linear flow
- **No compensating transactions** -- preview-before-action prevents the need for rollback

---

## 6. Answers to Spike Research Questions

### Q1: How should alty orchestrate multi-step workflows?
**Keep imperative orchestration.** Evolve SessionTracker into a WorkflowCoordinator that formalizes step ordering and preconditions. The human drives the workflow; events notify about state changes.

### Q2: What patterns work for hybrid CLI + MCP with human-in-the-loop?
**Checkpoint + resume pattern.** Each step completes, marks downstream steps ready, returns available actions to the user. User explicitly invokes the next step. No automatic progression.

### Q3: Watermill Router/CQRS or custom coordinator?
**Neither as saga infrastructure.** Keep Watermill for event notifications (Tier 1/2/3). Use a custom WorkflowCoordinator for workflow state. These are separate concerns.

### Q4: GoChannel + BlockPublishUntilSubscriberAck and sagas?
**Incompatible.** Synchronous delivery + human approval gates = deadlock risk. This confirms that events should be notifications, not workflow drivers.

### Q5: Migration path?
**Three phases:** (1) Formalize WorkflowCoordinator from SessionTracker, (2) Integrate with MCP tools, (3) Optional persistence for daemon mode. No architectural rewrite needed.

---

## 7. Follow-Up Tickets

**Decision (2026-03-10):** No epic needed. The scope is small (2 tickets for immediate work), this is evolution of existing code (not a new system), and Phase 3 is explicitly deferred until MCP daemon mode becomes a priority. Standalone tickets with proper dependencies are sufficient.

### alty-vf2: Evolve SessionTracker into WorkflowCoordinator

**Ticket ID:** `alty-vf2`
**Type:** Task
**Priority:** P2
**Bounded Context:** shared/domain
**Status:** Open, Ready

**Description:** Rename and extend `SessionTracker` to `WorkflowCoordinator` with formal step status tracking (Pending/Ready/InProgress/Completed/Skipped), `CanExecute()` guard, `BeginStep()`/`CompleteStep()`/`SkipStep()` lifecycle methods, and `SessionContext` to hold DomainModel + StackProfile + ProjectDir. Pure domain changes, no external deps. Update event subscribers (Tier 2) to use new coordinator. Absorb `ModelStore` into `SessionContext`.

**AC:**
- [ ] StepStatus enum with 5 states
- [ ] WorkflowCoordinator with CanExecute, BeginStep, CompleteStep, SkipStep
- [ ] SessionContext holds DomainModel, StackProfile, ProjectDir
- [ ] Tier 2 event subscribers updated
- [ ] ModelStore functionality absorbed (Put/Get → SetSessionContext/SessionContext)
- [ ] 90%+ test coverage on state transitions
- [ ] All quality gates pass

### alty-6q5: Integrate WorkflowCoordinator with MCP Tools

**Ticket ID:** `alty-6q5`
**Type:** Task
**Priority:** P2
**Bounded Context:** mcp, composition
**Depends on:** alty-vf2
**Status:** Open, Blocked by alty-vf2

**Description:** Update MCP tool handlers in `tools_bootstrap.go` to use `WorkflowCoordinator.CanExecute()` before processing and `BeginStep()`/`CompleteStep()` around execution. Add MCP resource `session_status` returning available actions. Remove `ModelStore`. Update tool descriptions with precondition documentation.

**AC:**
- [ ] MCP tools return clear error when precondition not met
- [ ] MCP tools report BeginStep/CompleteStep lifecycle
- [ ] session_status resource shows available actions
- [ ] ModelStore removed (session context in coordinator)
- [ ] Integration test: full workflow sequence via MCP tools
- [ ] All quality gates pass

---

## 8. Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| WorkflowCoordinator becomes complex (many steps, branching) | Low | Medium | Monitor step count; adopt qmuntal/stateless if > 10 steps |
| Session state loss on CLI crash | Low | Low | CLI sessions are minutes long; user re-runs |
| MCP daemon mode needs persistence | Medium | Medium | Phase 3 design ready; JSON file persistence is simple |
| Over-engineering: coordinator adds complexity without value | Low | Medium | Coordinator is ~100 lines domain code; net simplification by absorbing ModelStore |

---

## References

- [Saga Pattern Demystified](https://blog.bytebytego.com/p/saga-pattern-demystified-orchestration) -- Orchestration vs choreography comparison
- [Saga and Process Manager](https://event-driven.io/en/saga_process_manager_distributed_transactions/) -- When to use each pattern
- [Watermill Issue #7: Saga Support](https://github.com/ThreeDotsLabs/watermill/issues/7) -- Still open, no built-in support
- [Watermill CQRS Docs](https://watermill.io/docs/cqrs/) -- Event chaining example
- [Temporal Human-in-the-Loop](https://docs.temporal.io/ai-cookbook/human-in-the-loop-python) -- Signal-based approval pattern
- [qmuntal/stateless](https://github.com/qmuntal/stateless) -- Go FSM library (BSD-2, v1.8.0, evaluated but not recommended)
- `internal/shared/domain/session_tracker.go` -- Current SessionTracker implementation
- `internal/mcp/tools_bootstrap.go` -- Current imperative tool orchestration
- `internal/mcp/model_store.go` -- Session-scoped DomainModel cache
- `internal/composition/event_subscribers.go` -- Tier 1/2/3 event wiring
- `internal/shared/infrastructure/eventbus/bus.go` -- GoChannel config (BlockPublishUntilSubscriberAck)
- `docs/ARCHITECTURE.md:51-53` -- "Human-in-the-loop" design principle
- `docs/PRD.md:119` -- "alty init with preview" (P0 requirement)
- `docs/research/20260307_6_watermill_cqrs_deep_dive.md` -- Prior Watermill CQRS research
- `docs/research/20260309_tier3_autonomous_effects_design.md` -- Tier 3 subscriber design
