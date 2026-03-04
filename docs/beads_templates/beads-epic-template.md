# Beads Epic Template

Use this template when creating an epic with beads:

```bash
bd create "Epic: <Title>" -p 0
```

---

> **Before Starting:** Always groom the epic first. Ensure the goal is clear,
> success metrics are measurable, scope is well-defined, and child tasks are planned.

## Goal / Problem

High-level problem statement and desired outcome.

## DDD Alignment

Which bounded context(s) does this epic affect? Reference `docs/DDD.md`:

| Bounded Context | Impact |
|----------------|--------|
| [Context name] | What changes |

Verify with `/architecture-docs` commands:

| Check | Command |
|-------|---------|
| Domain model | `/architecture-docs domain` |
| Architecture | `/architecture-docs components` |

## Success Metrics

- Metric 1 (measurable)
- Metric 2

## Scope

**In scope**

- Item 1
- Item 2

**Out of scope**

- Item 1

## Phases / Milestones

1. Phase 1 - Research / design
2. Phase 2 - Implementation
3. Phase 3 - Validation / rollout

## Child Tasks (planned)

Create child tasks under this epic:

```bash
bd create "Task title" --parent <epic-id>
```

## Dependencies

- Dependency 1
- Dependency 2

## Risks / Unknowns

- Risk 1
- Unknown 1

## Acceptance Criteria

- [ ] All child tasks closed (`bd close <id>`) — each child must have passed quality gates and QA before close
- [ ] Documentation updated where required
- [ ] For all code changes (in child tasks), quality gates were run before each task was closed:
  - [ ] `<lint-command>` (linting)
  - [ ] `<type-check-command>` (type checking)
  - [ ] `<test-runner> --coverage --min-coverage=80` (tests with 80% coverage)
- [ ] QA was done for each child task before close (happy path, edge cases, error handling, no regressions)
