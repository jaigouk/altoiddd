# Developer — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## Source layout (Go)

```
internal/
├── {context}/
│   ├── domain/             # Core business logic (ZERO external deps)
│   ├── application/        # Use cases, command/query handlers, ports
│   └── infrastructure/     # Adapters for external concerns
├── shared/                 # Shared kernel
│   ├── domain/             # DomainModel, BoundedContext, sentinel errors, value objects, events
│   ├── application/        # Shared ports (FileWriter)
│   └── infrastructure/     # Event bus, LLM client, persistence
├── composition/            # Composition root (DI wiring)
└── integration/            # Cross-context integration tests
cmd/
├── alto/main.go            # CLI entry point (Cobra)
└── alto-mcp/main.go        # MCP server entry point
```

## CQRS-lite (Go-specific)

- Watermill GoChannel for event dispatch (local) / NATS (distributed)

## Linting Rules (golangci-lint v2)

These linters are enforced — your code MUST pass:

| Linter | What it checks |
|--------|---------------|
| errcheck | No ignored errors |
| errorlint | `errors.Is`/`errors.As` not type assertion |
| wrapcheck | Errors from external packages wrapped with `%w` |
| contextcheck | `context.Context` propagated correctly |
| noctx | `exec.CommandContext` not `exec.Command` |
| revive | No name stutter (`pkg.PkgFoo`), exported types documented |
| gocritic | No `os.Exit` after `defer` |
| exhaustive | Switch on enums covers all cases |
| testifylint | `assert.Len`, `assert.Empty`, `assert.ErrorIs` idioms |
| gci | Import order: stdlib | third-party | local |
| gofumpt | Stricter gofmt formatting |
| staticcheck | ST1005 (lowercase errors), SA1012 (no nil context) |

## Quality Gates (Go)

```bash
go build ./...           # Compile check
go test ./... -v -race   # Tests with race detector
go vet ./...             # Static analysis
golangci-lint run        # Meta-linter
```
