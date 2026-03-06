# Go DDD Project Setup Best Practices for Python-to-Go Migration

**Date:** 2026-03-07
**Type:** Spike Research
**Status:** Final
**Related:** `20260306_4_go_ecosystem_parity_evaluation.md`, `20260307_2_go_team_development_patterns.md`

## Research Questions

1. Monorepo vs separate repo for Go rewrite
2. Go module path naming conventions
3. Go DDD project structure best practices (2025-2026)
4. Go CLI+MCP project patterns (cmd/ directory)
5. golangci-lint v2 best config for DDD projects
6. go-arch-lint vs depguard vs internal/ for DDD boundaries

---

## 1. Monorepo vs Separate Repo for Go Rewrite

### Decision: Separate Repo (Recommended)

When migrating a Python CLI+MCP project (alty-cli) to Go, the evidence strongly favors a
**separate repository** for the Go rewrite.

### Real-World Migration Case Studies

#### Khan Academy (Python to Go, 2019-2025)

Khan Academy's "Goliath" project is the most well-documented large-scale Python-to-Go
migration. Key facts:

- **Scale:** 1 million lines of Python to 500,000+ lines of Go (as of June 2025)
- **Duration:** 3.5 years, ~100 engineers
- **Strategy:** Incremental migration via separate Go services
- **Bridge:** GraphQL Federation -- Python monolith and Go services coexisted behind an
  Apollo gateway, each serving part of the GraphQL schema
- **Repo strategy:** Separate services with separate codebases (the new Go code ran
  in separate processes by necessity)
- **Key insight:** "New Go code would have to run in a separate process at least from
  our existing Python" -- the language boundary forced service separation

**Source:** [Khan Academy: Half a Million Lines of Go](https://blog.khanacademy.org/half-a-million-lines-of-go/), [Go + Services = One Goliath Project](https://blog.khanacademy.org/go-services-one-goliath-project/), [Beating the Odds](https://blog.khanacademy.org/beating-the-odds-khan-academys-successful-monolith%E2%86%92services-rewrite/)

#### Industry Pattern

The general industry pattern for language migrations is one of three approaches:

1. **Complete rewrite in new repo** -- clean start, no legacy interference
2. **Gradual migration via services** -- both languages run concurrently (Khan Academy approach)
3. **Wrapper approach** -- Python calls Go (or vice versa) via FFI/subprocess

For a CLI tool like alty, the **complete rewrite in a new repo** is the most practical
because:

- No service communication needed (it is a CLI, not a distributed system)
- Python source can be read in the old repo as a reference
- Go tooling (go.mod, golangci-lint, go-arch-lint) does not need to coexist with Python
  tooling (pyproject.toml, ruff, mypy)
- CI/CD pipelines are clean -- no conditional logic for "is this Python or Go?"

**Source:** [Honeybadger: Migrate from Python to Go](https://www.honeybadger.io/blog/migrate-from-python-golang/)

### Comparison Table

| Factor | Same Repo (go/ subdir) | Separate Repo |
|--------|----------------------|---------------|
| Reference Python source during migration | Easy (side by side) | Must switch repos / use split view |
| CI pipeline complexity | HIGH (dual toolchains) | LOW (Go only) |
| go.mod cleanliness | Clean (go.mod in go/ subdir) | Clean (go.mod at root) |
| Python tooling interference | ruff/mypy may scan go/ | None |
| Git history | Shared (misleading: Python commits in Go repo) | Clean Go-only history |
| Module path | `github.com/org/project/go` (unusual) | `github.com/org/project` (standard) |
| Dependency graph | Two languages in one graph | Clean single-language graph |
| IDE experience | Confusing (Python + Go in one workspace) | Clean Go workspace |
| Template development | Shared templates directory | Templates bundled in Go binary or embedded |

### Recommendation

**Use a separate repository for the Go rewrite.** The Python source code is a reference
document during migration, not a runtime dependency. Keep `alty-cli` (Python) and create
`alty` or `alty-go` as the new repository. The Python repo can be archived once migration
is complete.

If the team wants to reference Python source during development, use a workspace with both
repos cloned side by side, or use git worktrees.

---

## 2. Go Module Path Conventions

### Module Path Requirements

Go module paths must follow these rules (from the Go Modules Reference):

1. The leading path element must contain at least one dot (`.`)
2. Only lowercase ASCII letters, digits, dots, and dashes allowed
3. Cannot start with a dash

**Source:** [Go Modules Reference](https://go.dev/ref/mod)

### Options for alty

| Module Path | Pros | Cons | When to Use |
|-------------|------|------|-------------|
| `github.com/alty-cli/alty` | Standard convention; go install works; vanity URL possible | Requires GitHub repo for downloads | Open source or public distribution |
| `git.example.com/org/alty` | Works with any Git host; GOPRIVATE-compatible | Requires GOPRIVATE env var setup; custom domain needed | Private Git server |
| `alty.dev/cli` | Short, professional vanity URL; decoupled from hosting | Requires HTTP server with go-import meta tags | Professional open-source project |
| `example.com/alty` | Placeholder; spec-compliant | Not real; cannot go install | Documentation and examples only |

### Key Rules

1. **Module path = repository URL** is the standard convention. This enables `go install`
   and `go get` to work without configuration.

2. **For private repos**, set `GOPRIVATE=git.example.com/*` so the Go toolchain skips the
   public module proxy and checksum database.

3. **Major version suffix**: For v2+, the module path must end with `/v2` (e.g.,
   `github.com/alty-cli/alty/v2`). This is enforced by the toolchain. For v0/v1, no suffix.

4. **Internal imports** follow the module path:
   ```go
   import "github.com/alty-cli/alty/internal/bootstrap/domain"
   import "github.com/alty-cli/alty/internal/shared/events"
   ```

5. **Vanity import paths** (e.g., `alty.dev/cli`) require an HTTP server that responds to
   `?go-get=1` with a `<meta name="go-import">` tag pointing to the actual repository.
   This decouples the import path from the hosting provider.

### Recommendation

Use `github.com/alty-cli/alty` if the project will be hosted on GitHub. If using a private
Git server, use the server's domain (e.g., `git.yourorg.com/alty/alty`). Both patterns are
fully standard.

**Source:** [Go Modules Reference](https://go.dev/ref/mod), [go.mod file reference](https://go.dev/doc/modules/gomod-ref), [DigitalOcean: Private Go Modules](https://www.digitalocean.com/community/tutorials/how-to-use-a-private-go-module-in-your-own-project), [Taking Control of Go Module Paths](https://www.n16f.net/blog/taking-control-of-your-go-module-paths/)

---

## 3. Go DDD Project Structure Best Practices (2025-2026)

### 3.1 Recommended Structure for alty

Combining patterns from ThreeDotsLabs Wild Workouts, Damiano Petrungaro, and the official
Go module layout guide:

```
alty/
  go.mod                           # module github.com/alty-cli/alty
  go.sum
  Makefile                         # Quality gates: build, test, lint, audit
  .golangci.yml                    # golangci-lint v2 config
  .go-arch-lint.yml                # Architecture boundary rules
  cmd/
    alty/                          # CLI binary entry point
      main.go                     # Minimal: wires deps, runs root Cobra command
    alty-mcp/                      # MCP server binary entry point
      main.go                     # Minimal: wires deps, starts MCP server
  internal/                        # Compiler-enforced: no external imports
    bootstrap/                     # Bounded Context: Project Bootstrap
      domain/
        project/                   # Aggregate: Project
          project.go               # Entity + value objects
          repository.go            # Port (interface)
          errors.go                # Domain errors (sentinels)
        guided_discovery/          # Aggregate: Guided Discovery
          session.go
          question.go
          repository.go
      application/
        commands/
          init_project.go          # Command handler
        queries/
          get_project_status.go    # Query handler
      infrastructure/
        persistence/
          filesystem_repo.go       # Implements project.Repository
        external/
          git_adapter.go           # Git operations
    knowledge/                     # Bounded Context: Knowledge Base
      domain/
        knowledge_base/            # Aggregate: KnowledgeBase
          knowledge_base.go
          entry.go                 # Value object
          repository.go
      application/
      infrastructure/
    tooltranslation/               # Bounded Context: Tool Translation
      domain/
      application/
      infrastructure/
    ticketpipeline/                # Bounded Context: Ticket Pipeline
      domain/
      application/
      infrastructure/
    shared/                        # Shared Kernel (cross-context)
      events/                      # Domain events shared across contexts
        event.go                   # Base event interface/struct
      valueobjects/                # Shared value objects
        subdomain_classification.go
      errors/                      # Common domain error types
        invariant.go
  templates/                       # Embedded templates (go:embed)
    claude_md/
    cursor/
    beads/
  docs/
  tests/                           # Integration tests (unit tests live next to code)
    integration/
```

**Sources:** [ThreeDotsLabs Wild Workouts](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example), [Damiano Petrungaro: DDD in Go](https://www.damianopetrungaro.com/posts/ddd-how-i-structure-idiomatic-golang-services/), [Go Module Layout](https://go.dev/doc/modules/layout), [Go Project Structure Practices & Patterns (2025)](https://dev.to/rosgluk/go-project-structure-practices-patterns-22l5)

### 3.2 ThreeDotsLabs Wild Workouts Pattern Analysis

Wild Workouts organizes by **bounded context first**, then DDD layers within each context:

```
internal/
  trainer/                         # Bounded context
    domain/hour/                   # Aggregate: Hour (per-aggregate subdirectory)
      hour.go
      repository.go
    app/                           # Application layer
    adapters/                      # Infrastructure adapters
    ports/                         # Port interfaces
    service/                       # Service wiring
    main.go                        # Entry point (each context is a microservice)
  trainings/                       # Bounded context
    domain/training/               # Aggregate: Training
      training.go
      user.go
      repository.go
      cancel.go
      cancel_balance.go
      reschedule.go
    app/
    adapters/
    ports/
    service/
```

**Key observations:**

1. **Domain events are NOT in a separate package** -- they live within aggregate packages
   or are handled via Watermill's event bus at the application/infrastructure layer
2. **Each bounded context has its own `go.mod`** -- Wild Workouts uses a multi-module repo
   (each service is independently deployable)
3. **Repository interfaces in domain** (`domain/hour/repository.go`) -- the Port pattern
4. **Behavior files per use case** -- `cancel.go`, `reschedule.go` are separate files within
   the aggregate package, not separate packages

**Source:** [Wild Workouts GitHub](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)

### 3.3 `internal/` vs Flat Package Structure

| Approach | Enforcement | DDD Layer Control | Complexity |
|----------|-------------|-------------------|------------|
| `internal/` at root | Compiler-enforced: no external module can import | Blocks external access only; no intra-module layer rules | Low |
| `internal/` + go-arch-lint | Compiler + YAML rules | Full DDD layer enforcement | Medium |
| Flat packages (no internal) | None | Relies purely on linter rules | Low initially, risky at scale |
| Multiple modules (workspace) | Module boundaries enforced by compiler | Each BC is its own module | High (multi-module complexity) |

**Recommendation:** Use `internal/` at root for external boundary enforcement, plus
go-arch-lint for intra-module DDD layer rules. This is the most practical balance for
alty's scope (single CLI tool, not microservices).

### 3.4 Where Domain Events Live

Three patterns observed in the Go DDD ecosystem:

| Pattern | Example | Pros | Cons |
|---------|---------|------|------|
| **Per-aggregate** | `internal/bootstrap/domain/project/events.go` | Cohesion; events close to aggregate | Hard to share event types across contexts |
| **Shared kernel** | `internal/shared/events/` | Easy cross-context subscriptions | Risk of coupling; shared kernel grows |
| **Per-context** | `internal/bootstrap/domain/events/` | Middle ground; BC-scoped | Still needs shared types for cross-BC events |

**Recommendation for alty:** Use the **per-aggregate + shared kernel hybrid**:

- Events that are specific to one aggregate live in that aggregate's package (e.g.,
  `ProjectCreated` in `internal/bootstrap/domain/project/events.go`)
- Events that cross bounded context boundaries are defined in `internal/shared/events/`
  (e.g., `TicketClosed` used by both TicketPipeline and Freshness contexts)
- The shared event types are minimal: just the event struct + topic constant

This matches the Wild Workouts approach (aggregate-local) with the Watermill CQRS
pattern (shared event bus with typed events).

### 3.5 Constructor Patterns

#### `NewXxx() (T, error)` -- Standard Pattern

This is the primary constructor pattern for domain entities with invariants:

```go
// internal/bootstrap/domain/project/project.go
func NewProject(name string, description string) (*Project, error) {
    if name == "" {
        return nil, ErrEmptyProjectName
    }
    if len(description) < 20 {
        return nil, ErrDescriptionTooShort
    }
    return &Project{
        id:          uuid.New().String(),
        name:        name,
        description: description,
        status:      StatusDraft,
        createdAt:   time.Now(),
    }, nil
}
```

**When to use:** Any domain entity or value object with invariants that must be validated.
This is the default pattern for all domain types.

**Source:** [DDD Entity in Go (Panayiotis Kritiotis)](https://pkritiotis.io/ddd-entity-in-go/)

#### `MustNewXxx()` -- Panic Pattern

The `Must*` pattern is a Go stdlib convention where the function panics on error instead
of returning one:

```go
// Only for package-level initialization with compile-time-known values
var defaultConfig = MustNewConfig("production", 8080)

func MustNewConfig(env string, port int) Config {
    cfg, err := NewConfig(env, port)
    if err != nil {
        panic(fmt.Sprintf("invalid config: %v", err))
    }
    return cfg
}
```

**When to use:** ONLY for:
- Package-level variable initialization with hardcoded values
- Test fixtures where failure means a programming error
- Values known at compile time (like `regexp.MustCompile`)

**NEVER use for:** User input, runtime data, or anything that could legitimately fail.

**Source:** [Go Style Decisions (Google)](https://google.github.io/styleguide/go/decisions.html), [On Errors in Golang: Must Pattern](https://journal.petrausch.info/post/2020/05/must-pattern/), [Applied Go Conventions](https://appliedgo.net/spotlight/conventions-in-go/)

#### Functional Options Pattern

For complex constructors with many optional parameters:

```go
type ProjectOption func(*Project)

func WithStatus(s Status) ProjectOption {
    return func(p *Project) { p.status = s }
}

func NewProject(name string, opts ...ProjectOption) (*Project, error) {
    if name == "" {
        return nil, ErrEmptyProjectName
    }
    p := &Project{name: name, status: StatusDraft}
    for _, opt := range opts {
        opt(p)
    }
    return p, nil
}
```

**When to use:** When constructors have 3+ optional parameters. Common in infrastructure
adapters (HTTP clients, database connections) but less common in domain entities.

**Source:** [Go Constructor Patterns](https://programmerscareer.com/go-function-option-patterns/)

### 3.6 Error Handling for Domain Invariants

The Go DDD convention is to use **sentinel errors** (package-level `var` errors) for
domain invariants:

```go
// internal/bootstrap/domain/project/errors.go
package project

import "errors"

var (
    ErrEmptyProjectName     = errors.New("project name cannot be empty")
    ErrDescriptionTooShort  = errors.New("description must be at least 20 characters")
    ErrProjectAlreadyClosed = errors.New("project is already closed")
    ErrInvalidTransition    = errors.New("invalid status transition")
)
```

Callers check with `errors.Is()`:

```go
project, err := project.NewProject("", "short")
if errors.Is(err, project.ErrEmptyProjectName) {
    // Handle specific domain error
}
```

For errors with context (e.g., "field X was invalid"), use `fmt.Errorf` wrapping:

```go
return nil, fmt.Errorf("creating project: %w", ErrEmptyProjectName)
```

Infrastructure errors are translated to domain errors at the adapter boundary:

```go
// internal/bootstrap/infrastructure/persistence/fs_repo.go
func (r *FSRepo) FindByID(ctx context.Context, id string) (*project.Project, error) {
    data, err := os.ReadFile(filepath.Join(r.dir, id+".json"))
    if errors.Is(err, os.ErrNotExist) {
        return nil, project.ErrProjectNotFound  // Translate to domain error
    }
    if err != nil {
        return nil, fmt.Errorf("reading project %s: %w", id, err)
    }
    // ...
}
```

**Source:** [JetBrains: Secure Go Error Handling (2026)](https://blog.jetbrains.com/go/2026/03/02/secure-go-error-handling-best-practices/), [marselester/ddd-err](https://github.com/marselester/ddd-err)

---

## 4. Go CLI+MCP Project Patterns

### 4.1 Single Binary with Subcommands vs Separate Binaries

| Approach | Example | Pros | Cons |
|----------|---------|------|------|
| **Single binary, subcommand** | `alty serve` for MCP | One binary to distribute; shared domain code; simpler CI | Binary size includes both CLI+MCP deps; `alty serve` confusing for CLI-only users |
| **Separate binaries** | `cmd/alty/` + `cmd/alty-mcp/` | Clean separation; each binary includes only needed deps; clear purpose | Two binaries to distribute; shared code via `internal/` |
| **Single binary, auto-detect** | `alty` (detects if invoked via stdio) | Simplest distribution; one binary | Complex startup logic; harder to debug |

**Source:** [Go Module Layout](https://go.dev/doc/modules/layout), [Cobra Issue #641](https://github.com/spf13/cobra/issues/641)

### 4.2 Recommended Pattern for alty

Use **separate binaries** with shared `internal/` packages:

```
cmd/
  alty/                            # CLI binary
    main.go                        # Wires Cobra commands, runs CLI
  alty-mcp/                        # MCP server binary
    main.go                        # Wires MCP server, starts stdio/SSE
```

Both binaries share the same `internal/` domain and application code. The difference is
only at the presentation layer:

- `cmd/alty/main.go` wires Cobra commands that call application use cases
- `cmd/alty-mcp/main.go` wires MCP server tools that call the same application use cases

This is the standard Go pattern for projects with multiple entry points. Users install
each binary separately:

```bash
go install github.com/alty-cli/alty/cmd/alty@latest
go install github.com/alty-cli/alty/cmd/alty-mcp@latest
```

### 4.3 Cobra + MCP Go SDK Coexistence

Cobra and the MCP Go SDK do not conflict because they operate at different layers:

- **Cobra** handles CLI argument parsing and command routing
- **MCP Go SDK** handles the MCP protocol (tools, resources, prompts) over stdio/SSE

They share no initialization, no global state, and no conflicting dependencies. Each
binary's `main.go` simply imports one framework:

```go
// cmd/alty/main.go
package main

import (
    "github.com/alty-cli/alty/internal/cli"
)

func main() {
    cli.Execute()  // Cobra root command
}
```

```go
// cmd/alty-mcp/main.go
package main

import (
    "github.com/alty-cli/alty/internal/mcpserver"
)

func main() {
    mcpserver.Run()  // MCP server over stdio
}
```

The MCP Go SDK (official `github.com/modelcontextprotocol/go-sdk`, Apache 2.0, v1.0)
uses `mcp.NewServer()` and `server.ServeStdio()` -- a completely independent startup
path from Cobra.

**Source:** [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk), [MCP Go Guide](https://mcpcat.io/guides/building-mcp-server-go/), [Cobra User Guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md)

### 4.4 Alternative: Single Binary with `serve` Subcommand

If single-binary distribution is a priority:

```go
// cmd/alty/main.go -- single binary with Cobra
// alty init, alty doc-health      -> CLI commands
// alty serve                      -> starts MCP server

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Start the MCP server",
    RunE: func(cmd *cobra.Command, args []string) error {
        return mcpserver.Run()
    },
}
```

This pattern works but conflates two concerns in one binary. The separate binary approach
is cleaner for DDD because CLI presentation and MCP presentation are distinct adapters
in the infrastructure layer, and separating them reduces coupling.

---

## 5. golangci-lint v2 Best Config for DDD Projects

### 5.1 Complete Configuration

```yaml
# .golangci.yml -- golangci-lint v2 format for DDD Go projects
version: "2"

# --- Linters ---
linters:
  default: standard                # govet, errcheck, staticcheck, unused, gosimple, ineffassign, typecheck
  enable:
    # Error handling (critical for AI-generated code)
    - errorlint                    # errors.Is/As enforcement
    - wrapcheck                    # Ensure errors from external packages are wrapped

    # Code quality
    - revive                       # Extensible linter (replaces golint)
    - gocritic                     # Opinionated quality checks
    - exhaustive                   # Switch/enum exhaustiveness
    - misspell                     # Spelling in comments/strings

    # Context propagation
    - noctx                        # HTTP requests must use context.Context
    - contextcheck                 # context.Context propagation

    # Resource management
    - bodyclose                    # Unclosed HTTP response bodies

    # Testing
    - testifylint                  # Testify best practices

    # Import control (DDD layer enforcement)
    - depguard                     # Package import restrictions

  settings:
    govet:
      enable-all: true

    revive:
      rules:
        - name: exported
          arguments: [checkPrivateReceivers]
        - name: var-naming
        - name: indent-error-flow
        - name: error-return          # error must be last return value
        - name: unexported-return     # Exported func must not return unexported type

    depguard:
      rules:
        # Rule: Domain layer must not import application or infrastructure
        domain:
          files:
            - "**/internal/**/domain/**/*.go"
          deny:
            - pkg: "github.com/alty-cli/alty/internal/**/application"
              desc: "Domain layer must not import application layer"
            - pkg: "github.com/alty-cli/alty/internal/**/infrastructure"
              desc: "Domain layer must not import infrastructure layer"
            - pkg: "github.com/spf13/cobra"
              desc: "Domain layer must not import CLI framework"
            - pkg: "github.com/mark3labs/mcp-go"
              desc: "Domain layer must not import MCP framework"
            - pkg: "database/sql"
              desc: "Domain layer must not import database packages"
            - pkg: "net/http"
              desc: "Domain layer must not import HTTP packages"

        # Rule: Application layer must not import infrastructure
        application:
          files:
            - "**/internal/**/application/**/*.go"
          deny:
            - pkg: "github.com/alty-cli/alty/internal/**/infrastructure"
              desc: "Application layer must not import infrastructure layer"
            - pkg: "github.com/spf13/cobra"
              desc: "Application layer must not import CLI framework"

        # Rule: Block deprecated packages everywhere
        deprecated:
          files:
            - $all
          deny:
            - pkg: "io/ioutil"
              desc: "Deprecated since Go 1.16; use io and os directly"
            - pkg: "github.com/pkg/errors"
              desc: "Deprecated; use stdlib errors + fmt.Errorf with %w"

  exclusions:
    rules:
      - path: '_test\.go'
        linters:
          - errcheck
          - gocritic
          - wrapcheck

# --- Formatters ---
formatters:
  enable:
    - gofumpt                      # Stricter formatting than gofmt
    - gci                          # Import grouping

  settings:
    gci:
      sections:
        - standard                 # stdlib
        - default                  # third-party
        - prefix(github.com/alty-cli/alty)  # project imports

# --- Output ---
output:
  formats:
    text:
      path: stdout
      print-linter-name: true
```

### 5.2 Key Linter Roles

| Linter | Role | DDD Relevance |
|--------|------|---------------|
| `errcheck` (standard) | Catch unchecked error returns | AI agents skip errors 2x more often |
| `staticcheck` (standard) | Deprecated API detection (SA1019) | Catches `io/ioutil`, old patterns |
| `errorlint` | Enforce `errors.Is()`/`errors.As()` | Domain error matching correctness |
| `revive` | Go-specific style rules | `error-return`, `unexported-return` |
| `depguard` | Package import restrictions | DDD layer boundary enforcement |
| `noctx` + `contextcheck` | context.Context propagation | All I/O methods must accept context |
| `testifylint` | Testify usage patterns | Correct `require` vs `assert` usage |
| `exhaustive` | Switch exhaustiveness | Ensures all domain Status enum cases handled |
| `wrapcheck` | Error wrapping from external packages | Forces infrastructure boundary wrapping |
| `gofumpt` | Strict formatting | Consistent code style |
| `gci` | Import grouping | stdlib / third-party / local separation |

### 5.3 golangci-lint v2 Format Changes

The v2 config format (introduced in golangci-lint v2.0, early 2026) has several structural
differences from v1:

| v1 | v2 | Change |
|----|----|----|
| `linters-settings:` | `linters.settings:` | Nested under `linters` |
| `issues.exclude-rules:` | `linters.exclusions.rules:` | Nested under `linters` |
| `run.issues-exit-code:` | Removed | Always non-zero on issues |
| Formatters mixed with linters | `formatters:` section | Separate top-level section |
| No `version:` field | `version: "2"` required | Must declare format version |

Use `golangci-lint migrate` to auto-convert v1 configs to v2.

**Source:** [golangci-lint Configuration](https://golangci-lint.run/docs/configuration/file/), [golangci-lint Migration Guide](https://golangci-lint.run/docs/product/migration-guide/), Context7 `/golangci/golangci-lint`

---

## 6. go-arch-lint vs depguard vs internal/ for DDD Boundaries

### 6.1 Comparison

| Feature | `internal/` (compiler) | depguard (golangci-lint) | go-arch-lint (standalone) |
|---------|----------------------|------------------------|--------------------------|
| **Enforcement** | Compile-time | Lint-time (CI) | Lint-time (CI) |
| **What it blocks** | External module imports only | Specific package imports per file pattern | Component-to-component dependencies |
| **Granularity** | Binary: inside/outside `internal/` | File glob + package prefix matching | Component (directory) level |
| **DDD layer rules** | Cannot enforce domain->app->infra | Can enforce via deny rules per file glob | First-class: `mayDependOn` syntax |
| **Config format** | None (automatic) | YAML within `.golangci.yml` | Dedicated `.go-arch-lint.yml` |
| **Separate install** | No (Go compiler built-in) | No (golangci-lint built-in) | Yes (separate binary) |
| **Visual output** | None | None | SVG dependency graph |
| **License** | N/A | GPL-3.0 (depguard lib) | MIT |
| **Stars** | N/A | N/A | 453 |
| **Version** | N/A | v2 (bundled with golangci-lint) | v1.14.0 (Nov 2025) |

**Sources:** [go-arch-lint GitHub](https://github.com/fe3dback/go-arch-lint), [depguard GitHub](https://github.com/OpenPeeDeeP/depguard), [Go Internal Directories](https://go.dev/doc/modules/layout)

### 6.2 What Each Tool Catches

```
SCENARIO: Domain package imports infrastructure package

internal/ (compiler)  -->  DOES NOT CATCH (both are inside internal/)
depguard              -->  CATCHES (deny rule on domain files importing infra packages)
go-arch-lint          -->  CATCHES (domain component has no mayDependOn infrastructure)

SCENARIO: External module tries to import internal domain types

internal/ (compiler)  -->  CATCHES (compile error)
depguard              -->  NOT APPLICABLE (checks your code, not external code)
go-arch-lint          -->  NOT APPLICABLE (checks your code, not external code)

SCENARIO: Application layer accidentally imports Cobra (CLI framework)

internal/ (compiler)  -->  DOES NOT CATCH (Cobra is external, not about internal/)
depguard              -->  CATCHES (deny "github.com/spf13/cobra" in application files)
go-arch-lint          -->  CATCHES (application component has no canUse for cobra vendor)
```

### 6.3 Can They Be Combined?

**Yes, and they should be.** The three tools are complementary:

1. **`internal/`** -- free, zero-config, compiler-level baseline. Prevents external
   modules from importing your domain types. Always use.

2. **depguard** -- integrated into golangci-lint (no extra install). Best for
   blocking specific packages (deprecated packages, framework leakage). Good for
   simple deny rules.

3. **go-arch-lint** -- dedicated architecture tool. Best for expressing DDD component
   relationships via `mayDependOn` and `canUse`. Generates visual dependency graphs.
   More expressive than depguard for architecture rules.

### 6.4 Recommended Combination for alty

| Layer | Tool | Config |
|-------|------|--------|
| **External boundary** | `internal/` | Automatic (directory placement) |
| **Framework leakage** | depguard | Deny Cobra/MCP/sql/http in domain+app layers |
| **DDD layer rules** | go-arch-lint | Components + `mayDependOn` |
| **Deprecated packages** | depguard | Deny `io/ioutil`, `github.com/pkg/errors` |
| **Vendor dependencies** | go-arch-lint | `canUse` per component |

### 6.5 go-arch-lint Configuration for alty

```yaml
# .go-arch-lint.yml
version: 3
workdir: internal

allow:
  depOnAnyVendor: false            # Strict: must declare vendor deps explicitly
  deepScan: true                   # Advanced AST analysis

exclude:
  - shared/generated               # Skip generated code

components:
  # Bounded contexts
  bootstrap-domain:       { in: bootstrap/domain/** }
  bootstrap-application:  { in: bootstrap/application/** }
  bootstrap-infrastructure: { in: bootstrap/infrastructure/** }

  knowledge-domain:       { in: knowledge/domain/** }
  knowledge-application:  { in: knowledge/application/** }
  knowledge-infrastructure: { in: knowledge/infrastructure/** }

  tooltranslation-domain:       { in: tooltranslation/domain/** }
  tooltranslation-application:  { in: tooltranslation/application/** }
  tooltranslation-infrastructure: { in: tooltranslation/infrastructure/** }

  ticketpipeline-domain:       { in: ticketpipeline/domain/** }
  ticketpipeline-application:  { in: ticketpipeline/application/** }
  ticketpipeline-infrastructure: { in: ticketpipeline/infrastructure/** }

  # Shared kernel
  shared-events:         { in: shared/events }
  shared-valueobjects:   { in: shared/valueobjects }
  shared-errors:         { in: shared/errors }

  # Presentation adapters
  cli:                   { in: cli/** }
  mcpserver:             { in: mcpserver/** }

commonComponents:
  - shared-events
  - shared-valueobjects
  - shared-errors

vendors:
  cobra:      { in: github.com/spf13/cobra/** }
  mcp-sdk:    { in: github.com/modelcontextprotocol/go-sdk/** }
  watermill:  { in: github.com/ThreeDotsLabs/watermill/** }
  testify:    { in: github.com/stretchr/testify/** }
  anthropic:  { in: github.com/anthropics/anthropic-sdk-go/** }

deps:
  # Domain layers: depend on NOTHING (except shared kernel via commonComponents)
  bootstrap-domain: {}
  knowledge-domain: {}
  tooltranslation-domain: {}
  ticketpipeline-domain: {}

  # Application layers: depend on their domain only
  bootstrap-application:
    mayDependOn: [bootstrap-domain]
    canUse: [watermill]
  knowledge-application:
    mayDependOn: [knowledge-domain]
    canUse: [watermill]
  tooltranslation-application:
    mayDependOn: [tooltranslation-domain]
    canUse: [watermill]
  ticketpipeline-application:
    mayDependOn: [ticketpipeline-domain]
    canUse: [watermill]

  # Infrastructure layers: depend on domain + application
  bootstrap-infrastructure:
    mayDependOn: [bootstrap-domain, bootstrap-application]
    anyVendorDeps: true
  knowledge-infrastructure:
    mayDependOn: [knowledge-domain, knowledge-application]
    anyVendorDeps: true
  tooltranslation-infrastructure:
    mayDependOn: [tooltranslation-domain, tooltranslation-application]
    anyVendorDeps: true
  ticketpipeline-infrastructure:
    mayDependOn: [ticketpipeline-domain, ticketpipeline-application]
    anyVendorDeps: true

  # Presentation adapters: depend on application layers
  cli:
    mayDependOn:
      - bootstrap-application
      - knowledge-application
      - tooltranslation-application
      - ticketpipeline-application
    canUse: [cobra]
  mcpserver:
    mayDependOn:
      - bootstrap-application
      - knowledge-application
      - tooltranslation-application
      - ticketpipeline-application
    canUse: [mcp-sdk]
```

---

## Summary and Recommendations

### Decision Matrix

| Question | Recommendation | Confidence |
|----------|---------------|------------|
| 1. Repo strategy | **Separate repo** for Go rewrite | HIGH |
| 2. Module path | `github.com/alty-cli/alty` (or private Git server equivalent) | HIGH |
| 3. Project structure | `internal/` with bounded contexts, per-aggregate domain packages | HIGH |
| 4. CLI+MCP pattern | **Separate binaries** (`cmd/alty/`, `cmd/alty-mcp/`) | HIGH |
| 5. golangci-lint config | v2 format with depguard, errorlint, revive, gofumpt, gci | HIGH |
| 6. Boundary enforcement | **All three combined**: `internal/` + depguard + go-arch-lint | HIGH |

### Key Finding

The combination of `internal/` (compiler), depguard (golangci-lint), and go-arch-lint
provides three complementary layers of DDD boundary enforcement that are strictly stronger
than anything achievable in Python. The compiler alone catches circular imports, unused
imports, and external access violations -- things that require third-party tools in Python.

### Biggest Risk

**go-arch-lint maintenance.** The tool has 453 stars and one primary maintainer. If it
goes unmaintained, depguard alone can provide basic layer enforcement (deny rules), but
without the expressive `mayDependOn` syntax or visual dependency graphs.

### Follow-up Tasks

- [ ] Create ticket: Scaffold Go project with recommended DDD structure (cmd/, internal/, Makefile)
- [ ] Create ticket: Configure golangci-lint v2 with DDD-focused linter set
- [ ] Create ticket: Configure go-arch-lint with bounded context dependency rules
- [ ] Create ticket: Set up separate cmd/alty/ (Cobra) and cmd/alty-mcp/ (MCP SDK) entry points
- [ ] Create ticket: Define domain event types in shared/events/ and per-aggregate packages
- [ ] Create ticket: Port domain models from Python to Go with NewXxx constructors and sentinel errors

## References

- [Khan Academy: Half a Million Lines of Go (2025)](https://blog.khanacademy.org/half-a-million-lines-of-go/)
- [Khan Academy: Go + Services = One Goliath Project](https://blog.khanacademy.org/go-services-one-goliath-project/)
- [Khan Academy: Technical Choices Behind a Successful Rewrite](https://blog.khanacademy.org/technical-choices-behind-a-successful-rewrite-project/)
- [Khan Academy: Beating the Odds](https://blog.khanacademy.org/beating-the-odds-khan-academys-successful-monolith%E2%86%92services-rewrite/)
- [Honeybadger: Migrate from Python to Go](https://www.honeybadger.io/blog/migrate-from-python-golang/)
- [Go Modules Reference](https://go.dev/ref/mod)
- [go.mod File Reference](https://go.dev/doc/modules/gomod-ref)
- [Go Official Module Layout](https://go.dev/doc/modules/layout)
- [DigitalOcean: Private Go Modules](https://www.digitalocean.com/community/tutorials/how-to-use-a-private-go-module-in-your-own-project)
- [ThreeDotsLabs Wild Workouts](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
- [ThreeDotsLabs: DDD Lite in Go](https://threedots.tech/post/ddd-lite-in-go-introduction/)
- [Damiano Petrungaro: DDD in Go](https://www.damianopetrungaro.com/posts/ddd-how-i-structure-idiomatic-golang-services/)
- [DDD Entity in Go (Panayiotis Kritiotis)](https://pkritiotis.io/ddd-entity-in-go/)
- [Go Project Structure Practices & Patterns (2025)](https://dev.to/rosgluk/go-project-structure-practices-patterns-22l5)
- [go-arch-lint GitHub (MIT, 453 stars, v1.14.0)](https://github.com/fe3dback/go-arch-lint)
- [depguard GitHub (GPL-3.0, v2)](https://github.com/OpenPeeDeeP/depguard)
- [golangci-lint Configuration](https://golangci-lint.run/docs/configuration/file/)
- [golangci-lint Migration Guide v1 to v2](https://golangci-lint.run/docs/product/migration-guide/)
- [golangci-lint Linters List](https://golangci-lint.run/docs/linters/)
- [Google Go Style Decisions](https://google.github.io/styleguide/go/decisions.html)
- [Go MCP SDK (official)](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Go Guide (MCPcat)](https://mcpcat.io/guides/building-mcp-server-go/)
- [Cobra User Guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md)
- [JetBrains: Secure Go Error Handling (2026)](https://blog.jetbrains.com/go/2026/03/02/secure-go-error-handling-best-practices/)
- [On Errors in Golang: Must Pattern](https://journal.petrausch.info/post/2020/05/must-pattern/)
- [Applied Go Conventions](https://appliedgo.net/spotlight/conventions-in-go/)
- [Go Constructor Patterns](https://programmerscareer.com/go-function-option-patterns/)
- [marselester/ddd-err](https://github.com/marselester/ddd-err)
- [Graphite: Multi-Language Monorepos](https://graphite.com/guides/managing-multiple-languages-in-a-monorepo)
- [Earthly: Golang Monorepo](https://earthly.dev/blog/golang-monorepo/)
