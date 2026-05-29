# /launch-team — alto-specific addenda

## Ticket prefix

alto tickets use the `alty-cli-` prefix. Example invocations:

```
/launch-team alty-cli-1wu                                  # single ticket
/launch-team alty-cli-cgm alty-cli-dfd                     # multiple tickets
/launch-team alty-cli-cgm alty-cli-dfd alty-cli-yl0
```

Dev agent names follow the pattern `dev-alty-cli-<id>` (e.g. `dev-alty-cli-1wu`).

## Project artifact paths

- `docs/DDD.md` — domain model, bounded contexts, ubiquitous language
- `docs/ARCHITECTURE.md` — technical architecture
- `docs/PRD.md` — product requirements
- `.golangci.yml` — lint config v2 (must pass with 0 issues)
- `arch-go.yml` — DDD layer enforcement (domain cannot import application/infrastructure)

## Go quality gates (must pass at every checkpoint)

```bash
go build ./...                                   # 0 errors
go vet ./...                                     # 0 errors
golangci-lint run ./...                          # 0 issues (v2 strict config)
go test ./... -race                              # all pass, >=80% coverage on new domain code
```

## Go-specific design-decision categories (Step 5)

- Go types and their shapes (value objects, entities, aggregates)
- Interface contracts (port signatures, return types, context-arg position)
- Patterns to follow (reference existing code with file:line)
- Domain events emitted or consumed (Watermill GoChannel)
- Bounded context boundaries (which package owns what)
- Constraints (what NOT to do — especially DDD layer rules)

## Enforced principles (Go-specific additions)

- **DDD layer rules** — `internal/{context}/domain/` has ZERO external deps; dependencies
  flow inward: infrastructure -> application -> domain. Enforced by `arch-go`.
- **CQRS-lite** — events route through Watermill GoChannel.
- **Testify idioms** — `assert.Len`, `assert.Empty`, `assert.ErrorIs`, `require.Error` for
  preconditions, `assert.InDelta` for floats (testifylint enforces).

## Handoff doc paths

Single-ticket waves: `.notes/handoff-alty-cli-<id>.md` (e.g. `handoff-alty-cli-1wu.md`).
Multi-ticket waves: `.notes/handoff-wave-<n>.md` or `.notes/handoff-<short-name>.md`.
