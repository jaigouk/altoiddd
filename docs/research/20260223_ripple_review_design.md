---
date: 2026-02-23
topic: Ticket Freshness and Ripple Review Design
status: complete
type: spike
ticket: vibe-seed-k7m.12
---

# Ticket Freshness and Ripple Review Design

> **Spike:** k7m.12 -- Design the ripple review system for event-driven ticket freshness
> **Timebox:** 4 hours
> **Decision:** Approach 2 -- Ripple Flag + Context Diff with human-in-the-loop (approved)

## Summary

This document provides the complete design for vibe-seed's ticket freshness system: the
data model, the close-time hook, the pick-up-time flow, the `vs ticket-health` command,
two-tier ticket generation rules, and the after-close protocol that vibe-seed generates
into every bootstrapped project's CLAUDE.md.

The design uses only existing beads capabilities (labels, comments, dependencies) and
requires no schema changes to beads. The existing `bin/bd-ripple` script is the
foundation; this design formalizes and extends it.

---

## 1. Ripple Review Data Model

### 1.1 Fields and Storage

Beads does not support custom fields per ticket. All freshness metadata is stored using
beads' native features: **labels** and **comments**.

| Concept | Storage Mechanism | Format |
|---------|------------------|--------|
| review_needed flag | `bd label add <id> review_needed` | Label on ticket |
| Triggering ticket ID | Embedded in comment text | `**Triggered by:** \`<closed-id>\`` |
| Context diff | Comment body | Markdown text (see section 1.3) |
| last_reviewed | Comment with specific prefix | `**Reviewed:** <ISO-date> by <actor>` |
| Flag stacking | Multiple comments, one per trigger | Each closure adds a new comment |

### 1.2 Why Not Custom Fields

Beads v0.55.4 has a fixed schema (title, description, status, priority, type, labels,
dependencies, comments, close_reason, etc.). Adding custom fields would require upstream
changes or a sidecar file. Using labels + comments works today with no beads changes.

The tradeoff: `last_reviewed` is stored as a comment (searchable via `bd list
--desc-contains` on comments) rather than as a sortable date field. The `vs ticket-health`
command will parse comments to extract dates. This is acceptable for the expected scale
(tens of tickets, not thousands).

### 1.3 Context Diff Comment Format

When `bd-ripple` flags a ticket, the comment MUST follow this exact format:

```markdown
**Ripple review needed** -- `<closed-ticket-id>` (<closed-ticket-title>) was closed.

**What changed:** <context-summary>

**Review checklist:**
- [ ] Description still accurate?
- [ ] Acceptance criteria still valid?
- [ ] Dependencies still correct?
- [ ] Estimates still realistic?

Review this ticket against the new context. Update if needed, or dismiss if still valid.
```

The `<context-summary>` comes from one of two sources (in priority order):
1. The second argument to `bd-ripple`: `bin/bd-ripple <id> "summary text"`
2. The closed ticket's `close_reason` field: `bd close <id> --reason "what was produced"`

### 1.4 Flag Cleared Comment Format

When a flag is cleared, a review comment MUST be added:

```markdown
**Reviewed:** 2026-02-23 by <actor>
**Triggered by:** `<closed-ticket-id>`
**Verdict:** <unchanged|updated|dismissed>
**Changes:** <one-line summary of what was updated, or "No changes needed">
```

This creates an audit trail. The `vs ticket-health` command parses `**Reviewed:**` lines
to determine `last_reviewed` timestamps.

### 1.5 Domain Model Alignment

From `docs/DDD.md` section 5, the RippleReview aggregate:

| DDD Concept | Implementation |
|------------|----------------|
| `RippleReview` (Aggregate) | One `bd-ripple` invocation = one ripple review |
| `ContextDiff` (Value Object) | The `**What changed:**` section of the comment |
| `FreshnessFlag` (Value Object) | The `review_needed` label + the comment |
| `TicketClosed` (Domain Event) | `bd close <id>` -- triggers the ripple |
| `TicketFlagged` (Domain Event) | `bd label add <id> review_needed` |
| `FlagCleared` (Domain Event) | `bd label remove <id> review_needed` + review comment |

### 1.6 Invariants (from DDD.md, verified)

1. **Non-empty context diff** -- `bd-ripple` MUST have a context summary (from argument or
   close_reason). If both are empty, the script MUST prompt for one or abort.
   *Current status: partially implemented -- falls back to "Ticket closed (no close reason
   provided)" which violates this invariant.*

2. **Only open tickets flagged** -- `bd-ripple` only flags tickets with status `open` or
   `in_progress`. Closed tickets are skipped.
   *Current status: implemented in `bin/bd-ripple` lines 79, 93.*

3. **Flag stacking** -- A ticket can accumulate multiple `review_needed` flags from
   different closures. Each closure adds a new comment. The label is idempotent (adding
   it twice is a no-op).
   *Current status: implemented in `bin/bd-ripple` lines 126-135.*

4. **No auto-clear** -- Clearing `review_needed` requires explicit human review. No agent
   or script may auto-clear without presenting the review to the user first.
   *Current status: enforced by CLAUDE.md grooming checklist step 2.*

---

## 2. Close-Time Hook Design

### 2.1 Trigger

The ripple review is triggered by running `bin/bd-ripple <closed-ticket-id>` after
`bd close <id>`. This is a manual step in the after-close protocol, not a beads hook.

**Why not a beads post-close hook:** Beads hooks are git hooks (pre-commit, post-merge,
etc.), not lifecycle hooks on ticket status changes. There is no `bd close` hook mechanism
in beads v0.55.4. The after-close protocol in CLAUDE.md is the enforcement mechanism --
agents run it automatically after every `bd close`.

### 2.2 Dependency Graph Traversal

The current `bin/bd-ripple` script traverses two relationship types:

1. **Siblings** -- Children of the same parent epic (via `bd children <parent-id>`)
2. **Dependents** -- Tickets that have a `blocks` dependency on the closed ticket
   (via the `dependents` field in `bd show <id> --json`)

This covers the two most common staleness vectors:
- Siblings: closing k7m.9 (killer features) affects k7m.2, k7m.3, k7m.4 (sibling spikes)
- Dependents: closing a spike unblocks implementation tickets whose specs may have changed

### 2.3 Missing Relationship Types

The current script does NOT traverse:
- `related` dependencies (soft relationships)
- `discovered-from` dependencies
- Cross-epic dependencies (ticket in epic A depends on ticket in epic B)

**Design decision:** Add `related` to the traversal. The `related` type is explicitly
informational, but "informational" means "context may have changed." `discovered-from` is
backward-looking (the discovered ticket was created from the source) and does not indicate
staleness flow. Cross-epic dependencies are already captured by `blocks` if properly set.

### 2.4 Updated Traversal Algorithm

```
Input: closed_ticket_id, context_summary
Output: list of flagged ticket IDs

1. ticket = bd show closed_ticket_id --json
2. Validate: ticket.status == "closed" (warn if not)
3. Validate: context_summary is non-empty (abort if empty)

4. candidates = {}

5. # Siblings (children of same parent)
   parent_id = find parent-child dependency in ticket.dependencies
   if parent_id:
     for child in bd children parent_id:
       if child.id != closed_ticket_id and child.status in (open, in_progress):
         candidates[child.id] = child

6. # Dependents (tickets this one blocks)
   for dep in ticket.dependents:
     if dep.type == "blocks" and dep.status in (open, in_progress):
       candidates[dep.id] = dep

7. # Related tickets (soft relationships -- both directions)
   for dep in ticket.dependencies:
     if dep.type == "related" and dep.status in (open, in_progress):
       candidates[dep.id] = dep
   for dep in ticket.dependents:
     if dep.type == "related" and dep.status in (open, in_progress):
       candidates[dep.id] = dep

8. For each candidate:
     - Add review_needed label (idempotent)
     - Add context diff comment (always, even if already flagged -- stacking)
     - Print summary line

9. Return list of flagged IDs
```

### 2.5 Changes Required to `bin/bd-ripple`

| Change | Current Behavior | Required Behavior |
|--------|-----------------|-------------------|
| Empty context guard | Falls back to "no close reason provided" | Abort with error message: "Context summary required. Use: `bd-ripple <id> 'summary'` or close with `bd close <id> --reason 'summary'`" |
| Related traversal | Not traversed | Add `related` deps (both directions) |
| Comment format | Basic markdown | Structured format with review checklist (section 1.3) |
| Exit code | Always 0 | Exit 1 on error, 0 on success |
| Machine-readable output | None | Add `--json` flag for `vs ticket-health` integration |

---

## 3. Pick-Up-Time Flow Design

### 3.1 When an Agent Picks Up a Flagged Ticket

The grooming checklist in CLAUDE.md (step 2: Freshness Check) defines the agent behavior.
This section specifies the exact steps the agent MUST follow.

```
Agent claims ticket <id> for work:

1. Run: bd label list <id>
2. If "review_needed" is present:
   a. Run: bd comments <id>
   b. Find all comments matching "**Ripple review needed**"
   c. For each ripple comment:
      - Read the "What changed" section
      - Read the triggering ticket's close_reason if more context needed
   d. Compare ticket description + AC against the new context
   e. Draft suggested updates (may be "no changes needed")
   f. Present to user:
      ---
      ## Freshness Review Required

      This ticket was flagged for review because:
      - `k7m.9` (Competitive analysis) was closed
        What changed: "Research found 4+1 feature strategy; P0 features changed"

      ### Suggested Updates
      - [ ] Update AC #3 to reflect new P0 feature list
      - [ ] Add edge case for new "rescue mode" scenario
      - OR: No changes needed -- context does not affect this ticket

      **Approve updates? (y/n/edit)**
      ---
   g. User approves, edits, or dismisses
   h. If approved: apply updates via bd update <id> --description/--acceptance
   i. Clear flag: bd label remove <id> review_needed
   j. Add review comment (section 1.4 format)
3. If "review_needed" is NOT present:
   - Proceed with normal grooming checklist (step 3: PRD traceability, etc.)
```

### 3.2 Multiple Stacked Flags

When a ticket has multiple ripple comments (flagged by multiple closures), the agent MUST
review ALL of them in a single review session. The review comment should reference all
triggering tickets:

```markdown
**Reviewed:** 2026-02-23 by developer-agent
**Triggered by:** `k7m.9`, `k7m.2`
**Verdict:** updated
**Changes:** Updated AC to reflect new feature list (k7m.9) and new question framework (k7m.2)
```

### 3.3 Dismissal

If the agent determines no changes are needed, it still clears the flag and adds a
review comment:

```markdown
**Reviewed:** 2026-02-23 by developer-agent
**Triggered by:** `k7m.9`
**Verdict:** unchanged
**Changes:** No changes needed -- competitive analysis findings do not affect this ticket's scope
```

This ensures the audit trail is complete even when no edits are made.

---

## 4. `vs ticket-health` Command Design

### 4.1 Purpose

Read-only report showing the freshness state of the project's ticket backlog. No writes,
no mutations. Maps to the `TicketHealthPort` protocol in the application layer.

### 4.2 Command Interface

```
vs ticket-health [--epic <epic-id>] [--json]

Options:
  --epic <epic-id>    Scope to a specific epic (default: all epics)
  --json              Machine-readable JSON output
```

### 4.3 Report Format (Human-Readable)

```
=== Ticket Health Report ===

Flagged for review: 3 tickets
  k7m.3  Multi-tool config design    (flagged by: k7m.9, 2 days ago)
  k7m.4  CLI+MCP design              (flagged by: k7m.9, 2 days ago)
  k7m.11 Ticket pipeline design      (flagged by: k7m.5, 1 day ago)

Oldest unreviewed: k7m.8 (last reviewed: 2026-02-15, 8 days ago)

By epic:
  vibe-seed-k7m (Phase 1 Foundation)
    Total: 12 | Open: 7 | Flagged: 3 | Closed: 5
    Freshness: 71% (5 of 7 open tickets reviewed within 7 days)

Summary:
  Review 3 flagged tickets before starting new work.
  Run: bd query label=review_needed
```

### 4.4 Report Format (JSON)

```json
{
  "flagged_tickets": [
    {
      "id": "k7m.3",
      "title": "Multi-tool config design",
      "flagged_by": ["k7m.9"],
      "flagged_at": "2026-02-21T14:30:00Z",
      "days_flagged": 2
    }
  ],
  "oldest_unreviewed": {
    "id": "k7m.8",
    "last_reviewed": "2026-02-15",
    "days_since_review": 8
  },
  "epics": [
    {
      "id": "vibe-seed-k7m",
      "title": "Phase 1 Foundation",
      "total": 12,
      "open": 7,
      "flagged": 3,
      "closed": 5,
      "freshness_pct": 71
    }
  ],
  "summary": {
    "total_flagged": 3,
    "total_open": 7,
    "freshness_pct": 71
  }
}
```

### 4.5 Data Sources

| Data Point | Source | Method |
|-----------|--------|--------|
| Flagged tickets | `bd list --label review_needed --json` | Label query |
| Triggering ticket ID | Parse comments for `**Triggered by:**` | Comment parsing |
| Flagged date | Parse comments for `**Ripple review needed**` timestamp | Comment parsing |
| Last reviewed | Parse comments for `**Reviewed:**` prefix | Comment parsing |
| Epic membership | `bd children <epic-id> --json` | Dependency query |
| Freshness percentage | (open - flagged) / open * 100 | Calculated |

### 4.6 Freshness Percentage Definition

```
freshness_pct = ((open_tickets - flagged_tickets) / open_tickets) * 100
```

A ticket is "fresh" if:
- It has no `review_needed` label, AND
- Its `last_reviewed` date (from review comments) is within 7 days

Thresholds:
- 90-100%: Healthy
- 70-89%: Acceptable
- Below 70%: Action needed -- review flagged tickets before starting new work

### 4.7 Implementation Notes

The `vs ticket-health` command is a thin CLI adapter over the `TicketHealthPort` protocol.
The port calls beads CLI commands (`bd list`, `bd children`, `bd comments`) via an
infrastructure adapter (the Beads ACL layer). The domain logic is the comment parsing and
freshness calculation.

Since this is a **read-only query** (not a command), it does not go through the
RippleReview aggregate. It is a separate read model (TicketHealthReport) as defined in
DDD.md section 6.

---

## 5. Two-Tier Ticket Generation Rules

### 5.1 Definitions

| Tier | Criteria | Detail Level |
|------|----------|-------------|
| **Near-term** | In the current or next epic; has no unresolved blockers, or blockers are in progress | Full: AC, TDD phases, SOLID mapping, edge cases, design section |
| **Far-term** | In a future epic; has unresolved blockers in earlier epics | Stub: title, one-sentence goal, epic link, formal dependencies |

### 5.2 Classification Algorithm

```
For each generated ticket:
  1. Determine epic (bounded context → epic)
  2. Determine dependency depth:
     depth = max hops from any root ticket (no dependencies) to this ticket
  3. If depth <= 2: near-term → full detail
  4. If depth > 2 AND subdomain == Core: near-term → full detail (always)
  5. If depth > 2 AND subdomain == Supporting: far-term → standard detail
  6. If depth > 2 AND subdomain == Generic: far-term → stub

  Override: if ticket.dependencies are all closed → promote to near-term
```

The depth threshold of 2 means: tickets that are at most 2 dependency hops from ready
work get full detail. This typically captures the current sprint's work plus the next
sprint's work.

### 5.3 Stub Ticket Format

```markdown
## Goal / Problem

<One sentence describing the outcome needed.>

## Background / Context

This is a far-term stub ticket. Full specification will be added when this ticket is
promoted to near-term (all blockers resolved or in progress).

## DDD Alignment

| Aspect | Detail |
|--------|--------|
| Bounded Context | <context-name> |
| Layer | <domain/application/infrastructure> |

## Dependencies

- Blocked by: <list of blocking ticket IDs>
```

### 5.4 Promotion (Stub to Full)

When a stub ticket's blockers are resolved (detected by ripple review), the stub is
promoted to full detail. This is a **manual** step in the grooming checklist:

1. Agent picks up a stub ticket that is now unblocked
2. Agent reads the DDD artifacts and closed blocker outputs
3. Agent generates full AC, TDD phases, SOLID mapping
4. Agent presents the expanded ticket to the user for approval
5. User approves the expansion
6. Agent updates the ticket via `bd update <id> --description`

The `promote_stub(ticket_id)` command from the TicketPlan aggregate (DDD.md) maps to
this flow.

---

## 6. After-Close Protocol Specification

### 6.1 Purpose

The after-close protocol is the exact sequence of steps that MUST run after every
`bd close`. It is embedded in the generated CLAUDE.md for every bootstrapped project.
Agents follow it automatically -- they do not wait for the user to ask.

### 6.2 The Protocol (4 Steps)

```markdown
# After-close protocol (run automatically after every `bd close`):

## Step 1: Ripple Review
bin/bd-ripple <closed-id> "<what this ticket produced>"
# If no summary provided, use the close_reason from bd close --reason

## Step 2: Review Flagged Tickets
bd query label=review_needed
# For each flagged ticket:
#   1. Read the ripple comment (bd comments <id>)
#   2. Assess whether description/AC need updates
#   3. Present suggested changes to user
#   4. On approval: update ticket, clear label, add review comment
#   5. On dismissal: clear label, add "unchanged" review comment

## Step 3: Follow-Up Work
# If closing produced new work items:
#   1. Create tickets with bd create (NEVER empty descriptions)
#   2. Use beads-ticket-template.md for tasks, beads-spike-template.md for research
#   3. Set formal dependencies with bd dep add
#   4. If the new ticket is far-term, use stub format (section 5.3)

## Step 4: Groom Next
bd ready
# Pick the highest-priority ready ticket, then run the grooming checklist:
#   1. Template compliance
#   2. Freshness check (label list → review_needed?)
#   3. PRD traceability (/prd-traceability <id>)
#   4. DDD alignment
#   5. Ubiquitous language
#   6. TDD & SOLID phases documented
#   7. Acceptance criteria testable
# Present grooming results to user and ask if they want to start.
```

### 6.3 CLAUDE.md Section (Verbatim for Generated Projects)

This is the exact text that `vs init` generates into the bootstrapped project's CLAUDE.md.
It replaces the current inline comments with a formal, numbered protocol.

```markdown
## After-Close Protocol

After every `bd close <id>`, run these steps automatically. Do not wait for the user to ask.

### 1. Ripple Review
```bash
bin/bd-ripple <closed-id> "<what this ticket produced>"
```
This flags open dependents and siblings with `review_needed` and adds a context diff
comment explaining what changed.

### 2. Review Flagged Tickets
```bash
bd query label=review_needed
```
For each flagged ticket:
1. Read the ripple comments: `bd comments <id>`
2. Compare the ticket's description and AC against the new context
3. Draft suggested updates (or "no changes needed")
4. **Present suggestions to the user for approval** -- never auto-update
5. If approved: apply updates, then clear the flag:
   ```bash
   bd update <id> --description "<updated description>"
   bd label remove <id> review_needed
   bd comments add <id> "**Reviewed:** <date> by <agent>\n**Triggered by:** \`<closed-id>\`\n**Verdict:** <updated|unchanged|dismissed>\n**Changes:** <summary>"
   ```

### 3. Follow-Up Tickets
If closing produced new work:
1. Create tickets using the appropriate template (never empty descriptions):
   - Tasks/Features: `docs/beads_templates/beads-ticket-template.md`
   - Spikes: `docs/beads_templates/beads-spike-template.md`
2. Set formal dependencies: `bd dep add <new-id> <related-id>`
3. Far-term tickets use stub format (title + one-sentence goal + dependencies only)

### 4. Groom Next Ticket
```bash
bd ready
```
Pick the highest-priority ready ticket and run the full grooming checklist:
1. **Template compliance** -- description follows the beads template
2. **Freshness check** -- `bd label list <id>` for `review_needed`
3. **PRD traceability** -- `/prd-traceability <id>` to verify capability coverage
4. **DDD alignment** -- bounded context boundaries respected
5. **Ubiquitous language** -- terms match `docs/DDD.md` glossary
6. **TDD & SOLID** -- RED/GREEN/REFACTOR phases documented
7. **Acceptance criteria** -- testable checkboxes, edge cases, coverage >= 80%

Present grooming results and ask the user if they want to start the ticket.
```

### 6.4 Enforcement Mechanism

The protocol is enforced by the CLAUDE.md instruction: "run automatically, don't wait
for user to ask." AI agents (Claude Code, Cursor agents, etc.) read CLAUDE.md at
session start and follow it. There is no technical hook that prevents an agent from
skipping the protocol -- the enforcement is convention-based, like all CLAUDE.md rules.

The `vs ticket-health` command provides a **detection** mechanism: if flagged tickets
accumulate without being reviewed, the freshness percentage drops, making the problem
visible.

---

## 7. Updated Ticket Template Guidance

### 7.1 Additions to `beads-ticket-template.md`

The existing ticket template (`docs/beads_templates/beads-ticket-template.md`) already
includes the Risks/Dependencies section with the critical warning about formal `bd dep add`.
No structural changes needed.

However, the following guidance should be added to the template header:

```markdown
> **Freshness:** If this ticket has a `review_needed` label, read the ripple comments
> (`bd comments <id>`) before starting work. Present review results to the user and
> clear the flag before claiming the ticket.
```

### 7.2 Stub Ticket Sections (New Template Variant)

For far-term stubs, the ticket template reduces to:

```markdown
## Goal / Problem
<One sentence.>

## DDD Alignment
| Aspect | Detail |
|--------|--------|
| Bounded Context | <name> |
| Layer | <domain/application/infrastructure> |

## Dependencies
- Blocked by: <ticket-ids>

> **Stub ticket.** Full specification will be added when blockers are resolved.
> Do not start work on this ticket until it has been promoted to full detail.
```

---

## 8. Integration Points

### 8.1 With PRD Traceability

Ripple review and PRD traceability are complementary safety nets:

| Safety Net | What It Catches | When It Runs |
|-----------|----------------|--------------|
| Ripple review | Freshness decay -- "did a dependency change?" | After `bd close` (event-driven) |
| PRD traceability | Completeness gaps -- "was a PRD capability never assigned?" | During grooming (step 3) |

Both run during the grooming checklist. Ripple review is step 2 (freshness), PRD
traceability is step 3 (completeness). Neither replaces the other.

### 8.2 With `vs doc-health`

`vs doc-health` (Knowledge Base context) tracks document freshness using time-based
`last_reviewed` dates in a doc registry TOML file. `vs ticket-health` (Ticket Freshness
context) tracks ticket freshness using event-based ripple flags.

They are separate commands in separate bounded contexts. The open question in DDD.md
section 8 ("How does `vs doc-health` relate to Ticket Freshness?") is answered: they are
separate contexts with separate mechanisms (time-based vs event-based). Both report
freshness but on different artifact types (docs vs tickets).

### 8.3 With Beads (Infrastructure)

The Ticket Freshness bounded context interacts with Beads through an Anticorruption Layer
(ACL). The ACL translates between domain concepts and beads CLI commands:

| Domain Concept | Beads Command |
|---------------|---------------|
| Flag ticket | `bd label add <id> review_needed` |
| Clear flag | `bd label remove <id> review_needed` |
| Add context diff | `bd comments add <id> "<comment>"` |
| Read flags | `bd label list <id>` |
| Query flagged | `bd query label=review_needed` or `bd list --label review_needed` |
| Read context | `bd comments <id>` |
| Find siblings | `bd children <parent-id>` |
| Find dependents | `bd show <id> --json` (parse dependents array) |
| Find related | `bd show <id> --json` (parse dependencies with type=related) |

---

## 9. Risks and Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| Comment parsing is fragile | Medium | Use exact format prefixes (`**Reviewed:**`, `**Ripple review needed**`). Test parsing with edge cases. |
| Agents skip the protocol | High | `vs ticket-health` detects accumulating flags. Include protocol in generated CLAUDE.md verbatim. |
| Over-flagging (too many ripples) | Low | Only siblings + dependents + related are flagged. Epics with 5-15 tickets produce manageable ripple counts. |
| Empty context diffs | Medium | `bd-ripple` aborts if context is empty. Require `--reason` on `bd close`. |
| Flag fatigue (always flagged) | Low | Two-tier generation means far-term stubs have nothing to review. Only near-term tickets with real AC trigger meaningful reviews. |
| Comment bloat | Low | At current scale (tens of tickets, not thousands), comment volume is manageable. Each ripple adds ~5 lines. |

---

## 10. Implementation Sequence

### 10.1 What Already Exists

- `bin/bd-ripple` -- Working bash script with sibling + dependent traversal, label + comment
- CLAUDE.md -- After-close protocol section (inline comments, not formal numbered steps)
- Grooming checklist -- Step 2 freshness check already defined
- DDD.md -- RippleReview aggregate with 4 invariants fully specified

### 10.2 Changes Needed (Follow-Up Tickets)

| # | Task | Type | Bounded Context | Depends On |
|---|------|------|----------------|------------|
| 1 | Update `bin/bd-ripple`: add `related` traversal, enforce non-empty context, structured comment format, `--json` flag, exit codes | Task | Ticket Freshness | None |
| 2 | Update `CLAUDE.md` after-close protocol to formal 4-step numbered format (section 6.3) | Task | Bootstrap (template) | None |
| 3 | Add freshness guidance to ticket template header (section 7.1) | Task | Bootstrap (template) | None |
| 4 | Create stub ticket template variant (section 7.2) | Task | Bootstrap (template) | None |
| 5 | Implement `vs ticket-health` CLI command with comment parsing and freshness report | Task | Ticket Freshness | k7m.4 (CLI framework) |
| 6 | Implement `TicketHealthPort` protocol and Beads ACL adapter | Task | Ticket Freshness | 5 |
| 7 | Write domain model for RippleReview aggregate (Python, with invariant enforcement) | Task | Ticket Freshness | k7m.4 |
| 8 | Write domain model for TicketHealthReport read model (Python, freshness calculation) | Task | Ticket Freshness | 7 |

Tasks 1-4 are quick template/script updates (no Python code, infrastructure only).
Tasks 5-8 are Phase 7 implementation work (Python domain + application + infrastructure).

---

## 11. Recommendation

The ripple review system is well-served by the current approach: a bash script
(`bd-ripple`) operating on beads labels and comments, enforced by CLAUDE.md convention.

**Immediate actions** (can be done now, no architecture work):
1. Fix `bd-ripple` to enforce non-empty context (invariant 1 violation)
2. Add `related` dependency traversal to `bd-ripple`
3. Formalize the after-close protocol in CLAUDE.md (section 6.3)

**Phase 7 actions** (requires CLI framework from k7m.4):
4. Implement `vs ticket-health` as a Python command with comment parsing
5. Implement RippleReview domain model with proper invariant enforcement
6. Generate the after-close protocol into bootstrapped projects' CLAUDE.md

The key finding is that **no beads schema changes are needed**. Labels + comments are
sufficient for the data model at the expected scale. The `vs ticket-health` command
provides the detection layer, and the CLAUDE.md protocol provides the enforcement layer.

---

## References

- `docs/DDD.md` section 5 (RippleReview aggregate) and section 6 (Event Storming)
- `docs/PRD.md` section 4 scenario 6 (Ticket Freshness) and section 5 P0 capability C20
- `docs/research/20260222_backlog_freshness_ticket_decay.md` (prior research)
- `docs/research/20260222_cli_mcp_design.md` section 2 (command tree) and section 4 (ports)
- `docs/reference/beads-knowledge.md` (labels, comments, dependencies CLI reference)
- `bin/bd-ripple` (existing implementation)
- CLAUDE.md after-close protocol section
