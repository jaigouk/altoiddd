# Beads Epic Template

Use this template when creating an epic with beads. Epics are large bodies of
work broken into child tickets that run in **execution waves**. Typical epic
duration: 1-3 months. Longer than that → split into sequential epics with
milestones.

## Command

```bash
bd create --title="Epic: <Title>" --type=epic --priority=<0-4> --description="<paste body below>"
```

---

> **Before Starting:** Always groom the epic first. Ensure the goal is clear,
> success metrics are measurable, scope is well-defined, and child tasks are
> grouped into waves before any of them is claimed.

> **Waves are the source of truth.** Once this epic has children, the "Child
> Tasks" table and "Execution Waves" block tell every agent which tickets can
> run in parallel and which must wait. Use [`/design-ticket --epic=<id>`](../commands/design-ticket.md)
> when adding children — it places them in the right wave and updates these
> sections automatically.

## Template

```markdown
<Brief 1-2 sentence summary of what this epic delivers.>

## Business Value

As a <user type>, I want <capability> so that <business outcome>.

## DDD Alignment

Which bounded context(s) does this epic affect? Reference `docs/DDD.md`:

| Bounded Context | Impact |
|-----------------|--------|
| [Context name]  | What changes |

Verify with `/architecture-docs` commands:

| Check        | Command                          |
|--------------|----------------------------------|
| Domain model | `/architecture-docs domain`      |
| Architecture | `/architecture-docs components`  |

## Scope

### In Scope
- Item 1
- Item 2

### Out of Scope
- Item 1
- Item 2

## Child Tasks

<!-- Wave column groups tickets that can run in parallel. See "Execution
     Waves" below for the rules. Keep this table and the Execution Waves
     block in lockstep — they MUST agree. -->

| Wave | ID            | Task              | Priority | Status | Blocked By              |
|------|---------------|-------------------|----------|--------|-------------------------|
| 1    | <prefix>-001  | Foundation task   | P1       | Open   | —                       |
| 2    | <prefix>-002  | Track A           | P2       | Open   | <prefix>-001            |
| 2    | <prefix>-003  | Track B           | P2       | Open   | <prefix>-001            |
| 3    | <prefix>-004  | Integration       | P2       | Open   | <prefix>-002, <prefix>-003 |

## Execution Waves

<!-- Group children into waves so parallel work is obvious. Tickets in the
     same wave can be picked up by different agents or humans without
     coordination beyond what's documented here. -->

```
                       time →

   ┌─────────┐
   │ Wave 0  │  (optional) siblings that should ship before the epic
   │ (opt)   │  but aren't children (e.g. a doc cleanup that improves
   │         │  later analysis quality)
   └────┬────┘
        │
   ┌────▼────┐
   │ Wave 1  │  <prefix>-001 — foundation; everything blocks on this
   │ ~Xd     │
   └────┬────┘
        │
        ├────────────────────┐
        ▼                    ▼
   ┌─────────┐         ┌─────────┐
   │ Wave 2a │         │ Wave 2b │   PARALLEL — same dep set, disjoint files
   │ ~Xd     │         │ ~Xd     │
   │ <p>-002 │         │ <p>-003 │
   └────┬────┘         └────┬────┘
        └─────────┬─────────┘
                  ▼
             ┌─────────┐
             │ Wave 3  │  <prefix>-004 — integration; needs Wave 2 outputs
             │ ~Xd     │
             └─────────┘
```

### Wave Rules

> Kept in sync with [`/design-ticket --epic=<id>`](../commands/design-ticket.md) Phase 3.

Two tickets belong in the **same wave** iff ALL true:

1. **Same dependency set** — both depend on the same in-epic parent ticket(s),
   OR both have no in-epic deps (both Wave 1).
2. **Disjoint file scope** — they don't both create or modify the same files.
   (Check each child's `Files in Scope` block.)
3. **No semantic ordering** — neither produces an artefact the other needs
   before starting.
4. **Parallel-session-safe** — could be picked up by two different agents (or
   two different humans) without coordination beyond what's documented here.

**Wave order**:

- Wave 1 = children with no in-epic deps.
- Wave N+1 = depends on at least one ticket in Wave N.
- A ticket's wave = `max(wave of each of its deps within the epic) + 1`.
- All Wave N tickets must close before any Wave N+1 ticket starts.
- Within a wave, tickets run in parallel.

**Wave 0** (optional): siblings NOT under the epic that should ship before
any in-epic work. Document why in this section, not as a child.

### Why this wave layout?

<!-- One short paragraph per wave explaining the reasoning. New readers
     should understand the parallelism story without reading every child. -->

- **Wave 1**: <why this is the foundation everything blocks on>
- **Wave 2a / 2b**: <why these are parallel-safe — usually "touch disjoint paths">
- **Wave 3**: <why this integrates Wave 2 outputs>

### File-coordination risks

<!-- If parallel siblings share a single coordination point, document it
     here AND mirror it in the Risks & Mitigations table below. Common
     pattern: one ticket ships with a config line commented out, the
     sibling uncomments it during its smoke test. -->

- **<prefix>-002 ↔ <prefix>-003**: <single shared file and how it's resolved>

## Dependency Graph

<!-- ASCII showing dep edges. Same data as the Child Tasks table but easier
     to scan when there are many edges. -->

```
        Wave 1                Wave 2 (parallel)         Wave 3
        ┌──────────┐          ┌──────────┐
        │ <p>-001  │ ──┬───→  │ <p>-002  │ ──┐
        └──────────┘   │      └──────────┘   │      ┌──────────┐
                       │                      ├───→ │ <p>-004  │
                       │      ┌──────────┐    │     └──────────┘
                       └───→  │ <p>-003  │ ──┘
                              └──────────┘

       (sibling, ship anytime — improves later analysis)
       <sib>-001
```

## Architecture

<!-- Optional: Show system architecture changes. Reference docs/ARCHITECTURE.md
     for the full picture; this diagram only shows what the epic changes. -->

```
[Component A]
    │
    ├── Service 1 ──→ Database
    │
    └── Service 2 ──→ Storage
```

## Success Metrics

<!-- How do we measure success? Be specific and measurable. -->

- Metric 1: <target value>
- Metric 2: <target value>

## Acceptance Criteria (Epic Level)

<!-- Business outcomes, not implementation details. -->

- [ ] Users can <do something>
- [ ] System achieves <performance target>
- [ ] <Business KPI> improves by <X%>
- [ ] All child tickets closed (`bd close <id>`) — each child must have passed
      quality gates and QA before close
- [ ] For every code-changing child, quality gates ran before close:
  - [ ] `<lint-command>` (linting)
  - [ ] `<type-check-command>` (type checking)
  - [ ] `<test-runner> --coverage --min-coverage=80` (tests with ≥ 80% coverage)
- [ ] Documentation updated where required

## Timeline

- **Start**:              YYYY-MM-DD
- **Target Completion**:  YYYY-MM-DD
- **Duration**:           X weeks

## Risks & Mitigations

| Risk   | Impact | Likelihood | Mitigation       |
|--------|--------|------------|------------------|
| Risk 1 | High   | Medium     | Mitigation plan  |

<!-- Mirror every entry from "File-coordination risks" above into this
     table so risk reviews don't miss them. -->

## References

- Link 1 — `<url>` — Description
- Link 2 — `<url>` — Description
```

---

## Best Practices

### Epic Duration

- Keep epics to 1-3 months maximum.
- If longer, split into sequential epics with milestones.

### Acceptance Criteria

- Focus on **business outcomes**, not features.
- Use measurable targets (99.9% uptime, < 200 ms response).
- Leave implementation details to child tasks.

### Child Tasks

- Break down by **vertical slices** (end-to-end functionality).
- Each task should be completable in 1-5 days.
- Define clear dependencies between tasks.
- Assign each task a **wave number** in the Child Tasks table.

### Execution Waves

- Define waves so parallel work is obvious to anyone picking up the epic.
- Same wave ⇒ same dep set + disjoint file scope + no semantic ordering.
- Document **why** each wave is grouped that way (one sentence per wave).
- Flag any single-file coordination points between parallel siblings —
  these go in both the "File-coordination risks" subsection AND the
  Risks & Mitigations table.
- Use [`/design-ticket --epic=<id>`](../commands/design-ticket.md) when
  adding new children — it places them in the right wave and updates
  these sections automatically.

### Prioritization

| Priority | Meaning                          |
|----------|----------------------------------|
| P0       | Critical blocker                 |
| P1       | High priority, needed soon       |
| P2       | Medium priority, standard work   |
| P3       | Low priority, nice to have       |
| P4       | Backlog, someday/maybe           |

---

## Example: Infrastructure Epic

```markdown
Deploy new monitoring stack to improve observability across the k3s cluster.

## Business Value

As an SRE, I want centralized monitoring and alerting so that I can detect
and resolve issues before they impact users.

## DDD Alignment

| Bounded Context | Impact                                       |
|-----------------|----------------------------------------------|
| Observability   | New Prometheus + Grafana adapters and dashboards |

## Scope

### In Scope
- Deploy Prometheus + Grafana
- Configure alerting rules
- Create dashboards for key services

### Out of Scope
- Log aggregation (separate epic)
- Distributed tracing (future work)

## Child Tasks

| Wave | ID      | Task               | Priority | Status | Blocked By |
|------|---------|--------------------|----------|--------|------------|
| 1    | mon-001 | Deploy Prometheus  | P1       | Open   | —          |
| 2    | mon-002 | Deploy Grafana     | P2       | Open   | mon-001    |
| 2    | mon-003 | Configure alerts   | P2       | Open   | mon-001    |
| 3    | mon-004 | Create dashboards  | P3       | Open   | mon-002    |

## Execution Waves

```
   ┌─────────┐
   │ Wave 1  │  mon-001 — Prometheus (data source for everything)
   │ ~2 d    │
   └────┬────┘
        │
        ├──────────────────┐
        ▼                  ▼
   ┌─────────┐        ┌─────────┐
   │ Wave 2a │ PARALL │ Wave 2b │
   │ ~3 d    │        │ ~2 d    │
   │ mon-002 │        │ mon-003 │
   │ Grafana │        │ Alerts  │
   └────┬────┘        └─────────┘
        │
        ▼
   ┌─────────┐
   │ Wave 3  │  mon-004 — Dashboards built on Grafana
   │ ~3 d    │
   └─────────┘
```

### Why this wave layout?

- **Wave 1**: Prometheus must exist before Grafana or alert rules can query it.
- **Wave 2a / 2b**: Grafana and alert rules touch disjoint files
  (`grafana/values.yaml` vs `prometheus-rules/*.yaml`) and don't reference
  each other — clean parallel.
- **Wave 3**: Dashboards need Grafana running.

## Success Metrics

- Alert latency < 30 seconds
- Dashboard load time < 2 seconds
- 100% of critical services monitored

## Acceptance Criteria (Epic Level)

- [ ] All k3s nodes have metrics scraped
- [ ] Alerts fire for CPU > 80%, Memory > 90%
- [ ] On-call can diagnose issues using dashboards alone
- [ ] No false-positive alerts for 7 consecutive days
- [ ] All child tickets closed with passing quality gates
```

---

## Related

- [`alto-scaffold/commands/design-ticket.md`](../commands/design-ticket.md) — places new children into the right wave and keeps this template's Child Tasks + Execution Waves blocks in sync
- [`alto-scaffold/templates/beads-ticket-template.md`](beads-ticket-template.md) — task body structure (must include a `Files in Scope` block consumed by wave assignment)
- [`alto-scaffold/templates/beads-spike-template.md`](beads-spike-template.md) — spike body structure
- [`alto-scaffold/templates/beads-bug-template.md`](beads-bug-template.md) — bug body structure

## Sources

- [How to Write an Epic in Agile](https://www.parallelhq.com/blog/how-to-write-epic-in-agile)
- [Agile Epics Definitive Guide](https://monday.com/blog/rnd/agile-epics/)
- [SAFe Epic Best Practices](https://agileseekers.com/blog/how-to-write-effective-safe-epics-format-criteria-best-practices)
- [Epics and Features](https://www.harness.io/harness-devops-academy/epics-and-features-the-cornerstone-of-agile-success)
