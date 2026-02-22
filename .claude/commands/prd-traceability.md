---
name: prd-traceability
description: Check ticket coverage against PRD capabilities using RLM pattern
allowed-tools: Read, Grep, Glob
---

# /prd-traceability <ticket-id-or-scope>

Verify that tickets cover all relevant PRD capabilities. Uses the **RLM pattern**: PRD capabilities are addressable variables mapped to tickets.

## Why This Exists

Ripple review catches **freshness** — "a dependency changed, is this ticket still valid?"
PRD traceability catches **gaps** — "was a PRD capability never assigned to any ticket?"

These are complementary. Ripple review won't catch a capability that was missing from the start.

## Usage

```
/prd-traceability <ticket-id>     # Check one ticket against PRD
/prd-traceability epic:<epic-id>  # Check all tickets in an epic against PRD
/prd-traceability all             # Full coverage matrix
```

## RLM Approach

**Do NOT search.** Instead:
1. Read the PRD capability map below (the "variable")
2. Read the ticket(s) via `bd show <id>`
3. Cross-reference: which capabilities should this ticket cover?
4. Report gaps

## PRD Capability Map (Documents as Variables)

### Source of Truth

| Document | Section | Content |
|----------|---------|---------|
| `docs/PRD.md` → Section 5 | Must Have (P0) | All P0 capabilities |
| `docs/PRD.md` → Section 5 | Should Have (P1) | All P1 capabilities |
| `docs/DDD.md` → Section 3 | Subdomain Classification | Core/Supporting/Generic |
| `docs/DDD.md` → Section 4 | Bounded Contexts | Context boundaries |

### P0 Capability → Bounded Context → Expected Ticket Coverage

| ID | PRD Capability | Bounded Context | Expected Ticket Scope |
|----|---------------|-----------------|----------------------|
| C1 | CLI tool (`vs`) | Bootstrap | CLI command tree, subcommands |
| C2 | MCP server | Bootstrap | MCP tool schemas, shared ports |
| C3 | `.vibe-seed/` project directory | Bootstrap | Directory structure, config.toml |
| C4 | `vs init` with preview | Bootstrap | Preview, confirm, file safety |
| C5 | Global settings detection | Bootstrap | Tool detection, conflict resolution |
| C6 | Existing project adoption (`vs init --existing`) | Rescue | Branch safety, gap report, scaffolding |
| C7 | Gap analysis | Rescue | Scan, compare, report |
| C8 | Guided project bootstrap | Guided Discovery | Conversational flow, question phases |
| C9 | DDD question framework | Guided Discovery | 10 questions, dual register, persona detection |
| C10 | Artifact generation | Domain Model | PRD, DDD.md, ARCHITECTURE.md from answers |
| C11 | Agent personas | Tool Translation | Developer, researcher, tech-lead, PM, QA agents |
| C12 | Beads integration | Ticket Pipeline | Epic/spike/ticket templates |
| C13 | Quality gates | Architecture Testing | ruff + mypy + pytest enforcement |
| C14 | Fitness function generation | Architecture Testing | import-linter + pytestarch from bounded context map |
| C15 | Domain story to ticket pipeline | Ticket Pipeline | DDD artifacts → ordered beads tickets with formal `bd dep add` (not text-only deps) |
| C16 | Complexity budget | Domain Model | Core/Supporting/Generic classification + treatment levels |
| C17 | Multi-tool support | Tool Translation | Claude Code, Cursor, Antigravity, OpenCode configs |
| C18 | Knowledge base (RLM) | Knowledge Base | Addressable docs, DDD patterns, tool conventions |
| C19 | Doc maintenance commands | Knowledge Base | `vs doc-health`, `vs doc-review` |
| C20 | Ticket freshness & ripple review | Ticket Freshness | Close → flag → context diff → review flow |
| C25 | Template-enforced ticket creation | Ticket Pipeline + Tool Translation + Ticket Freshness | Every ticket created (manual or generated) MUST use beads templates; generated CLAUDE.md enforces this in grooming checklist step 1 and after-close protocol step 2. Tickets: k7m.12 (after-close protocol design), k7m.20 (generated tickets use templates), k7m.21 (generated CLAUDE.md includes enforcement) |

### P1 Capability → Expected Ticket Scope

| ID | PRD Capability | Expected Ticket Scope |
|----|---------------|----------------------|
| C21 | Rescue mode structural migration | Implicit BC detection, anemic model scan, migration tickets |
| C22 | Tool knowledge versioning | Current + 3 previous major versions per tool |
| C23 | Knowledge base drift detection | Convention changes between versions, code vs doc divergence |
| C24 | Spike workflow | Guided spike creation, ADR output |

## Implementation

When `/prd-traceability <target>` is invoked:

### Step 1: Determine Scope

- If `<ticket-id>`: check that one ticket's deliverables/AC against capabilities it should cover
- If `epic:<id>`: check all tickets in the epic collectively cover all relevant capabilities
- If `all`: full coverage matrix

### Step 2: Read Sources

```
Read docs/PRD.md (Section 5 — Capabilities)
Read the ticket description(s) via bd show <id>
```

### Step 3: Cross-Reference

For each PRD capability in the map above:
1. Which ticket(s) should cover it? (Use the Bounded Context column)
2. Does that ticket's deliverables/AC mention this capability?
3. If not → **GAP**

### Step 4: Report

```
============================================================
PRD TRACEABILITY REPORT: <scope>
============================================================

COVERED  C8  Guided project bootstrap
  → vibe-seed-k7m.4 (deliverable: CLI command tree for vs init)

COVERED  C9  DDD question framework
  → vibe-seed-k7m.4 (deliverable: vs guide design)

GAP      C19 Doc maintenance commands
  → No ticket deliverable mentions vs doc-health or vs doc-review
  → Should be in: CLI/MCP design spike (k7m.4)

============================================================
Coverage: 18/20 P0 capabilities (90%)
Gaps: 2 capabilities with no ticket coverage
============================================================
```

## When to Run

- **During grooming** — Step 2 of the grooming checklist references this command
- **After generating tickets** — Verify the ticket pipeline covered everything
- **Before closing an epic** — Ensure all PRD capabilities have been addressed

## Maintaining This Map

When the PRD changes (new capabilities added, priorities shifted):
1. Update the capability map in this file
2. Run `/prd-traceability all` to find new gaps
3. Create or update tickets to cover new capabilities
