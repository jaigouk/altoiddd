---
name: design-ticket
description: Design a beads ticket (task / spike / bug) with type-aware validation, DDD/SOLID review, and optional wave placement within an epic
kind: command
phase: groom
when_to_use: When designing a single beads ticket and verifying it against the architecture before creating it — optionally placing it in the right execution wave of an epic
tools: Agent, Bash, Read, Grep, Glob
bash_substitution_policy: quoted
license: Apache-2.0
---

# /design-ticket <description> [--type=task|spike|bug] [--epic=<id>]

Design a single beads ticket with built-in architecture, DDD, and type-aware
validation. Uses a two-agent chain: a type-matched designer drafts the ticket
from codebase analysis; `tech-lead` reviews it against `docs/DDD.md` and
`docs/ARCHITECTURE.md` before creation. When `--epic=<id>` is given, the new
ticket is placed in the correct execution wave of that epic and the epic body
is updated to reflect the addition.

## Why This Exists

Creating tickets from rough bullet points produces vague descriptions that
fail during implementation. `/groom` catches this AFTER creation — but by
then the ticket exists with wrong assumptions baked in. This command prevents
bad tickets from being created in the first place, and adds two things `/groom`
cannot:

1. **Type-aware analysis.** A `task` needs trace-the-callers; a `spike` needs
   a sharp question and an exit criterion; a `bug` needs reproduction steps
   and a root-cause hypothesis. One pipeline can't validate all three.
2. **Wave placement inside an epic.** When a ticket is a child of an epic,
   its wave decides whether it runs in parallel with siblings or blocks them.
   Getting the wave wrong silently serialises work that could have been
   parallel — or worse, creates two parallel tickets touching the same file.

## Usage

```bash
# Standalone task
/design-ticket StorytellingFlow domain strategy for RAPID/THOROUGH modes

# Standalone spike (researches a question; output is a decision + artefact)
/design-ticket --type=spike Evaluate in-process vs broker-based event bus for our async dispatch

# Standalone bug (with repro context)
/design-ticket --type=bug doc-health crashes when README.md is empty

# Child of an existing epic — places it in the right wave
/design-ticket --epic=<epic-id> --type=task Implement CLI adapter for the prompter port
```

If `--type` is omitted, default is `task`. If `--epic` is omitted, no wave
assignment runs (Phase 3 is skipped).

## Process

### Phase 1 — Type-Aware Design

Launch a subagent matched to the ticket type. Each agent reads the repo state
relevant to that type and produces a body matching the right template.

| Type | Subagent | Reads | Template |
|------|----------|-------|----------|
| `task` | `feature-dev:code-architect` | `alto-scaffold/templates/beads-ticket-template.md`, `docs/DDD.md`, `docs/ARCHITECTURE.md`, existing code in the relevant bounded context (ports, handlers, adapters, domain types), spike reports referenced from the description | `alto-scaffold/templates/beads-ticket-template.md` |
| `spike` | `researcher` | `alto-scaffold/templates/beads-spike-template.md`, prior research under `docs/research/`, upstream docs via WebFetch, existing ADRs | `alto-scaffold/templates/beads-spike-template.md` |
| `bug` | `feature-dev:code-explorer` + `feature-dev:code-reviewer` | `alto-scaffold/templates/beads-bug-template.md`, `alto-scaffold/templates/bug-rca-template.md`, git log around suspected files, recent commits, related tests, logs/metrics references | `alto-scaffold/templates/beads-bug-template.md` |

**For `task`, the architect prompt is:**

> Design a detailed implementation ticket for: {user's description}
>
> Read and analyze:
> - `alto-scaffold/templates/beads-ticket-template.md` — MANDATORY template structure (read FIRST)
> - `docs/DDD.md` — bounded contexts, ubiquitous language glossary
> - `docs/ARCHITECTURE.md` — layer rules, planned file layout, ADRs
> - Existing code in the relevant bounded context (ports, handlers, adapters, domain types)
> - Any spike research reports referenced in the user's description
> - If `--epic=<id>` is given: `bd show <epic-id>` plus all of its current children, to understand sibling scope
>
> Produce a ticket following `beads-ticket-template.md` EXACTLY. Include ALL sections:
> - Goal / Problem
> - Background / Context (with verified file:line references)
> - DDD Alignment table
> - Design section (data models, method signatures, sequence flow)
> - SOLID Mapping
> - TDD Workflow (RED/GREEN/REFACTOR with specific test names)
> - Steps
> - Acceptance Criteria
> - Edge Cases
> - Quality Gates
> - Pre-Implementation Validation (verified against actual code — cite file:line)
> - Risks / Dependencies
>
> Key requirements:
> - Every file:line claim must be verified by reading the actual file
> - Every method signature must match what's actually in the code
> - Constructor chains must be traced through the composition root
> - Name every dependency (no "TBD" or "to be determined")
> - List every file the ticket will create or modify under a `Files in Scope` block (used by Phase 3 for wave placement)

**For `spike`**, populate at least: Problem statement, Research questions, Investigation plan, Exit criteria ("we will know we're done when…"), Time box, Likely outcomes, Follow-up intents (tickets that should exist if the spike says "yes").

**For `bug`**, populate at least: Reproduction (exact commands), Expected vs Actual, Suspected root cause (file:line or specific behaviour, not "something with X"), Files in scope, Acceptance Criteria (must include a regression test), Severity vs priority justification.

Skip sections that genuinely don't apply (e.g. "Storage" for a doc-only
ticket) but mark them `N/A` explicitly — never silently drop a template
section.

### Phase 2 — Tech Lead Review

Launch a `tech-lead` agent (or `feature-dev:code-reviewer` for bugs) with
the right checklist per type:

> Review this ticket design for DDD/SOLID/CQRS-lite compliance:
>
> {paste designer output}
>
> Check against:
> 1. **DDD layer rules**: Does the ticket respect infrastructure → application → domain?
> 2. **Bounded context boundaries**: Does it stay within one context? Any cross-context leakage?
> 3. **Ubiquitous language**: Do all type/method names match `docs/DDD.md` glossary?
> 4. **Port/adapter pattern**: Are ports in application, adapters in infrastructure?
> 5. **SOLID violations**: Any god objects, leaky abstractions, or concrete dependencies?
> 6. **Scope creep**: Does it try to do too much? (>8 new files, >3 dirs, >1 bounded context → consider split)
> 7. **Missing dependencies**: Are there tickets that should exist but don't?
> 8. **Type-specific gates** (apply only the row matching `--type`):
>
> | Type | Gates |
> |------|-------|
> | `task` | Existing patterns reused? Names follow `docs/DDD.md`? Secrets not hardcoded? Quality gates listed? |
> | `spike` | Sharp Yes/No question? Time box ≤ 1 week? Concrete exit criteria? No-go output produces an artefact? |
> | `bug` | Repro works on a fresh checkout? Root-cause hypothesis is concrete (file:line)? Acceptance includes a regression test? Severity matches priority (cluster-down = P0, blocks one user = P2, cosmetic = P3)? |
>
> Output one of: **APPROVED**, **APPROVED WITH FIXES** (list fixes), or **NEEDS REDESIGN** (explain why).

### Phase 3 — Wave Assignment (only if `--epic=<id>`)

Skip this phase when there is no `--epic` flag.

1. **Load the epic's current child set**:
   ```bash
   bd show "<epic-id>"
   bd list --status=open | grep "<epic-id>"     # children of this epic
   ```

2. **Determine this ticket's wave** by the wave-grouping rules below:
   - Identify which existing ticket(s) this new one will depend on (from the
     Risks / Dependencies section of the Phase-1 output).
   - Find their wave numbers in the epic's "Execution Waves" section
     (or, if the epic has no waves yet, in its "Phases / Milestones" section
     — see "Adding Waves to an Epic Without Them" below).
   - This ticket's wave = `max(deps' waves) + 1`, unless it has no in-epic
     deps → Wave 1.

3. **Find parallel-safe siblings in the same wave** using the `Files in Scope`
   block from Phase 1:
   - Same dep set + disjoint file scope + no semantic ordering → same wave.
   - If a sibling already in the candidate wave touches files this ticket
     will also touch, either:
     - **(a)** Move this ticket to the next wave (clean, but slows the epic), or
     - **(b)** Document the coordination point in the epic's
       "Risks & Mitigations" table (explicit handoff between parallel tickets,
       e.g. "Ticket A lands with line X commented out; ticket B uncomments it
       during its smoke test").

4. **Update the epic's body**:
   - Add this ticket to the "Child Tasks" table with the wave number.
   - Update the "Execution Waves" block to include this ticket in the right
     wave.
   - Update the "Dependency Graph" if new edges were created.
   - If a new coordination risk surfaced, add a row to "Risks & Mitigations".

5. **Wire dependencies in beads**:
   ```bash
   bd dep add "<new-ticket-id>" "<parent-dep-id>"   # for each in-epic dep
   ```

### Phase 4 — Apply Fixes and Create

If tech-lead said **APPROVED WITH FIXES**:
- Apply the fixes to the ticket body.
- Show the user the changes for approval.

If tech-lead said **NEEDS REDESIGN**:
- Show the user the tech-lead's concerns.
- Ask whether to redesign or proceed anyway.

Once approved by the user:

```bash
bd create \
  --title="<short imperative title>" \
  --type="<task|spike|bug>" \
  --priority="<0-4>" \
  --description="<body from Phase 1, validated in Phase 2>"

# If --epic was provided:
bd update "<new-id>" --parent="<epic-id>"
bd dep add "<new-id>" "<dep-id>"           # for each in-epic dep

# If the epic body was updated in Phase 3:
bd update "<epic-id>" --description="<updated epic body>"
bd export -o .beads/issues.jsonl            # persist title + dep updates
```

Add labels per project convention (e.g. context label):

```bash
bd label add "<new-id>" "<context-label>"
```

### Phase 5 — Report

```
================================================================
TICKET DESIGN REPORT
================================================================
TITLE:           <ticket title>
TYPE:            <task | spike | bug>
BEADS ID:        <created id>
PARENT EPIC:     <epic id | none>
WAVE:            <N | n/a>
SCOPE:           <N> files to create / modify

DESIGNER:        <code-architect | researcher | code-explorer+code-reviewer>
  Files read:           <N>
  Signatures verified:  <N>
  Claims cited:         <N> (all with file:line)

TECH-LEAD REVIEW:
  DDD compliance:         [PASS | FIX APPLIED]
  Layer rules:            [PASS | FIX APPLIED]
  Ubiquitous language:    [PASS | FIX APPLIED]
  SOLID:                  [PASS | FIX APPLIED]
  Scope:                  [OK | SPLIT RECOMMENDED]
  Type-specific gates:    [PASS | FIX APPLIED]

WAVE PLACEMENT (if --epic):
  Wave assigned:          [N | n/a]
  File-coordination risk: [NONE | FLAGGED in epic risks]
  Dependencies wired:     [N edges]

EPIC UPDATE (if --epic):
  Child Tasks table:       [updated | n/a]
  Execution Waves block:   [updated | n/a]
  Dependency Graph:        [updated | n/a]

VERDICT: CREATED / NEEDS USER INPUT / NEEDS REDESIGN
================================================================
```

## Wave Grouping Rules

When designing tickets that are children of an epic, place each ticket in the
correct execution wave so parallel work is obvious.

### Definitions

- **Wave 1**: child tickets with no dependencies on other children of the same
  epic. Ready immediately when the epic starts.
- **Wave N+1**: depends on at least one ticket in Wave N. Cannot start until
  all of its dependencies are closed.
- A ticket's wave = `max(wave of each of its deps within the epic) + 1`. If it
  has no in-epic deps, it's Wave 1.

### Two tickets belong in the SAME wave iff ALL are true

1. **Same dependency set** — both depend on the same in-epic parent ticket(s),
   OR both have no in-epic deps (both Wave 1).
2. **Disjoint file scope** — their `Files in Scope` blocks do not overlap on
   any file they both create or modify.
3. **No semantic ordering** — neither produces an artefact the other needs
   before starting. (E.g. "domain strategy for THOROUGH mode" depends on
   "StorytellingFlow port exists" is a semantic ordering — they can't be in
   the same wave even if their files are disjoint.)
4. **Parallel-session-safe** — could be picked up by two different agents (or
   two different humans) without coordination beyond what's documented in the
   epic body.

If a sibling exists in the same wave that violates rules 2 or 3, decide
between:

- **Move this ticket to the next wave**: cleanest, but slows the epic.
- **Document the coordination point in the epic's "Risks & Mitigations"
  table**: explicit handoff between parallel tickets.

### Wave ordering

Waves run **strictly sequentially at the wave boundary**. All Wave N tickets
must be closed before any Wave N+1 ticket can start. Within a wave, tickets
run in parallel.

### Wave naming

Number waves starting from 1. An optional **Wave 0** can hold sibling tickets
that are NOT children of the epic but should ship before any in-epic work
(e.g. a doc cleanup that prevents stale references during later analysis).

### Adding Waves to an Epic Without Them

If the parent epic uses `beads-epic-template.md`'s "Phases / Milestones"
structure but has no "Execution Waves" block yet, add one when placing the
first wave-aware ticket. Append this block under the epic's existing "Phases
/ Milestones" section:

```markdown
## Execution Waves

> Waves run sequentially at the wave boundary; tickets within a wave run in parallel.

### Wave 1 — <theme: e.g. "ports and value objects">
- `<child-id>` — <one-line summary>
- `<child-id>` — <one-line summary>

### Wave 2 — <theme: e.g. "handlers and adapters">
- `<child-id>` — <one-line summary> (depends on Wave 1)

### Dependency Graph
```
<child-A> ──→ <child-C>
<child-B> ──→ <child-C>
```
```

Mention the addition in the user-approval step of Phase 4.

## Rules

1. **One ticket per `/design-ticket` invocation.** Design quality drops when batching.
2. **Designer reads code, tech-lead reads docs.** Don't mix roles.
3. **Templates are not optional.** Every ticket must conform to its
   type's template under `alto-scaffold/templates/`. Mark `N/A` sections
   explicitly; don't drop them silently.
4. **Every claim is cited.** No "verified" without file:line.
5. **User approves before creation.** Never auto-create tickets.
6. **Epic stays in sync.** When `--epic` is given and the design changes the
   epic (wave placement, deps, risks), update the epic body in the same step.
   A child without a parent update is half-done work.
7. **Wave assignment uses the epic's existing waves as ground truth.** Don't
   invent new waves — fit into the existing structure or document why a new
   wave is needed.

## Self-contained-execution requirement

Tickets designed by this command must be **executable by a single agent
with no inter-agent messaging**. `/launch-team` defaults to
sequential-orchestrator mode (no SendMessage between spawned agents — see
[`launch-team.md`](launch-team.md) "Two execution modes" for why), so
each ticket has to stand on its own when handed to a `developer`
subagent.

Concretely this means:

1. **`Files in Scope` is the source of truth
   for scope.** The developer reads only what the ticket says it owns.
   If a file is needed but not listed, the developer will refuse to
   touch it.
2. **Acceptance Criteria must be machine-verifiable.** Each AC item
   should be checkable by a command (preferred) or a one-line file:line
   assertion. Vague AC ("works correctly", "looks good") cannot be
   self-verified by the developer and will surface as ambiguity to QA.
3. **Verification commands must be self-executable.** The developer runs
   them as part of Phase 2; if they need data the ticket doesn't
   provide, the developer is stuck.
4. **Prerequisites must be enumerable.** The developer checks them at
   the start of Phase 2. Missing prereq → block, not guess.
5. **Design section spells out the contract.** Port signatures, struct
   shapes, sentinel errors — written into the ticket body, not left for
   a TL to broadcast at runtime. In sequential mode there is no TL
   broadcast.

Tickets that fail this bar still work in team-mode (a TL agent can
clarify mid-flight), but they break in sequential mode. Default to
self-contained.

The tech-lead review in Phase 2 must add this as gate #9 (any type):
"Is this ticket self-contained enough to hand to a single developer
spawn? If a sequential-mode developer would need to ask a clarifying
question to proceed, the ticket fails this gate."

## Product Dogfooding Note

This command implements what alto's ImplementabilityValidator (Ticket Pipeline
context) should do automatically for users' generated tickets. The
designer→tech-lead chain maps to:

- ImplementabilityValidator (section completeness, unspecified dependencies)
- CodebasePortScanner (method signature verification)
- DesignTraceResult (structured findings)

The wave-grouping rules map to a future `WavePlanner` component in the same
context (input: epic + new ticket; output: wave number + coordination-risk
findings). When implementing the Ticket Pipeline context, reference this
command as the UX model.

## References

- [`alto-scaffold/templates/beads-ticket-template.md`](../templates/beads-ticket-template.md) — task body structure
- [`alto-scaffold/templates/beads-spike-template.md`](../templates/beads-spike-template.md) — spike body structure
- [`alto-scaffold/templates/beads-bug-template.md`](../templates/beads-bug-template.md) — bug body structure
- [`alto-scaffold/templates/beads-epic-template.md`](../templates/beads-epic-template.md) — epic body structure (Execution Waves block is appended by Phase 3 when missing)
- [`alto-scaffold/commands/launch-team.md`](launch-team.md) — execution modes (sequential default vs team opt-in); informs the self-contained-execution requirement above
- [`alto-scaffold/commands/groom.md`](groom.md) — what to run AFTER design but BEFORE claim, to re-validate against current state
