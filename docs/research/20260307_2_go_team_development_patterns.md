# Go Team Development Patterns for AI Coding Agent Workflows

**Date:** 2026-03-07
**Type:** Spike Research
**Status:** Final

## Research Question

What are the best practices for team-based Go development with AI coding agents? Specifically:

1. What are the standard Go project scaffolding conventions for DDD projects?
2. What quality gate commands replace ruff/mypy/pytest in Go?
3. How do Go table-driven tests translate pytest parametrize and fixtures?
4. What should a tech-lead agent check in Go DDD code reviews?
5. What are the most common mistakes AI agents make when writing Go code, and how to mitigate them?

## Summary

Go provides stronger compile-time guarantees than Python, which means the quality gate
pipeline is simpler but the tooling choices are different. The compiler itself replaces
~80% of what ruff+mypy do for Python. For DDD projects, Go's `internal/` package
mechanism provides compiler-enforced layer boundaries that are impossible to achieve in
Python without third-party tools. Table-driven tests are the Go equivalent of pytest
parametrize, and testify's suite package maps directly to pytest fixtures. AI agents
writing Go code commonly hallucinate package names, skip error checks, and misuse
`context.Context` -- all of which can be caught by the recommended quality gate
pipeline.

---

## 1. Go Project Scaffolding Conventions for DDD Projects

### 1.1 Directory Structure

The recommended DDD project structure for Go combines the
[official Go module layout](https://go.dev/doc/modules/layout) with the
[ThreeDotsLabs wild-workouts pattern](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
and [Damiano Petrungaro's idiomatic Go DDD](https://www.damianopetrungaro.com/posts/ddd-how-i-structure-idiomatic-golang-services/).

```
project-root/
  go.mod
  go.sum
  Makefile
  .golangci.yml
  cmd/
    app/
      main.go               # Entry point: wires dependencies, starts server/CLI
  internal/                  # Compiler-enforced: cannot be imported by other modules
    domain/                  # Core business logic -- ZERO external dependencies
      order/                 # One package per aggregate root
        order.go             # Entity + Value Objects + domain errors
        repository.go        # Repository interface (port)
        service.go           # Domain service (stateless business operations)
        events.go            # Domain events
      shared/                # Shared kernel (value objects used across aggregates)
        money.go
        errors.go            # Domain-level sentinel errors
    application/             # Use cases / orchestration
      command/               # Write operations
        create_order.go      # Command handler
      query/                 # Read operations
        get_order.go         # Query handler
    infrastructure/          # Adapters for external concerns
      persistence/
        postgres/
          order_repo.go      # Implements domain.order.Repository
      messaging/
        watermill/
          publisher.go       # Event publisher adapter
      external/
        anthropic/
          client.go          # External API adapter
  tests/                     # Integration tests (unit tests live next to code)
    integration/
      order_test.go
```

**Source:** [ThreeDotsLabs wild-workouts](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example), [Go official module layout](https://go.dev/doc/modules/layout), [Damiano Petrungaro DDD in Go](https://www.damianopetrungaro.com/posts/ddd-how-i-structure-idiomatic-golang-services/)

### Key conventions from Wild Workouts

The Wild Workouts project demonstrates these critical DDD patterns in Go:

- **Domain types use unexported fields** with constructor validation:

```go
// internal/domain/hour/hour.go
type Hour struct {
    hour         time.Time    // unexported = immutable from outside
    availability Availability // unexported = enforced via methods
}

func NewHour(hour time.Time) (*Hour, error) {
    if !isValidHour(hour) {
        return nil, ErrInvalidHour
    }
    return &Hour{hour: hour, availability: Available}, nil
}
```

- **Repository interfaces defined in the domain package** (port pattern):

```go
// internal/domain/hour/repository.go
type Repository interface {
    GetOrCreateHour(ctx context.Context, time time.Time) (*Hour, error)
    UpdateHour(
        ctx context.Context,
        hourTime time.Time,
        updateFn func(h *Hour) (*Hour, error),
    ) error
}
```

- **Thin handlers** that contain no domain logic -- just orchestration:

```go
// gRPC handler: 18 lines, no domain logic, just orchestration
```

**Source:** [ThreeDotsLabs DDD Lite intro](https://threedots.tech/post/ddd-lite-in-go-introduction/)

### 1.2 Monorepo Pattern (Multiple Bounded Contexts)

For projects with multiple bounded contexts (like alto with Bootstrap, Knowledge,
ToolTranslation, etc.):

```
project-root/
  go.mod
  cmd/
    alto/main.go
  internal/
    bootstrap/              # Bounded context 1
      domain/
      application/
      infrastructure/
    knowledge/              # Bounded context 2
      domain/
      application/
      infrastructure/
    tooltranslation/        # Bounded context 3
      domain/
      application/
      infrastructure/
    shared/                 # Shared kernel across contexts
      events/
      valueobjects/
```

This uses the monorepo pattern from Petrungaro where each bounded context gets its own
subtree under `internal/`, with domain types at the root of each context.

**Source:** [Damiano Petrungaro DDD monorepo pattern](https://www.damianopetrungaro.com/posts/ddd-how-i-structure-idiomatic-golang-services/)

### 1.3 How `internal/` Enforces DDD Layer Boundaries

Go's `internal/` directory has **compiler-level enforcement**: packages inside
`internal/` can only be imported by code in the parent of `internal/`. This is
not a convention -- it is enforced by the Go compiler. This means:

| Enforcement | Go | Python |
|---|---|---|
| Layer boundary | Compiler-enforced via `internal/` | Convention-only (needs import-linter) |
| Cycle detection | Compiler-enforced (no circular imports) | Runtime error or convention |
| Visibility | `unexported` (lowercase) = private to package | `_prefix` = convention only |

Two strategic `internal/` placements enforce DDD:

1. **Root `internal/`** -- prevents any external module from importing domain, application, or infrastructure code
2. **`cmd/{binary}/internal/`** -- infrastructure adapters private to a specific binary, cannot leak between binaries

**Source:** [Go official docs on internal](https://go.dev/doc/modules/layout), [Go spec on import restrictions](https://pkg.go.dev/cmd/go#hdr-Internal_Directories)

### 1.4 go.mod Setup Conventions

```
module github.com/yourorg/yourproject

go 1.23

require (
    github.com/spf13/cobra v1.10.2
    github.com/stretchr/testify v1.11.1
)
```

Key conventions:
- Module path = repository URL (e.g., `github.com/yourorg/yourproject`)
- Use `go 1.23` or later for latest language features
- Run `go mod tidy` to keep dependencies clean
- Separate `tools.go` file with `//go:build tools` for dev-only dependencies

**Source:** [go.mod file reference](https://go.dev/doc/modules/gomod-ref)

### 1.5 Standard Makefile Targets

Based on patterns from [Alex Edwards](https://www.alexedwards.net/blog/a-time-saving-makefile-for-your-go-projects),
[Mohit Khare](https://www.mohitkhare.com/blog/go-makefile/), and
[Earthly](https://earthly.dev/blog/golang-makefile/):

```makefile
# ============================================================
# Go DDD Project Makefile
# ============================================================

.PHONY: help build test lint fmt vet audit clean run

# --- Build ---
build:                             ## Compile the binary
	go build -o bin/app ./cmd/app/

run: build                         ## Build and run
	./bin/app

clean:                             ## Remove build artifacts
	rm -rf bin/ coverage.out

# --- Quality Gates ---
test:                              ## Run all tests with race detector
	go test -race -count=1 ./...

test-cover:                        ## Run tests with coverage report
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

vet:                               ## Static analysis (compiler-adjacent)
	go vet ./...

lint:                              ## Run golangci-lint (meta-linter)
	golangci-lint run

fmt:                               ## Format with gofumpt
	gofumpt -w .

fmt-check:                         ## Check formatting (CI mode)
	gofumpt -d . | grep -q . && exit 1 || true

# --- Combined Gates ---
audit: vet lint test               ## Run ALL quality gates (CI pipeline)
	@echo "All quality gates passed."

# --- Dependencies ---
tidy:                              ## Clean up go.mod/go.sum
	go mod tidy

deps:                              ## Download dependencies
	go mod download

# --- Architecture ---
arch-lint:                         ## Check architecture boundaries
	go-arch-lint check

# --- Help ---
help:                              ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | \
	awk 'BEGIN {FS = ":.*## "}; {printf "  %-15s %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
```

**Source:** [Alex Edwards Makefile](https://www.alexedwards.net/blog/a-time-saving-makefile-for-your-go-projects), [Earthly Go Makefile](https://earthly.dev/blog/golang-makefile/)

---

## 2. Go Quality Gate Commands (Replacing ruff/mypy/pytest)

### 2.1 Quality Gate Mapping

| Python Gate | Go Equivalent | What It Catches |
|---|---|---|
| `ruff check .` (lint) | `golangci-lint run` | Style, bugs, complexity, security |
| `ruff check .` (imports) | `goimports` / `gci` | Import ordering and grouping |
| `ruff format .` | `gofumpt -w .` | Formatting (stricter than gofmt) |
| `mypy .` (type check) | `go build ./...` | Type errors (compile-time, 100% coverage) |
| `mypy .` (nil safety) | `go vet ./...` + `nilaway` linter | Nil pointer dereference detection |
| `pytest` | `go test ./...` | Unit and integration tests |
| `pytest --cov` | `go test -coverprofile=coverage.out ./...` | Test coverage |
| import-linter | `go-arch-lint check` or `arch-go` | Architecture boundary violations |
| pytestarch | `arch-go` | Architecture fitness functions |

### 2.2 `go build ./...` (Compile Check)

The Go compiler replaces the entire role of mypy and most of ruff's type-checking:

- **100% type safety** -- every variable, function parameter, and return type is checked at compile time
- **Unused imports** -- compiler error (not just a warning)
- **Unused variables** -- compiler error
- **Unreachable code** -- compiler warning via `go vet`
- **Interface satisfaction** -- compile-time check that implementations satisfy interfaces

```bash
go build ./...    # Compiles all packages, reports type errors
```

This single command catches errors that require ruff + mypy + careful Protocol usage in Python.

**Source:** [Go specification](https://go.dev/ref/spec), [Go vet documentation](https://pkg.go.dev/cmd/vet)

### 2.3 `go test ./...` (Test Runner)

```bash
go test ./...                              # Run all tests
go test -race ./...                        # With race condition detector
go test -count=1 ./...                     # Disable test caching
go test -race -count=1 -v ./...            # Verbose with race detector
go test -coverprofile=coverage.out ./...   # With coverage
go test -run TestSpecific ./internal/...   # Run specific test pattern
go test -short ./...                       # Skip long-running tests
```

Key differences from pytest:
- Tests live **next to the code they test** (`order.go` / `order_test.go`), not in a separate tree
- Test functions must start with `Test` and take `*testing.T`
- No test discovery magic -- naming convention is the discovery mechanism
- `-race` flag enables the race condition detector (no Python equivalent)

**Source:** [Go testing package](https://pkg.go.dev/testing)

### 2.4 `go vet ./...` (Static Analysis)

`go vet` is the official Go static analyzer, distributed with the Go toolchain:

```bash
go vet ./...    # Run all built-in analyzers
```

What it catches:
- Printf format string mismatches
- Unreachable code
- Suspicious variable shadowing
- Incorrect struct tags
- Copying sync.Mutex
- Invalid build tags

`go vet` is the baseline -- golangci-lint includes it plus 100+ additional analyzers.

**Source:** [go vet documentation](https://pkg.go.dev/cmd/vet)

### 2.5 `golangci-lint run` (Meta-Linter)

golangci-lint is the standard Go meta-linter. It wraps 100+ individual linters into
a single fast runner with parallel execution and caching.

| Property | Value |
|---|---|
| GitHub | [github.com/golangci/golangci-lint](https://github.com/golangci/golangci-lint) |
| Latest version | v2.3.0 (March 2026) |
| License | GPL-3.0 (dev tool, does not affect project license) |
| Config file | `.golangci.yml` |

**Source:** [golangci-lint docs](https://golangci-lint.run/docs/configuration/), Context7 `/golangci/golangci-lint`

#### Recommended `.golangci.yml` for DDD Projects

```yaml
# .golangci.yml -- golangci-lint v2 format
version: "2"

linters:
  default: standard              # Includes govet, staticcheck, errcheck, etc.
  enable:
    - revive                     # Extensible linter (replaces golint)
    - gocritic                   # Opinionated quality checks
    - errorlint                  # errors.Is/As usage enforcement
    - exhaustive                 # Ensure all enum cases handled
    - noctx                      # HTTP requests without context.Context
    - bodyclose                  # Unclosed HTTP response bodies
    - nilaway                    # Nil pointer dereference detection
    - testifylint                # Testify usage best practices
    - contextcheck               # context.Context propagation
    - misspell                   # Spelling in comments/strings

  settings:
    govet:
      enable-all: true
    revive:
      rules:
        - name: exported
          arguments:
            - checkPrivateReceivers
        - name: var-naming
        - name: indent-error-flow
        - name: error-return       # Error should be last return value
        - name: unexported-return  # Exported func should not return unexported type

formatters:
  enable:
    - gofumpt                    # Stricter formatting than gofmt
    - goimports                  # Auto-fix imports
    - gci                        # Import grouping (stdlib / third-party / local)

  settings:
    gci:
      sections:
        - standard               # stdlib
        - default                # third-party
        - prefix(github.com/yourorg/yourproject)  # local

output:
  formats:
    text:
      path: stdout
      print-linter-name: true
```

### 2.6 gofumpt vs goimports (Formatting)

| Tool | Purpose | Relationship | Maintained By |
|---|---|---|---|
| `gofmt` | Canonical Go formatter | Ships with Go toolchain | Go team |
| `gofumpt` | Stricter superset of gofmt | Additional rules, backward compatible | [mvdan](https://github.com/mvdan/gofumpt) |
| `goimports` | gofmt + auto-manage imports | Adds/removes imports | Go tools team |
| `gci` | Import grouping/ordering | Stdlib / third-party / local sections | Community |

**Recommendation: Use gofumpt + gci together via golangci-lint.** gofumpt handles
formatting (is a superset of gofmt, backward compatible), and gci handles import
grouping. Both are integrated into golangci-lint v2 as formatters.

gofumpt is becoming the de facto standard due to its integration with `gopls` (the Go
language server) and its backward compatibility with gofmt.

**Source:** [gofumpt GitHub](https://github.com/mvdan/gofumpt), [golangci-lint formatters docs](https://golangci-lint.run/docs/formatters/)

### 2.7 Architecture Boundary Enforcement

Two tools enforce DDD layer boundaries via YAML configuration:

#### go-arch-lint (fe3dback)

| Property | Value |
|---|---|
| GitHub | [github.com/fe3dback/go-arch-lint](https://github.com/fe3dback/go-arch-lint) |
| Stars | 303 |
| License | MIT |
| Latest version | v1.14.0 (November 2025) |
| Config | `.go-arch-lint.yml` (version 3) |

```yaml
# .go-arch-lint.yml
version: 3
workdir: internal
components:
  domain:       { in: domain/** }
  application:  { in: application/** }
  infrastructure: { in: infrastructure/** }
  shared:       { in: shared/** }

commonComponents:
  - shared

deps:
  domain: {}                     # Domain depends on NOTHING
  application:
    mayDependOn:
      - domain                   # Application depends on domain only
  infrastructure:
    mayDependOn:
      - domain                   # Infrastructure depends on domain + application
      - application
```

#### arch-go (fdaines)

| Property | Value |
|---|---|
| GitHub | [github.com/arch-go/arch-go](https://github.com/arch-go/arch-go) |
| Stars | ~250 |
| License | MIT |
| Latest version | v2.1.2 (February 2026) |
| Config | `arch-go.yml` |

```yaml
# arch-go.yml
dependenciesRules:
  - package: "**.domain.**"
    shouldNotDependsOn:
      internal: ["**.application.**", "**.infrastructure.**"]
  - package: "**.application.**"
    shouldOnlyDependsOn:
      internal: ["**.domain.**"]
  - package: "**.infrastructure.**"
    shouldOnlyDependsOn:
      internal: ["**.domain.**", "**.application.**"]
```

**Recommendation:** go-arch-lint (fe3dback) is the better choice for DDD projects.
Its component-based configuration is more expressive, the `mayDependOn` syntax maps
directly to DDD layer rules, and it has active maintenance. arch-go is a solid
alternative with a different (dependency-rule-based) approach.

**Source:** [go-arch-lint GitHub](https://github.com/fe3dback/go-arch-lint), [arch-go GitHub](https://github.com/arch-go/arch-go)

---

## 3. Go Table-Driven Test Patterns

### 3.1 Python pytest.parametrize vs Go Table-Driven Tests

**Python:**
```python
@pytest.mark.parametrize("a,b,expected", [
    (10, 2, 5.0),
    (-10, 2, -5.0),
    (7, 2, 3.5),
])
def test_divide(a, b, expected):
    assert divide(a, b) == expected
```

**Go equivalent:**
```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name     string
        a, b     float64
        expected float64
        wantErr  bool
    }{
        {"positive division", 10, 2, 5, false},
        {"negative division", -10, 2, -5, false},
        {"decimal result", 7, 2, 3.5, false},
        {"divide by zero", 10, 0, 0, true},
        {"zero numerator", 0, 5, 0, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Divide(tt.a, tt.b)

            if tt.wantErr {
                require.Error(t, err)
                return
            }

            require.NoError(t, err)
            assert.InDelta(t, tt.expected, result, 0.0001)
        })
    }
}
```

**Source:** Context7 `/stretchr/testify`

Key differences:
- Go test cases are **anonymous structs** (typed, compiler-checked), not tuples
- Each test case has an explicit `name` field for sub-test identification
- `t.Run(tt.name, ...)` creates named subtests (like pytest's parametrize IDs)
- `require` stops the test immediately on failure; `assert` continues
- No decorator magic -- the pattern is explicit in the test function

### 3.2 pytest Fixtures vs Go Patterns

| pytest Pattern | Go Equivalent | Mechanism |
|---|---|---|
| `@pytest.fixture` | testify `suite.Suite` with `SetupTest()` | Method on suite struct |
| `conftest.py` (module scope) | `TestMain(m *testing.M)` | Per-package setup/teardown |
| `@pytest.fixture(scope="session")` | `suite.SetupSuite()` / `suite.TearDownSuite()` | Once per suite |
| `@pytest.fixture(scope="function")` | `suite.SetupTest()` / `suite.TearDownTest()` | Before/after each test |
| `yield` in fixture | `t.Cleanup(func() {...})` | Deferred cleanup |
| `tmp_path` | `t.TempDir()` | Auto-cleaned temp directory |

#### testify Suite Pattern (replaces conftest.py fixtures)

```go
type OrderServiceSuite struct {
    suite.Suite
    mockRepo *MockOrderRepository
    service  *OrderService
}

// SetupTest runs before EACH test (like @pytest.fixture scope="function")
func (s *OrderServiceSuite) SetupTest() {
    s.mockRepo = new(MockOrderRepository)
    s.service = NewOrderService(s.mockRepo)
}

// TearDownTest runs after EACH test
func (s *OrderServiceSuite) TearDownTest() {
    s.mockRepo.AssertExpectations(s.T())
}

func (s *OrderServiceSuite) TestCreateOrder() {
    s.mockRepo.On("Save", mock.AnythingOfType("*Order")).Return(nil)

    order, err := s.service.CreateOrder("item-1", 100)

    s.Require().NoError(err)
    s.Equal("item-1", order.ItemID)
}

func (s *OrderServiceSuite) TestCreateOrderFailure() {
    s.mockRepo.On("Save", mock.Anything).Return(errors.New("db error"))

    _, err := s.service.CreateOrder("item-1", 100)

    s.Error(err)
    s.Contains(err.Error(), "db error")
}

// Entry point -- required to run the suite
func TestOrderServiceSuite(t *testing.T) {
    suite.Run(t, new(OrderServiceSuite))
}
```

**Source:** Context7 `/stretchr/testify`

#### TestMain (replaces conftest.py session fixtures)

```go
// order_test.go -- runs once per package
func TestMain(m *testing.M) {
    // Setup: equivalent to @pytest.fixture(scope="session")
    db := setupTestDatabase()

    // Run all tests in this package
    code := m.Run()

    // Teardown: cleanup after all tests
    db.Close()

    os.Exit(code)
}
```

#### t.Cleanup (replaces yield in fixtures)

```go
func TestWithTempFile(t *testing.T) {
    // t.TempDir() auto-cleans (like pytest tmp_path)
    dir := t.TempDir()

    f, err := os.CreateTemp(dir, "test-*")
    require.NoError(t, err)

    // t.Cleanup runs after the test (like yield in a fixture)
    t.Cleanup(func() {
        f.Close()
    })

    // ... test using f
}
```

### 3.3 testify assert vs require

| Package | Behavior | Use When |
|---|---|---|
| `assert` | Records failure, **continues** test | Multiple independent checks (want to see all failures) |
| `require` | Records failure, **stops** test immediately | Setup preconditions (no point continuing if setup fails) |

**Convention:** Use `require` for setup and preconditions, `assert` for verifications.
This maps to pytest where `assert` in a fixture would stop the test, but multiple
asserts in the test body all run.

### 3.4 Mock Pattern (replaces unittest.mock / pytest-mock)

```go
// Define the interface (port)
type OrderRepository interface {
    Save(ctx context.Context, order *Order) error
    FindByID(ctx context.Context, id string) (*Order, error)
}

// Generate the mock (testify)
type MockOrderRepository struct {
    mock.Mock
}

func (m *MockOrderRepository) Save(ctx context.Context, order *Order) error {
    args := m.Called(ctx, order)
    return args.Error(0)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id string) (*Order, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*Order), args.Error(1)
}
```

Alternatively, use `mockery` to auto-generate mocks from interfaces:
```bash
go install github.com/vektra/mockery/v2@latest
mockery --all --with-expecter
```

**Source:** Context7 `/stretchr/testify`, [mockery](https://github.com/vektra/mockery)

### 3.5 testify vs stdlib testing: When to Use Each

| Factor | stdlib `testing` | testify |
|---|---|---|
| Zero dependencies | Yes | No (third-party) |
| Assertion readability | `if got != want { t.Errorf(...) }` | `assert.Equal(t, want, got)` |
| Mocking | Manual | `mock.Mock` with expectations |
| Suite lifecycle | `TestMain` only | `SetupTest`/`TearDownTest`/`SetupSuite` |
| DDD recommendation | Small packages, value objects | Service tests, integration tests |

**Recommendation:** Use testify for domain service tests and anything with mocked
dependencies. Use stdlib `testing` for simple value object tests where assertions
are trivial.

---

## 4. Go DDD Code Review Checklist for Tech-Lead Agent

### 4.1 Layer Violations

| Check | How to Detect | Severity |
|---|---|---|
| Domain imports infrastructure | `go-arch-lint check` or manual import review | CRITICAL |
| Domain imports application | `go-arch-lint check` or manual import review | CRITICAL |
| Application imports infrastructure directly (not via port) | Check for concrete type in application layer | HIGH |
| Infrastructure types leaked into domain signatures | Check return types and parameters of domain methods | HIGH |
| Framework types in domain (e.g., `*sql.DB`, `*http.Request`) | Search domain packages for framework imports | CRITICAL |
| Circular imports | `go build ./...` (compiler catches this) | CRITICAL (compiler enforced) |

### 4.2 Error Handling

| Check | How to Detect | Rule |
|---|---|---|
| Unchecked error returns | `errcheck` linter via golangci-lint | Every error MUST be checked |
| Using `==` instead of `errors.Is()` | `errorlint` linter via golangci-lint | Always use `errors.Is()` for sentinel errors |
| Using type assertion instead of `errors.As()` | `errorlint` linter via golangci-lint | Always use `errors.As()` for error types |
| Error wrapping loses context | Manual review | Use `fmt.Errorf("context: %w", err)` |
| Domain errors defined in infrastructure | Manual review | Errors belong in the domain package |
| Infrastructure errors leaking to domain | Manual review | Wrap infra errors in domain errors at the adapter boundary |
| `error` is not last return value | `revive` linter `error-return` rule | Convention: `func F() (Result, error)` |

**Domain error pattern for code review:**

```go
// GOOD: Domain errors in domain package
// internal/domain/order/errors.go
var (
    ErrOrderNotFound     = errors.New("order not found")
    ErrInvalidQuantity   = errors.New("quantity must be positive")
    ErrAlreadyCancelled  = errors.New("order already cancelled")
)

// GOOD: Infrastructure wraps at boundary
// internal/infrastructure/persistence/postgres/order_repo.go
func (r *OrderRepo) FindByID(ctx context.Context, id string) (*order.Order, error) {
    row := r.db.QueryRowContext(ctx, "SELECT ...", id)
    if err := row.Scan(&...); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, order.ErrOrderNotFound  // Translate to domain error
        }
        return nil, fmt.Errorf("query order %s: %w", id, err)  // Wrap with context
    }
    return o, nil
}
```

**Source:** [JetBrains Go error handling best practices (2026)](https://blog.jetbrains.com/go/2026/03/02/secure-go-error-handling-best-practices/), [ThreeDotsLabs DDD error handling issue](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/issues/41)

### 4.3 Interface Satisfaction

| Check | How to Detect | Rule |
|---|---|---|
| Implicit interface satisfaction not verified | Add compile-time check | `var _ Repository = (*PostgresRepo)(nil)` |
| Interfaces too large (ISP violation) | Count methods on interface | Prefer 1-3 methods per interface |
| Interface defined in implementation package | Manual review | Interfaces belong where they are USED (domain/application), not where they are implemented |
| Missing `context.Context` as first parameter | `noctx` linter | All I/O operations take `context.Context` |

**Compile-time interface satisfaction check:**

```go
// internal/infrastructure/persistence/postgres/order_repo.go

// Compile-time check: PostgresRepo must implement order.Repository
var _ order.Repository = (*PostgresRepo)(nil)

type PostgresRepo struct {
    db *sql.DB
}
```

This is a zero-cost assertion that fails at compile time if `PostgresRepo` does not
implement all methods of `order.Repository`.

### 4.4 Immutability and Encapsulation

| Check | How to Detect | Rule |
|---|---|---|
| Exported fields on domain entities | Manual review / `revive` | Domain entity fields MUST be unexported |
| Missing constructor (New function) | Manual review | Entities need `NewX()` with validation |
| Setter methods on value objects | Manual review | Value objects are immutable -- no setters |
| Direct struct literal creation bypassing constructor | Manual review | Only `New*()` should create domain objects |
| Mutable slices/maps returned from domain methods | Manual review | Return copies, not references |

**Correct pattern:**

```go
// GOOD: Unexported fields, constructor with validation
type Order struct {
    id        string
    items     []OrderItem  // unexported = cannot modify from outside
    status    OrderStatus
    createdAt time.Time
}

func NewOrder(id string, items []OrderItem) (*Order, error) {
    if len(items) == 0 {
        return nil, ErrEmptyOrder
    }
    return &Order{
        id:        id,
        items:     append([]OrderItem{}, items...),  // Defensive copy
        status:    StatusPending,
        createdAt: time.Now(),
    }, nil
}

// Items returns a COPY, not a reference
func (o *Order) Items() []OrderItem {
    return append([]OrderItem{}, o.items...)
}
```

### 4.5 Full Tech-Lead Review Checklist

```
## Go DDD Code Review Checklist

### Architecture
- [ ] No domain imports of application or infrastructure packages
- [ ] No framework types in domain signatures
- [ ] Repository interfaces defined in domain, implemented in infrastructure
- [ ] Compile-time interface satisfaction checks present
- [ ] `go-arch-lint check` passes

### Error Handling
- [ ] Every error return is checked
- [ ] errors.Is() used for sentinel errors (not ==)
- [ ] errors.As() used for error type checks (not type assertion)
- [ ] Errors wrapped with context: fmt.Errorf("...: %w", err)
- [ ] Domain errors defined in domain package
- [ ] Infrastructure errors translated to domain errors at boundary

### Encapsulation
- [ ] Domain entity fields are unexported
- [ ] Constructor functions (NewX) validate invariants
- [ ] Value objects have no setter methods
- [ ] Slices/maps returned as defensive copies
- [ ] No direct struct literal creation of domain types outside package

### context.Context
- [ ] First parameter of all I/O methods
- [ ] Propagated through the call chain (not created mid-chain)
- [ ] context.Background() only in main() or test setup
- [ ] context.TODO() has a comment explaining why

### Testing
- [ ] Table-driven tests for functions with multiple input scenarios
- [ ] require for setup preconditions, assert for verifications
- [ ] Mocks only for ports/interfaces, not for domain logic
- [ ] Edge cases covered (nil, empty, boundary values)
- [ ] No test pollution (tests do not depend on execution order)
```

---

## 5. Anti-Hallucination Guide for Go AI Agents

### 5.1 Common AI Mistakes When Writing Go Code

Based on research from [IEEE Spectrum](https://spectrum.ieee.org/ai-coding-degrades),
[Stack Overflow (2026)](https://stackoverflow.blog/2026/01/28/are-bugs-and-incidents-inevitable-with-ai-coding-agents/),
[CodeRabbit report](https://www.coderabbit.ai/blog/state-of-ai-vs-human-code-generation-report),
and [ACM on package hallucinations](https://cacm.acm.org/news/nonsense-and-malicious-packages-llm-hallucinations-in-code-generation/):

#### Category 1: Hallucinated Packages (CRITICAL)

AI agents invent non-existent packages or reference packages that were renamed, moved,
or never existed. In one study, 5.2% of AI-generated code samples referenced
hallucinated packages (GPT models), rising to 21.7% for open-source models.

**Common Go hallucinations:**
- `github.com/pkg/errors` (deprecated in favor of stdlib `errors` + `fmt.Errorf %w`)
- Inventing subpackages of popular libraries (e.g., `github.com/gin-gonic/gin/middleware`)
- Using `golang.org/x/` packages that do not exist
- Referencing old import paths after module renames

**Mitigation:**
```makefile
# In Makefile audit target
audit: vet lint test
	go mod tidy        # Removes unused, errors on missing
	go mod verify      # Verifies checksums of downloaded modules
```

The Go compiler itself catches hallucinated imports immediately: `go build ./...`
will fail if any import does not exist. This is Go's strongest anti-hallucination
defense.

#### Category 2: Wrong Function Signatures (HIGH)

AI confuses method signatures across versions or between similar packages:
- Wrong parameter order (especially `context.Context` not first)
- Wrong return types (returning `error` vs `(T, error)`)
- Using methods that do not exist on a type
- Confusing methods between similar types (`*sql.DB` vs `*sql.Tx`)

**Mitigation:** `go build ./...` catches ALL of these at compile time. Go's static
type system is the primary defense. No runtime surprises.

#### Category 3: Deprecated APIs (MEDIUM)

AI models are trained on older code and may use deprecated patterns:
- `io/ioutil` (deprecated since Go 1.16; use `io` and `os` directly)
- `golang.org/x/net/context` (use `context` from stdlib since Go 1.7)
- `github.com/pkg/errors` (stdlib `errors` + `fmt.Errorf %w` since Go 1.13)
- Old-style error checking (`err.Error() == "..."` instead of `errors.Is()`)

**Mitigation:** Enable these golangci-lint linters:
```yaml
linters:
  enable:
    - staticcheck    # SA1019 flags deprecated stdlib usage
    - gomodguard     # Block specific modules (e.g., github.com/pkg/errors)
```

#### Category 4: Not Handling Errors (CRITICAL)

AI-generated code nearly doubles the rate of missing error checks. The AI writes
code that compiles but silently ignores errors:

```go
// BAD: AI often generates this
result, _ := someFunction()   // Ignoring error!

// GOOD: Always handle errors
result, err := someFunction()
if err != nil {
    return fmt.Errorf("some context: %w", err)
}
```

**Mitigation:** The `errcheck` linter (included in golangci-lint `standard` preset)
catches ALL unhandled error returns. This is the single most important linter for
AI-generated Go code.

#### Category 5: Wrong Struct Initialization (MEDIUM)

AI creates struct literals with wrong field names or types, or uses positional
initialization (fragile):

```go
// BAD: Positional initialization (breaks when fields are added)
order := Order{"id-1", []Item{}, StatusPending, time.Now()}

// GOOD: Named field initialization
order := Order{
    id:     "id-1",
    items:  []Item{},
    status: StatusPending,
}

// BEST: Constructor with validation (for domain types)
order, err := NewOrder("id-1", items)
```

**Mitigation:** `govet` analyzer `composites` checks for unkeyed struct literals.
golangci-lint includes this by default.

#### Category 6: context.Context Misuse (HIGH)

AI frequently gets context usage wrong:

```go
// BAD: Creating context in the middle of a call chain
func (s *Service) Process(id string) error {
    ctx := context.Background()  // Should be passed in!
    return s.repo.FindByID(ctx, id)
}

// GOOD: Accept and propagate context
func (s *Service) Process(ctx context.Context, id string) error {
    return s.repo.FindByID(ctx, id)
}

// BAD: Using context.TODO() without explanation
ctx := context.TODO()

// GOOD: context.TODO() with explanation
ctx := context.TODO() // TODO: propagate from HTTP handler after #123
```

**Mitigation:** Enable `noctx` and `contextcheck` linters in golangci-lint.

#### Category 7: Concurrency Mistakes (HIGH)

AI-generated code is 2x more likely to misuse concurrency primitives:
- Data races (accessing shared data without sync)
- Goroutine leaks (starting goroutines without cleanup)
- Channel misuse (sending on closed channels, missing select default)

**Mitigation:** `go test -race ./...` (race detector) catches most data races.
Enable `gocritic` linter for common concurrency anti-patterns.

### 5.2 Anti-Hallucination Quality Gate Pipeline

The recommended pipeline catches AI mistakes in order of severity:

```bash
# Step 1: Compile (catches hallucinated imports, wrong signatures, type errors)
go build ./...

# Step 2: Static analysis (catches error handling, context misuse, deprecated APIs)
golangci-lint run

# Step 3: Tests with race detector (catches logic errors, concurrency bugs)
go test -race -count=1 ./...

# Step 4: Architecture check (catches layer violations)
go-arch-lint check
```

This pipeline is strictly stronger than the Python equivalent because Steps 1 catches
issues that Python can only find at runtime (or with mypy + strict mode + 100%
annotation coverage, which is rarely achieved).

### 5.3 What to Include in Agent Instructions (CLAUDE.md)

When configuring AI coding agents for a Go DDD project, include these rules in the
project's `CLAUDE.md` or equivalent:

```markdown
## Go-Specific Rules

### Error Handling
- ALWAYS check error returns. NEVER use `_` to discard errors.
- Use `errors.Is()` for sentinel errors, `errors.As()` for error types.
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Define domain errors in the domain package, not infrastructure.

### context.Context
- ALWAYS pass context.Context as the first parameter of I/O functions.
- NEVER create context.Background() except in main() or TestMain().
- ALWAYS propagate the context received from the caller.

### Imports
- Do NOT use `io/ioutil` (deprecated since Go 1.16).
- Do NOT use `github.com/pkg/errors` (use stdlib errors + fmt.Errorf).
- Do NOT use `golang.org/x/net/context` (use stdlib context).
- VERIFY that all imported packages exist before writing code.

### Struct Initialization
- Use named fields in struct literals (never positional).
- Use constructor functions (NewX) for domain types.
- Domain entity fields MUST be unexported (lowercase).

### Testing
- Use table-driven tests with named subtests.
- Use `require` for setup, `assert` for verification.
- Run `go test -race ./...` to catch data races.

### Quality Gates (all must pass)
- `go build ./...`          # Zero compile errors
- `golangci-lint run`       # Zero lint errors
- `go test -race ./...`     # All tests pass, no races
- `go-arch-lint check`      # No architecture violations
```

---

## Recommendation

For alto's Go project templates, generate the following quality gate pipeline as
the default:

1. **Makefile** with `build`, `test`, `lint`, `vet`, `fmt`, `audit`, `arch-lint`,
   `clean`, and `help` targets
2. **`.golangci.yml`** (v2 format) with the DDD-focused linter configuration above
3. **`.go-arch-lint.yml`** with layer boundary rules matching the DDD structure
4. **`CLAUDE.md`** with the Go-specific agent instructions above

The compile step (`go build ./...`) is the single most powerful anti-hallucination
tool because it catches hallucinated imports, wrong function signatures, and type
errors instantly. Combined with `errcheck` (via golangci-lint) for unchecked errors
and `-race` for concurrency bugs, this pipeline catches the vast majority of
AI-generated Go code mistakes before they reach code review.

## References

- [ThreeDotsLabs wild-workouts-go-ddd-example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example) -- Reference Go DDD project
- [ThreeDotsLabs DDD Lite introduction](https://threedots.tech/post/ddd-lite-in-go-introduction/) -- DDD patterns in Go
- [Go official module layout](https://go.dev/doc/modules/layout) -- Official directory structure guide
- [Damiano Petrungaro: DDD in Go](https://www.damianopetrungaro.com/posts/ddd-how-i-structure-idiomatic-golang-services/) -- Idiomatic Go DDD with internal/ pattern
- [golangci-lint docs](https://golangci-lint.run/docs/configuration/) -- Linter configuration
- [golangci-lint migration guide v1 to v2](https://golangci-lint.run/docs/product/migration-guide/) -- v2 format changes
- [gofumpt](https://github.com/mvdan/gofumpt) -- Stricter Go formatter
- [go-arch-lint](https://github.com/fe3dback/go-arch-lint) -- Architecture boundary linter (MIT, 303 stars)
- [arch-go](https://github.com/arch-go/arch-go) -- Architecture testing tool (MIT, ~250 stars, v2.1.2)
- [testify](https://github.com/stretchr/testify) -- Go testing toolkit (MIT, 26k stars)
- [Alex Edwards: Makefile for Go](https://www.alexedwards.net/blog/a-time-saving-makefile-for-your-go-projects) -- Makefile patterns
- [Earthly: Golang Makefile](https://earthly.dev/blog/golang-makefile/) -- Makefile conventions
- [JetBrains: Secure Go Error Handling (2026)](https://blog.jetbrains.com/go/2026/03/02/secure-go-error-handling-best-practices/) -- Error handling patterns
- [IEEE Spectrum: AI Coding Degrades](https://spectrum.ieee.org/ai-coding-degrades) -- AI code quality issues
- [Stack Overflow: AI Coding Agent Bugs (2026)](https://stackoverflow.blog/2026/01/28/are-bugs-and-incidents-inevitable-with-ai-coding-agents/) -- AI bug patterns
- [CodeRabbit: AI vs Human Code Report](https://www.coderabbit.ai/blog/state-of-ai-vs-human-code-generation-report) -- AI generates 1.7x more bugs
- [ACM: LLM Hallucinations in Code](https://cacm.acm.org/news/nonsense-and-malicious-packages-llm-hallucinations-in-code-generation/) -- Package hallucination risks
- [Simon Willison: Hallucinations in Code](https://simonwillison.net/2025/Mar/2/hallucinations-in-code/) -- Hallucination analysis
- [Go testing package](https://pkg.go.dev/testing) -- Official test documentation
- [Go vet](https://pkg.go.dev/cmd/vet) -- Official static analysis

## Follow-up Tasks

- [ ] Create ticket: Implement Go project template with Makefile, `.golangci.yml`, `.go-arch-lint.yml`, and DDD directory structure
- [ ] Create ticket: Add Go-specific CLAUDE.md template section with anti-hallucination rules
- [ ] Create ticket: Add Go quality gate commands to alto's config generation (replaces ruff/mypy/pytest equivalents)
- [ ] Create ticket: Implement Go table-driven test template generation for domain entities
- [ ] Create ticket: Add Go DDD code review checklist to tech-lead agent persona
