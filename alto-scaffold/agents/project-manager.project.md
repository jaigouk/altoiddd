# Project Manager — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## Ticket templates (this project)

- Epic: `alto-scaffold/templates/beads-epic-template.md`
- Task: `alto-scaffold/templates/beads-ticket-template.md`
- Spike: `alto-scaffold/templates/beads-spike-template.md`

## Go Quality Gates Reference

When creating/grooming tickets, reference these quality gates:

```bash
go build ./...           # Compile check
go test ./... -v -race   # Tests with race detector
go vet ./...             # Static analysis
golangci-lint run        # Meta-linter
```

## Go Project Structure

```
internal/{context}/domain/        # Domain layer per bounded context
internal/{context}/application/   # Application layer per bounded context
internal/{context}/infrastructure/ # Infrastructure layer per bounded context
internal/shared/domain/           # Shared kernel (errors, events, VOs, DDD types)
cmd/alto/                         # CLI entry point (Cobra)
cmd/alto-mcp/                     # MCP server entry point
```

## Go Ticket Conventions

- Tickets organized by DDD bounded context
- Each ticket specifies which `internal/{context}/` it affects
- Acceptance criteria include: `go build` passes, `go test -race` passes, `golangci-lint run` passes
- TDD required: RED/GREEN/REFACTOR phases documented
- BDD naming: `TestSubject_WhenCondition_ExpectOutcome`
- CQRS-lite: commands vs queries separated in ticket design
- Domain tests are the majority (fast, pure, no mocks)

## Enforced Principles (Go-specific row)

| Principle | Ticket Requirement |
|-----------|-------------------|
| Linting | `golangci-lint run` in quality gates |
