# Tech Lead — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## DDD Layer Paths (Go layout)

- `internal/{context}/domain/` — ZERO external deps (compiler-enforced via `internal/`)
- `internal/{context}/application/` — depends on domain + ports only
- `internal/{context}/infrastructure/` — implements ports, external deps allowed
- `internal/shared/domain/` — shared kernel (errors, value objects, events, DDD types)

## CQRS-lite (Go-specific)

- Watermill GoChannel for event dispatch (where applicable)

## Layer Violation Detection (Go grep recipes)

```bash
# Check domain files don't import application or infrastructure
grep -r "internal/.*application\|internal/.*infrastructure" internal/*/domain/ internal/shared/domain/

# Check application files don't import infrastructure
grep -r "internal/.*infrastructure" internal/*/application/
```

## Quality Gate Enforcement (Go)

```bash
go build ./...                                    # Compile check
go test ./... -v -race -coverprofile=coverage.out  # Tests + race detector
go vet ./...                                      # Static analysis
golangci-lint run                                 # Meta-linter
go tool cover -func=coverage.out                  # Verify >= 80%
```

## Linting Enforcement (golangci-lint v2)

golangci-lint v2 config in `.golangci.yml`. Key linters:

| Linter | Purpose |
|--------|---------|
| errcheck | No ignored errors |
| errorlint | Proper error wrapping/matching |
| wrapcheck | External errors wrapped with `%w` |
| contextcheck | Context propagation |
| noctx | `exec.CommandContext` required |
| revive | No name stutter, exported docs |
| gocritic | No `os.Exit` after `defer` |
| exhaustive | Enum switches complete |
| testifylint | Testify idioms |
| gci | Import ordering |
| gofumpt | Strict formatting |
| depguard | Package dependency rules |

**`fieldalignment` is disabled** (memory optimization, not correctness).

DDD layer enforcement also handled by `arch-go` (MIT) — see `arch-go.yml`.
