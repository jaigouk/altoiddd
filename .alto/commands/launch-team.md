---
name: launch-team
description: Generate a team launch prompt from one or more beads tickets
allowed-tools: Agent, Bash, Read, Grep, Glob
---

# /launch-team <ticket-id> [ticket-id...]

Generate a ready-to-paste prompt for launching a multi-agent team from one or more tickets.

## Usage

```
/launch-team alty-cli-1wu                                  # single ticket
/launch-team alty-cli-cgm alty-cli-dfd                     # multiple tickets
/launch-team alty-cli-cgm alty-cli-dfd alty-cli-yl0
```

## Process

### Step 1 — Gather Ticket Context

For each ticket ID in the input:

```bash
bd show <ticket-id>
```

Extract from each ticket:
- Title, type, priority
- Description (goal, design, steps)
- Acceptance criteria
- Files to create or modify (from description and steps)
- Bounded context(s) touched

If a ticket lacks acceptance criteria or file paths, warn the user:
`"<ticket-id> has no acceptance criteria — consider running /groom <ticket-id> first."`

### Step 2 — Read Referenced Code

For every file path mentioned in the tickets, read it to understand:
- Current signatures and interfaces
- Existing patterns to follow
- What already exists (so agents don't recreate it)

Also read:
- `.claude/CLAUDE.md` — project conventions, workflow, enforced principles
- `docs/DDD.md` — domain model, bounded contexts, ubiquitous language
- `docs/ARCHITECTURE.md` — technical architecture
- `docs/PRD.md` — product requirements (for capability traceability)

### Step 3 — Build File Ownership Map

From the tickets, map each file to exactly one dev. If two tickets touch the same file, flag it:

```
CONFLICT: internal/discovery/domain/session.go claimed by <ticket-1> AND <ticket-2>
```

Ask the user to resolve conflicts before proceeding.

### Step 4 — Assign Team Roles

Map tickets to dev agents. Fixed roles (all live in `.claude/agents/`):
- **tech-lead** — coordinator, interface contracts, architecture compliance, final quality gates
- **qa-engineer** — test verification, edge cases, QA reports
- **white-hacker** — security review

One **developer** agent per ticket. Name them `dev-<ticket-id>` (e.g., `dev-alty-cli-1wu`).

For a single ticket, the team is: tech-lead + 1 dev + qa-engineer + white-hacker.

#### Hard cap: 5 active agents per wave

**Never spawn more than 5 agents in parallel.** Running 10–11 agents has frozen
the host with OOM. The cap is a system constraint, not a guideline — exceeding
it crashes the session and loses unsaved progress for every teammate.

The three fixed roles (TL + QA + WH) consume 3 slots. That leaves **2 developer
slots per active wave**. Spike variants where `developer` is replaced by
`researcher` follow the same cap (1 researcher per wave is typical; never more
than 2).

If the input list has more tickets than slots, split into **waves**:

| Tickets in input | Waves | Per-wave team size |
|------------------|-------|--------------------|
| 1                | 1     | 4 (TL + 1 dev + QA + WH) |
| 2                | 1     | 5 (TL + 2 devs + QA + WH) |
| 3                | 2     | wave 1 = 5, wave 2 = 4 |
| 4                | 2     | wave 1 = 5, wave 2 = 5 |
| 5+               | ⌈N/2⌉ | each wave ≤ 5; respect dep order |

Wave-split rules:
1. **Respect `bd dep` order.** A blocked ticket must land in a wave AFTER its blocker.
2. **Keep tickets that share files in the same wave** (the file-ownership map in Step 3 catches this).
3. **Generate ONE prompt per wave.** Each wave gets its own fenced code block; label them `WAVE 1`, `WAVE 2`, … with the ticket IDs included in the wave's `Tickets:` header.
4. **Hand-off between waves runs through `/handoff`.** Phase 7's `.notes/handoff-<slug>.md` is the entry point the next wave reads first.
5. **Tell the user about the split** in the Step 7 preamble (e.g. `"3 tickets → 2 waves; paste wave 1 first, then wave 2 after the handoff lands"`).

### Step 5 — Extract Design Decisions

From the tickets and code read in Step 2, extract settled design decisions:
- Go types and their shapes (value objects, entities, aggregates)
- Interface contracts (port signatures, return types, context-arg position)
- Patterns to follow (reference existing code with file:line)
- Domain events emitted or consumed (Watermill GoChannel)
- Bounded context boundaries (which package owns what)
- Constraints (what NOT to do — especially DDD layer rules)

### Step 6 — Generate the Prompt

Output the following prompt as a fenced code block the user can copy-paste.
Fill in ALL placeholders with actual data from Steps 1-5.

````markdown
Create a team for: <ticket titles, comma-separated>

## Reference Files (read before starting)

- `.claude/CLAUDE.md` — project conventions, enforced principles (DDD/TDD/SOLID/CQRS-lite), quality gates
- `docs/DDD.md` — domain model, bounded contexts, ubiquitous language (canonical terms)
- `docs/ARCHITECTURE.md` — technical architecture, port/adapter layout
- `docs/PRD.md` — product requirements (for capability traceability)
- `.golangci.yml` — lint config v2 (must pass with 0 issues)
- `arch-go.yml` — DDD layer enforcement (domain cannot import application/infrastructure)

## Tickets

<For each ticket, include:>
### <ticket-id> — <title>

<Full ticket description from bd show>

## Team Roster (max 5 agents — OOM cap)

This wave runs at most 5 agents in parallel. TL + QA + WH = 3 slots fixed;
that leaves 2 dev slots. Do NOT exceed.

| Name | Agent | Ticket | Key Files |
|------|-------|--------|-----------|
| tech-lead | tech-lead | (coordinator) | reviews all |
| qa-engineer | qa-engineer | (reviewer) | tests + reports |
| white-hacker | white-hacker | (security) | security review |
| dev-<ticket-id> | developer | <ticket-id> | <files from ownership map> |

## Execution Phases

Phase 1 — TL reads all tickets, publishes interface contracts
          (Go port interface signatures, struct shapes, sentinel error types,
          context-arg conventions, domain events emitted/consumed, bounded
          context ownership). Devs WAIT until TL broadcasts contracts.
Phase 2 — Devs implement in parallel using Red/Green/Refactor TDD.
          Self-verify quality gates before reporting.
          Report completion to qa-engineer + white-hacker (NOT tech-lead).
Phase 3 — QA + White Hacker review independently.
          Report findings to tech-lead (NOT devs).
Phase 4 — TL triages findings, assigns fixes to specific devs.
Phase 5 — Fix cycle: dev fixes → re-verify with QA + WH → TL confirms.
          Max 3 rounds per issue — TL escalates to user after that.
Phase 6 — TL runs final quality gates, closes tickets via `bd close`,
          then runs the After-Close Protocol from CLAUDE.md:
            a) `bin/bd-ripple <closed-id> "<what shipped>"` to flag dependents
            b) `bd query label=review_needed` and review each flagged ticket
            c) Compatibility check (read sources + dependent design, cite file:line)
            d) Present any suggested updates to the user — never auto-apply
Phase 7 — TL invokes the `/handoff` skill to write a session-summary doc to
          `<repo-root>/.notes/handoff-<wave-or-ticket-id>.md`. Include: tickets
          closed, files changed (paths + line counts), unresolved findings,
          newly-discovered follow-up issues filed via `bd create`, and
          suggested next-session entry points. Print the absolute path of the
          saved file so the user can open it.

## Communication Rules

- ALL communication via SendMessage — text output is invisible to teammates
- Devs report completion to QA + White Hacker (not TL)
- QA + WH report findings to TL (not devs)
- TL assigns fixes to devs (authority over triage)
- Peer-to-peer for clarifications (devs ↔ QA/WH directly)
- Acknowledge messages before starting work
- Escalate blockers to TL immediately — no silent waiting
- Max 3 fix rounds per issue — TL escalates to user after that

## Quality Gates (must pass at every checkpoint)

```bash
go build ./...                                   # 0 errors
go vet ./...                                     # 0 errors
golangci-lint run ./...                          # 0 issues (v2 strict config)
go test ./... -race                              # all pass, ≥80% coverage on new domain code
```

## Enforced Principles (non-negotiable)

These must hold in every PR. If any change weakens them, escalate before merging:

- **DDD layer rules** — `internal/{context}/domain/` has ZERO external deps;
  dependencies flow inward: infrastructure → application → domain.
  Enforced by `arch-go`.
- **Ubiquitous Language** — names match `docs/DDD.md` glossary; do NOT introduce synonyms.
- **Value Objects first** — default to immutable VOs; entities only when identity is needed.
- **One aggregate per transaction** — reference other aggregates by ID, not by pointer.
- **Port/Adapter** — handlers depend on port interfaces in the application layer,
  never on concrete adapters from infrastructure.
- **TDD required** — RED (failing test) → GREEN (minimal code) → REFACTOR.
- **CQRS-lite** — command handlers and query handlers are structurally separated;
  queries have no side effects; events route through Watermill GoChannel.
- **Wrapped errors** — `fmt.Errorf("doing X: %w", err)`; lowercase, no punctuation.
- **Testify idioms** — `assert.Len`, `assert.Empty`, `assert.ErrorIs`, `require.Error`
  for preconditions, `assert.InDelta` for floats (testifylint enforces).
- **No git commit/push** without explicit user permission (CLAUDE.md Git Rules).
- **No GitHub CLI** — repo is on private Git server; don't use `gh`.

## Design Decisions

<Settled decisions extracted from tickets and code — concrete, not placeholder>

## File Ownership (no conflicts)

| File | Owner | Ticket |
|------|-------|--------|
| <file> | dev-<ticket-id> | <ticket-id> |

DO NOT modify files not in your ownership table.

## Existing Code (DO NOT recreate)

<List files/packages that already exist and must not be overwritten>
````

### Step 7 — Present to User

Show the generated prompt(s) inside fenced code blocks. If Step 4 split the
input into multiple waves, output ONE block per wave, in order.

Prefix each wave with:

```
TEAM LAUNCH PROMPT — WAVE <n> of <total>
Tickets: <ids in this wave>
Team size: <N> (TL + <N> devs + QA + WH)   ← must be ≤ 5
Files touched: <N>
Conflicts: <none | list>
Next wave entry point: .notes/handoff-<slug>.md  (omit on final wave)

Copy the prompt below and paste it into a new session to launch this wave.
```

For a single-wave run drop the `WAVE <n> of <total>` suffix and the
`Next wave entry point` line.

## Rules

1. **Never launch the team yourself.** Only generate the prompt — the user decides when and where to paste it.
2. **Read actual code.** Every file reference in the prompt must be verified by reading the file.
3. **Flag undergroomed tickets.** Warn if a ticket lacks acceptance criteria or file paths.
4. **Resolve file conflicts.** If two tickets own the same file, stop and ask the user.
5. **No placeholders in output.** Every `<placeholder>` must be filled with real data. If you can't fill it, say what's missing.
6. **Always cite the Enforced Principles in the prompt.** Teams that don't see them will accidentally break DDD/TDD/SOLID guarantees.
7. **End every wave with `/handoff`.** Phase 7 in the generated prompt mandates a `.notes/handoff-<slug>.md` write-up. The handoff slug should be the ticket ID for a single-ticket wave (e.g., `handoff-alty-cli-1wu.md`), or a short wave name (`handoff-wave-1.md`, `handoff-discovery-redesign.md`) for a multi-ticket wave.
8. **Hard cap: 5 active agents per wave.** Running 10–11 in parallel has frozen the host (OOM). TL + QA + WH = 3 fixed slots, leaving 2 dev slots. More tickets → more waves. This is a host constraint, not a preference — never override it.
