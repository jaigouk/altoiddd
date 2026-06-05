# Beads Ticket Template

Use this template for **implementation work** (follows Red/Green/Refactor). When creating a task with beads:

```bash
bd create "Task title"
# Or as a child of an epic:
bd create "Task title" --parent <epic-id>
```

---

> **Before Starting:** Always groom the ticket first. Ensure the goal is clear,
> acceptance criteria are testable, and steps are well-defined before assigning work.

> **Freshness:** If this ticket has a `review_needed` label, read the ripple comments
> (`bd comments <id>`) before starting work. Present review results to the user and
> clear the flag before claiming the ticket.

## Goal / Problem

Describe the user/system problem and the outcome needed.

## Background / Context

- Links to research, docs, or prior decisions.
- **Pattern reference:** Existing file to follow (e.g., similar module patterns)

## DDD Alignment

| Aspect | Detail |
|--------|--------|
| Bounded Context | Which context does this belong to? |
| Ubiquitous Language | Key domain terms used |
| Layer | Domain / Application / Infrastructure |

Use `/architecture-docs <topic>` to verify alignment.

## Design

### Data Models

| Model | Type | Purpose |
|-------|------|---------|
| `ModelName` | Entity / Value Object / Aggregate | Description |

### Port / Function Signatures

| Symbol | Signature | Layer |
|--------|-----------|-------|
| `PortName.method` | `method(<args>) -> <return>` (use the project's language syntax — see `<asset>.project.md` overlay for examples) | application |

### Sentinel Errors / Domain Failure Modes

| Failure | Identifier | When raised |
|---------|-----------|-------------|
| `<Name>` | `<language-specific declaration>` (e.g. Go `var ErrXxx = errors.New(...)`, Python `class XxxError(DomainError): ...`, TypeScript `class XxxError extends DomainError {}`) | invariant violation / not-found / etc. |

### Sequence / Flow

```
Component A          Component B          Component C
    |                     |                    |
    |-- message --------->|                    |
    |                     |-- action --------->|
```

## SOLID Mapping

| Principle | Implementation |
|-----------|----------------|
| **S**ingle Responsibility | One class, one job |
| **O**pen/Closed | Extend via composition/registry |
| **L**iskov Substitution | Subtypes honor contracts |
| **I**nterface Segregation | Focused Protocol with single method |
| **D**ependency Inversion | Depend on Protocol, not concrete class |

## TDD Workflow

### RED Phase

Write failing tests first. Example test signatures:

```
# <test-dir>/domain/test_feature.<ext>
test_happy_path: Description of expected behavior.
test_error_condition: Description of error handling.
```

Run: `<test-runner> <test-dir>/domain/test_feature.<ext>` → should FAIL

### GREEN Phase

1. Create `<source-dir>/domain/models/feature.<ext>` (or appropriate layer)
2. Define models
3. Implement minimal logic to pass tests

Run: `<test-runner> <test-dir>/domain/test_feature.<ext>` → should PASS

### REFACTOR Phase

- Clean up code, improve naming
- Ensure all quality gates pass
- Verify ubiquitous language matches `docs/DDD.md`

## Steps

1. Step 1 - What will be changed and why.
2. Step 2 - What will be changed and why.
3. Step 3 - What will be changed and why.

## Files in Scope

Source of truth for what this ticket owns. `/design-ticket --epic` reads
this block during wave placement; a single-agent `developer` spawn
refuses to touch any file not listed here.

| Path | Action | Owner / Notes |
|------|--------|---------------|
| `<path>` | NEW \| MODIFY \| RENAME \| DELETE | optional context (e.g. "follows pattern in `<other-path>`") |

## Prerequisites

Checked at the start of Phase 2. Missing prereq → block, do not guess.

- [ ] Required tooling / runtime installed (per `<asset>.project.md`)
- [ ] Dependency tickets closed (cross-link with `bd dep add`)
- [ ] Required environment variables / secrets present
- [ ] Test fixtures or data available

## Verification

Self-executable commands the developer runs to confirm the work is
correct. These must succeed without the developer needing to ask a
clarifying question — see "Self-contained-execution" gate in
`/design-ticket`.

```bash
<command 1>     # what it verifies
<command 2>     # what it verifies
```

## Acceptance Criteria

Each AC item must be checkable by a command (preferred) or a one-line
file:line assertion. Vague AC ("works correctly", "looks good") will
fail the self-contained-execution gate.

- [ ] Criterion 1 (testable, measurable — cite the verification command or file:line)
- [ ] Criterion 2
- [ ] Criterion 3

## Edge Cases

| Case | Input | Expected Output |
|------|-------|-----------------|
| Empty input | `""` or `None` | Return default / raise error |
| Invalid data | Malformed input | Return validation error |
| Not found | Missing resource | Return `None` or specific error |
| Duplicate | Already exists | Idempotent success or error |

## Quality Gates

Only close when all gates pass **and** edge cases are tested.

```bash
<lint-command>
<type-check-command>
<test-runner> --coverage --min-coverage=80
```

- [ ] Lint passes
- [ ] Type check passes
- [ ] All tests pass with >= 80% coverage
- [ ] Edge cases have test coverage

## Pre-Implementation Validation

Before claiming this ticket, trace the implementation end-to-end:

- [ ] Every dependency in the Design section resolves to a concrete interface (no "magic happens here")
- [ ] Port signatures in the SOLID/ISP section match the sequence diagram
- [ ] New constructor parameters won't break existing tests (or updates are listed in Steps)
- [ ] External libraries/APIs are specified (not just "does web search" — which library? which port?)

If any check fails, the ticket needs updating before work begins.

## QA Before Close

- [ ] Happy path works as expected
- [ ] Edge cases covered (see Edge Cases section)
- [ ] Error handling tested
- [ ] No regressions in existing functionality
- [ ] Domain layer has no external dependencies

## Commit Message Format

```
<type>: <description>

Types: feat / fix / test / refactor / docs / chore
```

Do **not** add AI attribution trailers to commit messages.

## Risks / Dependencies

- Risk 1
- Dependency 1

> **IMPORTANT:** Dependencies listed here are documentation only. You MUST also set
> formal dependencies with `bd dep add <this-ticket> <depends-on>` so that
> `bd blocked` / `bd ready` / ripple review can see them. Text-only deps are invisible
> to the dependency graph.
