# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

<!-- CUSTOMIZE: Replace with your project description -->
> **TODO:** Describe your project here. What problem does it solve? Who is it for?

## Quick Reference

```bash
# Quality gates (run before completing any task)
uv run ruff check .              # Lint (auto-fix: --fix)
uv run mypy .                    # Type check
uv run pytest                    # Tests

# Issue tracking (Beads)
bd ready                         # Find available work
bd show <id>                     # View details
bd update <id> --status in_progress
bd close <id>
bin/bd-ripple <id>               # Flag dependents after close (ripple review)
bd query label=review_needed     # See tickets needing review
bd update <id> --remove-label review_needed  # Clear flag after review
bd label add <id> <label>        # Add label to issue
bd label remove <id> <label>     # Remove label from issue
bd export                        # Export Dolt DB → JSONL (manual sync)
# NOTE: bd sync is deprecated. Git hooks handle Dolt↔JSONL sync automatically.

```

## After-Close Protocol

After every `bd close <id>`, run these steps automatically. Do not wait for the user to ask.

### 1. Ripple Review
```bash
bin/bd-ripple <closed-id> "<what this ticket produced>"
```
This flags open dependents and siblings with `review_needed` and adds a context diff comment.

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
   bd comments add <id> "**Reviewed:** <date>\n**Triggered by:** \`<closed-id>\`\n**Verdict:** <updated|unchanged|dismissed>\n**Changes:** <summary>"
   ```

### 3. Follow-Up Tickets
If closing produced new work:
1. Create tickets using the appropriate template (never empty descriptions):
   - Tasks/Features: `docs/beads_templates/beads-ticket-template.md`
   - Spikes: `docs/beads_templates/beads-spike-template.md`
   - Far-term stubs: `docs/beads_templates/beads-stub-template.md`
2. Set formal dependencies: `bd dep add <new-id> <depends-on-id>`
3. Verify with `bd blocked` that the graph is correct
4. **Spike audit:** If the closed ticket was a spike, verify its research report's follow-up intents were all created as tickets. Orphaned follow-ups (defined in the report but never created) represent lost work items.

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

## Development Rules

- **TDD required** — Write test first, then implementation
- **DDD + SOLID enforced** — Domain logic in `src/domain/`, no framework leakage
- **Python 3.12+** with `uv` package manager
- **No personal info** — No real names, emails, paths, or hardware specs in code
- **Do not commit/push** without explicit user permission
- **Do not proceed** to next ticket without explicit user permission
- **Dogfooding rule** — When you encounter a process problem (missing templates, broken workflow, enforcement gap), fix it for yourself AND capture it as a ticket. Use PRD traceability to find the related ticket, or create a new one. If the problem affects users, update the PRD.

## Project Lifecycle

Projects follow this progression. Do NOT skip steps:

```
1. README.md        → Initial idea (a few sentences)
2. docs/PRD.md      → Refined requirements (use docs/templates/PRD_TEMPLATE.md)
3. DDD Artifacts    → Domain stories, bounded contexts, ubiquitous language
                      (use docs/templates/DDD_STORY_TEMPLATE.md)
4. Architecture     → Technical design informed by DDD
                      (use docs/templates/ARCHITECTURE_TEMPLATE.md)
5. Spikes           → Time-boxed research for unknowns (docs/spikes/)
6. Implementation   → Beads tickets with DDD + TDD + SOLID
```

## Architecture

<!-- CUSTOMIZE: Fill in after completing DDD and architecture phases -->

### DDD Layer Structure

```
src/
├── domain/              # Core business logic (NO external dependencies)
│   ├── models/          # Entities, Value Objects, Aggregates
│   ├── services/        # Domain Services (stateless business operations)
│   └── events/          # Domain Events
├── application/         # Use cases / orchestration
│   ├── commands/        # Write operations (Command handlers)
│   ├── queries/         # Read operations (Query handlers)
│   └── ports/           # Interfaces (Protocols) for infrastructure
├── infrastructure/      # Adapters for external concerns
│   ├── persistence/     # Database, file storage implementations
│   ├── messaging/       # Message bus, event publishing
│   └── external/        # External API clients, third-party integrations
└── tests/
    ├── domain/          # Unit tests for domain logic
    ├── application/     # Unit tests for use cases
    ├── infrastructure/  # Integration tests for adapters
    └── integration/     # End-to-end tests
```

**Layer Rules:**
- `domain/` has ZERO external dependencies (no frameworks, no DB, no HTTP)
- `application/` depends on `domain/` and `ports/` (interfaces only)
- `infrastructure/` implements `ports/` and depends on external libraries
- Dependencies flow inward: infrastructure → application → domain

### Key Documents

| Document | Purpose |
|----------|---------|
| `docs/PRD.md` | Product requirements |
| `docs/DDD.md` | Domain model, bounded contexts, ubiquitous language |
| `docs/ARCHITECTURE.md` | Technical architecture |
| `docs/architecture/*.md` | Detailed architecture sections |

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
6. **TDD & SOLID Compliance** — RED/GREEN/REFACTOR phases documented
7. **Acceptance Criteria** — Testable checkboxes, edge cases, coverage >= 80%

If incomplete, update via `bd update <id> --description` before claiming.

## Coding Standards

> **Reference**: Rules enforced via `pyproject.toml`. Auto-fix: `uv run ruff check --fix .`

### TDD (Test-Driven Development)

| Phase           | Action                     |
| --------------- | -------------------------- |
| RED             | Write failing test first   |
| GREEN           | Minimal code to pass       |
| REFACTOR        | Clean up, tests stay green |

### SOLID Principles

| Principle                 | Rule                     | Example                                   |
| ------------------------- | ------------------------ | ----------------------------------------- |
| **S**ingle Responsibility | One class, one job       | `OrderValidator` only validates orders     |
| **O**pen/Closed           | Extend via composition   | `Notifier(channels=[email, slack])`        |
| **L**iskov Substitution   | Subtypes honor contracts | `PostgresRepo(Repository)` same interface  |
| **I**nterface Segregation | Focused interfaces       | `Protocol` with single method              |
| **D**ependency Inversion  | Depend on abstractions   | `def process(repo: Repository)` not `PostgresRepo` |

### DDD Principles

| Principle | Rule |
|-----------|------|
| **Ubiquitous Language** | Class/method names = domain expert terminology |
| **Value Objects first** | Default to immutable value objects; entities only when identity needed |
| **Rich Domain Model** | Business logic lives in domain objects, not services |
| **Aggregate boundaries** | One aggregate per transaction; reference others by ID |
| **Repositories for Roots** | Only aggregate roots get repositories |

### Python Conventions

**Import Order** (ruff I):

```python
from __future__ import annotations    # 1. Future
import sys                            # 2. Stdlib
from pydantic import BaseModel        # 3. Third-party
from src.domain.models import Order   # 4. Local
```

**Type Annotations** (mypy strict):

```python
def process(data: dict[str, Any]) -> list[str]: ...
def get_value(key: str) -> str | None: ...
class Config:
    DEFAULTS: ClassVar[dict[str, int]] = {}
```

**Naming**:

- Classes: `PascalCase`
- Functions/variables: `snake_case`
- Constants: `UPPER_SNAKE_CASE`
- Private: `_underscore_prefix`
- Unused: `_underscore_prefix`

### Avoid

```python
# Mutable default              # Use None instead
def f(items=[]):               def f(items=None):
    ...                            items = items or []

# Broad except                 # Specific exceptions
except Exception:              except (ValueError, KeyError):

# Magic values                 # Constants/Enums
if status == "active":         if status == Status.ACTIVE:

# Anemic domain model          # Rich domain model
order.status = "cancelled"     order.cancel(reason=reason)
```

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
- Commit format: `<type>: <description>` (feat/fix/test/refactor/docs/chore)

## Tooling

- **Beads** (`bd`) — Issue tracking in `.beads/issues.jsonl`
- **Context7** — MCP server for library docs
- **Templates** — `docs/beads_templates/` (epic, spike, ticket)
- **Doc Templates** — `docs/templates/` (PRD, DDD Story, Architecture)
