---
name: groom
description: Deep-groom a ticket — enforced implementation simulation, scope check, split detection
kind: command
phase: groom
when_to_use: When deep-grooming a ticket before claiming — implementation simulation, scope check, split detection
tools: Read, Grep, Glob, Bash
bash_substitution_policy: quoted
license: Apache-2.0
---

# /groom <ticket-id>

Deep-groom a single ticket before claiming it. Delegates to existing tools where they exist, then runs the **implementation simulation** — the step that catches "magic happens here" gaps.

## Why This Exists

The grooming checklist in CLAUDE.md has 8 steps. Steps 1-7 are mechanical checks. Step 8 (Implementation Simulation) is the one that gets skipped and catches every real gap: missing adapters, signature mismatches, scope overload. This command makes Step 8 rigorous and unavoidable.

## Relation to `/design-ticket`

`/design-ticket` produces a verified ticket BEFORE creation. `/groom` is what you run AFTER creation but BEFORE claiming the work — it re-validates the ticket against the **current** state of the code, which may have drifted since the ticket was designed. For freshly-designed tickets, Phases 1-3 are usually no-ops; Phase 4 (Implementation Simulation) earns its keep by catching code drift since design. For older tickets, the full grooming pass matters. See [`design-ticket.md`](design-ticket.md) for the upstream design flow.

## Usage

```
/groom <ticket-id>           # Deep-groom one ticket
```

Always groom ONE ticket at a time. Never batch-groom.

## Process

### Phase 1 — Load Context

```bash
bd show <ticket-id>
bd comments <ticket-id>
bd dep list <ticket-id>
bd label list <ticket-id>
```

### Phase 2 — Delegate to Existing Tools

| Check | Tool | Action |
|-------|------|--------|
| Freshness | `bd label list <id>` | If `review_needed` → read ripple comments, resolve before proceeding |
| PRD traceability | `/prd-traceability <id>` | Cross-reference ticket AC against PRD capabilities |
| Template compliance | Manual | Compare description against the matching template: `beads-ticket-template.md` (task), `beads-epic-template.md` (epic), `beads-spike-template.md` (spike), `beads-bug-template.md` (bug — see also `/rca`). All under `alto-scaffold/templates/`. |

If template sections are missing → draft them before proceeding.

### Phase 3 — Quick Checks

Verify (can be done from the ticket description alone):

- [ ] **DDD alignment** — ticket stays within its bounded context; no cross-context leakage
- [ ] **Ubiquitous language** — class/method names match `docs/DDD.md` glossary
- [ ] **TDD phases** — RED/GREEN/REFACTOR with specific test names and file paths
- [ ] **SOLID mapping** — concrete implementations, not generic placeholders
- [ ] **AC testability** — every acceptance criterion is testable, not vague

### Phase 4 — Implementation Simulation (THE CRITICAL STEP)

**This is not optional. Read actual code. Trace actual chains.**

#### 4a. Read every referenced source file

For EVERY file the ticket mentions or depends on, use the Read tool:
- Port interfaces (`application/ports.go` or handler-local interfaces)
- Existing handlers in the same bounded context (pattern reference)
- Domain types used in signatures (entities, value objects, result types)
- Infrastructure adapters (existing or to-be-created)
- Composition root (project-specific path — see `groom.project.md`)

**Do NOT say "verified" without citing the file and line you read.**

#### 4b. Trace the constructor chain

For each new struct the ticket creates, write out the full chain:

```
NewXxxHandler(port)
  → port type: XxxPort interface at {application_layer}/ports.<ext>:NN
  → methods: Foo(ctx, string) (Result, error) — confirmed line NN
  → adapter: XxxAdapter at {infrastructure_layer}/xxx_adapter.<ext>
  → adapter constructor: NewXxxAdapter(dep1, dep2)
  → dep1 comes from: ...
  → wired in NewApp() at {composition_root}:NN  # project-specific
  → imports needed: xxxapp, xxxinfra
```

If any link in this chain is "TBD" or unresolved → **FAIL**.

#### 4c. Verify every method signature

For each method the ticket will call:
- Read the actual interface definition
- Compare parameter types and return types **exactly**
- Flag mismatches (e.g., ticket says `string` but port has `*time.Time`)

#### 4d. Check for interface mismatches

Read `{composition_root}/adapters.<ext>` (see project overlay). If adapter method signatures don't match the port, an adapter bridge is needed. The ticket must specify this.

#### 4e. Cross-reference within the ticket

- Does the Design data model table match the port signatures you just read?
- Does the sequence diagram match what the methods actually do?
- Do the Steps include updating existing tests that will break?
- Does the TDD section have tests for every AC item?

### Phase 5 — Scope Check

Count from the ticket:

| Metric | Threshold | Action |
|--------|-----------|--------|
| New files to create | > 5 | Flag for split |
| New functions/methods | > 20 | Flag for split |
| Bounded contexts touched | > 2 | Flag for split |
| Security + business logic | mixed | Must split |

If ANY threshold exceeded → recommend specific split.

### Phase 6 — Report

```
================================================================
GROOMING REPORT: <ticket-id>
================================================================

TEMPLATE:          [PASS | FAIL — missing: <sections>]
FRESHNESS:         [CLEAN | STALE]
PRD TRACEABILITY:  [COVERED | GAP — <ids>]
DDD/LANGUAGE:      [PASS | FAIL]
TDD/SOLID/AC:      [PASS | FAIL]
CLAIM VERIFICATION: [VERIFIED | MISMATCH | NO CLAIMS | UNVERIFIED]

IMPLEMENTATION SIMULATION:
  Files read:       <N> (list each with path)
  Constructor chain: [COMPLETE | BROKEN at <link>]
  Signature mismatches: <N>
    - <method>: ticket says X, port has Y (file:line)
  Missing adapters: <N>
  Missing imports:  <N>
  Bridges needed:   <N>

SCOPE:
  New files: <N>  |  Functions: <N>  |  Contexts: <N>
  [OK | SPLIT RECOMMENDED — <reason>]

================================================================
VERDICT: [READY | NEEDS UPDATE | NEEDS SPLIT]
================================================================
```

## Rules

1. **One ticket at a time.** Never batch-groom multiple tickets in one pass.
2. **Read before you verify.** Every claim must cite the file:line you actually read.
3. **Trace before you approve.** Write out the constructor chain. No shortcuts.
4. **Split before you bloat.** Two focused tickets > one that gets re-scoped mid-implementation.
5. **No false confidence.** If unsure about any step, mark FAIL and say what's unclear.
