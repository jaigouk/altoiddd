# QA Engineer — alto Go addenda

## Project language / runtime

- **Go 1.26+** with modules

## Test Commands (Go)

```bash
go test ./internal/domain/... -v -race                    # Domain only
go test ./internal/application/... -v -race               # Application only
go test ./internal/infrastructure/... -v -race            # Infrastructure only
go test ./... -v -race -coverprofile=coverage.out         # All + coverage
go tool cover -func=coverage.out                          # Coverage by function
go tool cover -html=coverage.out -o coverage.html         # Visual coverage
go test ./... -bench=. -benchmem                          # Benchmarks
```

## Quality Gates (Go)

```bash
go build ./...           # Compile check
go test ./... -v -race   # Tests with race detector
go vet ./...             # Static analysis
golangci-lint run        # Meta-linter
```
