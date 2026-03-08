# CLAUDE.md

This file provides guidance to Claude Code when working on the alty project.

## Project Overview

alty is a guided project bootstrapper that turns a simple idea (4-5 sentences) into a structured, production-ready project. It enforces DDD + TDD + SOLID before AI coding tools start writing code. It works with Claude Code, Cursor, Roo Code, and OpenCode.

**Key interfaces:** CLI (`cmd/alty`) and MCP server (planned).

## Quick Reference

```bash
# Quality gates (run before completing any task)
go build ./...                   # Build
go vet ./...                     # Vet
go test ./... -race              # Tests with race detector
golangci-lint run ./...          # Lint (auto-fix: --fix)

# CLI testing
go run ./cmd/alty help           # Show commands
go run ./cmd/alty init           # Test new project flow
go run ./cmd/alty doc-health     # Test doc health check

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
bd export                        # Export Dolt DB -> JSONL (manual sync)
# NOTE: bd sync is deprecated. Git hooks handle Dolt<->JSONL sync automatically.
```

## Enforced Principles

These are non-negotiable. Every ticket, every PR, every code change.

### DDD (Domain-Driven Design)

| Principle | Enforcement |
|-----------|-------------|
| **Ubiquitous Language** | Type/method names = domain expert terminology |
| **Value Objects first** | Default to immutable value objects; entities only when identity needed |
| **Rich Domain Model** | Business logic lives in domain objects, not services |
| **Aggregate boundaries** | One aggregate per transaction; reference others by ID |
| **Bounded Contexts** | Each context has its own `domain/`, `application/`, `infrastructure/` |
| **Layer Rules** | Dependencies flow inward: infrastructure -> application -> domain |

### TDD (Test-Driven Development)

| Phase | Action |
|-------|--------|
| RED | Write failing test first |
| GREEN | Minimal code to pass |
| REFACTOR | Clean up, tests stay green |

### BDD (Behavior-Driven Development)

| Principle | Enforcement |
|-----------|-------------|
| **Given/When/Then** | Integration tests use this structure |
| **User behavior** | Test observable behavior, not implementation |
| **Acceptance criteria** | Written as BDD scenarios before implementation |

### SOLID

| Principle | Rule | Go Enforcement |
|-----------|------|----------------|
| **S**ingle Responsibility | One struct, one job | One handler per use case |
| **O**pen/Closed | Extend via composition | New adapters, not modifying existing |
| **L**iskov Substitution | Subtypes honor contracts | All adapters pass port interface tests |
| **I**nterface Segregation | Focused interfaces | One interface per concern (ISP) |
| **D**ependency Inversion | Depend on abstractions | Handlers depend on port interfaces, never concrete adapters |

### CQRS-lite

| Principle | Enforcement |
|-----------|-------------|
| **Separation** | Command handlers and query handlers are structurally separated |
| **Commands** | May return domain objects (not strict "error only") |
| **Queries** | Have no side effects |
| **Event bus** | Watermill GoChannel routes events asynchronously |
| **Not event sourcing** | State is mutable, events are notifications |

### Linting

All enforced by `.golangci.yml` v2 config:

| Category | Linters | Why |
|----------|---------|-----|
| Error handling | `errcheck`, `errorlint`, `wrapcheck` | All errors checked and wrapped |
| Context propagation | `noctx`, `contextcheck` | Context passed through all layers |
| Code quality | `revive`, `gocritic`, `exhaustive`, `staticcheck` | Idiomatic Go |
| Testing | `testifylint` | Idiomatic testify assertions |
| DDD boundaries | `depguard` | Domain cannot import application/infrastructure |
| Formatting | `gci`, `gofumpt` | Consistent import ordering and formatting |

**Rule: `golangci-lint run ./...` must report 0 issues before any task is complete.**

## Architecture

### Project Structure

```
alty/
├── cmd/alty/                    # CLI entry point (Cobra)
├── internal/
│   ├── bootstrap/               # Bootstrap bounded context
│   │   ├── domain/              # Entities, VOs, aggregates
│   │   ├── application/         # Handlers + port interfaces
│   │   └── infrastructure/      # Adapters
│   ├── discovery/               # Discovery bounded context
│   ├── challenge/               # DDD Challenge bounded context
│   ├── fitness/                 # Architecture Fitness bounded context
│   ├── dochealth/               # Doc Health bounded context
│   ├── knowledge/               # Knowledge Base bounded context
│   ├── rescue/                  # Rescue Mode bounded context
│   ├── research/                # Research bounded context
│   ├── ticket/                  # Ticket Pipeline bounded context
│   ├── tooltranslation/         # Tool Translation bounded context
│   ├── shared/                  # Shared kernel
│   │   ├── domain/              # Shared VOs, events, errors, DDD types
│   │   ├── application/         # Shared ports (FileWriter)
│   │   └── infrastructure/      # Event bus, LLM client, persistence
│   ├── composition/             # Composition root (DI wiring)
│   └── integration/             # Cross-context integration tests
├── docs/
│   ├── PRD.md                   # Product requirements
│   ├── templates/               # PRD, DDD Story, Architecture templates
│   ├── beads_templates/         # Epic, spike, ticket templates
│   ├── spikes/                  # Research spike definitions
│   └── research/                # Spike output reports
├── .claude/
│   ├── CLAUDE.md                # This file
│   ├── agents/                  # Agent personas
│   └── commands/                # Slash commands
├── .golangci.yml                # Lint config (v2, strict)
├── go.mod / go.sum              # Go module
└── Makefile                     # Build targets
```

### Layer Rules

- `internal/{context}/domain/` has ZERO external dependencies
- `internal/{context}/application/` depends on domain + ports only
- `internal/{context}/infrastructure/` implements ports, external deps allowed
- `internal/shared/` is the shared kernel (errors, VOs, events, DDD types)
- Dependencies flow inward: infrastructure -> application -> domain
- **Enforced by `depguard` in `.golangci.yml`**

### Key Documents

| Document | Purpose |
|----------|---------|
| `README.md` | Public-facing description |
| `docs/PRD.md` | Product requirements |
| `docs/DDD.md` | Domain model, bounded contexts, ubiquitous language |
| `docs/ARCHITECTURE.md` | Technical architecture |

## Development Rules

- **TDD required** -- Write test first, then implementation
- **DDD + SOLID enforced** -- Domain logic in `internal/{context}/domain/`, no framework leakage
- **Go 1.26+** with modules
- **Do not commit/push** without explicit user permission
- **Do not proceed** to next ticket without explicit user permission
- **Dogfooding rule** -- When we encounter a process problem, fix it for ourselves AND for the product. Update the relevant ticket via `/prd-traceability` to find it, or create a new ticket if none exists.

## What alty IS and IS NOT

**IS:** The architect that runs before builders. It produces blueprints, guardrails, and structured tickets for AI coding tools to execute.

**IS NOT:** Another AI coding tool. It does not write application code. It produces project structure, domain models, configs, and tickets.

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
2. **Surface review** — Compare the ticket's description and AC against the new context. Check for stale counts, renamed types, changed tool names, outdated assumptions.
3. **Compatibility check** (MANDATORY for dependent tickets) — This is a mini implementation simulation scoped to the interface between the closed ticket and the flagged ticket. You MUST:
   - Read every source file the closed ticket created or modified
   - Read the flagged ticket's design section (constructors, ports, adapters)
   - Trace the interface: do the flagged ticket's assumed methods/types/constructors still work with what was actually delivered?
   - Check: Are package-private symbols assumed to be accessible? Are constructors assumed to exist? Do method signatures match?
   - **Cite file:line for every claim.** Do NOT say "verified" without showing what you read.
   - If ANY link in the chain is broken, the ticket NEEDS UPDATE — not just text fixes, but design changes.
4. Draft suggested updates (or "no changes needed")
5. **Present suggestions to the user for approval** -- never auto-update
6. If approved: apply updates, then clear the flag

### 3. Follow-Up Tickets
If closing produced new work:
1. Create tickets using the appropriate template (never empty descriptions)
2. Set formal dependencies: `bd dep add <new-id> <depends-on-id>`
3. Verify with `bd blocked` that the graph is correct
4. **Spike audit:** If the closed ticket was a spike, verify its research report's follow-up intents were all created as tickets.

### 4. Groom Next Ticket
```bash
bd ready
```
Pick the highest-priority ready ticket and run the full grooming checklist.

## Workflow

### Agent Selection

| Ticket Type | Agent | Purpose |
|-------------|-------|---------|
| Spike / ADR | `researcher` | Library evaluation, research reports |
| Task / Bug | `developer` | TDD implementation |
| Task (tests) | `qa-engineer` | Coverage, edge cases |
| Review | `tech-lead` | Architecture compliance, code review |
| Planning | `project-manager` | Tickets, backlog grooming |

### Ticket Grooming Checklist

Before claiming a ticket:

1. **Template Compliance** -- Description follows the beads template
2. **Freshness Check** -- `bd label list <id>` for `review_needed`
3. **PRD Traceability** -- `/prd-traceability <id>` to verify capability coverage
4. **DDD Alignment** -- Bounded context boundaries respected
5. **Ubiquitous Language** -- Names match `docs/DDD.md` glossary
6. **TDD & SOLID** -- RED/GREEN/REFACTOR phases documented
7. **Acceptance Criteria** -- Testable, edge cases, coverage >= 80%
8. **Implementation Simulation** -- Mentally trace: constructor -> deps -> calls -> returns. No "magic happens here" steps.

## Go Conventions

### Import Order (enforced by gci)

```go
import (
    "context"                    // 1. Standard library
    "fmt"

    "github.com/stretchr/testify/assert"  // 2. Third-party

    "github.com/alty-cli/alty/internal/shared/domain/errors"  // 3. Local
)
```

### Naming

- Types/Interfaces: `PascalCase` (`DiscoverySession`, `ToolDetector`)
- Functions/methods: `PascalCase` exported, `camelCase` unexported
- Constants: `PascalCase` (`QualityGateLint`, `SubdomainCore`)
- Packages: `lowercase` single word (`domain`, `application`, `infrastructure`)

### Error Handling

```go
// Always wrap errors with context
if err := h.repo.Save(ctx, entity); err != nil {
    return fmt.Errorf("saving entity: %w", err)
}

// Error strings: lowercase, no punctuation
fmt.Errorf("invalid session state: %w", ErrInvariantViolation)

// Use errors.As for type assertions on errors
var exitErr *exec.ExitError
if errors.As(err, &exitErr) { ... }

// Use context.TODO() instead of nil context
session, err := adapter.GetSession(context.TODO(), id)
```

### Port/Adapter Pattern

```go
// Port (interface in application layer)
type ToolDetector interface {
    Detect(projectDir string) ([]string, error)
}

// Adapter (concrete in infrastructure layer)
type FileSystemToolDetector struct { ... }
func (d *FileSystemToolDetector) Detect(projectDir string) ([]string, error) { ... }

// Handler (depends on port, never adapter)
type BootstrapHandler struct {
    toolDetection ToolDetector  // interface, not concrete
}
```

### Testing

```go
// Use testify idioms (enforced by testifylint)
assert.Len(t, items, 3)           // not assert.Equal(t, 3, len(items))
assert.Empty(t, items)            // not assert.Len(t, items, 0)
assert.NotEmpty(t, items)         // not assert.True(t, len(items) > 0)
assert.ErrorIs(t, err, ErrFoo)   // not assert.True(t, errors.Is(err, ErrFoo))
require.Error(t, err)            // not assert.Error(t, err) for preconditions
assert.InDelta(t, 42.0, val, 0)  // not assert.Equal(t, 42.0, val)
```

## Quality Gates

**All must pass before task completion:**

| Gate | Command | Requirement |
|------|---------|-------------|
| Build | `go build ./...` | Zero errors |
| Vet | `go vet ./...` | Zero errors |
| Lint | `golangci-lint run ./...` | Zero issues |
| Tests | `go test ./... -race` | All pass |

**If any fail, you are NOT DONE.**

## Git Rules

- NEVER commit without explicit user request
- NEVER add Co-Authored-By lines
- NEVER amend unless explicitly asked
- Stage specific files, not `git add -A`
- Commit format: `<type>: <description>` (feat/fix/test/refactor/docs/chore)
- No GitHub -- repo is on private Git server. Do not use `gh` CLI.

## Tooling

- **Beads** (`bd`) -- Issue tracking in `.beads/issues.jsonl`
- **Context7** -- MCP server for library docs
- **Templates** -- `docs/beads_templates/` (epic, spike, ticket)
- **Doc Templates** -- `docs/templates/` (PRD, DDD Story, Architecture)
- **golangci-lint v2** -- Strict lint config in `.golangci.yml`
- **Watermill** -- Event bus (GoChannel for local, NATS for distributed)
- **Cobra** -- CLI framework
