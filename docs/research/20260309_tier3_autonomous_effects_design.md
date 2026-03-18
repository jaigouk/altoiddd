# Research: Tier 3 Autonomous Event Side-Effects Design

**Date:** 2026-03-09
**Spike Ticket:** alto-4pp.11
**Status:** Final

## Summary

This spike investigates Tier 3 event subscribers — those that perform autonomous side-effects without user approval. After cataloging all 10 domain events and classifying their potential reactions, we recommend:

1. **Two events qualify for autonomous effects:** `TicketFlagged` and `FlagCleared`
2. **Beads integration:** Option C (BeadsLabelWriter port + subprocess adapter)
3. **Middleware stack:** Recoverer + Retry (3 attempts) with log-and-continue failure policy
4. **No blocking failures:** Autonomous subscribers must never halt the workflow

## Research Questions

1. Which domain events can safely trigger autonomous side-effects (no user approval needed)?
2. How should Go code integrate with the beads CLI for TicketFlagged → label management?
3. What is the right adapter pattern for beads-Go integration (subprocess vs direct JSONL)?
4. What are the failure modes and rollback strategies for autonomous subscribers?

---

## 1. Event Catalog and Classification

### Complete Event Inventory

| Event | Bounded Context | Source File | Current Subscribers |
|-------|-----------------|-------------|---------------------|
| `DiscoveryCompletedEvent` | Discovery | `discovery/domain/discovery_events.go` | Tier 1 (log), Tier 2 (→ artifact_generation ready) |
| `BootstrapCompletedEvent` | Bootstrap | `bootstrap/domain/bootstrap_events.go` | Tier 1 (log) |
| `DomainModelGenerated` | Shared | `shared/domain/events/events.go` | Tier 1 (log), Tier 2 (→ fitness/tickets/configs ready) |
| `GapAnalysisCompleted` | Shared | `shared/domain/events/events.go` | Tier 1 (log) |
| `ConfigsGenerated` | Shared | `shared/domain/events/events.go` | Tier 1 (log) |
| `TicketPlanApproved` | Ticket | `ticket/domain/ticket_events.go` | Tier 1 (log), Tier 2 (→ ripple_review ready) |
| `TicketFlagged` | Ticket | `ticket/domain/ticket_freshness_events.go` | Tier 1 (log) |
| `FlagCleared` | Ticket | `ticket/domain/ticket_freshness_events.go` | Tier 1 (log) |
| `FitnessTestsGenerated` | Fitness | `fitness/domain/fitness_events.go` | Tier 1 (log), Tier 2 (→ fitness completed) |
| `ConfigsGeneratedEvent` | ToolTranslation | `tooltranslation/domain/tool_config.go` | Tier 1 (log) |

### Classification Criteria

**Safe Autonomous (Tier 3):** Infrastructure bookkeeping only. No user-visible artifacts. No files written. No tickets created. Failure is acceptable (log and continue).

**Needs User Approval:** Produces files, documents, tickets, or other user-visible artifacts. alto's "Preview Before Action" principle applies.

**Not Applicable:** No side-effect reaction defined or needed.

### Classification Results

| Event | Classification | Potential Autonomous Effect | Rationale |
|-------|---------------|----------------------------|-----------|
| `TicketFlagged` | **SAFE AUTONOMOUS** | `bd label add <ticketID> review_needed` | Infrastructure bookkeeping. Label is metadata, not user-visible artifact. Failure = label missing (acceptable). |
| `FlagCleared` | **SAFE AUTONOMOUS** | `bd label remove <ticketID> review_needed` | Same as above. |
| `BootstrapCompletedEvent` | NOT APPLICABLE | (none) | Session log already handled by slog. No additional side-effect needed. |
| `ConfigsGenerated` | NEEDS APPROVAL | Write to `.alto/knowledge/tools/` | Cache update produces files. User should approve cache writes. |
| `DiscoveryCompletedEvent` | NEEDS APPROVAL | Artifact generation | Already handled by Tier 2 readiness → handler invocation. |
| `DomainModelGenerated` | NEEDS APPROVAL | Multiple downstream workflows | Triggers fitness/tickets/configs — all need approval. |
| `TicketPlanApproved` | NOT APPLICABLE | Ripple review | Already marks ripple_review ready; review itself requires user action. |
| `FitnessTestsGenerated` | NOT APPLICABLE | (none) | Already marks fitness completed. Tests written require prior approval. |
| `GapAnalysisCompleted` | NOT APPLICABLE | (none) | Analysis output shown to user. No autonomous effect needed. |
| `ConfigsGeneratedEvent` | NEEDS APPROVAL | Config file writes | Files require prior approval before write. |

### Key Insight

Only **label operations** qualify as safe autonomous effects:
- Labels are infrastructure metadata, not user-visible artifacts
- `review_needed` is a workflow signal, not content
- Failure simply means manual label management (existing workflow)

---

## 2. Beads Integration Analysis

### Existing Patterns in Codebase

**Pattern A: SubprocessGateRunner** (fitness infrastructure)
```go
// internal/fitness/infrastructure/subprocess_gate_runner.go:49-98
command := exec.CommandContext(execCtx, cmd[0], cmd[1:]...)
command.Dir = r.projectDir
output, err := command.CombinedOutput()
```
- Runs external commands with timeout
- Captures combined stdout/stderr
- Maps exit codes to domain results
- 300-second default timeout

**Pattern B: BeadsTicketReader** (ticket infrastructure)
```go
// internal/ticket/infrastructure/beads_ticket_reader.go:153-170
cmd := exec.CommandContext(ctx, "bd", "query", "label=review_needed")
output, err := cmd.Output()
```
- Runs `bd` CLI commands with 10-second timeout
- Parses stdout for structured data
- Silently returns empty on error (graceful degradation)

### Integration Options

| Option | Approach | Pros | Cons |
|--------|----------|------|------|
| **A: Subprocess only** | Call `bd label add/remove` directly | Simple, uses existing CLI, no format coupling | Subprocess overhead (~50ms), requires `bd` in PATH |
| **B: Direct JSONL** | Read/write `.beads/issues.jsonl` | Fastest, no subprocess | Couples to beads format, must handle locking, bypasses beads hooks |
| **C: Port + Adapter** | Define `BeadsLabelWriter` port, implement subprocess adapter | DDD-compliant, testable, future-proof | More code initially |

### Recommendation: Option C

**Define a port interface in `ticket/application/ports.go`:**
```go
// BeadsLabelWriter manages ticket labels in the beads issue tracker.
type BeadsLabelWriter interface {
    AddLabel(ctx context.Context, ticketID, label string) error
    RemoveLabel(ctx context.Context, ticketID, label string) error
}
```

**Implement subprocess adapter in `ticket/infrastructure/beads_label_adapter.go`:**
```go
type BeadsLabelAdapter struct {
    timeout time.Duration
}

func (a *BeadsLabelAdapter) AddLabel(ctx context.Context, ticketID, label string) error {
    ctx, cancel := context.WithTimeout(ctx, a.timeout)
    defer cancel()
    cmd := exec.CommandContext(ctx, "bd", "label", "add", ticketID, label)
    _, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("adding label %s to %s: %w", label, ticketID, err)
    }
    return nil
}
```

**Rationale:**
1. Follows existing DDD/ports pattern (`SubprocessGateRunner`, `BeadsTicketReader`)
2. Testable via mock adapter
3. Future flexibility (could add direct JSONL adapter if performance matters)
4. No coupling to beads internal format

---

## 3. Failure Mode Analysis

### Subscriber Failure Scenarios

| Scenario | Impact | Mitigation |
|----------|--------|------------|
| `bd label add` fails (bd not found) | Label missing | Log error, continue. User can manually `bd label add`. |
| `bd label add` fails (beads DB locked) | Label missing | Retry 3x with backoff, then log and continue. |
| `bd label add` times out | Label missing | 5-second timeout. Log and continue. |
| Subscriber panics | Subscriber crashes | Recoverer middleware catches panic, logs stack trace, continues. |
| Event malformed (bad JSON) | Event dropped | JSONMarshaler returns error, message Nack'd, logged. |

### Failure Policy

**Autonomous subscribers must never block the workflow.**

- Retry: 3 attempts with exponential backoff (100ms → 200ms → 400ms)
- Timeout: 5 seconds per attempt
- On exhaustion: Log error at WARN level, acknowledge message, continue
- No poison queue needed (label operations are idempotent; retry at next event)

### Watermill Middleware Stack

Based on prior research (`docs/research/20260307_6_watermill_cqrs_deep_dive.md`) and Context7 documentation:

```go
// For Tier 3 autonomous subscriber handlers
router.AddMiddleware(
    middleware.Recoverer,  // Catch panics, convert to errors
    middleware.Retry{
        MaxRetries:      3,
        InitialInterval: 100 * time.Millisecond,
        MaxInterval:     500 * time.Millisecond,
        Multiplier:      2.0,
        Logger:          watermillLogger,
    }.Middleware,
)
```

**Not recommended for Tier 3:**
- `middleware.PoisonQueue` — Label ops are idempotent; dead-letter adds complexity
- `middleware.CircuitBreaker` — CLI runs are short-lived; circuit state doesn't persist
- `middleware.Timeout` — Already handled by subprocess context timeout

---

## 4. Implementation Design

### Wiring in Composition Root

```go
// internal/composition/event_subscribers.go

// ===========================================================================
// Tier 3 — Autonomous side-effects (infrastructure bookkeeping only)
// ===========================================================================

labelWriter := ticketinfra.NewBeadsLabelAdapter(5 * time.Second)

// TicketFlagged → add review_needed label
errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *ticketdomain.TicketFlagged) error {
    if err := labelWriter.AddLabel(ctx, evt.TicketID(), "review_needed"); err != nil {
        logger.WarnContext(ctx, "autonomous.label_add_failed",
            "event", "TicketFlagged",
            "ticket_id", evt.TicketID(),
            "error", err,
        )
        // Do NOT return error — acknowledge and continue
    }
    return nil
}))

// FlagCleared → remove review_needed label
errs = append(errs, eventbus.SubscribeTyped(sub, func(ctx context.Context, evt *ticketdomain.FlagCleared) error {
    if err := labelWriter.RemoveLabel(ctx, evt.TicketID(), "review_needed"); err != nil {
        logger.WarnContext(ctx, "autonomous.label_remove_failed",
            "event", "FlagCleared",
            "ticket_id", evt.TicketID(),
            "error", err,
        )
    }
    return nil
}))
```

### Key Design Decisions

1. **Errors are logged, not propagated.** Autonomous subscribers return `nil` even on failure. This ensures the event is acknowledged and the workflow continues.

2. **Idempotency by design.** `bd label add` on an already-labeled ticket is a no-op. `bd label remove` on an unlabeled ticket is a no-op. Safe for retry.

3. **Separate adapter, not inline subprocess.** The `BeadsLabelAdapter` encapsulates subprocess details, making the subscriber testable with a mock.

4. **WarnContext, not ErrorContext.** Failed label operations are operational anomalies, not application errors. WARN level indicates "human should investigate if frequent."

---

## 5. Answers to Research Questions

### Q1: Which events can safely trigger autonomous side-effects?

**Answer:** Only `TicketFlagged` and `FlagCleared`.

These events manage the `review_needed` label, which is infrastructure metadata (workflow signal), not a user-visible artifact. All other events either:
- Produce files/documents (violates "Preview Before Action")
- Already have adequate handling (Tier 1 logging, Tier 2 readiness)
- Have no defined side-effect

### Q2: How should Go code integrate with beads CLI?

**Answer:** Define a `BeadsLabelWriter` port interface in `ticket/application/ports.go`, implemented by a subprocess adapter in `ticket/infrastructure/beads_label_adapter.go`.

This follows the existing patterns (`SubprocessGateRunner`, `BeadsTicketReader`) and provides:
- DDD compliance (port/adapter separation)
- Testability (mock adapter in tests)
- Future flexibility (could add JSONL adapter if performance matters)

### Q3: Subprocess vs direct JSONL read/write?

**Answer:** Subprocess for Tier 3 implementation.

Direct JSONL access would couple alto to beads internal format and bypass beads hooks (git sync, validation). The subprocess overhead (~50ms) is negligible for label operations that occur once per ticket close.

### Q4: Failure modes and rollback strategies?

**Answer:** Log-and-continue with retry.

- 3 retries with exponential backoff (100ms → 400ms)
- 5-second timeout per attempt
- On exhaustion: Log at WARN level, acknowledge message, continue
- No rollback needed (label operations are idempotent)
- User can manually `bd label add/remove` if automation fails

---

## 6. Follow-Up Tasks

### Implementation Tickets

1. **Create `BeadsLabelWriter` port and adapter**
   - Port interface in `ticket/application/ports.go`
   - Subprocess adapter in `ticket/infrastructure/beads_label_adapter.go`
   - Unit tests for adapter (mock exec, verify command construction)

2. **Wire Tier 3 subscribers in composition root**
   - Add `TicketFlagged → AddLabel` subscriber
   - Add `FlagCleared → RemoveLabel` subscriber
   - Integration test: publish event → verify label via `bd label list`

3. **Add retry middleware to event subscriber wiring**
   - Currently using `BlockPublishUntilSubscriberAck: true` with no retry
   - Consider adding Watermill Router + middleware for Tier 3 handlers only

### Deferred Decisions

- **ConfigsGenerated → cache update:** Deferred. Cache management should go through existing handlers with user approval, not autonomous effects.
- **Direct JSONL access:** Deferred. Subprocess is sufficient. Revisit if performance profiling shows label ops are a bottleneck.

---

## References

- [internal/fitness/infrastructure/subprocess_gate_runner.go](../../internal/fitness/infrastructure/subprocess_gate_runner.go) — Subprocess pattern reference
- [internal/ticket/infrastructure/beads_ticket_reader.go](../../internal/ticket/infrastructure/beads_ticket_reader.go) — Beads CLI integration pattern
- [internal/composition/event_subscribers.go](../../internal/composition/event_subscribers.go) — Current Tier 1/2 wiring
- [docs/research/20260307_6_watermill_cqrs_deep_dive.md](./20260307_6_watermill_cqrs_deep_dive.md) — Middleware patterns, retry configuration
- [Watermill Middleware Docs](https://watermill.io/docs/middlewares/) — Official middleware reference
