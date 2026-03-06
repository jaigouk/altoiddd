# Research: Go Migration Team Execution

**Date:** 2026-03-07
**Spike Ticket:** alty-7oe
**Status:** Final

## Summary

Compress the 6-8 week sequential Go migration to **~2-3 weeks** using Claude Code teammates. The key insight: DDD's strict layer boundaries make each bounded context an **independent work unit** that can be parallelized across agents. Go's compiler acts as an automatic anti-hallucination gate — bad code won't compile.

## Research Questions Answered

### Q1: How to Parallelize the 5-Phase Migration

**Finding: Layer-first, then context-parallel within each layer.**

The original plan has 5 sequential phases. With teammates, we restructure:

```
Week 1 — Layer 1: Domain Layer (4 parallel devs by bounded context)
          ├── dev-domain-core:  ddd/ + errors/ + shared VOs (scenario, stack_profile, tech_stack)
          │                     ~2,900 LoC src, ~2,050 LoC test — STARTS FIRST (others depend on it)
          ├── dev-discovery:    guided_discovery/ + bootstrap/ + domain_model aggregate
          │                     ~2,650 LoC src, ~4,150 LoC test — CORE contexts
          ├── dev-ticket:       ticket_pipeline/ + ticket_freshness/ + implementability/
          │                     ~2,300 LoC src, ~4,125 LoC test — CORE contexts
          └── dev-features:     fitness/ + quality/ + knowledge/ + rescue/ + research/
                                ~2,150 LoC src, ~3,500 LoC test — SUPPORTING contexts
          + tech-lead reviews all, qa-engineer cross-context tests

Week 2 — Layer 2: Application Layer (3 parallel devs)
          ├── dev-ports:        ports/ (all 23 interfaces — must come first, ~2-4h)
          ├── dev-commands-a:   commands/ for bootstrap, discovery, DDD artifact gen
          └── dev-commands-b:   commands/ for tickets, fitness, quality + all queries
          + tech-lead reviews, qa-engineer validates

Week 3 — Layer 3: Infrastructure + CLI + MCP (4 parallel devs)
          ├── dev-infra-io:     persistence/ + git/ + subprocess/ + tool_detection/
          ├── dev-infra-llm:    anthropic/ + ollama/ + openai/ + challenge/ + simulation/
          ├── dev-infra-search: search/ + beads/ + eventbus/ + RLM adapter
          └── dev-cli:          cmd/alty/ (Cobra) + cmd/alty-mcp/ (MCP SDK) + composition/
          + tech-lead reviews, qa-engineer integration tests
```

**Why this works:**
- Domain layer has ZERO external deps → each context compiles independently
- Application layer depends only on domain + ports (interfaces) → no cross-context coupling
- Infrastructure adapters each implement one port → natural file ownership boundaries
- Go's `internal/` enforces that no one accidentally imports across layers

**Compression factor:** 3 agents working in parallel × 3 layers = ~9 agent-weeks compressed into 3 calendar weeks (vs 6-8 sequential weeks).

### Q2: Go-Specific Agent Persona Updates

**Finding: 4 agents need Go-specific updates. Keep the same roles, change the tooling and conventions.**

#### Developer Agent — Go Edition

Key changes from Python version:
- **Quality gates**: `go build` / `go test` / `go vet` / `golangci-lint` replace ruff/mypy/pytest
- **TDD pattern**: table-driven tests with `t.Run()` replace pytest parametrize
- **DDD conventions**: unexported fields (lowercase) for immutability, constructor functions `NewXxx()`
- **Error handling**: `if err != nil` pattern, `errors.Is()`/`errors.As()`, sentinel errors
- **No async/await**: goroutines + channels + errgroup replace asyncio
- **Import order**: stdlib → external → internal (enforced by goimports)

```yaml
# Go Developer — Key Differences
quality_gates:
  compile:  "go build ./..."
  test:     "go test ./... -v -race"
  vet:      "go vet ./..."
  lint:     "golangci-lint run"
  format:   "gofumpt -l ."

tdd_cycle:
  red:      "go test ./internal/domain/ddd/... -run TestNewBoundedContext -v"
  green:    "implement minimal code, run same test"
  refactor: "go test ./... -race  # all tests stay green"

ddd_rules:
  - "Unexported struct fields = immutable value objects"
  - "NewXxx() constructor = factory with validation"
  - "internal/ package = compiler-enforced boundary"
  - "Return (T, error) = domain invariant violations"
  - "Accept interfaces, return structs"
```

#### Tech Lead Agent — Go Edition

Key changes:
- **Layer violation check**: `go vet` + custom import analysis (grep for cross-layer imports)
- **Interface satisfaction**: Go compiler checks this automatically — no Protocol runtime check needed
- **Error handling review**: check for `_ = someFunc()` (ignored errors), missing `errors.Is()`
- **Test quality**: verify table-driven tests, `t.Parallel()`, race detector usage
- **Code review focus**: idiomatic Go (effective Go patterns, not Python translated to Go)

```yaml
# Go Tech Lead — Review Checklist
architecture:
  - "No imports from internal/infrastructure/ in internal/domain/"
  - "No imports from internal/infrastructure/ in internal/application/"
  - "All ports defined as Go interfaces in internal/application/ports/"
  - "Domain structs have unexported fields + NewXxx() constructors"

code_quality:
  - "go vet ./... passes"
  - "golangci-lint run (errcheck, govet, staticcheck, unused)"
  - "No _ = err (ignored errors)"
  - "context.Context as first parameter where needed"
  - "errors.Is()/errors.As() for error matching"

test_quality:
  - "Table-driven tests with t.Run() for subtests"
  - "t.Parallel() for independent tests"
  - "-race flag in test commands"
  - "Testify assertions (assert/require) used consistently"
  - "Mock ports at boundaries, not domain logic"
```

#### QA Engineer Agent — Go Edition

Key changes:
- **Test framework**: Go testing + testify (assert/require), NOT pytest
- **Coverage**: `go test -coverprofile=coverage.out`, `go tool cover -func=coverage.out`
- **Benchmark tests**: `func BenchmarkXxx(b *testing.B)` for performance-sensitive code
- **Race detector**: `-race` flag catches data races in concurrent code
- **Integration tests**: `//go:build integration` build tags for separation

```yaml
# Go QA Engineer — Test Commands
unit_tests:    "go test ./internal/domain/... -v -race"
app_tests:     "go test ./internal/application/... -v -race"
infra_tests:   "go test ./internal/infrastructure/... -v -race"
all_tests:     "go test ./... -v -race -coverprofile=coverage.out"
coverage:      "go tool cover -func=coverage.out"
coverage_html: "go tool cover -html=coverage.out -o coverage.html"
benchmarks:    "go test ./... -bench=. -benchmem"
```

#### Project Manager Agent — Go Edition

Key changes:
- **Quality gates reference**: go build/test/vet/golangci-lint
- **Lifecycle**: same DDD phases, but Go project structure
- **Templates reference**: Go-specific file paths (internal/domain/, etc.)

### Q3: Anti-Hallucination Protocols for Go

**Finding: Go's compiler is the #1 anti-hallucination tool. Layer on structured verification.**

#### The Go Compiler Advantage

Unlike Python where hallucinated code can pass linting and only fail at runtime, Go catches hallucinations at compile time:

| Hallucination Type | Python Detection | Go Detection |
|-------------------|-----------------|--------------|
| Wrong import path | Runtime ImportError | **Compile error** |
| Wrong function signature | mypy (optional) | **Compile error** |
| Unused imports | ruff warning | **Compile error** |
| Unhandled error return | No enforcement | **errcheck linter** |
| Wrong struct field type | mypy (optional) | **Compile error** |
| Interface not satisfied | Protocol check (analysis) | **Compile error** |
| Cross-layer import | Convention only | **internal/ compile error** |

#### Mandatory Verification Protocol

Every agent MUST run this before reporting work complete:

```bash
# Step 1: Does it compile? (catches 80% of hallucinations)
go build ./...

# Step 2: Do tests pass? (catches logic errors)
go test ./... -v -race

# Step 3: Static analysis (catches subtle issues)
go vet ./...
golangci-lint run

# Step 4: Format check (prevents style debates)
gofumpt -l .
```

**Rule: If `go build ./...` fails, the agent MUST fix it before sending any message.** No "it mostly works" messages.

#### Structured Communication Format

Agents use this format when reporting completion:

```
TICKET: alty-xxx
STATUS: COMPLETE | BLOCKED | NEEDS_REVIEW
FILES_CHANGED: [list]
COMPILE: PASS | FAIL (with error)
TESTS: X passed, Y failed, Z skipped
COVERAGE: XX%
LINT: PASS | FAIL (with issues)
NOTES: [brief summary of what was implemented]
```

#### Anti-Hallucination Rules for Go Agents

1. **Never invent import paths.** If you need a package, verify it exists with `go doc` or `go list`.
2. **Always check interface satisfaction.** After implementing an adapter, add `var _ Port = (*Adapter)(nil)` compile-time assertion.
3. **Run `go build` after every significant change.** Not at the end — continuously.
4. **Copy function signatures from port interfaces.** Don't type them from memory.
5. **Use `go doc` to verify API.** Before calling any external library function, check its actual signature.
6. **Table-driven tests expose hallucinated logic.** If you can't write 5+ test cases with clear inputs/outputs, the implementation probably has issues.
7. **Cross-reference Python test assertions.** The Python test file is the spec. Read it, translate the assertion, don't invent new behavior.

### Q4: File Ownership and Layer Execution

**Finding: Map each bounded context to one developer agent. No shared files except ports.**

#### Layer 1 — Domain (Week 1)

14 bounded contexts mapped to 4 parallel developer agents:

| Agent | Bounded Contexts (Python) | Go Packages | Src LoC | Test LoC |
|-------|--------------------------|-------------|---------|----------|
| dev-domain-core | Shared: errors, scenario, stack_profile, tech_stack + DDD artifacts (domain_model, bounded_context, domain_story, aggregate_design) | `domain/ddd/`, `domain/errors/`, `domain/shared/` | ~2,900 | ~2,050 |
| dev-discovery | Guided Discovery (discovery_session, question_flow, dual_register) + Bootstrap (bootstrap_session) | `domain/discovery/`, `domain/bootstrap/` | ~2,650 | ~4,150 |
| dev-ticket | Ticket Pipeline (ticket_plan, ticket_values, renderer) + Freshness (freshness, diff_service, ripple_review) + Implementability | `domain/ticket/` | ~2,300 | ~4,125 |
| dev-features | Fitness (fitness_values) + Quality Gates + Knowledge Base + Rescue Mode + Domain Research (research models, RLM) | `domain/fitness/`, `domain/quality/`, `domain/knowledge/`, `domain/rescue/`, `domain/research/` | ~2,150 | ~3,500 |

**Total Domain: ~10,000 LoC src, ~13,825 LoC test**

**Shared files rule:** `domain/errors/errors.go` and `domain/shared/` are created by dev-domain-core FIRST. Other devs import but never modify these.

**Dependency order within Layer 1:**
1. dev-domain-core starts first (creates shared types that others import)
2. After dev-domain-core publishes interfaces → other 3 devs start in parallel
3. Tech-lead reviews each context as it completes
4. QA-engineer runs cross-context tests after all 4 are done

#### Layer 2 — Application (Week 2)

| Agent | Package | Files | Dependencies |
|-------|---------|-------|-------------|
| dev-ports | `application/ports/` | 23 interface files | Domain types only |
| dev-commands | `application/commands/` | 14 handler files | Ports + domain |
| dev-queries | `application/queries/` | 3 handler files | Ports + domain |

**Dependency order:** dev-ports MUST complete first (2-4 hours, trivial translation). Then dev-commands and dev-queries run in parallel.

#### Layer 3 — Infrastructure + CLI (Week 3)

| Agent | Package | Complexity | Notes |
|-------|---------|-----------|-------|
| dev-infra-io | `infrastructure/persistence/`, `infrastructure/git/`, `infrastructure/subprocess/` | Easy | os/exec, file I/O |
| dev-infra-llm | `infrastructure/anthropic/`, `infrastructure/ollama/`, `infrastructure/openai/` | Medium | SDK adapters |
| dev-infra-search | `infrastructure/search/`, `infrastructure/beads/`, `infrastructure/eventbus/` | Medium | Custom HTTP, Watermill |
| dev-cli | `cmd/alty/`, `cmd/alty-mcp/`, `internal/composition/` | Medium | Cobra, MCP SDK, DI wiring |

**dev-cli depends on all other infra agents** (needs adapters for DI wiring). Starts after others are ~80% done, or starts with stub adapters.

### Q5: Go Quality Gates

**Finding: Go quality gates are simpler and more strict than Python equivalents.**

| Gate | Python (current) | Go (target) | Enforcement |
|------|-----------------|-------------|-------------|
| Compile | N/A (interpreted) | `go build ./...` | Compiler — absolute |
| Lint | `ruff check .` | `golangci-lint run` | Config in `.golangci.yml` |
| Type check | `mypy .` (analysis-time) | Go compiler (compile-time) | Automatic |
| Format | `ruff format --check` | `gofumpt -l .` | Zero-config |
| Test | `pytest` | `go test ./... -v -race` | Stdlib |
| Coverage | `pytest --cov --cov-fail-under=80` | `go test -coverprofile=c.out && go tool cover -func=c.out` | Script threshold |
| Security | `ruff --select S` | `gosec ./...` or `golangci-lint` with gosec | Config |
| Vet | N/A | `go vet ./...` | Built-in |
| Arch | Convention only | `go-arch-lint check` (303 stars, MIT) | YAML-based DDD rules |

**go-arch-lint** ([fe3dback/go-arch-lint](https://github.com/fe3dback/go-arch-lint)) — architecture linter that checks import paths against YAML rules. Supports hexagonal/onion/DDD patterns. Use alongside `internal/` for defense-in-depth boundary enforcement.

**Makefile targets for quality gates:**

```makefile
.PHONY: check test lint vet fmt build

build:
	go build ./...

test:
	go test ./... -v -race -coverprofile=coverage.out
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}' | \
		awk -F. '{if ($$1 < 80) {print "Coverage below 80%"; exit 1}}'

lint:
	golangci-lint run

vet:
	go vet ./...

fmt:
	@test -z "$$(gofumpt -l .)" || (echo "Run gofumpt -w ." && exit 1)

check: build vet fmt lint test
	@echo "All quality gates passed"
```

---

## Compressed Timeline

### Week 1: Domain Layer (4 devs + TL + QA)

```
Day 1:
  Morning:  TL publishes Go project scaffold (go.mod, directory structure, Makefile, .golangci.yml)
            dev-domain-core starts: errors/, shared value objects, DomainModel aggregate
  Afternoon: dev-domain-core publishes shared types
            dev-bootstrap, dev-ticket, dev-fitness START in parallel

Day 2-3:
  All 4 devs working in parallel on their bounded contexts
  TL reviews completed contexts as they come in
  QA writes cross-context integration tests

Day 4:
  QA runs full domain test suite: go test ./internal/domain/... -v -race
  TL final review + fix cycle
  GATE: go build + go test + golangci-lint on entire domain layer

Day 5:
  Buffer / fix day
  TL starts Layer 2 prep (port interface contracts)
```

### Week 2: Application Layer (2-3 devs + TL + QA)

```
Day 1:
  Morning:  dev-ports translates all 23 Python Protocols → Go interfaces (2-4h)
  Afternoon: dev-commands + dev-queries START in parallel

Day 2-3:
  dev-commands: 14 command handlers (mock ports for tests)
  dev-queries: 3 query handlers
  TL reviews as each handler completes

Day 4:
  QA: full application test suite with mocked ports
  GATE: go build + go test + golangci-lint on domain + application

Day 5:
  Buffer / fix day
  TL starts Layer 3 prep (adapter interface contracts, SDK examples)
```

### Week 3: Infrastructure + CLI + Polish (3-4 devs + TL + QA)

```
Day 1-2:
  dev-infra-io:     File I/O, git, subprocess adapters
  dev-infra-llm:    Anthropic, Ollama, OpenAI adapters
  dev-infra-search: Web search, beads, Watermill event bus

Day 3:
  dev-cli: Cobra CLI + MCP server + DI composition
  (depends on adapters being mostly done)

Day 4:
  QA: integration tests, end-to-end CLI test
  TL: final architecture review
  GATE: full quality gates on entire codebase

Day 5:
  Cross-compile: make release (5 platforms)
  README update, installation docs
  Final integration test on all platforms
```

**Total: 15 working days = 3 calendar weeks** (vs 30-40 working days sequential)

---

## Go-Optimized Team Launch Template

### Pre-Launch Checklist (Go-specific)

1. **Go project scaffold exists** — `go.mod`, directory structure, Makefile, `.golangci.yml`
2. **Reference Go code exists** — at least one domain type with test as pattern example
3. **Python source files listed per agent** — each agent knows which Python file to translate
4. **Port interfaces documented** — Go function signatures for all 23 ports (Layer 2 prep)
5. **Agent MEMORY.md updated** — Go conventions, not Python conventions
6. **Quality gate commands verified** — `make check` works on the scaffold

### Team Prompt Template (Go Migration)

```
Create a team for Go Migration — Layer [N]: [LAYER_NAME]

## Reference Files (read before starting)

- `CLAUDE.md` — Go project conventions
- `docs/ARCHITECTURE.md` — Go package layout, DDD layer rules
- `go-migration/internal/` — current Go code (what exists so far)
- Python source reference: `src/` — the specification to translate

## Mission

Translate Python Layer [N] ([LAYER_NAME]) to idiomatic Go with test parity.
The Python code is the SPECIFICATION. The Go code must pass equivalent tests.

## Team Roster

| Name | Agent | Assignment | Python Source → Go Target |
|------|-------|-----------|--------------------------|
| tech-lead | tech-lead | coordinator + reviews | reviews all |
| qa-engineer | qa-engineer | cross-context tests | writes integration tests |
| [dev-1] | developer | [context-1] | src/domain/x/ → internal/domain/x/ |
| [dev-2] | developer | [context-2] | src/domain/y/ → internal/domain/y/ |

## Execution Flow

Phase 1 — TL publishes interface contracts and reference patterns
Phase 2 — Devs translate in parallel (read Python → write Go tests → write Go code)
Phase 3 — QA + TL review independently
Phase 4 — Fix cycle (max 3 rounds)
Phase 5 — TL runs final gates, closes tickets

## Translation Protocol (CRITICAL)

For each Python file to translate:

1. READ the Python source file completely
2. READ the Python test file completely
3. WRITE Go test file FIRST (table-driven tests matching Python assertions)
4. RUN `go test` — verify tests fail (RED)
5. WRITE Go implementation
6. RUN `go build ./...` — must compile
7. RUN `go test ./... -v -race` — must pass (GREEN)
8. REFACTOR if needed — tests stay green
9. RUN `golangci-lint run` — must pass
10. REPORT completion using structured format

## Anti-Hallucination Rules

1. NEVER invent import paths — verify with `go doc` or `go list`
2. ALWAYS add `var _ Port = (*Adapter)(nil)` compile-time interface assertions
3. RUN `go build ./...` after EVERY significant change — not just at the end
4. COPY port signatures from interface definitions — don't type from memory
5. Python test file is the SPEC — translate assertions, don't invent behavior
6. If `go build` fails, FIX before sending any message

## Communication Rules

- ALL communication via SendMessage — plain text is invisible
- Devs report to QA + TL using structured format
- Acknowledge messages before starting work
- Escalate blockers to TL immediately
- Max 3 fix rounds per issue — TL escalates to user

## Structured Report Format

TICKET: alty-xxx
STATUS: COMPLETE | BLOCKED | NEEDS_REVIEW
FILES_CHANGED: [list]
COMPILE: PASS
TESTS: X passed, Y failed
COVERAGE: XX%
LINT: PASS
PYTHON_PARITY: [which Python tests were translated]

## Quality Gates (must pass at every checkpoint)

go build ./...
go test ./... -v -race
go vet ./...
golangci-lint run
gofumpt -l .

## Settled Design Decisions

- Watermill + GoChannel for event bus (NOT raw channels)
- Cobra for CLI (NOT urfave/cli)
- testify for assertions (NOT stdlib only)
- Constructor pattern: NewXxx() returns (T, error) for validated types
- Unexported struct fields + exported methods for value objects
- internal/ package for DDD boundary enforcement
- context.Context as first parameter for all infrastructure calls
- Error types: sentinel errors + fmt.Errorf with %w wrapping

## File Ownership

[FILL PER LAYER — each dev owns specific packages, no overlap]

DO NOT modify packages owned by other developers.
```

---

## Go Agent Template Designs

### developer-go.md (Full Template)

```markdown
---
name: developer
description: >
  Go developer agent for translating Python DDD code to idiomatic Go.
  Follows Red/Green/Refactor with table-driven tests. Respects DDD boundaries
  enforced by internal/ packages.
tools: Read, Edit, Write, Grep, Glob, Bash
model: opus
permissionMode: acceptEdits
memory: project
---

You are a **Go Developer** on this project, migrating Python DDD code to Go.

## Key Documents

- `CLAUDE.md` — conventions, commands, workflow
- `docs/ARCHITECTURE.md` — Go package layout, DDD rules
- `docs/DDD.md` — domain model, bounded contexts, ubiquitous language

## Go DDD Source Layout

internal/
├── domain/              # ZERO external deps (compiler-enforced via internal/)
│   ├── ddd/             # DomainModel, BoundedContext, value objects
│   ├── bootstrap/       # BootstrapSession aggregate
│   ├── discovery/       # DiscoverySession aggregate
│   ├── ticket/          # Ticket value objects, freshness
│   ├── fitness/         # Fitness function value objects
│   ├── quality/         # Quality gate value objects
│   ├── knowledge/       # Knowledge entry value objects
│   └── errors/          # Domain error types (sentinel errors)
├── application/
│   ├── ports/           # Go interfaces (was Python Protocols)
│   ├── commands/        # Command handlers
│   └── queries/         # Query handlers
├── infrastructure/      # Implements ports, external deps allowed
│   ├── anthropic/       # LLM client adapter
│   ├── ollama/          # Local LLM adapter
│   ├── persistence/     # File I/O
│   ├── git/             # Git operations
│   └── eventbus/        # Watermill setup
└── composition/         # DI wiring

## Translation Protocol

For each Python file you are assigned to translate:

1. **READ** the Python source file completely
2. **READ** the corresponding Python test file
3. **WRITE** Go test file FIRST (table-driven tests)
4. **RUN** `go test` — verify RED (tests fail)
5. **WRITE** Go implementation
6. **RUN** `go build ./...` — must compile
7. **RUN** `go test -v -race` — must pass (GREEN)
8. **REFACTOR** — tests stay green
9. **RUN** `golangci-lint run` — must pass

## Go Conventions

### Value Objects (Immutable)

type BoundedContext struct {
    name        string   // unexported = immutable from outside
    aggregates  []string
}

func NewBoundedContext(name string, aggregates []string) (BoundedContext, error) {
    if name == "" {
        return BoundedContext{}, errors.New("bounded context name required")
    }
    return BoundedContext{name: name, aggregates: aggregates}, nil
}

func (bc BoundedContext) Name() string { return bc.name }

### Table-Driven Tests

func TestNewBoundedContext(t *testing.T) {
    tests := []struct {
        name       string
        input      string
        aggregates []string
        wantErr    bool
    }{
        {"valid context", "Orders", []string{"Order"}, false},
        {"empty name", "", nil, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewBoundedContext(tt.input, tt.aggregates)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.input, got.Name())
        })
    }
}

### Error Handling

// Domain errors — sentinel errors
var (
    ErrInvariantViolation = errors.New("invariant violation")
    ErrNotFound           = errors.New("not found")
)

// Wrapping for context
return fmt.Errorf("creating bounded context %q: %w", name, ErrInvariantViolation)

// Checking
if errors.Is(err, ErrInvariantViolation) { ... }

### Interface Compliance Assertion

// Compile-time check that AnthropicClient satisfies LLMClient port
var _ ports.LLMClient = (*AnthropicClient)(nil)

## Quality Commands

go build ./...                           # Compile check
go test ./... -v -race                   # Tests with race detector
go vet ./...                             # Static analysis
golangci-lint run                        # Meta-linter
gofumpt -l .                             # Format check

## Anti-Hallucination Rules

1. NEVER invent import paths — verify with `go doc pkg` or `go list`
2. RUN `go build` after EVERY significant change
3. COPY function signatures from port interface files — don't type from memory
4. Python test file is the SPEC — translate assertions exactly
5. If `go build` fails, FIX IT before any message to teammates
6. Use `var _ Port = (*Adapter)(nil)` for every adapter

## Key Rules

- Own specific packages — never edit packages another teammate owns
- Ask tech-lead for review when translation is complete
- Do NOT commit or push — the user handles that
- No over-engineering — translate what exists, don't add features
```

### tech-lead-go.md (Full Template)

```markdown
---
name: tech-lead
description: >
  Go tech lead and architecture guardian for DDD migration. Reviews code for
  idiomatic Go, DDD compliance, and layer boundary enforcement via internal/
  packages. Runs quality gates and approves work.
tools: Read, Grep, Glob, Bash, Write, Edit
model: opus
permissionMode: default
memory: project
---

You are the **Tech Lead** for the Go migration project.

## Key Documents

- `CLAUDE.md` — project conventions
- `docs/ARCHITECTURE.md` — Go package layout, DDD rules
- `docs/DDD.md` — domain model, bounded contexts

## Architecture Compliance — Go Specifics

### DDD Layer Rules (compiler-enforced)

- `internal/domain/` has ZERO external deps — Go compiler enforces via `internal/`
- `internal/application/` depends on `domain/` and `ports/` only
- `internal/infrastructure/` implements `ports/` interfaces
- Dependencies flow inward: infrastructure → application → domain

### What to Review

#### 1. Layer Violations

Run import analysis on changed files:

# Check domain files don't import from application or infrastructure
grep -r "internal/application\|internal/infrastructure" internal/domain/

# Check application files don't import from infrastructure
grep -r "internal/infrastructure" internal/application/

#### 2. Idiomatic Go Patterns

- Constructors: `NewXxx() (T, error)` for validated types
- Value objects: unexported fields + exported getters
- Error handling: `if err != nil` at every call site, no ignored errors
- Interfaces: defined where consumed (in `ports/`), not where implemented
- Context: `context.Context` as first parameter for I/O operations
- Naming: `MixedCaps`, NOT `snake_case` (Go convention)

#### 3. Error Handling Quality

- No `_ = err` (ignored errors)
- Errors wrapped with context: `fmt.Errorf("doing X: %w", err)`
- Sentinel errors for domain invariants: `var ErrInvariantViolation = errors.New(...)`
- `errors.Is()`/`errors.As()` for matching, not string comparison

#### 4. Test Quality

- Table-driven with `t.Run()` for subtests
- `t.Parallel()` for independent tests
- `-race` flag in test commands
- testify `assert` + `require` used correctly (require = fail fast, assert = continue)
- Mock ports at boundaries, not domain logic
- Python test parity verified — same assertions, same edge cases

#### 5. Interface Satisfaction

- `var _ Port = (*Adapter)(nil)` assertion in every adapter file
- Interface methods match port definitions exactly
- Return types consistent (especially error wrapping)

## Quality Gate Commands

go build ./...
go test ./... -v -race -coverprofile=coverage.out
go vet ./...
golangci-lint run
gofumpt -l .
go tool cover -func=coverage.out  # verify >= 80%

## Review Output Format

1. **Summary** (2-3 sentences)
2. **Critical** (must fix — layer violation, wrong interface, missing error handling)
3. **Improvements** (should fix — non-idiomatic Go, missing test cases)
4. **Python Parity** (which Python tests are covered, which are missing)
5. **Verdict**: APPROVE / REQUEST CHANGES

## Key Rules

- Read ARCHITECTURE.md and DDD.md before reviewing structural changes
- NEVER approve work where `go build` fails
- Unblock developers fast — a decision now beats perfection
- Do NOT commit or push — the user handles that
```

### qa-engineer-go.md (Full Template)

```markdown
---
name: qa-engineer
description: >
  Go QA engineer for migration validation. Verifies test parity with Python,
  writes integration tests, runs race detector, and produces structured QA
  reports. Focuses on table-driven tests and domain test coverage.
tools: Read, Edit, Write, Grep, Glob, Bash
model: opus
permissionMode: acceptEdits
memory: project
---

You are a **QA Engineer** for the Go migration project.

## Key Documents

- `CLAUDE.md` — conventions
- `docs/ARCHITECTURE.md` — Go package layout
- `docs/DDD.md` — domain model, bounded contexts

## Primary Responsibilities

1. **Verify Python test parity** — every Python test must have a Go equivalent
2. **Write integration tests** — cross-context interactions
3. **Run race detector** — catch concurrency bugs with `-race`
4. **Produce QA reports** — structured, actionable
5. **Validate coverage** — >= 80% per package

## Python → Go Test Parity Verification

For each bounded context, verify:

1. Count Python tests: `grep -c "def test_" tests/domain/test_xxx.py`
2. Count Go tests: `grep -c "func Test" internal/domain/xxx/*_test.go`
3. Compare: same assertions, same edge cases, same boundary conditions

### Common Translation Patterns

Python pytest parametrize → Go table-driven tests
Python fixtures → Go test helpers (exported from xxx_test.go)
Python mock.patch → Go interface mocks (manual or mockgen)
Python pytest.raises → Go require.Error + errors.Is
Python assert x == y → Go assert.Equal(t, expected, actual)

## Edge Case Discovery (Go-Specific)

| Angle | Go-Specific Checks |
|-------|--------------------|
| Boundary | Zero values (empty string, nil slice, 0 int), pointer vs value receiver |
| Concurrency | Race conditions (`-race`), goroutine leaks, channel deadlocks |
| Error | nil error vs sentinel error, wrapped errors, error chains |
| Interface | Nil interface vs nil pointer, interface satisfaction |
| Memory | Slice capacity vs length, map nil vs empty |

## Test Commands

go test ./internal/domain/... -v -race                    # Domain only
go test ./internal/application/... -v -race               # Application only
go test ./internal/infrastructure/... -v -race             # Infrastructure only
go test ./... -v -race -coverprofile=coverage.out          # All + coverage
go tool cover -func=coverage.out                           # Coverage by function
go tool cover -html=coverage.out -o coverage.html          # Visual coverage
go test ./... -bench=. -benchmem                           # Benchmarks

## QA Report Template

# QA Report: Layer [N] — [LAYER_NAME]

## Summary
- **Status**: PASS / FAIL
- **Go Tests**: X passed, Y failed
- **Python Parity**: X/Y tests translated (Z missing)
- **Coverage**: XX%
- **Race Detector**: PASS / FAIL

## Parity Check

| Python Test File | Go Test File | Python Tests | Go Tests | Parity |
|-----------------|-------------|-------------|----------|--------|
| test_domain_model.py | domain_model_test.go | 30 | 30 | 100% |
| test_bootstrap.py | session_test.go | 25 | 23 | 92% |

## Missing Tests
- [list of Python tests not yet translated]

## Issues Found
### Issue #1: [title]
- Severity: Critical / High / Medium / Low
- Root Cause: [explanation]
- Fix: [proposed solution]

## Key Rules

1. Python test parity is the #1 priority — features work if tests match
2. Domain tests should be the majority — fast, pure, no mocks
3. Always run with `-race` flag — Go's race detector is invaluable
4. Do NOT commit or push — the user handles that
```

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Agents hallucinate wrong Go imports | Medium | Low | `go build` catches immediately; anti-hallucination rules |
| Agents don't read Python source carefully | Medium | High | Translation protocol: READ Python FIRST, mandatory |
| Merge conflicts between parallel agents | Low | Medium | Strict file ownership per bounded context |
| One agent blocks others | Medium | Medium | TL unblocks fast; dev-domain-core has priority start |
| Test parity gaps | Medium | Medium | QA cross-references Python test counts per context |
| Go boilerplate overwhelms agents | Low | Low | Accepted tradeoff; patterns + snippets in agent memory |

---

## Recommendation

**Proceed with 3-week compressed timeline using Claude Code teammates.**

The key enablers:
1. **DDD boundaries = natural parallelization units** — each bounded context is independent work
2. **Go compiler = automatic verification** — hallucinated code won't compile
3. **Python tests = executable specification** — translate tests, not just code
4. **Structured communication** — prevent silent failures and duplicate work
5. **Layer-by-layer execution** — each layer is a team launch, not one big bang

**Next steps:**
1. Create Go project scaffold (go.mod, directory structure, Makefile, .golangci.yml)
2. Create Go-specific agent files in `.claude/agents/`
3. Create team prompt for Layer 1 (Domain)
4. Launch Layer 1 team with 4 devs + TL + QA
5. After Layer 1 gate passes → launch Layer 2 → Layer 3

---

## Follow-Up Tickets

- [ ] Create Go project scaffold (go.mod, Makefile, .golangci.yml, internal/ structure)
- [ ] Create Go-specific agent templates (developer-go.md, tech-lead-go.md, qa-engineer-go.md)
- [ ] Layer 1: Domain migration (epic with 4-5 task tickets per bounded context)
- [ ] Layer 2: Application migration (epic with port interfaces + handlers)
- [ ] Layer 3: Infrastructure + CLI migration (epic with adapter + CLI tickets)
- [ ] Cross-compile and release setup

## References

- [ThreeDotsLabs Wild Workouts DDD Example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example) — reference Go DDD architecture
- [Effective Go](https://go.dev/doc/effective_go) — idiomatic patterns
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) — review checklist
- [Watermill](https://github.com/ThreeDotsLabs/watermill) — DDD event system
- [testify](https://github.com/stretchr/testify) — test assertions
- [golangci-lint](https://github.com/golangci/golangci-lint) — meta-linter (v2, with formatters section for gofumpt)
- [go-arch-lint](https://github.com/fe3dback/go-arch-lint) — architecture linter for DDD/hexagonal (303 stars, MIT)
- [Go Migration Evaluation](20260306_5_go_rewrite_evaluation.md) — parent spike research
- [Go Team Development Patterns](20260307_2_go_team_development_patterns.md) — quality gates, anti-hallucination, test patterns
