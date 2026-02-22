# CLAUDE.md

This file provides guidance to Claude Code when working on the vibe-seed project itself.

## Project Overview

vibe-seed is a guided project bootstrapper that turns a simple idea (4-5 sentences) into a structured, production-ready project. It enforces DDD + TDD + SOLID before AI coding tools start writing code. It works with Claude Code, Cursor, Antigravity, and OpenCode.

**Key interfaces:** CLI (`bin/vs`) and MCP server (planned).

## Quick Reference

```bash
# Quality gates (run before completing any task)
uv run ruff check .              # Lint (auto-fix: --fix)
uv run mypy .                    # Type check
uv run pytest                    # Tests

# CLI testing
bin/vs help                      # Show commands
bin/vs version                   # Show version
bin/vs init                      # Test new project flow
bin/vs init --existing           # Test existing project flow
bin/vs doc-health                # Test doc health check

# Issue tracking (Beads)
bd ready                         # Find available work
bd show <id>                     # View details
bd update <id> --status in_progress
bd close <id>
bin/bd-ripple <id>               # Flag dependents after close (ripple review)
bd query label=review_needed     # See tickets needing review
bd update <id> --remove-label review_needed  # Clear flag after review
bd sync                          # Sync with git

# After-close protocol (run automatically, don't wait for user to ask):
# 1. Ripple review: bin/bd-ripple <id> → review flagged tickets → fix → clear labels
# 2. Follow-up tickets: if closing produced new work, create tickets WITH descriptions
#    using beads-ticket-template.md or beads-spike-template.md — NEVER empty descriptions
#    IMPORTANT: Set formal deps with `bd dep add <ticket> <depends-on>` — text in
#    descriptions is invisible to bd blocked/bd ready/ripple review
# 3. Groom next: bd ready → pick next ticket → run grooming checklist (§ Ticket Grooming Checklist)
# 4. Report: present grooming results + ask user if they want to start
```

## Development Rules

- **TDD required** — Write test first, then implementation
- **DDD + SOLID enforced** — Domain logic in `src/domain/`, no framework leakage
- **Python 3.12+** with `uv` package manager
- **Do not commit/push** without explicit user permission
- **Do not proceed** to next ticket without explicit user permission
- **Dogfooding rule** — When we encounter a process problem (missing templates, broken workflow, enforcement gap), fix it for ourselves AND for the product. Update the relevant ticket via `/prd-traceability` to find it, or create a new ticket if none exists. Our users will hit the same problem — the PRD must capture it as a capability.

## What vibe-seed IS and IS NOT

**IS:** The architect that runs before builders. It produces blueprints, guardrails, and structured tickets for AI coding tools to execute.

**IS NOT:** Another AI coding tool. It does not write application code. It produces project structure, domain models, configs, and tickets.

## Project Lifecycle

This project follows its own process. Do NOT skip steps:

```
1. README.md        → Vision (done)
2. docs/PRD.md      → Requirements (done, pending review)
3. DDD Artifacts    → Domain stories, bounded contexts
4. Architecture     → Technical design informed by DDD
5. Spikes           → Time-boxed research for unknowns
6. Implementation   → Beads tickets with TDD + SOLID
```

## Architecture

### Project Structure

```
vibe-seed/
├── bin/vs                       # CLI entry point (bash)
├── src/
│   ├── domain/                  # Core business logic (NO external deps)
│   │   ├── models/              # Entities, Value Objects, Aggregates
│   │   ├── services/            # Domain Services
│   │   └── events/              # Domain Events
│   ├── application/             # Use cases / orchestration
│   │   ├── commands/            # Write operations
│   │   ├── queries/             # Read operations
│   │   └── ports/               # Interfaces for infrastructure
│   └── infrastructure/          # Adapters for external concerns
│       ├── persistence/         # File I/O, template rendering
│       ├── messaging/           # (future) MCP server
│       └── external/            # Git operations, tool detection
├── docs/
│   ├── PRD.md                   # Product requirements
│   ├── templates/               # PRD, DDD Story, Architecture templates
│   ├── beads_templates/         # Epic, spike, ticket templates
│   ├── spikes/                  # DDD reference, research spikes
│   └── research/                # Spike output reports
├── .claude/
│   ├── CLAUDE.md                # Template CLAUDE.md for bootstrapped projects
│   ├── agents/                  # Agent personas (template for bootstrapped projects)
│   └── commands/                # Slash commands (template for bootstrapped projects)
└── tests/                       # Mirrors src/ structure
```

### Layer Rules

- `domain/` has ZERO external dependencies (no frameworks, no DB, no HTTP)
- `application/` depends on `domain/` and `ports/` (interfaces only)
- `infrastructure/` implements `ports/` and depends on external libraries
- Dependencies flow inward: infrastructure → application → domain

### Two Kinds of Files

1. **vibe-seed's own code** — `src/`, `tests/`, `bin/vs` — the tool itself
2. **Template files** — `.claude/`, `docs/templates/`, `docs/beads_templates/` — files that get copied into bootstrapped projects

When editing template files, remember they are generic. No vibe-seed-specific references.

## Key Documents

| Document                             | Purpose                                 | Status                        |
| ------------------------------------ | --------------------------------------- | ----------------------------- |
| `README.md`                          | Public-facing description               | Done                          |
| `docs/PRD.md`                        | Product requirements                    | Approved                      |
| `.notes/killer-features-analysis.md` | Competitive analysis, 6 killer features | Done                          |
| `docs/spikes/ddd_reference.md`       | DDD pragmatic guide                     | Done                          |
| `docs/templates/`                    | PRD, DDD Story, Architecture templates  | Done                          |

## Current Epic: Phase 1 Foundation (vibe-seed-k7m)

```
k7m.9 (killer features) ✓ → k7m.6 (PRD review) ✓ → k7m.5 (DDD) ✓ → k7m.7 (Architecture)
k7m.2 (DDD questions) ✓ ───────────────────────────→ k7m.5 ✓        ↑
k7m.1 (KB spike) ─────────────────────────────────────────────────────┤
k7m.3 (multi-tool) ───────────────────────────────────────────────────┤
k7m.4 (CLI+MCP) ──────────────────────────────────────────────────────┤
k7m.10 (fitness function design) ──────────────────────────────────────┤
k7m.11 (ticket pipeline design) ───────────────────────────────────────┤
k7m.12 (ticket freshness design) ─────────────────────────────────────┘
k7m.8 (gap analysis) — independent
```

Run `bd ready` to see what's available. Run `bd show <id>` for ticket details.

## Killer Features (Differentiators)

These six features define vibe-seed's competitive advantage. Reference `.notes/killer-features-analysis.md` for full details.

1. **Architecture Fitness Functions** — Executable boundary tests from domain model
2. **Domain Story → Ticket Pipeline** — Auto-generate ordered beads tickets from DDD
3. **Rescue Mode** — `vs init --existing` with structural migration, not just overlay
4. **Tool-Native Context Translation** — One domain model → native configs per AI tool
5. **Complexity Budget** — Core/Supporting/Generic classification enforced in tickets and tests
6. **Living Knowledge Base** — Versioned, drift-detecting knowledge in `.vibe-seed/`
7. **Ticket Freshness & Ripple Review** — Event-driven staleness detection; flag dependents on close, context diff for agents, human approves updates

## Workflow

### Agent Selection

| Ticket Type  | Agent             | Purpose                              |
| ------------ | ----------------- | ------------------------------------ |
| Spike / ADR  | `researcher`      | Library evaluation, research reports |
| Task / Bug   | `developer`       | TDD implementation                   |
| Task (tests) | `qa-engineer`     | Coverage, edge cases                 |
| Review       | `tech-lead`       | Architecture compliance, code review |
| Planning     | `project-manager` | Tickets, backlog grooming            |

### Ticket Grooming Checklist

Before claiming a ticket:

1. **Template Compliance** — Description MUST follow the appropriate beads template:
   - Tasks/Features → `docs/beads_templates/beads-ticket-template.md` (Goal, DDD Alignment, Design, SOLID Mapping, TDD Workflow, Steps, AC, Edge Cases, Quality Gates)
   - Spikes → `docs/beads_templates/beads-spike-template.md` (Research Question, Timebox, Background, Investigation Approach, Expected Deliverables)
   - If the description is missing or doesn't follow the template, populate it BEFORE any other grooming step.
2. **Freshness Check** — Run `bd label list <id>`. If `review_needed` is present, read the ripple comments (`bd comments <id>`) to see what changed. Present suggested updates to the user for approval before starting work. Clear with `bd update <id> --remove-label review_needed` after review.
3. **PRD Traceability** — Run `/prd-traceability <id>` to cross-reference the ticket's deliverables/AC against PRD capabilities. Ripple review catches *freshness* (did something change?), but not *completeness* (was something missing from the start). The capability map in `.claude/commands/prd-traceability.md` maps each PRD P0/P1 item to bounded contexts and expected ticket scope.
4. **DDD Alignment** — Does the ticket respect bounded context boundaries?
5. **Ubiquitous Language** — Do class/method names match domain language?
6. **TDD & SOLID** — RED/GREEN/REFACTOR phases documented
7. **Acceptance Criteria** — Testable checkboxes, edge cases, coverage >= 80%

Update via `bd update <id> --description` if incomplete.

## Coding Standards

### TDD

| Phase    | Action                     |
| -------- | -------------------------- |
| RED      | Write failing test first   |
| GREEN    | Minimal code to pass       |
| REFACTOR | Clean up, tests stay green |

### SOLID

| Principle                 | Rule                     |
| ------------------------- | ------------------------ |
| **S**ingle Responsibility | One class, one job       |
| **O**pen/Closed           | Extend via composition   |
| **L**iskov Substitution   | Subtypes honor contracts |
| **I**nterface Segregation | Focused interfaces       |
| **D**ependency Inversion  | Depend on abstractions   |

### Python Conventions

```python
from __future__ import annotations    # 1. Future
import sys                            # 2. Stdlib
from pydantic import BaseModel        # 3. Third-party
from src.domain.models import Order   # 4. Local
```

- Classes: `PascalCase` | Functions/variables: `snake_case` | Constants: `UPPER_SNAKE_CASE`
- Use `ClassVar` for mutable class attributes
- Prefer `list[str]` over `List[str]`, `str | None` over `Optional[str]`

## Quality Gates

**All must pass before task completion:**

| Gate  | Command               | Requirement        |
| ----- | --------------------- | ------------------ |
| Lint  | `uv run ruff check .` | Zero errors        |
| Types | `uv run mypy .`       | Zero errors        |
| Tests | `uv run pytest`       | All pass, no skips |

**If any fail, you are NOT DONE.**

## Git Rules

- NEVER commit without explicit user request
- NEVER add Co-Authored-By lines
- NEVER amend unless explicitly asked
- Stage specific files, not `git add -A`

## Tooling

- **Beads** (`bd`) — Issue tracking in `.beads/issues.jsonl`
- **Context7** — MCP server for library docs
- **Templates** — `docs/beads_templates/` (epic, spike, ticket)
- **Doc Templates** — `docs/templates/` (PRD, DDD Story, Architecture)
- **No GitHub** — Repo is on private Git server. Do not use `gh` CLI.
