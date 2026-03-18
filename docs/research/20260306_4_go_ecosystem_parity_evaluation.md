# Go Ecosystem Parity Evaluation for Python CLI+MCP Migration

**Date:** 2026-03-06
**Type:** Spike Research
**Status:** Complete

## Research Question

What is the Go ecosystem parity for migrating a Python CLI+MCP project? For each component (CLI framework, MCP SDK, Anthropic SDK, web search, testing, linting, DDD patterns, cross-compilation), what are the best Go equivalents, their maturity, and gaps versus the Python stack?

## Executive Summary

The Go ecosystem has reached strong parity with Python for CLI+MCP+LLM projects as of early 2026. The critical path items (CLI framework, MCP SDK, Anthropic SDK) all have production-ready, permissively-licensed options. The main gaps are: (1) no Go equivalent of the Claude Agent SDK (only community forks wrapping the CLI subprocess), (2) no mature DuckDuckGo search library comparable to Python's `duckduckgo-search`, and (3) template scaffolding tools like Copier have no Go equivalent.

---

## 1. CLI Framework: Cobra vs urfave/cli

### Cobra (spf13/cobra)

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/spf13/cobra](https://github.com/spf13/cobra) |
| **Stars** | 43.4k |
| **Latest version** | v1.10.2 (Dec 2024) |
| **License** | Apache 2.0 |
| **Go requirement** | Go 1.16+ |
| **Used by** | Kubernetes (kubectl), Hugo, GitHub CLI, Docker, Terraform |

**Strengths:**
- Subcommand-based CLI with nested commands (maps well to `alto init`, `alto doc-health`, etc.)
- POSIX-compliant flags with persistent flags (global flags cascade to subcommands)
- Auto-generated help, man pages, shell completions (bash, zsh, fish, PowerShell)
- Built-in argument validation (`cobra.MinimumNArgs`, `cobra.ExactArgs`, etc.)
- `cobra-cli` scaffolding tool for generating command boilerplate
- Dominant industry standard -- largest community and most documentation

**Weaknesses:**
- More verbose than Python's Typer (no automatic type inference from function signatures)
- Flags defined via `init()` functions or explicit binding, not declarative annotations
- No automatic type conversion from function signatures -- must manually define `StringVarP`, `IntVarP`, etc.
- Slightly heavier binary impact than urfave/cli

**Source:** [Cobra User Guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md), [Context7 docs](/spf13/cobra)

### urfave/cli (v3)

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/urfave/cli](https://github.com/urfave/cli) |
| **Stars** | 23.8k |
| **Latest version** | v3.6.1 (Nov 2025) |
| **License** | MIT |
| **Go requirement** | Go 1.22+ |

**Strengths:**
- Declarative flag definition in command struct (no `init()` boilerplate)
- Simpler API for single-command or small CLI tools
- v3 is actively maintained with recent releases (Mar 2026)
- Action callback pattern is cleaner for simple commands

**Weaknesses:**
- Less dominant for large multi-subcommand CLIs
- Fewer integrations (no man page generation, fewer completion options)
- Smaller ecosystem of plugins and extensions

### Comparison with Python's Typer

| Feature | Typer (Python) | Cobra (Go) | urfave/cli v3 (Go) |
|---------|---------------|-------------|---------------------|
| Type inference from function signatures | Yes (automatic) | No (manual flag binding) | No (manual flag binding) |
| Auto-generated help | Yes | Yes | Yes |
| Shell completion | Yes (click) | Yes (bash/zsh/fish/PowerShell) | Yes (bash/zsh) |
| Nested subcommands | Yes (app groups) | Yes (native) | Yes |
| Argument validation | Limited | Built-in validators | Limited |
| Rich terminal output | Yes (rich integration) | No (use lipgloss/bubbletea) | No |
| POSIX flags | No (click-style) | Yes | Yes |

### Recommendation

**Cobra** is the clear choice for a DDD-structured CLI app with multiple subcommands. Its command tree model maps naturally to alto's `init`, `doc-health`, `version` subcommands. The Apache 2.0 license is permissive. urfave/cli is better suited for simpler tools.

---

## 2. MCP SDK for Go

### Official: modelcontextprotocol/go-sdk

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) |
| **Stars** | ~4,000 |
| **Latest version** | v1.4.0 (Feb 27, 2026) |
| **Stable release** | v1.0.0 (compatibility guarantee) |
| **License** | Apache 2.0 (new contributions), MIT (existing code) |
| **Maintained by** | Anthropic + Google (official) |

**Features:**
- Server AND client support (first-class types)
- Transports: stdio, SSE, streamable HTTP (via low-level `Transport` interface)
- MCP spec version: 2025-11-25
- Single core package `mcp` (mirrors `net/http` design philosophy)
- OAuth authentication primitives

**Maturity:** Production-ready (v1.0.0 reached, compatibility guarantee established).

**Source:** [Official go-sdk README](https://github.com/modelcontextprotocol/go-sdk), [pkg.go.dev](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp)

### Community: mark3labs/mcp-go

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) |
| **Stars** | 8,300 |
| **License** | MIT |
| **MCP spec** | 2025-11-25 (backward compatible to 2024-11-05) |
| **Imported by** | 400+ packages across 200+ modules |

**Features:**
- Full MCP specification implementation
- Primary transport: stdio (`ServeStdio`)
- Simpler API with less boilerplate
- More established community (higher star count than official SDK)

### Comparison with Python MCP SDK

| Feature | Python MCP SDK | Go Official SDK | Go mcp-go |
|---------|---------------|-----------------|-----------|
| Server support | Yes | Yes | Yes |
| Client support | Yes | Yes | Yes (limited) |
| stdio transport | Yes | Yes | Yes |
| SSE transport | Yes | Yes | Via custom transport |
| Streamable HTTP | Yes | Yes | Via custom transport |
| Maturity | Production | Production (v1.0) | Production |
| Stars | ~3k | ~4k | ~8.3k |

### Recommendation

Use the **official `modelcontextprotocol/go-sdk`** for new projects. It has v1.0 stability guarantees, is co-maintained by Google and Anthropic, and supports all transports. `mark3labs/mcp-go` is more popular but lacks official backing. Full parity with the Python MCP SDK.

---

## 3. Anthropic SDK for Go

### Official: anthropics/anthropic-sdk-go

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/anthropics/anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go) |
| **Stars** | 868 |
| **Latest version** | v1.26.0 (Feb 2026) |
| **License** | MIT |
| **Go requirement** | Go 1.22+ |
| **Commits** | 416 on main |

**Features (full parity with Python SDK):**
- Messages API (core)
- Streaming (`Messages.NewStreaming()` with event accumulation)
- Tool use / function calling (with automatic schema generation via `jsonschema`)
- Tool runners for automatic conversation loops
- Message batches
- Vision/Image support
- PDF support (beta)
- Code execution (beta)
- Computer Use v5 (beta)
- Structured output via `schemautil` package
- Proper error handling with `*anthropic.Error` type
- Request/response dumping for debugging

**Supported models (as of v1.26.0):**
- Claude Opus 4.5, Sonnet 4.5, Sonnet 4, Haiku 3.5
- All latest model variants

**Design patterns:**
- Functional options pattern (`anthropic.String()`, `anthropic.Int()`)
- Go 1.24+ `omitzero` semantics with `param.Opt[T]` for optional fields
- Union types as structs with "Of" prefixed fields (e.g., `OfTool`, `OfText`)

**Source:** [README](https://github.com/anthropics/anthropic-sdk-go), [Context7 docs](/anthropics/anthropic-sdk-go)

### Comparison with Python SDK

| Feature | Python `anthropic` | Go `anthropic-sdk-go` |
|---------|-------------------|----------------------|
| Messages API | Yes | Yes |
| Streaming | Yes | Yes |
| Tool use | Yes | Yes |
| Structured output | Yes (Pydantic) | Yes (JSON schema) |
| Message batches | Yes | Yes |
| Vision | Yes | Yes |
| Computer Use | Yes (beta) | Yes (beta) |
| Prompt caching | Yes | Yes |
| Beta features | Yes | Yes |
| Type safety | Runtime (Pydantic) | Compile-time (Go types) |

### Claude Agent SDK: severity1/claude-agent-sdk-go (Selected)

The Python `claude-agent-sdk` (v0.1.47) spawns the Claude CLI as a subprocess and provides a full agent loop with built-in tools. There is no official Go equivalent, but the community port by severity1 is the most mature option.

| Attribute | Value |
|-----------|-------|
| **GitHub** | [severity1/claude-agent-sdk-go](https://github.com/severity1/claude-agent-sdk-go) |
| **Stars** | 101 |
| **Version** | v0.6.12 (Jan 24, 2026) |
| **License** | MIT |
| **Go requirement** | Go 1.18+ |
| **Dependencies** | Zero external (stdlib only) |
| **Architecture** | CLI subprocess wrapper (same as Python SDK) |

**Features:**
- Two APIs: `Query()` for one-shot, `WithClient()` for interactive streaming sessions
- Full agent loop with tool use, MCP integration, permission callbacks, lifecycle hooks
- 80+ functional option constructors (`With*` pattern), idiomatic Go concurrency
- File checkpointing, structured output, sub-agents, plugins, stream diagnostics
- 20 comprehensive examples (beginner to expert)
- Claims "100% Python SDK feature parity" with ~40 Go-specific additions

**Risks:**
- Single maintainer (bus factor = 1), 14 forks provide some fallback
- Tracks Python SDK v0.1.22 — currently 25 releases behind (Python is at v0.1.47)
- Anthropic may release an official Go SDK (feature request #498 open)

**Requires:** Node.js + `npm install -g @anthropic-ai/claude-code` + `ANTHROPIC_API_KEY`

### Recommendation

**Dual-SDK strategy:**
1. **`anthropic-sdk-go`** (official, v1.26) for LLM port adapters (structured output, tool use, streaming)
2. **`severity1/claude-agent-sdk-go`** (MIT, v0.6.12) for agent workflows requiring Claude CLI tool execution

Pin `severity1/claude-agent-sdk-go` version carefully. Monitor for official Anthropic Go Agent SDK release.

---

## 4. Web Search Library (DuckDuckGo equivalent)

### Python Baseline: `duckduckgo-search` (ddgs)

The Python `duckduckgo-search` library by deedy5 is the gold standard: 8k+ stars, actively maintained, MIT license, supports text/images/news/maps/video search, proxy support, async API.

### Go Alternatives

| Library | GitHub | Stars | License | Last Active | Verdict |
|---------|--------|-------|---------|-------------|---------|
| [the-go-tool/websearch](https://github.com/the-go-tool/websearch) | the-go-tool/websearch | 15 | MIT | Low activity | Multi-engine (DDG, Google, Qwant), but barely maintained |
| [kuhahalong/ddgsearch](https://github.com/kuhahalong/ddgsearch) | kuhahalong/ddgsearch | ~0 | Unknown | Dec 2024 | New, zero community, configurable |
| [sap-nocops/duckduckgogo](https://github.com/sap-nocops/duckduckgogo) | sap-nocops/duckduckgogo | 4 | GPL-3.0 | Oct 2020 | Dead, GPL (not permissive) |
| [Struki84/ddgo](https://github.com/Struki84/ddgo) | Struki84/ddgo | Unknown | Unknown | Unknown | Simple wrapper, minimal |
| [Djarvur/ddg-search](https://github.com/Djarvur/ddg-search) | Djarvur/ddg-search | Unknown | Unknown | Unknown | CLI + library |

### Assessment

**This is the biggest ecosystem gap.** None of the Go DDG libraries come close to Python's `duckduckgo-search` in maturity, star count, or active maintenance. The best option (`the-go-tool/websearch`) has only 15 stars and appears minimally maintained.

### Mitigation Options

1. **Port the search logic**: DDG search is HTTP scraping -- implement a thin Go adapter that scrapes DDG directly (similar to what `duckduckgo-search` does in Python)
2. **Use SearXNG API**: Deploy a SearXNG instance and call its JSON API from Go (engine-agnostic)
3. **Use a paid search API**: Brave Search API, Tavily, or SearchAPI.io offer Go-friendly REST APIs
4. **Shell out to ddgs CLI**: Run the Python `ddgs` CLI as a subprocess (ugly but functional)

### Recommendation

Build a thin Go adapter for DuckDuckGo HTML scraping, or use a search API service. None of the existing Go libraries are production-viable.

---

## 5. Testing: Go testing + testify vs pytest

### Go Standard Library + testify

| Attribute | Value |
|-----------|-------|
| **testify GitHub** | [github.com/stretchr/testify](https://github.com/stretchr/testify) |
| **Stars** | 25.9k |
| **Latest version** | v1.11.1 (Aug 2025) |
| **License** | MIT |
| **Packages** | assert, require, mock, suite |

### Table-Driven Tests vs pytest.mark.parametrize

**Go table-driven tests:**
```go
tests := []struct {
    name     string
    input    string
    expected int
    wantErr  bool
}{
    {"empty input", "", 0, true},
    {"valid input", "hello", 5, false},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := process(tt.input)
        if tt.wantErr {
            require.Error(t, err)
            return
        }
        require.NoError(t, err)
        assert.Equal(t, tt.expected, got)
    })
}
```

**Python pytest.mark.parametrize:**
```python
@pytest.mark.parametrize("input,expected,raises", [
    ("", 0, True),
    ("hello", 5, False),
])
def test_process(input, expected, raises):
    if raises:
        with pytest.raises(ValueError):
            process(input)
    else:
        assert process(input) == expected
```

### Comparison

| Feature | Go testing + testify | pytest |
|---------|---------------------|--------|
| Table-driven / parametrize | Struct slices + `t.Run()` | `@pytest.mark.parametrize` |
| Type safety in test data | Compile-time (struct types) | Runtime only |
| Subtests | `t.Run("name", ...)` | Automatic from parametrize |
| Parallel tests | `t.Parallel()` | `pytest-xdist` |
| Assertions | `assert.Equal`, `require.NoError` | `assert` keyword |
| Mocking | `testify/mock` or `gomock` | `unittest.mock` or `pytest-mock` |
| Fixtures | `TestMain`, `suite.SetupTest` | `@pytest.fixture` (more powerful) |
| Coverage | `go test -cover` (built-in) | `pytest-cov` (plugin) |
| Benchmarks | `testing.B` (built-in) | `pytest-benchmark` (plugin) |
| Test discovery | File naming (`_test.go`) | File naming (`test_*.py`) |

### Mocking: testify/mock vs gomock

| Feature | testify/mock | gomock (mockgen) |
|---------|-------------|------------------|
| Code generation | Optional (mockery CLI) | Required (mockgen) |
| Type safety | Runtime assertion | Compile-time verification |
| Ease of use | Simpler, better error output | More powerful, more verbose |
| Embedded interfaces | Better support | Weaker support |
| Community adoption | Higher (bundled with testify) | Google-maintained |

### Assessment

Go testing is **at parity** with pytest for most needs. Key advantages: compile-time type safety in test structs, built-in benchmarking, built-in coverage. Key disadvantage: pytest's fixture system is more powerful and composable than Go's `TestMain`/suite approach. The testify library (25.9k stars, MIT) fills the assertion/mock gap effectively.

---

## 6. Linting / Type Checking: golangci-lint vs ruff + mypy

### golangci-lint

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/golangci/golangci-lint](https://github.com/golangci/golangci-lint) |
| **Stars** | 18.6k |
| **Latest version** | v2.11.1 (Mar 6, 2026) |
| **License** | GPL-3.0 (dev tool only, not distributed with your app) |
| **Linters included** | 100+ |

**Key included linters:**
- `errcheck` -- unchecked errors (unique to Go; Python has no equivalent)
- `staticcheck` -- 150+ static analysis checks (closest to mypy + ruff combined)
- `gosec` -- security-focused analysis
- `govet` -- Go's built-in static analysis
- `revive` -- configurable linter (successor to golint)
- `ineffassign` -- unused assignments
- `gosimple` -- code simplification suggestions
- `unused` -- dead code detection

### What Go Gets for Free (No Tooling Needed)

This is the most important comparison point. Go's compiler and type system **automatically enforce** many things that Python requires ruff + mypy for:

| Python Tool/Rule | Go Equivalent | Notes |
|-----------------|---------------|-------|
| mypy type checking | Go compiler | Compile-time, zero config |
| ruff: unused imports (F401) | Go compiler | Compiler error, not warning |
| ruff: undefined names (F821) | Go compiler | Compiler error |
| ruff: unused variables | Go compiler | Compiler error (`declared and not used`) |
| mypy: missing return types | Go compiler | Required by syntax |
| mypy: wrong argument types | Go compiler | Compile-time type checking |
| ruff: import sorting (I) | `goimports` | Standard tool, auto-formats |
| ruff: formatting | `gofmt` / `gofumpt` | Standard tool, canonical style |
| mypy: Protocol checking | Go interfaces | Structural typing, compile-time |

### What golangci-lint Adds Beyond the Compiler

| Linter | Catches |
|--------|---------|
| `errcheck` | Unchecked error returns (Go-specific) |
| `staticcheck` | Deprecated APIs, unreachable code, performance issues |
| `gosec` | Security vulnerabilities (SQL injection, crypto misuse) |
| `gocritic` | Code style and performance suggestions |
| `dupl` | Code duplication detection |
| `funlen` | Function length limits |
| `cyclop` | Cyclomatic complexity |

### License Note

golangci-lint is GPL-3.0, which is copyleft. However, as a **development tool** that is not distributed with the application binary, the GPL does not infect the project's own license. This is the same as using GCC (GPL) to compile proprietary code. It is safe to use in CI/CD and development.

**Source:** [golangci-lint GitHub issue #232](https://github.com/golangci/golangci-lint/issues/232)

### Comparison Summary

| Feature | ruff + mypy (Python) | golangci-lint + Go compiler |
|---------|---------------------|----------------------------|
| Type checking | mypy (runtime types) | Go compiler (compile-time) |
| Lint rules | ~800 rules (ruff) | 100+ linters, 1000+ rules |
| Speed | ruff: very fast; mypy: slow | golangci-lint: very fast (parallel) |
| Configuration | pyproject.toml | .golangci.yml |
| Auto-fix | ruff --fix | Limited (some linters) |
| Zero-config coverage | Poor (must install both) | Excellent (compiler does 80%) |

### Recommendation

Go's compiler eliminates ~80% of what ruff + mypy do in Python. golangci-lint covers the remaining 20% with security, style, and complexity checks. The tooling story is strictly better in Go -- fewer tools needed, faster, more reliable.

---

## 7. DDD in Go

### Go Interfaces as Ports (vs Python Protocols)

Go interfaces are **structurally typed** (like Python `Protocol`), meaning any type that implements the methods satisfies the interface -- no explicit declaration needed.

**Go port definition:**
```go
// Port (in domain or application layer)
type OrderRepository interface {
    FindByID(ctx context.Context, id OrderID) (*Order, error)
    Save(ctx context.Context, order *Order) error
}
```

**Python Protocol equivalent:**
```python
class OrderRepository(Protocol):
    def find_by_id(self, id: OrderID) -> Order | None: ...
    def save(self, order: Order) -> None: ...
```

**Key difference:** Go interfaces are satisfied implicitly (structural typing at compile time). Python Protocols are checked by mypy at analysis time but not enforced at runtime without explicit `isinstance()` checks. Go's approach is **strictly superior for ports** -- the compiler guarantees adapter compliance.

### Value Objects in Go

Go value objects are implemented as structs with value receivers (not pointers), validated at construction time:

```go
// Value Object -- immutable by convention
type Money struct {
    amount   int64  // unexported fields = immutable from outside
    currency string
}

// Factory function enforces invariants
func NewMoney(amount int64, currency string) (Money, error) {
    if currency == "" {
        return Money{}, errors.New("currency is required")
    }
    if amount < 0 {
        return Money{}, errors.New("amount must be non-negative")
    }
    return Money{amount: amount, currency: currency}, nil
}

// Value receivers -- cannot mutate
func (m Money) Amount() int64    { return m.amount }
func (m Money) Currency() string { return m.currency }
func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, errors.New("currency mismatch")
    }
    return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}
```

**Comparison with Python frozen dataclass:**
```python
@dataclass(frozen=True)
class Money:
    amount: int
    currency: str

    def __post_init__(self) -> None:
        if not self.currency:
            raise ValueError("currency is required")
```

| Feature | Go Value Object | Python frozen dataclass |
|---------|----------------|----------------------|
| Immutability | Unexported fields + value receivers | `frozen=True` |
| Validation | Factory function (`New...`) | `__post_init__` |
| Equality | Struct comparison (automatic) | `__eq__` (auto-generated) |
| Hash | Not hashable by default | Hashable when frozen |
| Encapsulation | Unexported fields (strong) | Underscore convention (weak) |

### DDD Project Structure in Go

Established Go DDD templates follow this layout:

```
cmd/
    myapp/
        main.go              # Entry point, wire dependencies
internal/
    domain/
        order/
            order.go          # Aggregate root
            order_test.go
            repository.go     # Port interface
            events.go         # Domain events
        money/
            money.go          # Value object
    application/
        order_service.go      # Use cases / orchestration
        ports.go              # All port interfaces (alternative)
    infrastructure/
        postgres/
            order_repo.go     # Adapter implementing OrderRepository
        http/
            handler.go        # HTTP adapter
```

**Sources:**
- [sklinkert/go-ddd](https://github.com/sklinkert/go-ddd) -- Opinionated DDD template (MIT)
- [RanchoCooper/go-hexagonal](https://github.com/RanchoCooper/go-hexagonal) -- Enterprise hexagonal framework
- [Practical DDD in Golang](https://www.ompluscator.com/article/golang/practical-ddd-value-object/)
- [DDD in Go - Citerus](https://www.citerus.se/part-2-domain-driven-design-in-go/)

### Assessment

Go is **well-suited for DDD** -- arguably better than Python. Go's structural interfaces are a natural fit for ports, compile-time type checking prevents interface drift, and the `internal/` package provides true encapsulation (not possible in Python). The lack of generics until Go 1.18 was a historical pain point, but modern Go (1.22+) has sufficient generics for repository patterns.

---

## 8. Cross-Compilation

### How It Works

Go cross-compilation is built into the compiler -- no additional toolchains needed (unlike C/C++/Rust):

```bash
# Build for Linux AMD64 from any host
GOOS=linux GOARCH=amd64 go build -o myapp-linux-amd64 ./cmd/myapp

# Build for macOS ARM64 (Apple Silicon) from any host
GOOS=darwin GOARCH=arm64 go build -o myapp-darwin-arm64 ./cmd/myapp

# Build for Windows AMD64 from any host
GOOS=windows GOARCH=amd64 go build -o myapp-windows-amd64.exe ./cmd/myapp
```

### Common Target Platforms

| GOOS | GOARCH | Platform |
|------|--------|----------|
| linux | amd64 | Linux x86-64 (servers, CI) |
| linux | arm64 | Linux ARM64 (AWS Graviton, Raspberry Pi 4) |
| darwin | amd64 | macOS Intel |
| darwin | arm64 | macOS Apple Silicon (M1/M2/M3/M4) |
| windows | amd64 | Windows x86-64 |
| windows | arm64 | Windows ARM (Surface Pro X) |

Full list: `go tool dist list` (dozens of GOOS/GOARCH combinations).

### Binary Sizes

| Build | Approximate Size |
|-------|-----------------|
| Hello World (unstripped) | ~2 MB |
| Hello World (stripped: `-ldflags="-s -w"`) | ~1.3 MB |
| Typical CLI app with Cobra | ~8-15 MB |
| Cobra + stripped + UPX compressed | ~3-5 MB |

**Size reduction techniques:**
- `go build -ldflags="-s -w"` -- strip debug symbols and DWARF info (~30-40% reduction)
- `CGO_ENABLED=0` -- pure Go, no C dependencies (also enables cross-compilation)
- UPX compression -- additional ~50% reduction (at cost of startup time)
- `go build -trimpath` -- remove filesystem paths from binary

### CGO Considerations

| Scenario | CGO_ENABLED | Cross-compilation | Notes |
|----------|------------|-------------------|-------|
| Pure Go dependencies | 0 (default for cross) | Works out of the box | Preferred |
| C library dependency (SQLite, etc.) | 1 | Requires cross-compiler | Use Zig as CC: `CC="zig cc -target x86_64-linux"` |
| DNS resolution on Linux | 0 | Uses pure Go resolver | Slightly different behavior from glibc |

**Critical rule:** `CGO_ENABLED=0` for maximum cross-compilation portability. If you need CGO, use Zig as the C compiler for cross-platform builds.

### Comparison with Python Distribution

| Aspect | Go | Python |
|--------|-----|--------|
| Single binary | Yes (statically linked) | No (requires Python runtime + venv) |
| Cross-compilation | Built-in (`GOOS/GOARCH`) | N/A (interpreted) |
| Binary size | 5-15 MB | N/A (use PyInstaller: 50-100MB+) |
| Startup time | ~10-50ms | ~200-500ms |
| Distribution | Copy binary | pip install / pipx / Docker |
| No runtime dependency | Yes | No (Python 3.12+ required) |

**Source:** [Go cross-compilation guide](https://www.digitalocean.com/community/tutorials/building-go-applications-for-different-operating-systems-and-architectures), [Binary size analysis](https://oneuptime.com/blog/post/2026-01-07-go-reduce-binary-size/view)

---

## Overall Parity Matrix

| Component | Python Tool | Go Equivalent | Parity | Risk |
|-----------|------------|---------------|--------|------|
| CLI framework | Typer | **Cobra** (Apache 2.0) | High | More verbose, but more powerful |
| MCP SDK | mcp (official) | **modelcontextprotocol/go-sdk** (Apache 2.0) | Full | v1.0 stable, co-maintained by Google |
| Anthropic SDK | anthropic (official) | **anthropic-sdk-go** (MIT) | Full | Official, v1.26.0, all features |
| Agent SDK | claude-agent-sdk | **severity1/claude-agent-sdk-go** (MIT) | Medium | 101 stars, v0.6.12, 25 releases behind Python; single maintainer |
| Web search | duckduckgo-search | **None mature** | Gap | Must build adapter or use paid API |
| Testing | pytest | **testing + testify** (MIT) | High | Fixtures less powerful, types better |
| Linting | ruff + mypy | **Go compiler + golangci-lint** | Superior | Go compiler does 80% for free |
| DDD patterns | Manual + Protocols | **Interfaces + internal/** | Superior | Structural typing + true encapsulation |
| RLM pattern | Custom Python (iterative search+reason) | **Native Go** | High | Pattern translates cleanly; no code-exec sandbox needed |
| Scaffolding | Copier | **None** | Gap | No Go equivalent of Copier's template update system |
| Distribution | pip/pipx/Docker | **Single binary** | Superior | Zero runtime dependencies |

---

## Gaps and Risks

### Critical Gaps

1. **No mature Go web search library** -- The Python `duckduckgo-search` (8k stars, MIT) has no Go equivalent. Mitigation: build thin DDG scraper or use paid search API (Brave, Tavily).

2. **Claude Agent SDK version lag** -- `severity1/claude-agent-sdk-go` (v0.6.12, MIT, 101 stars) is the best community port but tracks Python SDK v0.1.22 while Python is at v0.1.47. Single maintainer risk. Mitigation: pin version, monitor for official Anthropic Go Agent SDK (feature request #498 open).

3. **No Copier equivalent** -- Go has no template scaffolding tool comparable to Copier (with template updates). Mitigation: implement template rendering with Go's `text/template` stdlib, but lose template update tracking.

### Not a Gap: RLM Pattern

The RLM (Recursive Language Model) pattern is used in two ways:

1. **In code:** `RlmResearchAdapter` implements iterative search→LLM-synthesize→refined-search loop. This is a pure orchestration pattern (no code execution sandbox). Translates directly to Go — goroutines + channels could even parallelize the multi-area research.

2. **In commands:** `architecture-docs` uses RLM conceptually — documents as addressable variables via a knowledge map. This is a prompt pattern, not a code dependency. Works identically in Go.

The full RLM paper (Zhang et al., 2025) describes a REPL sandbox with AST validation and code execution. alto does NOT use this — it uses the iterative reasoning loop only. If the full REPL sandbox is needed later, Go provides stronger sandboxing options: `yaegi` (Go interpreter), Wasm sandboxes, or OS-level isolation via `exec.Command` with restricted permissions. See `docs/RLM.md` for the full pattern reference.

### Moderate Risks

4. **golangci-lint GPL-3.0 license** -- Not a distribution concern (dev tool only), but some organizations restrict GPL usage in CI. Most major Go projects use it without issue.

5. **Verbose CLI code** -- Cobra requires more boilerplate than Typer. Mitigated by `cobra-cli` scaffolding and Go's explicit style being a feature for maintainability.

6. **pytest fixture parity** -- Go's `TestMain` and `suite.SetupTest` are less composable than pytest fixtures. Not a blocker but requires different test organization patterns.

### Not a Gap: Multi-Tool Support (Tool-Agnostic Architecture)

alto is NOT locked to Claude Code. It supports multiple AI coding tools via config generation. All tools use JSON-based MCP config with nearly identical structure:

| Tool | Config Location | MCP Format | Status |
|------|----------------|------------|--------|
| Claude Code | `.claude/` | `claude_desktop_config.json` | Supported |
| Cursor | `.cursor/` | `mcp.json` | Supported |
| Roo Code | `.roo/mcp.json`, `.roo/rules/*.md`, `.roomodes` | JSON | To implement |
| OpenCode/Crush | `opencode.json` / `.crush.json` | JSON | To implement |

**Key finding:** OpenCode is written in **Go** (MIT, 11.3K stars, archived → continued as Crush by Charmbracelet, 21K stars). Crush uses FSL-1.1-MIT (2-year commercial restriction), but alto only generates config files for it — no license concern. The fact that a major AI coding tool is written in Go reinforces the Go migration case.

A single domain model in alto can represent MCP config and translate to each tool's native format with minimal per-tool adapter logic. This is already the design pattern in the Python codebase.

### Not a Gap: Local LLM Support

Go has excellent local LLM integration options:

| Option | Library | Stars | License | Approach |
|--------|---------|-------|---------|----------|
| **Ollama** (recommended) | `ollama/api` (official) | 164K | MIT | HTTP API, OpenAI-compatible `/v1/` endpoint |
| **Universal OpenAI** | `sashabaranov/go-openai` | 10.6K | Apache 2.0 | Works with Ollama, OpenAI, Anthropic, any OpenAI-compatible API |
| **Direct llama.cpp** (future) | `hybridgroup/yzma` | 341 | Apache 2.0 | No CGo needed (purego), version-synced with llama.cpp |

**Recommended tiered strategy:**
- **Tier 1:** `sashabaranov/go-openai` as universal LLM client (connects to Ollama, OpenAI, any compatible API)
- **Tier 2:** `ollama/api` for Ollama-specific features (model management, embeddings)
- **Tier 3 (future):** `hybridgroup/yzma` for single-binary embedded inference (no server needed)

Ollama itself is written in Go (164K stars, MIT). This is another strong signal for the Go ecosystem fit.

See full details: [docs/research/20260307_1_go_ai_tool_ecosystem_local_llm.md](20260307_1_go_ai_tool_ecosystem_local_llm.md)

### Advantages of Go Migration

7. **Single binary distribution** -- eliminates Python version, venv, pip dependency issues
8. **Compile-time type safety** -- catches errors that mypy misses (mypy is optional; Go compiler is mandatory)
9. **True encapsulation** -- `internal/` package prevents cross-boundary imports at compiler level
10. **Cross-compilation** -- trivial multi-platform builds vs Python's painful packaging story
11. **Startup time** -- 10-50ms vs 200-500ms for Python CLI
12. **Memory footprint** -- significantly lower than Python for long-running MCP server processes
13. **Ecosystem alignment** -- Both Ollama (164K stars) and OpenCode (11.3K stars) are written in Go. alto joins a Go-native AI tooling ecosystem.

---

## Recommendations

1. **CLI**: Use Cobra (Apache 2.0, 43k stars, industry standard)
2. **MCP**: Use official `modelcontextprotocol/go-sdk` (Apache 2.0, v1.0 stable)
3. **Anthropic API**: Use official `anthropic-sdk-go` (MIT, v1.26.0, full parity)
4. **Agent SDK**: Use `severity1/claude-agent-sdk-go` (MIT, v0.6.12, zero external deps); pin version, monitor for official Anthropic Go Agent SDK
5. **Local LLM**: Use `sashabaranov/go-openai` (Apache 2.0, 10.6k stars) as universal client + `ollama/api` for Ollama-specific features
6. **Web search**: Build custom DDG adapter or integrate Brave/Tavily API
7. **Testing**: Use `testing` + `testify` (MIT, 26k stars)
8. **Linting**: Use `golangci-lint` (GPL-3.0 dev tool) with errcheck, staticcheck, gosec enabled
8. **DDD**: Follow Go hexagonal architecture with `internal/` packages and interface ports
9. **Distribution**: Leverage `CGO_ENABLED=0` cross-compilation for all target platforms

---

## Sources

### CLI Frameworks
- [Cobra GitHub](https://github.com/spf13/cobra) -- 43.4k stars, Apache 2.0
- [urfave/cli GitHub](https://github.com/urfave/cli) -- 23.8k stars, MIT
- [Go CLI comparison](https://github.com/gschauer/go-cli-comparison)
- [JetBrains Go Ecosystem 2025](https://blog.jetbrains.com/go/2025/11/10/go-language-trends-ecosystem-2025/)

### MCP SDK
- [Official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) -- v1.0.0 stable, Apache 2.0
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) -- 8.3k stars, MIT
- [MCP Go SDK on pkg.go.dev](https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp)

### Anthropic SDK
- [anthropic-sdk-go GitHub](https://github.com/anthropics/anthropic-sdk-go) -- 868 stars, MIT, v1.26.0
- [Anthropic Client SDKs docs](https://docs.claude.com/en/api/client-sdks)

### Agent SDK (Community)
- [severity1/claude-agent-sdk-go](https://github.com/severity1/claude-agent-sdk-go)
- [schlunsen/claude-agent-sdk-go](https://pkg.go.dev/github.com/schlunsen/claude-agent-sdk-go)
- [dotcommander/agent-sdk-go](https://pkg.go.dev/github.com/dotcommander/agent-sdk-go)

### Web Search
- [the-go-tool/websearch](https://github.com/the-go-tool/websearch) -- 15 stars, MIT
- [kuhahalong/ddgsearch](https://github.com/kuhahalong/ddgsearch)
- [Python duckduckgo-search](https://github.com/deedy5/duckduckgo_search) -- baseline comparison

### Testing
- [testify GitHub](https://github.com/stretchr/testify) -- 25.9k stars, MIT, v1.11.1
- [Go Wiki: TableDrivenTests](https://go.dev/wiki/TableDrivenTests)
- [GoMock vs Testify comparison](https://www.codecentric.de/wissens-hub/blog/gomock-vs-testify)

### Linting
- [golangci-lint GitHub](https://github.com/golangci/golangci-lint) -- 18.6k stars, GPL-3.0
- [golangci-lint linters list](https://golangci-lint.run/docs/linters/)
- [Go linters guide](https://www.glukhov.org/post/2025/11/linters-for-go/)

### DDD in Go
- [sklinkert/go-ddd](https://github.com/sklinkert/go-ddd) -- MIT, opinionated DDD template
- [Practical DDD: Value Objects](https://www.ompluscator.com/article/golang/practical-ddd-value-object/)
- [DDD in Go - Citerus](https://www.citerus.se/part-2-domain-driven-design-in-go/)
- [Hexagonal Architecture in Go](https://dev.to/buarki/hexagonal-architectureports-and-adapters-clarifying-key-concepts-using-go-14oo)
- [DDD Value Objects in Go](https://dennisvis.dev/blog/ddd-in-go-value-objects)

### Cross-Compilation
- [DigitalOcean: Building Go for Different OS/Arch](https://www.digitalocean.com/community/tutorials/building-go-applications-for-different-operating-systems-and-architectures)
- [Go and CGo Cross-Compilation](https://ecostack.dev/posts/go-and-cgo-cross-compilation/)
- [Zig Makes Go Cross Compilation Work](https://dev.to/kristoff/zig-makes-go-cross-compilation-just-work-29ho)
- [Go Binary Size Reduction](https://oneuptime.com/blog/post/2026-01-07-go-reduce-binary-size/view)
