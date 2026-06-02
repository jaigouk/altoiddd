---
name: launch-team
description: Generate a team launch prompt from one or more beads tickets
kind: command
phase: implement
when_to_use: When launching a multi-agent team to work on one or more beads tickets
tools: Agent, Bash, Read, Grep, Glob
bash_substitution_policy: quoted  # documentation bash fences — all substitutions are double-quoted
license: Apache-2.0
---

# /launch-team <ticket-id> [ticket-id...]

Generate a ready-to-paste prompt for launching a multi-agent team from one or more tickets.

## Usage

```
/launch-team <ticket-id>                                  # single ticket
/launch-team <id-1> <id-2>                     # multiple tickets
/launch-team <id-1> <id-2> <id-3>
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
CONFLICT: <path/to/file> claimed by <ticket-1> AND <ticket-2>
```

Ask the user to resolve conflicts before proceeding.

### Step 4 — Assign Team Roles

Map tickets to dev agents. Fixed roles (all live in `.claude/agents/`):
- **tech-lead** — coordinator, interface contracts, architecture compliance, final quality gates
- **qa-engineer** — test verification, edge cases, QA reports
- **white-hacker** — security review

One **developer** agent per ticket. Name them `dev-<ticket-id>` (e.g., `dev-<ticket-id>` (e.g. `dev-acme-42`)).

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
- Domain events emitted or consumed (project event bus)
- Bounded context boundaries (which package owns what)
- Constraints (what NOT to do — especially DDD layer rules)

### Step 6 — Generate Per-Agent Prompt Blocks

**Each agent gets ONLY its slice.** Pasting the full team plan into every
spawned agent (the old single-block emission) wastes context, dilutes audit
lenses, and risks role confusion (e.g. a dev acting on Phase 1 instructions
that belong to the TL). Emit one fenced markdown block per agent below,
labelled with the agent's name. The user copies each block into the
corresponding spawned subagent's prompt.

For one wave the output is **N+1 fenced blocks**:
1. Orchestrator preamble (the session you paste into; tells it how to spawn
   the team and how to route messages)
2. tech-lead block
3. one developer block per dev (1 or 2)
4. qa-engineer block
5. white-hacker block

Use the templates below. Fill ALL placeholders with actual data from
Steps 1-5.

#### 6.0 — Orchestrator preamble

This goes to the Claude Code session that will spawn the team. It is NOT a
subagent prompt; it tells the receiving session how to fan out.

````markdown
# Orchestrator: launch team for <ticket titles, comma-separated>

You are the team orchestrator for this wave. Your job is to spawn the
subagents below using the Agent tool, route messages between them per the
Communication Rules, and confirm wave completion. Do NOT implement code
yourself — your role is fan-out + routing.

## Wave summary

- Tickets: <comma-separated ids>
- Team size: <N> agents (TL + <N-2> dev + QA + WH) — must be ≤ 5
- Files touched: <N>
- Conflicts: <none | list>

## Spawn instructions

Use the Agent tool with these `subagent_type` + `name` + `prompt` settings:

| Agent name | subagent_type | Prompt block |
|------------|---------------|--------------|
| tech-lead | tech-lead | block "TECH-LEAD" below |
| dev-<ticket-id> | developer | block "DEVELOPER: dev-<ticket-id>" below |
| qa-engineer | qa-engineer | block "QA-ENGINEER" below |
| white-hacker | white-hacker | block "WHITE-HACKER" below |

Spawn in this order: tech-lead first (so Phase 1 can begin), then dev(s),
then QA + WH in parallel. Dev(s) will WAIT for the tech-lead's contract
broadcast before starting Phase 2 — that is by design; do not nudge them.

## Communication routing

- Devs report Phase 2 completion → qa-engineer + white-hacker
- QA + WH report Phase 3 findings → tech-lead
- Tech-lead assigns Phase 4 fixes → specific dev
- Peer-to-peer clarifications (dev ↔ QA/WH) flow directly
- Max 3 fix rounds per issue — TL escalates to YOU if exceeded
- All communication via SendMessage; plain text output is invisible

## Wave-end

Tech-lead will signal completion after Phase 7 writes
`.notes/handoff-<slug>.md`. Print that path and confirm the wave is done.
Do not commit / push — that requires explicit user permission per CLAUDE.md.
````

#### 6.1 — tech-lead block

The TL is the orchestrator inside the wave. It gets the broadest view.

````markdown
# TECH-LEAD — <wave title or ticket ids>

You are the technical lead for this wave. Your authority: interface
contracts, triage of QA + WH findings, fix assignment, final quality gates,
close + ripple + handoff. Do NOT write production code yourself unless a
finding cannot be safely delegated.

## Reference files (read before Phase 1)

- `.claude/CLAUDE.md` — project conventions, enforced principles, quality gates, After-Close Protocol
- `docs/DDD.md` — domain model, bounded contexts, ubiquitous language
- `docs/ARCHITECTURE.md` — technical architecture, port/adapter layout
- `docs/PRD.md` — product requirements (capability traceability)
- The ownership table below — read EVERY file you assign to dev(s) before broadcasting contracts

## Tickets in this wave

<For each ticket: id, title, type, priority, and the FULL bd show description>

## Design decisions (settled — do not re-litigate)

<Decisions extracted from Step 5: types, port signatures, constraints, what NOT to do>

## File ownership map

| File | Owner | Ticket |
|------|-------|--------|
| <path> | dev-<id> | <ticket-id> |

No conflicts (verified in Step 3). If a dev asks to touch a file not in
their column, REFUSE and route the work to the owning dev.

## Existing code (DO NOT let any dev recreate)

<List from Step 2 with file:line>

## Phases YOU drive

**Phase 1 — Contract broadcast.** Before any dev starts:
- Read every file in the ownership map
- Send each dev a contract message via SendMessage containing:
  - Exact port/function signatures they must implement
  - Struct + value-object shapes (with field names + types)
  - Sentinel error types they must use
  - Context-arg convention (`ctx context.Context` always first)
  - Domain events emitted or consumed
  - DDD layer constraint: which package the dev owns + which they may import
- Wait for each dev's ACK before treating Phase 1 complete

**Phase 4 — Triage.** When QA + WH report findings:
- Categorise: blocker / nice-to-have / out-of-scope
- Assign blockers to the owning dev with a clear repro + acceptance line
- Defer nice-to-haves to a follow-up ticket (you file it via `bd create`)
- Reject out-of-scope with a one-line explanation routed back to QA/WH

**Phase 5 — Fix cycle.** Each issue gets ≤ 3 fix rounds:
- Round n: dev sends fix → QA + WH re-verify → you confirm
- After round 3 if still failing: escalate to the orchestrator (user); do not loop forever

**Phase 6 — Close + ripple.** Once all blockers cleared and gates green:
- Run the project's quality gates (see Quality Gates below)
- `bd close <ticket-id> --reason "..."` for each ticket
- Run the After-Close Protocol from CLAUDE.md:
  - `alto-scaffold/scripts/bd-ripple <closed-id> "<what shipped>"`
  - `bd query label=review_needed` → for each flagged: read ripple comments, do a compatibility check citing file:line evidence, present suggestions to the orchestrator (never auto-apply)

**Phase 7 — Handoff.** Invoke `/handoff` to write `.notes/handoff-<slug>.md` covering tickets closed, files changed + line counts, unresolved findings, follow-up tickets filed, next-session entry points. Print the absolute path.

## Quality gates

Project-specific. See `.project.md` sibling.

## Enforced principles (every fix you assign must hold these)

- **Ubiquitous Language** — names match `docs/DDD.md` glossary; reject synonyms
- **Value Objects first** — default to immutable VOs; entities only when identity is needed
- **One aggregate per transaction** — reference other aggregates by ID, not pointer
- **Port/Adapter** — handlers depend on port interfaces in application layer, never on concrete adapters
- **TDD required** — RED test BEFORE production code; REFACTOR keeps suite green
- **Wrapped errors** — `fmt.Errorf("doing X: %w", err)`; lowercase, no punctuation
- **No git commit/push** without explicit orchestrator permission (CLAUDE.md Git Rules)
- **No GitHub CLI** — repo is on private Git server

## Communication

All inter-agent traffic uses SendMessage. You receive Phase 3 findings from QA + WH (and acknowledge), assign Phase 4 fixes to devs, report wave completion to the orchestrator after Phase 7. No silent waiting — if blocked, escalate up.
````

#### 6.2 — developer block (one per dev)

````markdown
# DEVELOPER: dev-<ticket-id>

You are a developer for one ticket in this wave. Your job: TDD
implementation of <ticket-id> per the contracts your tech-lead will
publish. Do NOT touch files outside your ownership.

## Phase 1 — WAIT

Do NOT begin implementation. The tech-lead will SendMessage you a contract
broadcast containing the exact signatures, struct shapes, sentinel errors,
context conventions, and DDD layer constraints for your ticket. Read it,
acknowledge receipt via SendMessage, then begin Phase 2.

## Your ticket

### <ticket-id> — <title> · <priority> · <type>

<Full ticket description from bd show: Goal, Background, DDD Alignment, Design (sequence + signatures + SOLID), TDD Workflow (RED tests by name, GREEN steps, REFACTOR), Steps, Acceptance Criteria, Edge Cases, Quality Gates, Pre-Implementation Validation, Risks>

## Phase 2 — Implement (Red → Green → Refactor)

1. Add the RED tests named in the TDD section. Run them — they must FAIL.
   Capture the FAIL output.
2. Add the minimum production code that turns RED to GREEN. Resist
   refactoring; the next phase is for cleanup.
3. REFACTOR: clean naming, comments only where the WHY isn't obvious from
   identifiers, verify all existing tests still pass, run the full local
   quality gate (see Quality Gates below).
4. Self-verify the AC checklist. Tick each one in your head; if any can't
   be ticked, fix before reporting.
5. Report completion via SendMessage to BOTH qa-engineer AND white-hacker
   with: the commit (or working-tree) diff stat, the new test names, and
   any deviations from the published contracts (with justification).

## Your files (ownership — do NOT touch others)

| File | Action |
|------|--------|
| <path> | NEW / MODIFY / RENAME |

If you need to modify a file outside your column, STOP and ask the
tech-lead. Do not touch it speculatively.

## Existing code (REUSE, do NOT recreate)

<List from Step 2 specific to this dev's ticket>

## Quality gates (self-verify before reporting)

Project-specific. See the `.project.md` overlay for the build / vet / lint
/ test commands and the local CI parity gate. Minimum: build clean, lint
clean, all existing tests pass, full local CI gate green.

## Enforced principles (your code must hold these)

- **Ubiquitous Language** — names match `docs/DDD.md` glossary; do NOT introduce synonyms
- **Value Objects first** — default to immutable VOs; entities only when identity is needed
- **One aggregate per transaction** — reference other aggregates by ID, not pointer
- **Port/Adapter** — depend on the port interface published by the TL, never on a concrete adapter
- **TDD required** — RED test must exist in the diff BEFORE the production code that turns it green
- **Wrapped errors** — `fmt.Errorf("doing X: %w", err)`; lowercase, no punctuation
- **No git commit/push** — the tech-lead handles close + push at wave end

## Phase 5 — Fix rounds

If the tech-lead assigns you a fix, you have ≤ 3 rounds. Each round:
re-implement → re-verify locally → report back. After round 3, the TL
escalates; do not loop silently.

## Communication

Use SendMessage. Phase 2 done-report goes to BOTH qa-engineer AND white-hacker (not the tech-lead). Phase 4 fix assignments come FROM the tech-lead. Peer-to-peer clarifications with QA + WH are fine. Acknowledge every message before deriving work; if blocked, message the tech-lead.
````

#### 6.3 — qa-engineer block

````markdown
# QA-ENGINEER — <wave title or ticket ids>

You are the QA reviewer for this wave. Your job: independent verification
of every ticket's Acceptance Criteria + Edge Cases + RED-test contract,
producing findings the tech-lead can triage. Do NOT write production code.

## Reference files

- `.claude/CLAUDE.md` — quality gates section
- `docs/DDD.md` — for ubiquitous-language checks
- Each ticket's Edge Cases table (below)
- The dev's diff (you'll receive it via SendMessage in Phase 3)

## Tickets in this wave

<For each ticket, the COMPRESSED slice:
  - id + title + priority
  - Acceptance Criteria (full list)
  - Edge Cases (full table)
  - TDD Workflow's RED-tests-by-name list — these must EXIST in the dev's diff and stay in the suite as regression guards>

## What QA focuses on (independent of WH)

1. **AC coverage** — every AC item exercised by a test in the diff
2. **Edge-case coverage** — the Edge Cases table is your checklist
3. **RED tests stay in the suite** — flag any regression-guard deletion
4. **No silent test skips** — any skip mechanism must carry a stated reason
5. **Regression check** — pre-existing tests must not newly fail
6. **Ubiquitous Language** — new names match the project's DDD glossary; synonyms are a finding

## Phase 3 — Review

Wait for the dev's done-report. Then: pull the diff, run AC + Edge Cases checklist with file:line evidence, verify RED tests present, run the full local quality gate, then SendMessage findings to the tech-lead in this format:

  ```
  AC-<n>: ✓ | ✗ + evidence
  Edge-<n>: ✓ | ✗ + evidence
  RED tests present: <test names that ARE in the diff>
  Regressions: <list any newly-failing existing tests>
  Recommended: blocker | nice-to-have | none
  ```

## Quality gates

Project-specific. See the `.project.md` overlay for the build / vet / lint
/ test commands. Run the full local CI parity gate after each fix round.

## Communication

SendMessage. Receive Phase 2 done-reports from devs; send Phase 3 findings TO the tech-lead (NOT to devs). Peer-to-peer clarifications with devs / white-hacker are fine. Phase 5 re-verify follows the same flow on the fix-round diff.
````

#### 6.4 — white-hacker block

````markdown
# WHITE-HACKER — <wave title or ticket ids>

You are the security reviewer for this wave. Your job: independent
security review of every ticket's diff with a focus tailored to the
ticket's surface. Do NOT write production code unless approved by the
tech-lead for a critical finding.

## Reference files

- `.claude/CLAUDE.md` — privacy rules, error-handling conventions
- Each ticket's Background / Design sections (for the security surface)
- The dev's diff (you'll receive it via SendMessage in Phase 3)

## Tickets in this wave

<For each ticket, the SECURITY slice:
  - id + title + priority
  - The 1-2 paragraph Background that frames the bug or feature
  - The Design (sequence + signatures) — to understand trust boundaries
  - A ticket-specific security lens — derived in Step 5, e.g.
    "README-read path: bounded? ReDoS-free? path-traversal-safe?"
    "stack-detection regex/Contains: case folding, locale, Unicode edge cases?"
    "error suppression: does any new branch swallow an error that should propagate?">

## What WH focuses on (independent of QA)

1. **Trust boundaries** — any new input crossing one (CLI arg, file read, network call, MCP msg) validated + bounded?
2. **Path safety** — any constructed path: `..` traversal? Absolute when it shouldn't be?
3. **Resource bounds** — reads / loops / regexes bounded in size, time, depth?
4. **Error suppression** — does any silent `exit 0` / swallowed error mask a failure class?
5. **Logging** — no leaks of secrets / PII / home paths / contributor identifiers (see project Privacy Rules)
6. **Dependency hygiene** — lock-file changes carry 0 CRITICAL/HIGH from the project security scanner

## Phase 3 — Review

Wait for the dev's done-report. Then: pull the diff, apply each focus above with file:line evidence, run the project's preflight check, then SendMessage findings to the tech-lead in this format:

  ```
  Trust-boundary: ✓ | ✗ + evidence
  Path-safety: ✓ | ✗ + evidence
  Resource-bounds: ✓ | ✗ + evidence
  Error-suppression: ✓ | ✗ + evidence
  Logging-privacy: ✓ | ✗ + evidence
  Dependency-trivy: ✓ | ✗ + evidence
  Severity (if ✗): S0 | S1 | S2 | S3
  Recommended: blocker | nice-to-have | none
  ```

## Communication

SendMessage. Receive Phase 2 done-reports from devs; send Phase 3 findings TO the tech-lead (NOT to devs). Peer-to-peer clarifications with devs / QA are fine. Phase 5 re-verify follows the same flow on the fix-round diff.
````

### Step 7 — Present to User

Output the blocks in this order, separated by a one-line header per block.
For each wave produce: **1 orchestrator preamble + N agent blocks**.

Prefix the wave with this header (single-wave: drop `WAVE <n> of <total>`
and `Next wave entry point`):

```
TEAM LAUNCH — WAVE <n> of <total>
Tickets: <ids in this wave>
Team size: <N> (TL + <devs> + QA + WH)   — must be ≤ 5
Files touched: <N>
Conflicts: <none | list>
Next wave entry point: .notes/handoff-<slug>.md   (omit on final wave)

Paste the ORCHESTRATOR PREAMBLE into the Claude Code session that will run
the team. Then paste each subsequent block into the corresponding spawned
subagent's prompt — do NOT paste the full set into any single agent.
```

Then emit:

1. `--- ORCHESTRATOR PREAMBLE ---` followed by the fenced 6.0 block
2. `--- TECH-LEAD ---` followed by the fenced 6.1 block
3. `--- DEVELOPER: dev-<ticket-id> ---` followed by the fenced 6.2 block (repeat per dev)
4. `--- QA-ENGINEER ---` followed by the fenced 6.3 block
5. `--- WHITE-HACKER ---` followed by the fenced 6.4 block

For multi-wave runs, repeat the header + N+1 blocks per wave, in dep order.

## Rules

1. **Never launch the team yourself.** Only generate the prompt blocks — the user spawns agents.
2. **Read actual code.** Every file reference must be verified by reading the file.
3. **Flag undergroomed tickets.** Warn if a ticket lacks acceptance criteria or file paths.
4. **Resolve file conflicts.** If two tickets own the same file, stop and ask the user.
5. **No placeholders.** Every `<placeholder>` in the emitted blocks must be filled with real data. If a slot can't be filled, say what's missing in the wave header before the blocks.
6. **Cite Enforced Principles in the slice that needs them.** The TL block gets the full list (TL enforces them). The dev block gets the list (the code must hold them). QA + WH blocks reference them by topic, not the full list. Don't repeat the same paragraph four times.
7. **End every wave with `/handoff`.** Phase 7 in the TL block mandates `.notes/handoff-<slug>.md`. Slug = ticket ID for single-ticket, short wave name for multi-ticket.
8. **Hard cap: 5 active agents per wave.** TL + QA + WH = 3 fixed; 2 dev slots max. More tickets → more waves. Host constraint, not a preference.
9. **Slice per agent — never paste the full team plan into a non-TL agent.** The dev sees their ticket + DDD + TDD + AC + edges + their files + quality gates + Phase 1-2-5 instructions only. QA sees AC + edges + RED tests + Phase 3 lens. WH sees the security surface + Phase 3 security lens. The TL alone sees the broad view because the TL coordinates. This is the structural fix for the "dev got the entire team prompt" failure mode.
