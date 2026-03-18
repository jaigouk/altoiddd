---
date: 2026-03-04
topic: Ripple Review Spike Validation
status: final
type: spike
ticket: alto-2j7.2
---

# Ripple Review Spike Validation

> **Spike:** alto-2j7.2 -- Validate ripple review design against current codebase
> **Primary Reference:** `docs/research/20260223_ripple_review_design.md` (spike k7m.12)
> **Decision:** The spike design is substantially implemented. Remaining gaps are
> small and well-scoped for child tickets 2j7.9, 2j7.10, 2j7.11.

## Summary

The ripple review spike (k7m.12, dated 2026-02-23) specified 8 sections covering
the full lifecycle of event-driven ticket freshness. Since that spike was written,
the codebase has implemented most of the design. This report compares each section
of the spike against current code, classifying findings as ALIGNED, GAP, DRIFT,
or N/A.

**Overall implementation: approximately 75-80% of the spike design is realized.**

The remaining delta is concentrated in three areas:
1. `TicketHealthReport` missing epic-level and freshness-percentage data (spike section 4)
2. `bin/bd-ripple` comment format lacks the review checklist (spike section 1.3)
3. No infrastructure adapter for `TicketReaderProtocol` (the Beads ACL)

---

## Section-by-Section Comparison

### 1. Ripple Review Data Model (Spike Section 1)

| Spike Element | Current State | Status | Delta |
|---|---|---|---|
| `review_needed` label storage | `bin/bd-ripple` line 162: `bd update "$tid" --add-label review_needed` | ALIGNED | None |
| Triggering ticket ID in comment | `bin/bd-ripple` line 142: comment includes `` `$TICKET_ID` `` | ALIGNED | None |
| Context diff in comment body | `bin/bd-ripple` line 144: `**What changed:** $CONTEXT` | ALIGNED | None |
| `last_reviewed` stored in comment | Not parsed anywhere in Python. CLAUDE.md instructs format `**Reviewed:** <date>` | ALIGNED (convention) | No code parses it yet; `TicketHealthHandler` would need an adapter |
| Flag stacking (multiple comments) | `bin/bd-ripple` lines 157-174: always adds comment even if already flagged | ALIGNED | None |
| Comment format: review checklist | `bin/bd-ripple` line 142-146: NO checklist items present | GAP | Spike section 1.3 specifies `- [ ] Description still accurate?` etc. Current comment omits checklist. |
| `ContextDiff` VO | `src/domain/models/ticket_freshness.py` lines 15-35: frozen dataclass with non-empty invariant | ALIGNED | None |
| `FreshnessFlag` VO | `src/domain/models/ticket_freshness.py` lines 38-48: frozen dataclass with `context_diff` + `flagged_at` | ALIGNED | None |
| `RippleReview` aggregate | `src/domain/models/ripple_review.py`: full aggregate with 4 invariants, `flag_ticket`, `clear_flag` | ALIGNED | None |
| `TicketFlagged` event | `src/domain/events/ticket_freshness_events.py` lines 12-26 | ALIGNED | None |
| `FlagCleared` event | `src/domain/events/ticket_freshness_events.py` lines 29-42 | ALIGNED | None |
| Non-empty context diff (invariant 1) | `ContextDiff.__post_init__` raises `InvariantViolationError`; `bin/bd-ripple` lines 58-68 enforces non-empty | ALIGNED | None |
| Only open tickets flagged (invariant 2) | `bin/bd-ripple` lines 91-92: filters `status in ('open', 'in_progress')`; `RippleReview.flag_ticket` checks `is_open` | ALIGNED | None |
| Flag stacking (invariant 3) | Both bash (comment always added) and Python (`flag_ticket` appends) | ALIGNED | None |
| No auto-clear (invariant 4) | CLAUDE.md checklist step 2 enforces human approval; `RippleReview.clear_flag` requires explicit call | ALIGNED | None |

**Section 1 verdict: 12 ALIGNED, 1 GAP (comment format missing review checklist)**

---

### 2. Close-Time Hook Design (Spike Section 2)

| Spike Element | Current State | Status | Delta |
|---|---|---|---|
| Trigger: manual `bin/bd-ripple` after `bd close` | CLAUDE.md after-close protocol step 1 | ALIGNED | None |
| Sibling traversal (children of same parent) | `bin/bd-ripple` lines 73-98 | ALIGNED | None |
| Dependent traversal (`blocks` type) | `bin/bd-ripple` lines 100-112 | ALIGNED | None |
| Related traversal (both directions) | `bin/bd-ripple` lines 114-126: traverses `dependencies` + `dependents` for `related` type | ALIGNED | Was listed as GAP in spike "changes required" table but has since been implemented |
| Empty context guard: abort | `bin/bd-ripple` lines 58-68: aborts with error if empty or "Closed" | ALIGNED | Was listed as GAP in spike but has since been fixed |
| Structured comment format (section 1.3) | Comment body is basic markdown, no checklist | GAP | Same as section 1 finding |
| Exit codes | `set -euo pipefail` at top; explicit `exit 1` on errors, `exit 0` on no candidates | ALIGNED | None |
| `--json` flag for machine-readable output | Not implemented | GAP | Spike section 2.5 specifies `--json` flag for `alto ticket-health` integration |
| Deduplication of candidates | `bin/bd-ripple` lines 128-131: `sort -u` | ALIGNED | None |

**Section 2 verdict: 7 ALIGNED, 2 GAP (comment checklist format, --json flag)**

---

### 3. Pick-Up-Time Flow Design (Spike Section 3)

| Spike Element | Current State | Status | Delta |
|---|---|---|---|
| Grooming checklist step 2: freshness check | CLAUDE.md line 225 (project), `.claude/CLAUDE.md` line 171 (template) | ALIGNED | None |
| `bd label list <id>` to check for `review_needed` | CLAUDE.md grooming step 2 | ALIGNED | None |
| Read ripple comments: `bd comments <id>` | CLAUDE.md grooming step 2 | ALIGNED | None |
| Present suggested updates to user | CLAUDE.md: "Present suggestions to the user for approval" | ALIGNED | None |
| Clear flag + add review comment on approval | CLAUDE.md after-close protocol step 2, item 5 | ALIGNED | None |
| Review comment format (section 1.4) | CLAUDE.md specifies: `**Reviewed:** <date>\n**Triggered by:**...` | ALIGNED | None |
| Multiple stacked flags: review ALL in one session | Not explicitly stated in CLAUDE.md | GAP | Spike section 3.2 specifies agent reviews ALL stacked comments. CLAUDE.md only says "Read the ripple comments" (implicit but not explicit). |
| Dismissal: still adds review comment | CLAUDE.md step 2 item 5: mentions `<updated|unchanged|dismissed>` | ALIGNED | None |

**Section 3 verdict: 7 ALIGNED, 1 GAP (explicit stacking review instruction)**

---

### 4. `alto ticket-health` Command Design (Spike Section 4)

| Spike Element | Current State | Status | Delta |
|---|---|---|---|
| `alto ticket-health` CLI command registered | `src/infrastructure/cli/main.py` line 483: stub returning "not yet implemented" | ALIGNED (stub) | Command exists but is not wired to handler |
| `TicketHealthPort` protocol | `src/application/ports/ticket_health_port.py`: `report(project_dir) -> TicketHealthReport` | ALIGNED | None |
| `TicketReaderProtocol` | `src/application/queries/ticket_health_handler.py` lines 22-27: `read_open_tickets()`, `read_flags()` | ALIGNED | None |
| `TicketHealthHandler` query handler | `src/application/queries/ticket_health_handler.py`: builds report from reader | ALIGNED | None |
| Human-readable report format | Not implemented (CLI stub only) | GAP | Spike section 4.3 specifies formatted output |
| JSON output format | Not implemented | GAP | Spike section 4.4 specifies `--json` with structured data |
| `--epic` scoping flag | Not implemented | GAP | Spike section 4.2 specifies `--epic <epic-id>` |
| Freshness percentage calculation | Not present in domain model | GAP | Spike section 4.6 defines `freshness_pct = (open - flagged) / open * 100`; `TicketHealthReport` has no such property |
| Epic-level breakdown | Not present in domain model | GAP | Spike section 4.4 JSON has `epics` array; `TicketHealthReport` has no epic data |
| `_StubTicketHealth` in composition | `src/infrastructure/composition.py` line 137-140: stub raises `NotImplementedError` | ALIGNED (stub) | No real adapter |
| Beads ACL adapter for `TicketReaderProtocol` | Not implemented | GAP | No infrastructure adapter translates `bd list`/`bd comments` to `TicketReaderProtocol` |
| Comment parsing logic for `**Reviewed:**` | Not implemented | GAP | Spike section 4.5 specifies parsing comments for dates |

**Section 4 verdict: 4 ALIGNED (including stubs), 7 GAP**

---

### 5. Two-Tier Ticket Generation Rules (Spike Section 5)

| Spike Element | Current State | Status | Delta |
|---|---|---|---|
| `TicketDetailLevel` VO (full/standard/stub) | `src/domain/models/ticket_values.py` (imported by `TicketDetailRenderer`) | ALIGNED | None |
| `TicketDetailRenderer` service | `src/domain/services/ticket_detail_renderer.py`: `render(aggregate, detail_level, profile)` with full/standard/stub paths | ALIGNED | None |
| Full detail: AC, TDD phases, SOLID mapping, edge cases | `_render_full`: goal + ddd + design + solid + tdd + steps + ac + edge cases + quality gates | ALIGNED | None |
| Standard detail: core sections only | `_render_standard`: goal + ddd + steps + ac + quality gates | ALIGNED | None |
| Stub detail: minimal | `_render_stub`: goal + ac (one line) | DRIFT | Spike section 5.3 specifies DDD Alignment table + Dependencies section in stubs. `_render_stub` only has Goal + single AC checkbox. |
| Classification algorithm (depth-based) | Not implemented as code | N/A | This is ticket pipeline generation logic, not yet built |
| Stub promotion flow | Not implemented as code | N/A | Spike section 5.4 describes the manual grooming flow, not a code artifact |
| Stub template file | `docs/beads_templates/beads-stub-template.md`: Goal + DDD Alignment + Dependencies | ALIGNED | Template exists and matches spike section 7.2 |

**Section 5 verdict: 4 ALIGNED, 1 DRIFT (stub render content), 2 N/A**

---

### 6. After-Close Protocol (Spike Section 6)

| Spike Element | Current State | Status | Delta |
|---|---|---|---|
| 4-step protocol (Ripple, Review, Follow-Up, Groom) | Both CLAUDE.md files have all 4 steps | ALIGNED | None |
| Step 1: `bin/bd-ripple <id> "summary"` | CLAUDE.md step 1 | ALIGNED | None |
| Step 2: `bd query label=review_needed` + review flow | CLAUDE.md step 2 with 5 sub-items | ALIGNED | None |
| Step 3: follow-up tickets with templates | CLAUDE.md step 3 | DRIFT | Spike only mentions `beads-ticket-template.md` and `beads-spike-template.md`. Current CLAUDE.md also references `beads-stub-template.md` (codebase is ahead of spike). |
| Step 3: spike audit sub-step | CLAUDE.md step 3 item 4: "Spike audit: If closed ticket was a spike, verify follow-up intents" | DRIFT | This sub-step does not appear in the spike design at all. It was added post-spike. Codebase is ahead. |
| Step 4: `bd ready` + grooming checklist (7 items) | CLAUDE.md step 4 with all 7 grooming items | ALIGNED | None |
| Verbatim text matches spike section 6.3 | `.claude/CLAUDE.md` template matches spike's prescribed format almost exactly | ALIGNED | Minor wording differences but structurally identical |
| Enforcement: CLAUDE.md convention + ticket-health detection | CLAUDE.md convention is in place; `ticket-health` is stub only | ALIGNED (partial) | Detection layer not functional yet |

**Section 6 verdict: 5 ALIGNED, 2 DRIFT (both positive -- codebase ahead of spike)**

---

### 7. Ticket Template Guidance (Spike Section 7)

| Spike Element | Current State | Status | Delta |
|---|---|---|---|
| Freshness guidance in ticket template header | `docs/beads_templates/beads-ticket-template.md` line 16-18: has `> **Freshness:** If this ticket has a `review_needed` label...` | ALIGNED | None |
| Stub ticket template (section 7.2) | `docs/beads_templates/beads-stub-template.md`: matches spike format | ALIGNED | None |

**Section 7 verdict: 2 ALIGNED, 0 GAP**

---

### 8. Integration Points (Spike Section 8)

| Spike Element | Current State | Status | Delta |
|---|---|---|---|
| Ripple review + PRD traceability complementary | CLAUDE.md grooming checklist: step 2 (freshness) + step 3 (PRD traceability) | ALIGNED | None |
| `alto doc-health` vs `alto ticket-health` separate | Separate ports: `DocHealthPort` and `TicketHealthPort`; separate bounded contexts | ALIGNED | None |
| Beads ACL layer for Ticket Freshness | Not implemented (no infrastructure adapter) | GAP | Spike section 8.3 specifies the ACL mapping table. No adapter exists. |
| Follow-up intent / spike audit | `src/domain/models/follow_up_intent.py` + `SpikeFollowUpPort` | DRIFT | These were NOT in the spike design. They were added post-spike as an extension to the after-close protocol. Codebase is ahead. |

**Section 8 verdict: 2 ALIGNED, 1 GAP, 1 DRIFT (positive)**

---

## Consolidated Findings

### Status Summary

| Status | Count | Percentage |
|--------|-------|-----------|
| ALIGNED | 43 | 72% |
| GAP (missing from codebase) | 12 | 20% |
| DRIFT (codebase diverged from spike) | 4 | 7% |
| N/A (not applicable yet) | 2 | 3% |
| **Total** | **61** | **100%** |

### All Gaps (Spike specifies, codebase lacks)

| # | Gap | Spike Section | Impact | Effort |
|---|-----|--------------|--------|--------|
| G1 | Comment format missing review checklist | 1.3 | Low -- agents can still review without checklist items | Small (bash edit) |
| G2 | `bin/bd-ripple --json` flag | 2.5 | Medium -- needed for `ticket-health` integration | Small (bash edit) |
| G3 | Explicit stacking review instruction in CLAUDE.md | 3.2 | Low -- current wording is implicit | Trivial (doc edit) |
| G4 | Human-readable report format for `ticket-health` | 4.3 | High -- command is a stub | Medium (Python) |
| G5 | JSON output format for `ticket-health` | 4.4 | Medium -- needed for tooling | Medium (Python) |
| G6 | `--epic` scoping flag for `ticket-health` | 4.2 | Low -- can be added later | Small (Python) |
| G7 | `freshness_pct` property on `TicketHealthReport` | 4.6 | Medium -- core metric missing | Small (Python + test) |
| G8 | Epic-level breakdown data in `TicketHealthReport` | 4.4 | Medium -- needed for meaningful reports | Medium (Python) |
| G9 | Beads ACL adapter for `TicketReaderProtocol` | 4.7, 8.3 | High -- without this, nothing works end-to-end | Medium (Python) |
| G10 | Comment parsing for `**Reviewed:**` timestamps | 4.5 | Medium -- feeds `last_reviewed` | Medium (Python) |
| G11 | CLI wiring: connect `ticket-health` to `TicketHealthHandler` | 4.7 | High -- command is unusable | Small (Python) |
| G12 | `_render_stub` missing DDD Alignment + Dependencies sections | 5.3 | Low -- template file is correct but code render diverges | Small (Python) |

### All Drift (Codebase ahead of spike)

| # | Drift | Spike Section | Assessment |
|---|-------|--------------|-----------|
| D1 | After-close protocol references `beads-stub-template.md` | 6 | Positive -- spike only mentioned 2 templates, codebase has all 3 |
| D2 | Spike audit sub-step in after-close protocol | 6 | Positive -- follow-up intent auditing was added post-spike |
| D3 | `FollowUpIntent` + `FollowUpAuditResult` domain models | 8 | Positive -- extends Ticket Freshness context beyond spike scope |
| D4 | `SpikeFollowUpPort` | 8 | Positive -- extends the port surface |

All drift items are **positive** -- the codebase has grown beyond the spike design in
useful, aligned directions. No drift items conflict with the spike design.

---

## Recommendations for Child Tickets

### Ticket 2j7.9: Harden `bin/bd-ripple`

Based on gaps G1 and G2, the following changes should be in scope:

1. **Add review checklist to comment format** (G1): Update the `COMMENT` variable at
   `bin/bd-ripple` line 142 to include the 4-item checklist from spike section 1.3:
   ```
   - [ ] Description still accurate?
   - [ ] Acceptance criteria still valid?
   - [ ] Dependencies still correct?
   - [ ] Estimates still realistic?
   ```

2. **Add `--json` output flag** (G2): When `--json` is passed, output a JSON array of
   flagged ticket IDs and context summaries instead of human-readable text. This enables
   `alto ticket-health` to call `bd-ripple` programmatically.

3. **Consider NOT adding `--json` to bd-ripple**: The spike says the `--json` flag enables
   `ticket-health` integration, but the `TicketHealthHandler` uses `TicketReaderProtocol`
   which reads from `bd list`/`bd comments`, not from `bd-ripple` output. The `--json`
   flag on `bd-ripple` may be unnecessary. Recommend deferring unless a concrete consumer
   is identified.

### Ticket 2j7.10: Implement `TicketHealthReport` Enrichments

Based on gaps G7, G8, G10, and G12:

1. **Add `freshness_pct` computed property** (G7) to `TicketHealthReport`:
   ```python
   @property
   def freshness_pct(self) -> float:
       if self.total_open == 0:
           return 100.0
       return ((self.total_open - self.review_needed_count) / self.total_open) * 100
   ```

2. **Add epic-level data** (G8): Introduce an `EpicHealthSummary` VO:
   ```python
   @dataclass(frozen=True)
   class EpicHealthSummary:
       epic_id: str
       title: str
       total: int
       open: int
       flagged: int
       closed: int
       freshness_pct: float
   ```
   Add `epics: tuple[EpicHealthSummary, ...]` to `TicketHealthReport`.

3. **Enrich `_render_stub`** (G12) to include DDD Alignment and Dependencies sections,
   matching `docs/beads_templates/beads-stub-template.md`.

4. **Comment parsing** (G10): Add a domain service or utility that extracts `**Reviewed:**`
   dates from comment text. This is needed by the Beads ACL adapter.

### Ticket 2j7.11: Wire `ticket-health` CLI End-to-End

Based on gaps G4, G5, G6, G9, G11:

1. **Implement Beads ACL adapter** (G9) for `TicketReaderProtocol`:
   - `read_open_tickets()` calls `bd list --status open --json` + `bd list --status in_progress --json`
   - `read_flags()` calls `bd comments <id> --json` and parses `**Ripple review needed**` comments

2. **Wire CLI command** (G11): Replace the stub at `main.py` line 483 with actual
   handler invocation via `AppContext.ticket_health`.

3. **Human-readable output** (G4): Format the `TicketHealthReport` per spike section 4.3.

4. **JSON output** (G5): Add `--json` flag to `ticket-health` command.

5. **Epic scoping** (G6): Add `--epic <epic-id>` flag. This can be deferred to a later
   ticket if 2j7.11 is already large.

---

## Other Observations

### CLAUDE.md Template Already Complete

The `.claude/CLAUDE.md` template (for bootstrapped projects) already contains the
full 4-step after-close protocol matching spike section 6.3. Both alto's own `CLAUDE.md`
and the template `CLAUDE.md` are in sync and aligned with the spike. No template changes
needed (spike section 6 task 2 is DONE).

### Follow-Up Intent System is a Bonus

The `FollowUpIntent`, `FollowUpAuditResult`, and `SpikeFollowUpPort` domain models were
not part of the original spike. They extend the Ticket Freshness bounded context in a
useful direction (auditing spikes for orphaned work items). The CLAUDE.md after-close
protocol step 3 item 4 ("Spike audit") references this. This is a positive evolution
that the spike did not anticipate.

### No Beads Schema Changes Needed

The spike's key finding (section 11) remains valid: labels + comments are sufficient for
the data model. No upstream beads changes are required.

### Domain Model Quality is High

The `RippleReview` aggregate, value objects, and domain events are well-tested:
- `tests/domain/models/test_ripple_review.py`: 10 tests covering all 4 invariants
- `tests/domain/models/test_ticket_freshness.py`: 14 tests covering all VOs
- `tests/domain/events/test_ticket_freshness_events.py`: event dataclass tests
- `tests/application/queries/test_ticket_health_handler.py`: 7 tests with fake reader

---

## References

- Spike design: `docs/research/20260223_ripple_review_design.md`
- Domain models: `src/domain/models/ticket_freshness.py`, `src/domain/models/ripple_review.py`
- Domain events: `src/domain/events/ticket_freshness_events.py`
- Application handler: `src/application/queries/ticket_health_handler.py`
- Application port: `src/application/ports/ticket_health_port.py`
- Bash script: `bin/bd-ripple`
- CLAUDE.md (alto): `CLAUDE.md` (after-close protocol, grooming checklist)
- CLAUDE.md (template): `.claude/CLAUDE.md`
- Ticket template: `docs/beads_templates/beads-ticket-template.md`
- Stub template: `docs/beads_templates/beads-stub-template.md`
- DDD artifacts: `docs/DDD.md` sections 4-6
- Composition root: `src/infrastructure/composition.py`
