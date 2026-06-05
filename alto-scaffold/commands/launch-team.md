---
name: launch-team
description: Generate a team launch prompt from one or more beads tickets, or from a wave of an epic. Defaults to sequential-orchestrator mode (works in stock Claude Code); team-mode (SendMessage between spawned agents) is opt-in via --mode=team and requires a harness that exposes SendMessage to subagents.
kind: command
phase: implement
when_to_use: When launching a multi-agent team to work on one or more beads tickets, or to ship the next ready wave of an epic
tools: Agent, Bash, Read, Grep, Glob, SendMessage, ToolSearch
bash_substitution_policy: quoted
license: Apache-2.0
---

# /launch-team <ticket-id> [ticket-id...]
# /launch-team --epic=<epic-id> [--wave=<N>]
# /launch-team ... [--mode=sequential|team]

Generate a ready-to-paste prompt for launching a multi-agent team.

## Two execution modes

| Mode | When to use | Mechanism |
|------|-------------|-----------|
| **`sequential` (default)** | Stock Claude Code, most user-installed setups | Orchestrator session plays the TL role. Spawns `developer` / `qa-engineer` / `white-hacker` via the `Agent` tool one phase at a time. Each subagent runs to completion and returns its result. Orchestrator parses returns and routes follow-ups. **No SendMessage between subagents required.** |
| **`team` (opt-in)** | Environments where `SendMessage` is exposed to spawned subagents (verified via Step -1 harness probe) | Original protocol. TL is spawned as a separate agent. Spawned agents exchange messages directly via SendMessage. Requires the harness probe (Step -1) to pass. |

**Why sequential is the default.** A canary test in stock Claude Code
confirmed that spawned subagents do NOT get `SendMessage` in their
deferred-tool set, even when the orchestrator session can load it. The
orchestrator's `ToolSearch select:SendMessage` succeeds; the same call
inside a spawned subagent returns no function block. This isn't a bug
in either side — it's the harness's deferred-tool propagation boundary.
Sequential mode bypasses it by keeping all inter-agent routing in the
orchestrator session, which keeps SendMessage where it works.

Two input modes (independent of execution mode):

1. **Ticket-list mode** — `<ticket-id> [ticket-id...]` — pick any set of
   tickets and the command splits them into team rounds respecting
   `bd dep` order and the 5-agent cap.
2. **Epic-wave mode** — `--epic=<id> [--wave=<N>]` — read the epic's
   "Execution Waves" section and launch the next ready wave.

Use epic-wave mode whenever the work has been pre-planned via
[`/design-ticket --epic=<id>`](design-ticket.md) and lives in an epic
following [`beads-epic-template.md`](../templates/beads-epic-template.md).
Use ticket-list mode for ad-hoc launches of standalone tickets.

## Step -1 — Harness probe (only when `--mode=team`)

Skip when `--mode=sequential` (the default). Sequential mode does not
require SendMessage between spawned agents.

When `--mode=team` is requested, run a cheap canary before any other
work:

1. **Orchestrator self-probe** — Call
   `ToolSearch({query: "select:SendMessage"})` in the orchestrator
   session. If SendMessage doesn't load, abort: print the diagnostic
   below and exit with "team mode unavailable; re-run without
   `--mode=team` for sequential."

2. **Spawned-agent probe** — Spawn ONE canary subagent with the prompt:
   ```
   Run ToolSearch({query: "select:SendMessage"}) and report whether the
   SendMessage function block appears in the result. Then exit. Do NOT
   send any actual SendMessage.
   ```
   The orchestrator inspects the canary's return text for the substring
   `SendMessage function block` AND confirms the canary did NOT report
   "no match." If either check fails, abort.

3. **Two-canary cross-check** — Spawn a second canary that ALSO tries
   to call `SendMessage({to: "orchestrator", message: "canary-ok"})` and
   reports whether the call returned an error. If the orchestrator never
   receives the `canary-ok` message AND the canary reports a routing
   error, the inter-agent transport is broken.

If any of the three checks fails, print:

```
Harness probe failed. Diagnostic:
  - Orchestrator self-load of SendMessage:                <PASS | FAIL>
  - Canary 1 (deferred-tool surface in spawned agent):    <PASS | FAIL>
  - Canary 2 (actual SendMessage delivery):               <PASS | FAIL>

Conclusion: this Claude Code session's deferred-tool set for spawned
subagents does not include SendMessage. The team's inter-agent transport
doesn't exist, so the orchestrator-routed Phase 2/3/4 message loops
can't run.

Action: re-run without --mode=team to use sequential-orchestrator mode,
which routes all coordination through the orchestrator session (no
SendMessage between subagents required).
```

The canaries cost two cheap spawns and save ~140k tokens on a degraded
fanout. Do NOT skip the probe for `--mode=team`.

For `--mode=sequential` (default), proceed directly to Step 0 / Step 1
without probing — the orchestrator session itself is the only place
SendMessage needs to work, and that's already true in any harness that
can run this command.

## Prerequisites — harness facts every emitted prompt must respect

Three Claude Code harness behaviors break the team model unless the emitted
prompts explicitly account for them. The templates below already do — do
NOT remove these guardrails when generating the prompt.

1. **`SendMessage` is a deferred tool.** A spawned agent's tool list does
   NOT include `SendMessage` by default. The agent must load it on its
   first turn by calling `ToolSearch({query: "select:SendMessage"})`.
   Without this, the agent will conclude "no inter-agent transport in
   this harness" and bail. Every emitted block opens with a tool-loading
   preamble for this reason.

2. **Spawned agents are one-shot.** A spawned agent runs its reasoning
   loop, then **exits** when it has no more actions to take. A `Phase 1
   WAIT` instruction terminates the agent cleanly — there is no
   suspended process waiting for a later message. To resume a completed
   agent, the orchestrator (or TL) must call `SendMessage` against the
   agent's `agentId` (format `a...`), NOT its display name. Names work
   only while an agent is alive; `agentId`s work for resume.

   The orchestrator MUST record each agentId returned by the `Agent`
   tool call and keep it for the wave's lifetime. Don't rely on
   addressing by name for resume.

3. **Custom `subagent_type` values are not auto-registered.** The
   personas in `alto-scaffold/agents/*.md` (tech-lead, developer,
   qa-engineer, white-hacker, researcher, project-manager) are
   templates; they are NOT installed into `.claude/agents/` for the
   target project unless a separate setup step has run. Until that
   install step exists, the emitted blocks use `subagent_type: claude`
   (the catch-all) and embed the persona inline in the prompt. The
   per-block `subagent_type` cell in Step 6.0's spawn table reflects
   this default; flag the user only if they've explicitly installed
   the custom personas (check `ls .claude/agents/`).

If any of these assumptions break for a given target project, the
"Failure modes & fallback" section at the bottom of this skill
describes the escape hatch (single-agent execution mode by the TL).

## Team-Mode Communication Protocol

> **Canonical spec.** Every persona under `alto-scaffold/agents/`, every
> emitted wave prompt, and every orchestrator preamble references THIS
> section. Do not re-state the protocol elsewhere in shorter or
> diverging form. If the protocol changes, change it here and let the
> references catch the update.

### P1 — Loading the transport (first turn, before any other action)

`SendMessage` is a deferred tool. On its first turn, every spawned
subagent MUST call:

```
ToolSearch({query: "select:SendMessage"})
```

If `SendMessage` does not appear in the result, the agent MUST reply
with plain-text `"SendMessage unavailable; cannot participate in team
mode"` and exit. The orchestrator surfaces the harness mismatch and
falls back per Failure Mode 1.

The orchestrator additionally loads `TaskList` and `TaskStop` for
routing; persona agents need only `SendMessage`.

### P2 — Addressing peers

- **While the peer is alive** — address by display name: `SendMessage({to: "qa-engineer", message: "..."})`.
- **After the peer has exited** (status: completed) — address by `agentId` (format `a...-...`), NOT display name. The orchestrator maintains the name → agentId map.
- **Agents address peers by name** in their own SendMessage calls; the orchestrator translates name → agentId at resume time.

### P3 — One-shot exit semantics

Spawned agents are one-shot. When you reach a WAIT state (no contract
yet, no fix-assignment yet, no fresh finding yet), you EXIT cleanly.
There is no suspended process; the orchestrator resumes you with
SendMessage when the next message arrives. Do NOT loop, sleep, or poll.

### P4 — Routing chart (who → who, when, via whom)

```
                                  ┌──────────────────┐
                                  │   ORCHESTRATOR   │
                                  │ (session you run)│
                                  └────────┬─────────┘
                                           │ translates name→agentId
                                           │ for every cross-agent hop
                  ┌────────────────────────┼────────────────────────┐
                  ▼                        ▼                        ▼
           ┌──────────┐             ┌──────────┐            ┌──────────────┐
           │ TECH-LEAD│ ──────────▶ │   DEV(s) │            │ QA + WH      │
           │          │  P1 contract│          │            │              │
           │          │ ◀────────── │ Phase 1  │            │              │
           │          │   ACK       │ ACK then │            │              │
           │          │             │ exit     │            │              │
           │          │             │          │            │              │
           │          │             │ Phase 2  │ ─P2 done─▶ │ Phase 3      │
           │          │             │ implement│   report   │ review       │
           │          │             │          │            │              │
           │ P4 fix   │ ◀──────────────────────────P3 ──────│ findings     │
           │ assign   │ ──────────▶ │ Phase 5  │            │              │
           │          │   fix req   │ fix      │ ─P5 done─▶ │ re-verify    │
           │ P6 close │             │          │            │              │
           │ + ripple │             │          │            │              │
           │ P7 handoff             │          │            │              │
           └──────────┘             └──────────┘            └──────────────┘
```

Rules:
- Dev → QA + WH ONLY for Phase 2 done-report and Phase 5 re-verify request.
- QA + WH → TL ONLY (never directly to dev with a finding).
- TL → dev for Phase 1 contract + Phase 4 fix assignment.
- Peer-to-peer clarifications (dev ↔ QA/WH) flow directly while both are alive.
- Plain-text output goes ONLY to the orchestrator. Peers receive nothing unless addressed via SendMessage.

### P5 — Message format reference

Use these formats verbatim where possible — QA/WH findings and TL fix
assignments are parsed by the orchestrator (and by you, on resume) to
make routing deterministic.

**ACK (dev → TL, Phase 1):**

```
dev-<ticket-id> ready
```

One line. No prose. Triggers TL to send the contract.

**Contract broadcast (TL → dev, Phase 1):**

```
Contract for dev-<ticket-id>:
- Port/function signatures: <list with language-native signatures per project conventions>
- Struct / value-object shapes: <list with field names + types>
- Sentinel errors / domain failure modes: <list>
- Cancellation / context propagation: <language-specific convention — e.g. Go ctx context.Context first arg, Python asyncio cancel scope, TypeScript AbortSignal>
- Domain events: <emitted | consumed | none>
- DDD layer constraint: owns <path>; may import <path>; MUST NOT import <path>
- Ownership: <files this dev owns>
- ACK with "dev-<ticket-id> contract-acked" then begin Phase 2.
```

**Done-report (dev → QA + WH, Phase 2):**

```
dev-<ticket-id> done-report
- Diff stat: <files changed, +/- lines>
- New tests: <test names, RED→GREEN evidence>
- AC self-check: <each AC ✓/✗ with file:line>
- Deviations from contract: <none | list with justification>
- Ready for review.
```

**Findings (QA → TL, Phase 3):**

```
QA-findings for <ticket-id>
- AC-<n>: ✓ | ✗ + evidence (file:line)
- Edge-<n>: ✓ | ✗ + evidence
- RED tests present: <names in diff>
- Regressions: <newly-failing existing tests | none>
- Recommended: blocker | nice-to-have | none
```

**Findings (WH → TL, Phase 3):**

```
WH-findings for <ticket-id>
- Trust-boundary: ✓ | ✗ + evidence
- Path-safety: ✓ | ✗ + evidence
- Resource-bounds: ✓ | ✗ + evidence
- Error-suppression: ✓ | ✗ + evidence
- Logging-privacy: ✓ | ✗ + evidence
- Dependency-hygiene: ✓ | ✗ + evidence
- Severity (if ✗): S0 | S1 | S2 | S3
- Recommended: blocker | nice-to-have | none
```

**Fix assignment (TL → dev, Phase 4):**

```
Fix-request for dev-<ticket-id>, round <n>/3
- Finding: <one-line summary>
- Repro: <command or test that demonstrates>
- Acceptance: <how we know it's fixed>
- Re-report to qa-engineer + white-hacker when done.
```

**Escalation (TL → orchestrator):**

```
Escalation: <reason>
- Ticket: <id>
- What failed: <one line>
- Rounds attempted: <n>
- Recommended next step: <abandon | new ticket | user decision>
```

### P6 — What NOT to do

- Do NOT reply with prose like "I'll start working on this" — either ACK in the canonical format and exit, or proceed to the next phase.
- Do NOT broadcast contracts as a single dump to all agents — TL sends individually to each dev.
- Do NOT include diff bodies in SendMessage payloads. Send filename + line ranges; peers Read the actual diff.
- Do NOT auto-retry a SendMessage that returned an error — the orchestrator handles failed deliveries.
- Do NOT cite `.notes/handoff-*.md` paths from messages that will be quoted in committable artefacts (commit messages, ticket bodies, `bd close --reason`). `.notes/` is the gitignored scratchpad.

### P7 — Failure handling

- `SendMessage` returns an error → exit; the orchestrator will resume you with diagnostic context.
- A peer address fails (display name no longer alive AND no agentId given to you) → SendMessage the orchestrator with `routing-help: <peer name> unreachable, last message: <one line>` and exit.
- After 3 fix rounds on the same finding with no convergence, TL escalates via the P5 Escalation format. Devs do NOT loop silently.

## Usage

```bash
# Ticket-list mode
/launch-team <ticket-id>                                  # single ticket
/launch-team <id-1> <id-2>                                # multiple tickets
/launch-team <id-1> <id-2> <id-3>

# Epic-wave mode — recommended when work was planned via /design-ticket --epic
/launch-team --epic=<epic-id>                             # next ready wave
/launch-team --epic=<epic-id> --wave=2                    # specific wave
```

## Terminology — "epic wave" vs "team round"

Two distinct things called something-like-wave exist; this command keeps
them separate:

- **Epic wave** — defined in the epic body's "Execution Waves" section.
  Group of child tickets that can be worked on in parallel because they
  share the same deps + touch disjoint files. See
  [`beads-epic-template.md`](../templates/beads-epic-template.md).
- **Team round** — one spawned set of ≤ 5 agents (TL + ≤ 2 devs + QA + WH)
  produced by this command. The 5-agent cap is a host constraint, not a
  preference.

One epic wave becomes one or more team rounds depending on ticket count:

| Tickets in epic wave | Team rounds | Per-round team size |
|----------------------|-------------|----------------------|
| 1                    | 1           | 4 (TL + 1 dev + QA + WH) |
| 2                    | 1           | 5 (TL + 2 devs + QA + WH) |
| 3                    | 2           | round-a = 5, round-b = 4 |
| 4                    | 2           | round-a = 5, round-b = 5 |
| 5+                   | ⌈N/2⌉       | each round ≤ 5; tickets within a wave have no inter-deps so any order works |

Team rounds within ONE epic wave can spawn back-to-back (or in parallel if
the host can sustain it), but the **per-round user gate** still applies —
the orchestrator stops between rounds. Between epic waves the gate is
non-negotiable: Wave N+1 cannot begin until every ticket in Wave N is
closed.

## Epic-wave mode — what it does

When invoked with `--epic=<id>`, the command:

1. Reads the epic body via `bd show <epic-id>`. Requires the epic to
   follow `beads-epic-template.md` — specifically, it expects:
   - A **Child Tasks** table with a `Wave` column
   - An **Execution Waves** section (for the ASCII timeline + reasoning,
     used to enrich the round headers)
2. Resolves `--wave=<N>` to a target wave. If `--wave` is omitted, picks
   the **lowest open wave where every ticket in every prior wave is
   `closed`**.
3. Runs a **wave-readiness check** — fails if any ticket in waves < N is
   still open. Prints the blocker list and exits without emitting
   prompts.
4. Lists tickets assigned to wave N (from the Child Tasks table).
5. Applies the same per-ticket grooming check as ticket-list mode (each
   needs acceptance criteria + file paths; warn if missing).
6. Splits the wave's tickets into team rounds using the round-cap table
   above.
7. Emits one `.notes/next-round-<epic-id-suffix>-w<N><suffix>.md` file
   per round (e.g. `next-round-tzv-w2-a.md`, `next-round-tzv-w2-b.md`).
   Each file holds the orchestrator preamble + per-agent blocks for that
   round.

### Wave-readiness check details

A wave is **ready** iff:

- All tickets in all prior waves of the same epic are `closed` (or
  superseded).
- All `bd dep` edges that point into this wave from outside the epic are
  closed.
- Sibling tickets flagged as "Wave 0 / ship-before" in the epic body
  (the optional pre-epic siblings pattern) are closed.

If the user passes `--wave=N` for a not-ready wave, the command refuses
with a one-line summary of what's still open and the suggestion: run
`/launch-team --epic=<id>` (without `--wave`) to launch the actual next
ready wave instead.

### Fallback when the epic body doesn't follow the template

If the epic's body doesn't have a `Wave` column in Child Tasks or
doesn't have an Execution Waves section, the command falls back to
**ticket-list mode** using the epic's open children as the list. Prints
a warning suggesting the epic be migrated to the template via the
example in [`beads-epic-template.md`](../templates/beads-epic-template.md).

## Process

### Step 0 — Epic-wave resolution (only if `--epic=<id>`)

Skip this step when launching in ticket-list mode.

1. **Load the epic**:
   ```bash
   bd show "<epic-id>"
   ```
   Confirm `Type: epic`. If not, abort with "expected epic, got
   `<type>`".

2. **Parse the Child Tasks table** from the epic body. Each row needs at
   minimum: `Wave`, `ID`, `Status`. If the table lacks a `Wave` column,
   the epic is pre-template — fall back to **ticket-list mode** with the
   epic's open children as the input list (print a one-line warning).

3. **Pick the target wave**:
   - With `--wave=N` — use N as the target.
   - Without `--wave` — the **lowest open wave whose prior waves are all
     closed**. Equivalent to: smallest N such that there exists an open
     ticket with wave = N AND all tickets with wave < N have status =
     closed.

4. **Wave-readiness check**:
   - For each ticket in waves 1..N-1: status must be `closed` (or
     `archived` / superseded).
   - For each `bd dep` edge into a wave-N ticket from outside the epic:
     blocker must be `closed`.
   - Wave-0 siblings flagged in the epic body's "Execution Waves"
     section must be `closed`.

   If any check fails, print:
   ```
   Wave <N> of <epic-id> not ready. Blockers:
     - <ticket-id>  status=<open|in_progress>  wave=<M>
     - ...
   Suggested: /launch-team --epic=<epic-id>   (auto-picks the actual next ready wave)
   ```
   and exit.

5. **Resolve the wave's ticket list**: all wave-N rows from the Child
   Tasks table whose status is not `closed`. Pass this list to Step 1
   below as if it were the user's input.

6. **Carry epic context forward**: keep the epic ID, epic title, and the
   wave's "Why this wave layout?" + "File-coordination risks"
   sub-sections — these get embedded in the round headers (Step 7) so
   the spawned team sees the wave's rationale, not just the bare ticket
   list.

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

#### Hard cap: 5 active agents per team round

**Never spawn more than 5 agents in parallel.** Running 10–11 agents has frozen
the host with OOM. The cap is a system constraint, not a guideline — exceeding
it crashes the session and loses unsaved progress for every teammate.

The three fixed roles (TL + QA + WH) consume 3 slots. That leaves **2 developer
slots per team round**. Spike variants where `developer` is replaced by
`researcher` follow the same cap (1 researcher per round is typical; never
more than 2).

If the input list has more tickets than slots, split into **team rounds** (NOT
to be confused with epic waves — see "Terminology" above):

| Tickets in input | Team rounds | Per-round team size |
|------------------|-------------|---------------------|
| 1                | 1           | 4 (TL + 1 dev + QA + WH) |
| 2                | 1           | 5 (TL + 2 devs + QA + WH) |
| 3                | 2           | round-a = 5, round-b = 4 |
| 4                | 2           | round-a = 5, round-b = 5 |
| 5+               | ⌈N/2⌉       | each round ≤ 5; respect dep order |

Round-split rules:
1. **Respect `bd dep` order.** A blocked ticket must land in a round AFTER its blocker. (In epic-wave mode this is rarely needed within a wave — tickets in the same epic wave have no inter-deps by construction.)
2. **Keep tickets that share files in the same round** (the file-ownership map in Step 3 catches this).
3. **One file per round.** Emit each round to its own file:
   - Ticket-list mode: `.notes/next-round-<N><suffix>.md` (e.g. `next-round-1a.md`, `next-round-1b.md`).
   - Epic-wave mode: `.notes/next-round-<epic-id-suffix>-w<wave><suffix>.md` (e.g. `next-round-tzv-w2-a.md`, `next-round-tzv-w2-b.md`).
   A single file containing multiple rounds obscures the user gate between them. If the user requests inline output instead of files, still separate rounds with a `========== ROUND BOUNDARY ==========` line.
4. **Per-round user gate is non-negotiable.** Project CLAUDE.md ("Do not proceed to next ticket without explicit user permission") makes round N+1 launch a user decision. The orchestrator MUST stop after Phase 7 of round N (handoff written, ripple done, gates green) and surface the round N+1 launch as an offer. Do NOT auto-chain even if the round files exist side-by-side. This rule overrides any "spawn the next round" wording elsewhere in the emitted prompt.
5. **Hand-off between rounds runs through `/handoff`.** Phase 7's `.notes/handoff-<slug>.md` is the entry point the next round reads first. `.notes/` is the gitignored scratchpad — see Rule 10 for the privacy constraint on this path.
6. **Tell the user about the split** in the Step 7 preamble (e.g. `"3 tickets in epic wave 2 → 2 rounds; round-a lands first, then I'll stop and ask before round-b"`).
7. **Inter-epic-wave gate is stricter than inter-round gate.** Wave N+1 of an epic cannot launch until EVERY ticket in Wave N is `closed` (not just done locally). The wave-readiness check in Step 0 enforces this. A round split within Wave N does NOT release Wave N+1.

### Step 5 — Extract Design Decisions

From the tickets and code read in Step 2, extract settled design decisions:
- Go types and their shapes (value objects, entities, aggregates)
- Interface contracts (port signatures, return types, context-arg position)
- Patterns to follow (reference existing code with file:line)
- Domain events emitted or consumed (project event bus)
- Bounded context boundaries (which package owns what)
- Constraints (what NOT to do — especially DDD layer rules)

### Step 6 — Generate Per-Agent Prompt Blocks

Two emission formats. **Use Step 6-sequential by default**; Step 6-team
only when `--mode=team` was given AND the Step -1 harness probe passed.

---

### Step 6-sequential — DEFAULT — Orchestrator playbook + reusable agent prompts

In sequential mode the **orchestrator session itself plays the TL role**.
There is no separate `tech-lead` spawn. The orchestrator spawns
`developer` / `qa-engineer` / `white-hacker` via the `Agent` tool one
phase at a time, receives each subagent's return value, and routes
follow-ups by spawning the next subagent with the right context.

This works in stock Claude Code because:
- The orchestrator session has full tool access (Bash, Read, Edit, Agent, etc.)
- Spawned subagents return their final message synchronously
- No SendMessage between spawned agents is needed
- Phase coordination is just "wait for result, branch on it" in the orchestrator session

#### Sequential output structure

For each team round, emit ONE file:
- Epic-wave mode: `.notes/next-round-<epic-id-suffix>-w<wave><suffix>.md`
- Ticket-list mode: `.notes/next-round-<N><suffix>.md`

Each file contains:

1. **Round header** — same epic-context fields as team mode (epic id,
   wave number, rationale, file-coordination, tickets, conflicts).

2. **`--- ORCHESTRATOR PLAYBOOK ---`** — pasted into the user's session.
   This is a STEP-BY-STEP runbook the orchestrator executes itself, NOT
   a TL spawn prompt. It contains:
   - **Phase 1 — Contracts (orchestrator).** The orchestrator reads
     every file in the ownership map and the relevant docs (`docs/DDD.md`,
     `docs/ARCHITECTURE.md`, ticket bodies). Distils a contract per
     ticket: exact port signatures, struct shapes, sentinel errors,
     DDD layer constraint. The contract gets embedded directly into the
     developer prompt — there is no broadcast step.
   - **Phase 2 — Spawn developer.** Use the Agent tool with the
     `--- DEVELOPER PROMPT ---` block below. `subagent_type` is
     `developer` if installed in `.claude/agents/`, else `claude` with
     the persona inline. `run_in_background: false` for serial flow;
     `true` only when batching independent tickets.
   - **Phase 3 — Spawn QA + WH in parallel.** ONE Agent tool call
     containing both subagent spawns; both get the developer's return
     text + the relevant slice. Wait for both returns before triaging.
   - **Phase 4 — Triage (orchestrator).** Parse QA + WH returns.
     Categorise findings: blocker / nice-to-have / out-of-scope.
   - **Phase 5 — Fix rounds (≤ 3).** If blockers, spawn `developer`
     again with the `--- DEVELOPER FIX PROMPT ---` block, then re-spawn
     QA + WH for re-verify. Cap at 3 rounds; escalate to user otherwise.
   - **Phase 6 — Close request.** Surface `bd close <id>` to the user
     with a one-line reason. Do NOT auto-close (CLAUDE.md rule).
   - **Phase 7 — Handoff.** Write `.notes/handoff-<slug>.md` yourself;
     print the path to the user. Privacy rule still applies — don't
     cite `.notes/` paths from committable artefacts.

3. **`--- DEVELOPER PROMPT ---`** — persona-augmented prompt to pass to
   the `developer` subagent in Phase 2. Includes:
   - The full ticket body from `bd show <ticket-id>`.
   - The file ownership list (only this dev's files).
   - The contract distilled by the orchestrator in Phase 1 (port
     signatures, struct shapes, sentinel errors, DDD layer constraint).
   - The TDD instruction (RED test before production code).
   - Explicit return-value format the orchestrator will parse:
     ```
     === DEVELOPER RETURN ===
     status: ready-for-review | blocked | needs-clarification
     diff-stat: <files changed, +/- lines>
     verification-results:
       - <command>: <PASS | FAIL with output>
     ac-self-check:
       - AC-1: ✓ | ✗ + file:line
       - ...
     deviations: <none | list>
     blocked-on: <empty | what>
     === END ===
     ```

4. **`--- QA-ENGINEER PROMPT ---`** — pass after Phase 2. Includes:
   - Ticket's AC + Edge Cases + RED-tests-by-name list.
   - The developer's return text verbatim (so QA has full context).
   - Explicit return-value format:
     ```
     === QA RETURN ===
     ac-coverage:
       - AC-1: ✓ | ✗ + evidence (file:line)
     edge-coverage:
       - Edge-1: ✓ | ✗ + evidence
     red-tests-present: <names in diff | missing list>
     regressions: <none | list>
     recommended: blocker | nice-to-have | none
     === END ===
     ```

5. **`--- WHITE-HACKER PROMPT ---`** — pass in parallel with QA after
   Phase 2. Same shape as QA, with the security focus list (trust
   boundaries, path safety, resource bounds, error suppression, logging
   privacy, dependency hygiene) and return format mirroring the
   WH-findings P5 format.

6. **`--- DEVELOPER FIX PROMPT ---`** — pass during Phase 5 fix rounds.
   Includes:
   - Original ticket reference.
   - Round number (1/3, 2/3, 3/3).
   - Each finding's repro + acceptance.
   - Same return-value format as the original developer prompt.

The orchestrator playbook explicitly **records the return text from each
spawn** in `.notes/round-<slug>-phases.md` as it goes — this is the
audit trail (since there is no SendMessage history to fall back on).

#### subagent_type for sequential mode

Use the project's installed personas where they fit:
- `developer` — implementer
- `qa-engineer` — independent verification
- `white-hacker` — security review
- `tech-lead` — **NOT spawned in sequential mode** (the orchestrator IS the TL)

If a persona doesn't exist for your role (check `ls .claude/agents/`),
fall back to `claude` and embed the persona inline in the prompt.

#### Why no separate TL spawn

In team-mode, the TL is a spawned agent that needs SendMessage to route
between dev and QA/WH. In sequential mode there's no inter-spawn
messaging — the orchestrator session handles routing. Spawning a
separate TL would add a layer of indirection without adding parallelism.
Skip it.

#### Sequential output mapping to existing protocol surfaces

| Sequential phase | Team-mode equivalent | What's different |
|------------------|----------------------|------------------|
| Phase 1 (orchestrator distils contract) | TL Phase 1 broadcast | No SendMessage; contract is embedded in the developer prompt |
| Phase 2 (developer spawn, sync return) | Dev Phase 2 + done-report | Return value replaces SendMessage payload; format is explicit |
| Phase 3 (parallel QA+WH spawn) | QA/WH Phase 3 findings | Two subagent spawns in one Agent batch; returns parsed by orchestrator |
| Phase 4 (orchestrator triages) | TL Phase 4 triage | Orchestrator code does what TL did via SendMessage |
| Phase 5 (re-spawn dev with fix prompt) | Dev Phase 5 fix cycle | Same ≤ 3-round cap; fix request embedded in new spawn prompt |
| Phase 6 (close request to user) | TL Phase 6 close | Same gate — user must approve `bd close` |
| Phase 7 (write handoff) | TL Phase 7 handoff | Same `.notes/handoff-<slug>.md` output |

---

### Step 6-team — OPT-IN — Per-agent SendMessage blocks (legacy)

**ONLY emit this format when `--mode=team` was given AND Step -1's
harness probe passed.** In stock Claude Code the probe will fail and you
should fall back to Step 6-sequential.

**Each agent gets ONLY its slice.** Pasting the full team plan into every
spawned agent (the old single-block emission) wastes context, dilutes audit
lenses, and risks role confusion (e.g. a dev acting on Phase 1 instructions
that belong to the TL). Emit one fenced markdown block per agent below,
labelled with the agent's name. The user copies each block into the
corresponding spawned subagent's prompt.

**Terminology inside the agent-facing blocks.** The 6.0–6.4 templates
below use the word "wave" to mean "this team round — the work the
spawned agents are executing right now." That's intentional: agents
don't need to distinguish between epic-wave and team-round levels; from
their perspective there's just "the batch I'm working on." The
distinction matters only at the orchestrator level (Steps 0, 4, 7 and
the round/wave rules above). In epic-wave mode, the orchestrator
preamble (6.0) gets ADDITIONAL fields under "Wave summary" carrying the
epic context (epic id, epic wave number, wave rationale) so the TL sees
the broader plan — but the rest of the agent blocks stay mode-agnostic.

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
subagents below using the Agent tool, route resume-messages to them when
they hit Phase boundaries, and confirm wave completion. Do NOT implement
code yourself — your role is fan-out + routing.

## Tool-loading preamble (run this FIRST, before anything else)

Step 1 — Call `ToolSearch({query: "select:SendMessage,TaskList,TaskStop"})`
to load the inter-agent transport. If `SendMessage` does not appear in the
result, abort: print "team-launch infrastructure unavailable in this
harness; falling back to single-agent execution mode" and proceed per the
"Failure modes & fallback" section of `alto-scaffold/commands/launch-team.md`.

Step 2 — **Canary pre-flight.** ToolSearch loading SendMessage in YOUR
session does NOT guarantee that a spawned subagent can also load it (the
deferred-tool set per spawned agent can differ from the orchestrator's).
Before fanning out the full N-agent team, spawn ONE canary agent with a
trivial prompt: "Run `ToolSearch({query: \"select:SendMessage\"})`, then
SendMessage `orchestrator` with the literal string `canary-ok`, then
exit." If the canary either fails to load SendMessage OR fails to deliver
the `canary-ok` message, abort and fall back to single-agent execution mode.
This costs one cheap spawn and saves ~140k tokens on a degraded fanout.

## Wave summary

- Tickets: <comma-separated ids>
- Team size: <N> agents (TL + <N-2> dev + QA + WH) — must be ≤ 5
- Files touched: <N>
- Conflicts: <none | list>

<!-- The four lines below appear only when the orchestrator was invoked
     via epic-wave mode. Emit them; omit otherwise. -->
- Epic:                <epic-id> · <epic title>
- Epic wave:           <wave number> of <total waves in epic>
- Wave rationale:      <one-line "why this wave" from the epic's "Why this wave layout?" subsection>
- File-coordination:   <one-line from epic's "File-coordination risks" subsection, or "none">

<!-- If this team round is one of several within the same epic wave,
     also emit: -->
- Round position:      <n> of <N> within epic wave <wave>
- Next round:          <path to the sibling next-round-... file, or "final round of this wave">
- Next epic wave:      <id of next wave, or "epic complete after this wave">  (user gate still enforced regardless)

## Spawn instructions

Use the Agent tool with these `subagent_type` + `name` + `prompt` settings.
Default `subagent_type` is `claude` (catch-all) because the persona-typed
agents are not registered in `.claude/agents/` for most projects; the
persona content is embedded in each prompt block.

| Agent name | subagent_type | Prompt block |
|------------|---------------|--------------|
| tech-lead | claude | block "TECH-LEAD" below |
| dev-<ticket-id> | claude | block "DEVELOPER: dev-<ticket-id>" below |
| qa-engineer | claude | block "QA-ENGINEER" below |
| white-hacker | claude | block "WHITE-HACKER" below |

Spawn each agent with `run_in_background: true`. The `Agent` tool returns
an `agentId` (format `a...-...`) for each. **Record every agentId in a
table you keep for the duration of the wave** — you will need it to
resume agents after they exit on Phase boundaries (see "One-shot agent
semantics" below).

**Spawn all agents in one batched tool call (parallel).** Don't serialize
"TL first, then dev, then QA+WH" — all four/five agents are one-shot and
exit on WAIT, so the dev/QA/WH spawn order has no effect on what the TL
does in Phase 1 (the TL exits and gets resumed when the orchestrator
delivers Phase-1 messages anyway). Parallel spawn cuts wall-clock and
produces identical behavior. The serial-order hint was folklore from a
stateful-agent model that does not apply here.

## One-shot agent semantics — read carefully

Spawned agents run their reasoning loop and EXIT when they have no more
actions to take. A "Phase WAIT" in a dev block means the dev will
complete-and-exit immediately after acknowledging readiness — there is
no suspended process. This is expected; do not retry the spawn.

To wake a completed agent and feed it a new turn, call
`SendMessage({to: "<agentId>", message: "..."})`. Addressing by display
name (`to: "dev-alty-cli-2f9"`) only works while the agent is alive.
After an agent has emitted a `<task-notification>` with `status:
completed`, you MUST use its `agentId` to resume it.

Therefore: keep the agentId table from the spawn step. You will:
- Resume the TL with each Phase 3 finding from QA/WH (TL exits after
  broadcasting contracts).
- Resume each dev with the TL's contract message (the dev exits after
  ACKing readiness or after a fix-cycle iteration).
- Resume QA/WH with each dev's Phase 2 done-report.

The agents themselves address peers by name in their SendMessage calls;
you (the orchestrator) translate name → agentId when relaying.

## Communication routing

- Devs report Phase 2 completion → qa-engineer + white-hacker (via you)
- QA + WH report Phase 3 findings → tech-lead (via you)
- Tech-lead assigns Phase 4 fixes → specific dev (via you)
- Peer-to-peer clarifications (dev ↔ QA/WH) flow directly while both alive;
  otherwise via you
- Max 3 fix rounds per issue — TL escalates to YOU if exceeded
- All inter-agent communication via SendMessage; plain text output is
  invisible to peers (it goes only to you, the orchestrator)

## Wave-end

Tech-lead will signal completion after Phase 7 writes
`.notes/handoff-<slug>.md`. Print that path locally and confirm the wave
is done. Do not commit / push — that requires explicit user permission
per CLAUDE.md.

**Privacy:** `.notes/` is the project's gitignored scratchpad. Do NOT cite
the handoff path from anything that gets committed (commit messages,
ticket bodies, PR descriptions, code comments, `bd close --reason` text).
Print it to the user; keep it out of the repo's public surface.

**Do NOT auto-launch the next round.** Even if a sibling `next-round-…`
file exists, the per-round user gate from Step 4 (Round-split rule #4)
applies: stop, surface the offer, wait. See also "Failure mode 4 — team
mode degraded between rounds" for the cross-round consequence of
single-agent fallback.
````

#### 6.1 — tech-lead block

The TL is the orchestrator inside the wave. It gets the broadest view.

````markdown
# TECH-LEAD — <wave title or ticket ids>

You are the technical lead for this wave. Your authority: interface
contracts, triage of QA + WH findings, fix assignment, final quality gates,
close + ripple + handoff. Do NOT write production code yourself unless a
finding cannot be safely delegated.

## Tool-loading preamble (run this FIRST, before reading any file)

Call `ToolSearch({query: "select:SendMessage"})` to load the inter-agent
transport. If `SendMessage` does not appear in the result, abort: send a
single plain-text reply "SendMessage unavailable; team-mode broken — need
orchestrator decision" and exit. Do NOT attempt to compensate by writing
code yourself unless the orchestrator explicitly switches you to
single-agent execution mode.

## One-shot agent semantics

You are a one-shot agent: when you run out of actions you exit. If you are
WAITING for a dev's ACK or a QA/WH finding, just exit — the orchestrator
will resume you with a SendMessage when the message arrives. Do not loop
or sleep waiting for input.

When you address peers in SendMessage, use their display name
(`dev-<ticket-id>`, `qa-engineer`, `white-hacker`). The orchestrator
translates names to agentIds for resume.

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
  - Exact port/function signatures they must implement (in the project's language syntax)
  - Struct / value-object shapes (with field names + types)
  - Sentinel errors / domain failure modes they must use
  - Cancellation / context propagation convention (e.g. Go `ctx context.Context` first arg, Python async cancel scope, TypeScript `AbortSignal` — per project convention)
  - Domain events emitted or consumed
  - DDD layer constraint: which package / module the dev owns + which they may import
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

**Phase 7 — Handoff.** Invoke `/handoff` to write `.notes/handoff-<slug>.md` covering tickets closed, files changed + line counts, unresolved findings, follow-up tickets filed, next-session entry points. Print the path to the user. **Do NOT cite this path from any committable artefact** — `.notes/` is the gitignored scratchpad; mentioning it in commit messages, ticket bodies, PR descriptions, code comments, or `bd close --reason` leaks a local scratchpad path into the OSS repo's permanent surface. Reference shipped artefacts using repo-relative paths in the project's source tree (e.g. `<source-root>/<context>/...`, `docs/...`) in those committable surfaces, not the handoff scratch.

## Quality gates

Project-specific. See `.project.md` sibling.

## Enforced principles (every fix you assign must hold these)

- **Ubiquitous Language** — names match `docs/DDD.md` glossary; reject synonyms
- **Value Objects first** — default to immutable VOs; entities only when identity is needed
- **One aggregate per transaction** — reference other aggregates by ID, not pointer
- **Port/Adapter** — handlers depend on port interfaces in application layer, never on concrete adapters
- **TDD required** — RED test BEFORE production code; REFACTOR keeps suite green
- **Wrapped errors** — errors cross-layer carry context (language-specific syntax — Go `fmt.Errorf("doing X: %w", err)`, Python `raise ... from err`, TypeScript `throw new Err("...", { cause: err })`); messages lowercase, no trailing punctuation
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

## Tool-loading preamble (run this FIRST, before anything else)

Call `ToolSearch({query: "select:SendMessage"})` to load the inter-agent
transport. If `SendMessage` does not appear, send a plain-text reply
"SendMessage unavailable; cannot ACK tech-lead" and exit — the
orchestrator will surface the harness mismatch.

## One-shot agent semantics

You are a one-shot agent. When you reach a WAIT state (no contract yet,
or no fix assignment yet) you EXIT cleanly — the orchestrator will resume
you with a SendMessage when the next message arrives. This is normal;
do not loop, sleep, or poll. Your context and tool state are restored
on resume.

## Phase 1 — Acknowledge readiness, then exit

Do NOT begin implementation. SendMessage to `tech-lead` with a one-line
ACK ("dev-<ticket-id> ready"), then exit. The tech-lead will SendMessage
you a contract broadcast containing the exact signatures, struct shapes,
sentinel errors, context conventions, and DDD layer constraints for your
ticket. Reading it (on resume) is your trigger for Phase 2.

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
- **Wrapped errors** — errors cross-layer carry context (language-specific syntax — Go `fmt.Errorf("doing X: %w", err)`, Python `raise ... from err`, TypeScript `throw new Err("...", { cause: err })`); messages lowercase, no trailing punctuation
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

## Tool-loading preamble (run this FIRST, before anything else)

Call `ToolSearch({query: "select:SendMessage"})` to load the inter-agent
transport. If `SendMessage` does not appear, send a plain-text reply
"SendMessage unavailable; cannot route findings" and exit.

## One-shot agent semantics

You are a one-shot agent: exit cleanly while waiting for a dev's
done-report or a re-review request. The orchestrator resumes you when
the next message arrives.

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

## Tool-loading preamble (run this FIRST, before anything else)

Call `ToolSearch({query: "select:SendMessage"})` to load the inter-agent
transport. If `SendMessage` does not appear, send a plain-text reply
"SendMessage unavailable; cannot route findings" and exit.

## One-shot agent semantics

You are a one-shot agent: exit cleanly while waiting for a dev's
done-report or a re-review request. The orchestrator resumes you when
the next message arrives.

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
  Dependency-hygiene: ✓ | ✗ + evidence
  Severity (if ✗): S0 | S1 | S2 | S3
  Recommended: blocker | nice-to-have | none
  ```

## Communication

SendMessage. Receive Phase 2 done-reports from devs; send Phase 3 findings TO the tech-lead (NOT to devs). Peer-to-peer clarifications with devs / QA are fine. Phase 5 re-verify follows the same flow on the fix-round diff.
````

### Step 7 — Present to User

**One file per team round.** Write each round's full block set to its
own file. Print every path to the user.

File naming:

- Ticket-list mode: `.notes/next-round-<N><suffix>.md` — e.g.
  `next-round-1a.md`, `next-round-1b.md` for a 2-round split.
- Epic-wave mode: `.notes/next-round-<epic-id-suffix>-w<wave><suffix>.md`
  — e.g. `next-round-tzv-w2-a.md`, `next-round-tzv-w2-b.md` for Wave 2
  of `k3s-setup-tzv` split into two rounds. The `<epic-id-suffix>`
  strips the project prefix (e.g. `tzv` from `k3s-setup-tzv`) so paths
  stay short.

Within each round file, output the blocks in this order, separated by a
one-line header per block: **1 orchestrator preamble + N agent blocks**.

Prefix the round with this header (single-round + non-epic: drop the
`ROUND <n> of <total>` and `Next round entry point` lines; non-epic:
drop the `Epic` / `Epic wave` / `Wave rationale` / `File-coordination`
lines):

```
TEAM LAUNCH — ROUND <n> of <total>
Epic:                 <epic-id> · <epic title>     (epic-wave mode only)
Epic wave:            <N>                          (epic-wave mode only)
Wave rationale:       <one-line "why this wave" from epic body>   (epic-wave mode only)
File-coordination:    <one-line from epic's File-coordination risks subsection | none>   (epic-wave mode only)
Tickets:              <ids in this round>
Team size:            <N> (TL + <devs> + QA + WH)   — must be ≤ 5
Files touched:        <N>
Conflicts:            <none | list>
Next round entry pt:  .notes/handoff-<slug>.md   (omit on final round of the wave)
Next epic-wave gate:  after this round closes ALL wave-<N> tickets, ask user before launching wave <N+1>   (epic-wave mode, final round of wave only)
User-gate:            orchestrator stops after Phase 7 and asks before round <n+1>   (omit on final round)

Paste the ORCHESTRATOR PREAMBLE into the Claude Code session that will run
the team. Then paste each subsequent block into the corresponding spawned
subagent's prompt — do NOT paste the full set into any single agent.
```

Then emit. The block structure depends on the execution mode resolved
during emission:

**Sequential mode (default):**

1. `--- ORCHESTRATOR PLAYBOOK ---` — the runbook the user's session
   executes itself (Phase 1-7, including its own Agent-tool spawn calls).
2. `--- DEVELOPER PROMPT ---` — passed to each `developer` spawn in Phase 2 (repeat per dev).
3. `--- QA-ENGINEER PROMPT ---` — passed to the `qa-engineer` spawn in Phase 3.
4. `--- WHITE-HACKER PROMPT ---` — passed to the `white-hacker` spawn in Phase 3.
5. `--- DEVELOPER FIX PROMPT ---` — passed to the `developer` re-spawn during Phase 5 fix rounds (template — orchestrator fills in finding + round number at spawn time).

**Team mode (opt-in, only after Step -1 probe passes):**

1. `--- ORCHESTRATOR PREAMBLE ---` followed by the fenced 6.0 block
2. `--- TECH-LEAD ---` followed by the fenced 6.1 block
3. `--- DEVELOPER: dev-<ticket-id> ---` followed by the fenced 6.2 block (repeat per dev)
4. `--- QA-ENGINEER ---` followed by the fenced 6.3 block
5. `--- WHITE-HACKER ---` followed by the fenced 6.4 block

For multi-round runs, write one file per round (in dep order if any
deps exist) and print all paths. Do NOT concatenate rounds into a single
file — the user reads each one independently between launches.

In epic-wave mode, the orchestrator preamble's (team) or the playbook's
header (sequential) carries the epic context (epic id, wave number,
wave rationale) so the orchestrator knows the broader plan, not just
the per-round slice.

## Rules

1. **Never launch the team yourself.** Only generate the prompt blocks — the user spawns agents.
2. **Read actual code.** Every file reference must be verified by reading the file.
3. **Flag undergroomed tickets.** Warn if a ticket lacks acceptance criteria or file paths.
4. **Resolve file conflicts.** If two tickets own the same file, stop and ask the user.
5. **No placeholders.** Every `<placeholder>` in the emitted blocks must be filled with real data. If a slot can't be filled, say what's missing in the wave header before the blocks.
6. **Cite Enforced Principles in the slice that needs them.** The TL block gets the full list (TL enforces them). The dev block gets the list (the code must hold them). QA + WH blocks reference them by topic, not the full list. Don't repeat the same paragraph four times.
7. **End every round with `/handoff`.** Phase 7 in the TL block mandates `.notes/handoff-<slug>.md`. Slug = ticket ID for single-ticket, short round / wave name for multi-ticket.
8. **Hard cap: 5 active agents per team round.** TL + QA + WH = 3 fixed; 2 dev slots max. More tickets → more rounds. Host constraint, not a preference.
9. **Slice per agent — never paste the full team plan into a non-TL agent.** The dev sees their ticket + DDD + TDD + AC + edges + their files + quality gates + Phase 1-2-5 instructions only. QA sees AC + edges + RED tests + Phase 3 lens. WH sees the security surface + Phase 3 security lens. The TL alone sees the broad view because the TL coordinates. This is the structural fix for the "dev got the entire team prompt" failure mode.
10. **Handoff paths stay out of committable artefacts.** `.notes/handoff-<slug>.md` lives in the gitignored scratchpad. The TL prints the path to the user but must NOT cite it from `bd close --reason`, commit messages, ticket bodies, PR descriptions, or code comments. Reference shipped repo-relative paths under the project's source tree (e.g. `<source-root>/...`, `docs/...`) in those committable surfaces. Bake this into the Phase 7 instruction the TL block emits.
11. **Per-round user gate (cross-round control flow).** Project CLAUDE.md says "Do not proceed to next ticket without explicit user permission." The launch-team flow inherits that: after Phase 7 of round N, the orchestrator stops and offers round N+1 — it never auto-chains, even when the round files exist side-by-side. See Round-split rule #4 in Step 4 and Failure Mode 4 below for the degraded-mode caveat.
12. **Inter-epic-wave gate is stricter and lives in `bd close` status, not in user permission alone.** Wave N+1 of an epic launches ONLY when every ticket in Wave N has status `closed`. A round-split within Wave N does NOT release Wave N+1; the orchestrator MUST run the wave-readiness check (Step 0.4) before treating a new `--epic` invocation as ready. The user gate from Rule 11 still applies on top.
13. **Epic-wave mode respects the epic body as truth.** In epic-wave mode the wave's ticket list, dep edges, file-coordination risks, and rationale come from the epic body — not from re-deriving them at launch time. If the epic body is stale, fix it via `/design-ticket --epic=<id>` (which keeps the epic in sync) rather than reasoning around it here.

## Failure modes & fallback

### Failure mode 1 — `SendMessage` not loadable

Symptom: the orchestrator's tool-loading preamble (`ToolSearch
select:SendMessage`) returns no match, or the TL's first run reports
"SendMessage unavailable."

Cause: harness build does not expose `SendMessage` as a deferred tool.

Action: **abort team mode immediately.** Switch to single-agent
execution: keep the TL alive, send it a follow-up SendMessage (after
loading it — if even the orchestrator can't, fall through to the user's
own session) that drops the routing role and instructs the TL to
implement every ticket sequentially itself. The TL is the most
context-rich persona; it has already read the ownership-map files.

Example resume message:
```
Single-agent execution mode. Drop the dev/QA/WH personas. Implement
<ticket-1> first, then <ticket-2>. RED tests before production. Run
the full quality gate after each ticket. Do not commit, do not bd close
— report back with a diff stat, AC self-check, and suggested commit
messages.
```

### Failure mode 2 — custom `subagent_type` rejected

Symptom: `Agent` tool call with `subagent_type: tech-lead` (or other
custom persona) errors with "unknown agent type."

Cause: `.claude/agents/` is empty or missing the persona file.

Action: re-spawn with `subagent_type: claude` (the default in this
skill). The persona content is already embedded in the prompt; the
catch-all type is functionally equivalent for prompt-driven roles. The
emitted blocks already default to `claude` per the Prerequisites
section.

### Failure mode 4 — team mode degraded between rounds

Symptom: round N fell through to single-agent execution mode (Failure mode 1
or 3). Round N+1 is queued behind it in a sibling
`.notes/next-round-<id-suffix>-w<wave>b.md` file and was designed for a
5-agent paired execution.

Cause: the harness can't sustain team mode this session, but the round
N+1 design assumes it.

Action: **STOP before launching round N+1.** Surface the choice to the
user:
- The round N+1 design's value proposition (parallel paired execution,
  separate review lenses) collapses under single-agent fallback. Round
  N+1 becomes "one agent does both tickets serially" — equivalent to a
  single bigger prompt with no orchestration benefit.
- Options: (a) defer round N+1 to a session where team mode works,
  (b) launch round N+1 in single-agent execution mode now and accept the
  lost signal, (c) abandon the multi-round plan and re-groom as solo
  tickets.

The orchestrator MUST NOT silently launch round N+1 in degraded mode —
that spends tokens on a degraded design without telling the user the
design's value prop is gone. Note that the inter-epic-wave gate (Rule
12) is independent of this failure mode — even if round N+1 launches
successfully, Wave N+1 of an epic still requires every ticket in Wave N
to be `closed`.

### Failure mode 3 — agent exits without ACKing

Symptom: a dev or TL spawn completes (`status: completed`) but the
result is "needs input" or an apology without a SendMessage ACK.

Cause: the agent didn't run the tool-loading preamble (older prompt
template) or the preamble failed silently.

Action: send a resume message via `SendMessage({to: "<agentId>",
message: "Re-run your tool-loading preamble: ToolSearch
select:SendMessage, then ACK readiness."})`. If two consecutive resumes
fail to produce an ACK, fall back to single-agent execution mode per Failure
mode 1.

### When to escalate to the user

Escalate if:
- Two of the five agents have failed to ACK after a resume cycle.
- The TL has issued > 3 fix rounds for the same finding without
  convergence.
- Any agent reports a security finding rated S0 or S1.
- The harness probe at orchestrator-startup fails.

Print a one-line summary of what went wrong and ask the user whether to
fall back to single-agent execution mode, retry team mode after fixing the
harness, or abandon the wave.
