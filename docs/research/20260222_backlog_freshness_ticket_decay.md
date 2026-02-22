---
date: 2026-02-22
topic: Backlog Freshness and Ticket Decay Prevention
status: complete
author: researcher-agent
---

# Backlog Freshness and Ticket Decay Prevention

## Research Question

When a spike or epic is completed, its findings change the context for sibling and dependent
tickets that are still open. Those tickets become stale — their descriptions, acceptance
criteria, and priorities no longer reflect current understanding. What is the minimal effective
process to prevent this ticket decay without creating bureaucratic overhead?

## Context

This research is directly relevant to two areas of vibe-seed:

1. **Ticket pipeline generation** — The PRD requires auto-generating dependency-ordered beads
   tickets from DDD artifacts (PRD section 5, P0 capability "Domain story to ticket pipeline").
   Generated tickets must remain valid as spikes complete and findings accumulate.

2. **`vs doc-health` and maintenance philosophy** — The PRD explicitly requires detecting
   staleness in docs and tickets (section 5.3). This research informs how that detection
   and refresh cycle should work.

---

## 1. Agile/Scrum: Backlog Refinement After Spikes

### What the Literature Says

Scrum defines no mandatory cadence specifically for post-spike backlog updates; it
recommends that up to 10% of sprint capacity be spent on backlog refinement, with sessions
of 1-2 hours per week for a standard team.
Source: [Scrum.org — Product Backlog Refinement](https://www.scrum.org/resources/product-backlog-refinement)

The Agile Alliance explicitly notes that the biggest misuse of spikes is treating them as a
planning mechanism rather than an uncertainty-reduction tool, and that a spike should produce
actionable insight that feeds directly into refinement — meaning the affected stories must be
updated before the sprint that follows the spike.
Source: [Agile Alliance — Sizing Spikes with Story Points](https://agilealliance.org/the-practice-of-sizing-spikes-with-story-points/)

The practical pattern that emerges from multiple sources (Mountain Goat Software, Scrum.org,
Agile Academy) is:

- A spike should not be in the same sprint as the stories it unblocks, because you don't yet
  know what those stories should say.
- On spike completion, the team holds a **targeted refinement session** (not a full backlog
  review) covering only the stories that were directly affected by the spike findings.
- New stories discovered by the spike are added; outdated stories are revised or split;
  duplicates are merged.

Source: [Mountain Goat Software — What Are Agile Spikes?](https://www.mountaingoatsoftware.com/blog/spikes)
Source: [Scrum.org — Product Backlog Refinement (2/3)](https://www.scrum.org/resources/blog/product-backlog-refinement-explained-23)

### Gap

No Scrum literature provides a formal mechanism for identifying *which* tickets are affected
by a spike completion. The identification is assumed to happen through team memory and the
product owner's judgment. There is no prescribed checklist or dependency graph traversal.

---

## 2. Shape Up: The Betting Table and Stale Pitches

### Core Mechanism

Shape Up explicitly avoids maintaining a formal backlog. Shaped pitches are stored
informally in "decentralized lists" owned by individual stakeholders. The betting table
evaluates only a small number of pitches per cycle rather than a groomed backlog.

Source: [Shape Up — The Betting Table (Chapter 8)](https://basecamp.com/shapeup/2.2-chapter-08)
Source: [Shape Up — Place Your Bets (Chapter 9)](https://basecamp.com/shapeup/2.3-chapter-09)

### How Staleness Is Handled

Shape Up's answer to ticket decay is **not to have long-lived tickets at all**:

- If context changes, shaped work is not updated in place — it is **reshaped from scratch**
  before returning to the betting table. A pitch that no longer reflects current understanding
  is discarded and a new pitch is written.
- The circuit breaker principle applies: if a project didn't finish, it means shaping was
  wrong; the response is to reframe the problem, not to extend or patch the old work.
- "Important ideas come back" is the explicit strategy: if an idea is still valuable after
  context changes, it will naturally resurface. If it doesn't resurface, it probably wasn't
  as valuable as thought.

Source: [Shape Up — Place Your Bets](https://basecamp.com/shapeup/2.3-chapter-09)

### What This Means in Practice

The Shape Up mechanism is radical simplicity: keep the backlog near-empty. Only 2-5 shaped
pitches exist at any time, so the problem of stale tickets is structurally prevented by
volume. A team of 3-8 people betting on 6-week cycles simply cannot accumulate 50 stale tickets.

### Applicability to vibe-seed

vibe-seed generates backlogs with potentially 20-50 tickets from DDD artifacts. Shape Up's
anti-backlog approach doesn't translate directly, but the principle "don't update tickets in
place when context changes — regen from current facts" is useful for the ticket pipeline
design.

---

## 3. Kanban: WIP Limits, Replenishment, and Aging Work

### WIP Limits as Staleness Prevention

Kanban's primary answer to stale tickets is structural: **WIP limits prevent accumulation**.
When a column is at its limit, nothing new enters until something exits. This prevents the
backlog from growing faster than the team can process it.

Source: [Atlassian — Working with WIP limits](https://www.atlassian.com/agile/kanban/wip-limits)
Source: [Kanban University — Glossary](https://kanban.university/glossary/)

### The Seven Cadences

Kanban defines seven cadences, two of which directly address backlog freshness:

| Cadence | Frequency | Relevance to Staleness |
|---------|-----------|----------------------|
| Kanban Meeting | Daily (15 min) | Spots blocked/aging items immediately |
| Replenishment Meeting | Weekly or on-demand | Re-evaluates backlog options before pulling |
| Delivery Planning | Per delivery | Reviews scope |
| Service Delivery Review | Bi-weekly | Reviews flow efficiency |
| Operations Review | Monthly | Reviews process health |
| Risk Review | Monthly | Identifies blocked/expired items |
| Strategy Review | Quarterly | Realigns backlog with strategy |

Source: [ZenTao — Seven Cadences of Kanban](https://www.zentao.pm/blog/seven-cadences-of-kanban-1050.html)

The **Replenishment Meeting** is the key mechanism: items in the backlog are explicitly called
"options" in Kanban, not commitments. At each replenishment, options are re-evaluated against
current capacity and context. Items that are no longer relevant are discarded without ceremony.

### Aging Work Items

Kanban tracks "work item age" as a primary flow metric. Aging items (those that have been
in the backlog for a long time without being pulled) are treated as signals of a problem:
either they are blocked, no longer relevant, or the team is overloaded.

The recommended response: set an explicit **age policy** per work item type. If a ticket
exceeds its age threshold, it must be re-evaluated before it can be pulled into active work.

Source: [Kanban Tool — WIP Limits](https://kanbantool.com/kanban-wip-limits)
Source: [ProKanban — WIP: What It Is, What It Isn't](https://www.prokanban.org/blog/wip-what-it-is-what-it-isnt-and-why-it-still-matters)

### Applicability to vibe-seed

The "age policy" concept is directly applicable. A beads ticket older than N days since its
parent spike completed could be automatically flagged as requiring re-review before work begins.

---

## 4. AI-Assisted Staleness Detection

### Current State (2025-2026)

The market is immature. Available tools fall into three categories:

**Time-based staleness only:**
- GitHub Actions `actions/stale` — closes issues inactive for N days. Does not understand
  context or dependency relationships. Widely considered harmful for open source projects
  because it silently closes valid issues.
  Source: [GitHub Marketplace — Stale for Actions](https://github.com/marketplace/actions/stale-for-actions)
  Source: [Hacker News — GitHub stale bot considered harmful](https://news.ycombinator.com/item?id=28998374)

- Linear auto-archive — archives issues inactive for 30+ days; cycle rollover automatically
  moves incomplete items back to backlog. No contextual awareness.
  Source: [Linear Docs — Cycles](https://linear.app/docs/use-cycles)
  Source: [Linear Changelog — Auto-archive](https://linear.app/changelog/2021-04-15-auto-archive-cycles-and-projects-and-deleting-issues)

**Dependency-based visualization (no auto-detection):**
- Jira Align + Easy Agile Programs — visualize blocking/blocked-by relationships. Do NOT
  automatically flag when a dependency is resolved that the blocked ticket needs re-review.
  Source: [Easy Agile — Dependency Map](https://help.easyagile.com/easy-agile-programs/dependency-map)

**AI-generated backlog items (not staleness detection):**
- Jira AI (2025) — generates new work items from Loom videos; predicts sprint success rates;
  detects duplicate issues. Does not detect context decay in existing tickets.
  Source: [Atlassian Cloud Changes Jul 2025](https://confluence.atlassian.com/cloud/blog/2025/07/atlassian-cloud-changes-jul-7-to-jul-14-2025)
  Source: [Top Jira AI Apps 2025](https://appliger.com/top-jira-ai-apps/)

- Zenhub (2025) — AI-powered backlog insights for GitHub-based teams; sprint planning
  suggestions. Does not detect stale context from completed work.
  Source: [Zenhub — 7 Best Backlog Refinement Tools 2025](https://www.zenhub.com/blog-posts/the-7-best-backlog-refinement-tools-2025)

### The Gap

**No existing tool in 2025-2026 automatically detects that completing ticket A has changed
the context for open sibling tickets B, C, and D.** All staleness detection is purely
time-based (inactivity) rather than event-based (dependency resolution). This is the core
gap this research is trying to fill.

---

## 5. Spec-Driven Development: Kiro and Spec-Kit

### Kiro's Approach

Kiro uses a three-phase spec workflow: `requirements.md` → `design.md` → `tasks.md`. Each
spec is stored in `.kiro/specs/<feature-name>/`.

Source: [Kiro Docs — Specs](https://kiro.dev/docs/specs/)

**What Kiro does to handle change propagation:**

- Manual trigger only: when requirements or design change, the user navigates to `tasks.md`
  and manually clicks "Update tasks." Kiro then regenerates the task list to match the
  updated spec.
- A spec session can also ask Kiro to "Check which tasks are already complete" to auto-mark
  progress.
- Kiro's "steering documents" (always-included context files in `.kiro/steering/`) can carry
  project-level constraints that apply to all spec interactions.

Source: [Kiro Docs — Spec Best Practices](https://kiro.dev/docs/specs/best-practices/)
Source: [Kiro Docs — Steering](https://kiro.dev/docs/steering/)

**Known limitations (2025):**

- Steering files with `inclusion: always` are sometimes ignored in favour of task execution
  instructions (GitHub issue #2250, filed Sep 2025, unresolved).
  Source: [Kiro Issue #2250](https://github.com/kirodotdev/Kiro/issues/2250)

- There is no automated notification from a completed spec phase to related open tasks in
  other specs. Cross-spec ripple is entirely manual.
  Source: [Martin Fowler — Understanding SDD: Kiro, spec-kit, and Tessl](https://martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html)

- Martin Fowler notes that Spec-Kit raises a key open question: is it "spec-anchored" (specs
  stay authoritative over time) or merely "spec-first" (specs only used to generate initial
  code)? Neither Kiro nor Spec-Kit has answered this definitively in production use.

### Spec-Kit

GitHub Spec-Kit v0.1.4 (MIT, Sep 2025) uses a Constitution + Spec + Plan + Tasks structure.
There is no documented mechanism for propagating changes across specs.
Source: [Medium — Comprehensive Guide to SDD: Kiro, GitHub Spec-Kit, and BMAD](https://medium.com/@visrow/comprehensive-guide-to-spec-driven-development-kiro-github-spec-kit-and-bmad-method-5d28ff61b9b1)

### cc-sdd (community tool)

The `gotalab/cc-sdd` GitHub project implements Kiro-style SDD commands for Claude Code,
Codex, OpenCode, Cursor, Copilot, Gemini CLI, and Windsurf. It enforces
requirements→design→tasks workflow with steering, but does not implement cross-spec
propagation.
Source: [GitHub — gotalab/cc-sdd](https://github.com/gotalab/cc-sdd)

### Summary for SDD Tools

None of Kiro, Spec-Kit, BMAD, or cc-sdd implements automatic context propagation from a
completed spec to open tasks in sibling specs. The "Update tasks" mechanism in Kiro is
**within a single spec** only.

---

## 6. Event-Driven Backlog Management

No established methodology is called "event-driven backlog management" in the project
management literature. The closest formal concepts are:

### SAFe: Inspect and Adapt

At the end of each Program Increment (PI), SAFe runs an "Inspect and Adapt" workshop. The
output is a set of improvement backlog items added to the ART backlog. This is a periodic
(PI-cadence, ~10 weeks) rather than event-driven (immediate) update.

Source: [SAFe — Inspect and Adapt](https://framework.scaledagile.com/inspect-and-adapt)

The I&A event does handle the case where completed work reveals new requirements — those are
captured as improvement items. But the latency is one full PI, not immediate.

### Jira Automation Triggers

Jira supports "When issue transitions" automation rules. A rule like "When issue X moves to
Done, assign label 'needs-review' to all issues in epic Y that are still Open" is technically
possible using Jira's automation engine (2025 feature set).

This is the closest the tooling space comes to event-driven backlog management, but:
- It requires the epic/dependency relationship to be explicitly modeled in Jira beforehand.
- It produces a label change, not an actual re-evaluation of the ticket content.
- No tool then prompts an AI or human to assess whether the ticket description is still valid.

Source: [Atlassian — How to automatically transition parent issue based on sub-task status](https://support.atlassian.com/jira/kb/how-to-automatically-transition-the-parent-issue-based-on-the-sub-task-status/)

### The Theoretical Model

The concept most aligned with the research question is **change impact analysis (CIA)** from
software safety literature. A 2014 paper in Springer Nature describes CIA in agile contexts:
systematic tracing of requirement changes to identify all affected artifacts (code, tests,
documentation, and other requirements).

The paper's key finding: without explicit dependency tracing (a graph of what affects what),
CIA is conducted by "team memory" — which is unreliable, especially as the team grows or
knowledge is distributed across agents.

Source: [Springer — Agile Change Impact Analysis of Safety Critical Software](https://link.springer.com/chapter/10.1007/978-3-319-10557-4_48)
Source: [Capital One — 5 Questions to Help Guide Impact Analysis](https://www.capitalone.com/tech/software-engineering/best-practices-for-impact-analysis-and-effective-backlog-refinement/)

---

## 7. The "Context Rot" Problem — Named and Quantified

The phenomenon the research question describes has a name: **context rot** (analogous to
bit rot in software, but applied to the surrounding documentation and ticket context).

Key observations from the literature:

1. **Context rot is accelerating.** AI systems operationalize stale context — if documented
   context is outdated, AI agents will build on incorrect assumptions and compound the error.
   Source: [Medium — Context Engineering: Mitigating Context Rot in AI Systems](https://medium.com/ai-pace/context-engineering-mitigating-context-rot-in-ai-systems-21eb2c43dd18)

2. **Freshness is measurable.** "Backlog freshness" can be defined as the percentage of
   backlog items created or reviewed within a recency window (typically 30-90 days). Teams
   with high freshness scores tend to deliver more predictably.
   Source: [Medium — All You Need Is a Fresh Backlog](https://medium.com/agileinsider/all-you-need-is-a-fresh-backlog-e3e5bad717a7)

3. **Just-in-time detail is the key pattern.** Writing detailed descriptions for work not
   addressed for weeks leads to stale tickets and wasted effort. The lean practice is: keep
   early-stage items intentionally lightweight and add detail only as they rise in priority
   and are about to be worked on.
   Source: [Easy Agile — Essential Checklist for Effective Backlog Refinement](https://www.easyagile.com/blog/backlog-refinement)

4. **Oral decisions decay faster than written ones.** If scope, priority, or dependencies
   change, the update must be documented immediately in the ticket or linked documents.
   Source: [Plane.so — Backlog Grooming Best Practices](https://plane.so/blog/backlog-grooming-best-practices-for-agile-teams)

---

## 8. Synthesis: The Minimal Effective Process

Combining findings from all seven areas, the minimal effective process to prevent ticket
decay without bureaucratic overhead consists of four principles:

### Principle 1: Deferred Detail (Lean/Kanban)

Do not write detailed acceptance criteria for tickets more than 1-2 sprints away from
implementation. Tickets far from the top of the priority queue should carry only a title,
a one-line summary, and a set of tagged dependencies — no full specification.

Detailed specification is written (or regenerated) at the **last responsible moment**:
just before the ticket is pulled into active work.

This eliminates the decay problem for low-priority tickets because there is nothing to
decay yet.

### Principle 2: Event-Triggered Ripple Review (Novel Pattern)

When a ticket moves to Done, the system should:

1. Identify all open tickets that have a `depends-on` or `informed-by` relationship to
   the completed ticket.
2. Flag those tickets with a `needs-review` marker (or equivalent in beads).
3. Optionally, prompt the agent responsible for the dependent ticket to re-read the
   completed ticket's output and confirm whether the dependent ticket still accurately
   describes the work.

This is the event-driven approach. It has no established name in the literature but is
the logical extension of Jira's automation triggers to the full ticket lifecycle.

The key difference from existing tools: the trigger is the **semantic completion of a
dependency**, not time-based inactivity.

### Principle 3: Lightweight Freshness Metadata

Each ticket carries a `last_reviewed` date (analogous to the `last_reviewed` field in
vibe-seed's doc registry). When a spike or epic completes, all open sibling tickets in
the same epic receive a `review_needed: true` flag.

A `vs ticket-health` command (or integration with `vs doc-health`) surfaces tickets
where:
- `review_needed: true` (completion event triggered a review flag)
- `last_reviewed` is older than the `created_at` date of a dependency's completion
- The ticket's parent epic has had a completion event since `last_reviewed`

This is measurable and automatable without AI assistance.

### Principle 4: Shape Up's "No Backlog" Safety Valve

For tickets that cannot be refined (because insufficient information exists), the Shape Up
approach applies: keep the ticket as a **stub** (title + one sentence) rather than deleting
it. Stubs are cheap to maintain. When the context arrives, the stub is reshaped into a
full ticket just before implementation.

This prevents premature deletion of potentially valuable work while avoiding the cost of
maintaining stale specifications.

---

## 9. Recommended Pattern for vibe-seed

Based on the research, the recommended minimal process for vibe-seed's ticket lifecycle is:

### At Ticket Creation (domain story pipeline)

- Tickets are generated in two tiers:
  - **Near-term** (next 1-2 epics): full specification with acceptance criteria, TDD phases,
    SOLID mapping
  - **Far-term** (later epics): stub tickets — title, one-sentence summary, epic link,
    dependencies list, no detailed AC

### At Spike/Epic Completion

When a beads ticket closes, the `bd close` command (or a `vs` hook on ticket closure) should:

1. Query all open tickets with `depends_on` or `related_to` the closed ticket ID.
2. Set `review_needed: true` on each.
3. Print a summary: "Spike k7m.5 closed. 3 dependent tickets need review: k7m.8, k7m.9, k7m.12."

This is a 10-line shell or Python script. No AI required.

### At Ticket Pick-Up (before claiming in_progress)

The CLAUDE.md "Ticket Grooming Checklist" (already exists in the project) is extended
with one step:

```
5. **Freshness Check** — Is this ticket's context still valid?
   - Has any dependency been completed since this ticket was last reviewed?
   - If `review_needed: true`, re-read the completed dependency's output first.
   - Update AC, description, or estimates if the spike findings changed scope.
   - Set `last_reviewed` to today and clear `review_needed` flag.
```

### Cadence (Kanban-inspired)

- **On-demand**: Freshness check at ticket pick-up (before claiming in_progress). Zero overhead.
- **Weekly**: `vs ticket-health` run shows count of `review_needed` tickets and oldest
  `last_reviewed` dates. No action required unless count > 3 or oldest > 14 days.
- **Per epic completion**: Targeted refinement session for the N tickets flagged by the
  completion event. Should take 15-30 minutes maximum.

---

## 10. What No Tool Does Today (Gap Analysis)

| Capability | Jira | Linear | Kiro | Shape Up | vibe-seed (proposed) |
|-----------|------|--------|------|----------|----------------------|
| Time-based staleness detection | Yes (activity) | Yes (activity) | No | N/A | Partial (`last_reviewed`) |
| Event-based staleness (dependency completion) | Manual (automation rules) | No | No (within spec only) | No | Yes (proposed ripple review) |
| Freshness metadata per ticket | No | No | No | No | Yes (`review_needed` flag) |
| Deferred detail / stub tickets | Manual | Manual | Not designed for | Yes (natural) | Yes (two-tier generation) |
| Re-generation of ticket spec from updated context | No | No | Yes (within spec, manual trigger) | Reshape from scratch | Yes (vs ticket-refresh proposed) |
| Cross-epic ripple propagation | No | No | No | No | Yes (proposed) |

---

## 11. Recommendation

The minimal effective process is:

1. **Generate far-term tickets as stubs** — defer detailed AC until near-term.
2. **On every ticket close, flag dependents for review** — a simple dependency graph
   traversal and flag-setting operation, no AI required.
3. **Make freshness a first-class field** — `review_needed` and `last_reviewed` on every
   ticket.
4. **Embed the freshness check in the existing grooming checklist** — not a new process,
   just one additional step before claiming a ticket as in_progress.

This costs approximately 1-2 minutes per ticket pick-up and 15-30 minutes per epic
completion. For a team of one or a small AI agent workflow, this is sustainable.

The alternative — ignoring context decay — produces the "80-100% anti-pattern rate"
documented in AI-assisted codebases (Ox Security, 2025), because AI agents will faithfully
implement stale specifications.

---

## Sources

- [Scrum.org — Product Backlog Refinement](https://www.scrum.org/resources/product-backlog-refinement)
- [Scrum.org — Product Backlog Refinement (2/3)](https://www.scrum.org/resources/blog/product-backlog-refinement-explained-23)
- [Agile Alliance — Sizing Spikes with Story Points](https://agilealliance.org/the-practice-of-sizing-spikes-with-story-points/)
- [Mountain Goat Software — What Are Agile Spikes?](https://www.mountaingoatsoftware.com/blog/spikes)
- [Shape Up — The Betting Table (Chapter 8)](https://basecamp.com/shapeup/2.2-chapter-08)
- [Shape Up — Place Your Bets (Chapter 9)](https://basecamp.com/shapeup/2.3-chapter-09)
- [Atlassian — Working with WIP limits for kanban](https://www.atlassian.com/agile/kanban/wip-limits)
- [Kanban University — The Official Guide to The Kanban Method](https://kanban.university/kanban-guide/)
- [ZenTao — Seven Cadences of Kanban](https://www.zentao.pm/blog/seven-cadences-of-kanban-1050.html)
- [ProKanban — WIP: What It Is, What It Isn't](https://www.prokanban.org/blog/wip-what-it-is-what-it-isnt-and-why-it-still-matters)
- [Kanban Tool — WIP Limits](https://kanbantool.com/kanban-wip-limits)
- [GitHub Marketplace — Stale for Actions](https://github.com/marketplace/actions/stale-for-actions)
- [Hacker News — GitHub stale bot considered harmful](https://news.ycombinator.com/item?id=28998374)
- [Linear Docs — Cycles](https://linear.app/docs/use-cycles)
- [Linear Docs — Project Dependencies](https://linear.app/docs/project-dependencies)
- [Linear Changelog — Auto-archive](https://linear.app/changelog/2021-04-15-auto-archive-cycles-and-projects-and-deleting-issues)
- [Easy Agile — Dependency Map](https://help.easyagile.com/easy-agile-programs/dependency-map)
- [Atlassian Cloud Changes Jul 2025](https://confluence.atlassian.com/cloud/blog/2025/07/atlassian-cloud-changes-jul-7-to-jul-14-2025)
- [Zenhub — 7 Best Backlog Refinement Tools 2025](https://www.zenhub.com/blog-posts/the-7-best-backlog-refinement-tools-2025)
- [Kiro Docs — Specs](https://kiro.dev/docs/specs/)
- [Kiro Docs — Spec Best Practices](https://kiro.dev/docs/specs/best-practices/)
- [Kiro Docs — Steering](https://kiro.dev/docs/steering/)
- [Kiro Issue #2250 — Steering files with inclusion: always ignored](https://github.com/kirodotdev/Kiro/issues/2250)
- [Martin Fowler — Understanding SDD: Kiro, spec-kit, and Tessl](https://martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html)
- [GitHub — gotalab/cc-sdd](https://github.com/gotalab/cc-sdd)
- [SAFe — Inspect and Adapt](https://framework.scaledagile.com/inspect-and-adapt)
- [SAFe — Team Backlog](https://framework.scaledagile.com/team-backlog)
- [Atlassian — Automatically transition parent issue based on sub-task status](https://support.atlassian.com/jira/kb/how-to-automatically-transition-the-parent-issue-based-on-the-sub-task-status/)
- [Springer — Agile Change Impact Analysis of Safety Critical Software](https://link.springer.com/chapter/10.1007/978-3-319-10557-4_48)
- [Capital One — 5 Questions to Help Guide Impact Analysis](https://www.capitalone.com/tech/software-engineering/best-practices-for-impact-analysis-and-effective-backlog-refinement/)
- [Medium — Context Engineering: Mitigating Context Rot in AI Systems](https://medium.com/ai-pace/context-engineering-mitigating-context-rot-in-ai-systems-21eb2c43dd18)
- [Medium — All You Need Is a Fresh Backlog](https://medium.com/agileinsider/all-you-need-is-a-fresh-backlog-e3e5bad717a7)
- [Easy Agile — Essential Checklist for Effective Backlog Refinement](https://www.easyagile.com/blog/backlog-refinement)
- [Plane.so — Backlog Grooming Best Practices](https://plane.so/blog/backlog-grooming-best-practices-for-agile-teams)
- [Medium — Comprehensive Guide to SDD: Kiro, GitHub Spec-Kit, and BMAD](https://medium.com/@visrow/comprehensive-guide-to-spec-driven-development-kiro-github-spec-kit-and-bmad-method-5d28ff61b9b1)
