# /prd-traceability — alto-specific addenda

## PRD Capability Map — alto P0 + P1 tables

### P0 Capability → Bounded Context → Expected Ticket Coverage

| ID | PRD Capability | Bounded Context | Expected Ticket Scope |
|----|---------------|-----------------|----------------------|
| C1 | CLI tool (`vs`) | Bootstrap | CLI command tree, subcommands |
| C2 | MCP server | Bootstrap | MCP tool schemas, shared ports |
| C3 | `alto-scaffold/` project directory | Bootstrap | Directory structure, config.toml |
| C4 | `alto init` with preview | Bootstrap | Preview, confirm, file safety |
| C5 | Global settings detection | Bootstrap | Tool detection, conflict resolution |
| C6 | Existing project adoption (`alto init --existing`) | Rescue | Branch safety, gap report, scaffolding |
| C7 | Gap analysis | Rescue | Scan, compare, report |
| C8 | Guided project bootstrap | Guided Discovery | Conversational flow, question phases |
| C9 | DDD question framework | Guided Discovery | 10 questions, dual register, persona detection |
| C10 | Artifact generation | Domain Model | PRD, DDD.md, ARCHITECTURE.md from answers |
| C11 | Agent personas | Tool Translation | Developer, researcher, tech-lead, PM, QA agents |
| C12 | Beads integration | Ticket Pipeline | Epic/spike/ticket templates |
| C13 | Quality gates | Architecture Testing | ruff + mypy + pytest enforcement |
| C14 | Fitness function generation | Architecture Testing | import-linter + pytestarch from bounded context map |
| C15 | Domain story to ticket pipeline | Ticket Pipeline | DDD artifacts -> ordered beads tickets with formal `bd dep add` (not text-only deps) |
| C16 | Complexity budget | Domain Model | Core/Supporting/Generic classification + treatment levels |
| C17 | Multi-tool support | Tool Translation | Claude Code, Cursor, Roo Code, OpenCode configs |
| C18 | Knowledge base (RLM) | Knowledge Base | Addressable docs, DDD patterns, tool conventions |
| C19 | Doc maintenance commands | Knowledge Base | `alto doc-health`, `alto doc-review` |
| C20 | Ticket freshness & ripple review | Ticket Freshness | Close -> flag -> context diff -> review flow |
| C25 | Template-enforced ticket creation | Ticket Pipeline + Tool Translation + Ticket Freshness | Every ticket created (manual or generated) MUST use beads templates; generated CLAUDE.md enforces this in grooming checklist step 1 and after-close protocol step 2. Tickets: k7m.12 (after-close protocol design), k7m.20 (generated tickets use templates), k7m.21 (generated CLAUDE.md includes enforcement) |

### P1 Capability → Expected Ticket Scope

| ID | PRD Capability | Expected Ticket Scope |
|----|---------------|----------------------|
| C21 | Rescue mode structural migration | Implicit BC detection, anemic model scan, migration tickets |
| C22 | Tool knowledge versioning | Current + 3 previous major versions per tool |
| C23 | Knowledge base drift detection | Convention changes between versions, code vs doc divergence |
| C24 | Spike workflow | Guided spike creation, ADR output |

## Worked example report (alto)

```
============================================================
PRD TRACEABILITY REPORT: alto-k7m.4
============================================================

COVERED  C8  Guided project bootstrap
  -> alto-k7m.4 (deliverable: CLI command tree for alto init)

COVERED  C9  DDD question framework
  -> alto-k7m.4 (deliverable: alto guide design)

GAP      C19 Doc maintenance commands
  -> No ticket deliverable mentions alto doc-health or alto doc-review
  -> Should be in: CLI/MCP design spike (k7m.4)

============================================================
Coverage: 18/20 P0 capabilities (90%)
Gaps: 2 capabilities with no ticket coverage
============================================================
```

## Invocation

```bash
alto doc-health          # alto-specific doc-health invocation
```
