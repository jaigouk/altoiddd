# Research: Go Rewrite Evaluation

**Date:** 2026-03-06
**Spike Ticket:** alty-4fr
**Status:** Final

## Summary

Go rewrite is **RECOMMENDED**. The codebase is architected for portability: strict DDD layering, Protocol-based ports, zero async in domain, and 803+ tests that serve as executable specification. Go provides superior type safety (compile-time vs mypy analysis-time), single binary distribution, native concurrency, and true DDD boundary enforcement via `internal/` packages.

**Key numbers:**
- 135 source files, 15.6K LoC (domain: 7.3K, application: 3.5K, infrastructure: 5.7K)
- 157 test files, 30.3K LoC (2:1 test-to-code ratio)
- 23 Protocol-based ports → 23 Go interfaces
- 9 domain event types (frozen dataclasses → Go structs)
- 17 async files — ALL in infrastructure layer, zero in domain
- Estimated effort: 6-8 person-weeks for full parity

## Research Questions Answered

### 1. Concurrency: goroutines vs asyncio

**Finding: Go wins decisively. Async is localized, not systemic.**

Current async inventory (17 files, all infrastructure):
- LLM clients (Anthropic SDK calls)
- Research adapters (DuckDuckGo, RLM)
- Challenger/Simulator adapters
- MCP server (595 LoC, heaviest async user)

Domain layer has **ZERO async code**. Application handlers are sync orchestrators.

| Pattern | Python (current) | Go (target) |
|---------|-----------------|-------------|
| LLM API call | `async def structured_output()` | `func StructuredOutput() (T, error)` — sync is fine, or goroutine for concurrent |
| MCP tool handler | `@mcp.tool()` async | MCP Go SDK handler — goroutines for concurrent requests natively |
| Subprocess exec | `asyncio.create_subprocess_exec` | `exec.CommandContext()` — simpler, with context cancellation |
| Fan-out requests | `asyncio.gather()` | `errgroup.Group` — typed, with error propagation |

**Verdict:** Go's goroutines + errgroup provide simpler concurrency with less boilerplate than asyncio. The MCP server benefits most — goroutines handle concurrent sub-agent requests without async/await coloring.

### 2. Cross-compilation: single binary distribution

**Finding: Go eliminates the distribution problem entirely.**

| Aspect | Python (current) | Go (target) |
|--------|-----------------|-------------|
| User prerequisite | Python 3.12+ and `uv` | None (single binary) |
| Binary distribution | N/A (pip/pipx) | `GOOS=linux GOARCH=amd64 go build` |
| Binary size | N/A (PyInstaller: 50-100MB) | 8-15MB (Cobra + deps) |
| Startup time | 200-500ms | 10-50ms |
| CI matrix | Python version matrix | `GOOS/GOARCH` env vars |
| CGo needed? | N/A | No — all pure Go deps available |

Build script (single Makefile target):
```makefile
release:
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o dist/alty-linux-amd64
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o dist/alty-linux-arm64
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o dist/alty-darwin-amd64
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o dist/alty-darwin-arm64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/alty-windows-amd64.exe
```

### 3. Migration gap analysis

**Finding: ~70% translates mechanically, ~25% needs Go-specific libraries, ~5% needs redesign.**

| Layer | Python LoC | Translation Difficulty | Notes |
|-------|-----------|----------------------|-------|
| Domain models | 4,500 | Easy | Frozen dataclasses → Go structs with unexported fields |
| Domain services | 1,460 | Easy | Pure functions → Go methods/functions |
| Domain events | 237 | Trivial | Frozen dataclasses → Go structs |
| Domain errors | 28 | Trivial | Custom exceptions → Go error types |
| Application ports | 500 | Trivial | Protocol → Go interface |
| Application handlers | 2,500 | Easy | Orchestration logic, straightforward |
| Infra: persistence | 1,200 | Easy | File I/O → `os` package |
| Infra: external | 1,500 | Medium | SDK differences, search adapter gap |
| Infra: CLI | 500 | Easy | Typer → Cobra (well-documented) |
| Infra: MCP | 600 | Medium | FastMCP → Go MCP SDK (reimplement) |
| Infra: composition | 200 | Easy | DI wiring → constructor injection |
| **Tests** | **30,300** | **Medium** | pytest fixtures → table-driven tests |

### 4. Ecosystem parity

**Finding: Critical-path libraries have full Go equivalents. 1 gap (web search). Local LLM story is stronger in Go.**

| Component | Python | Go Equivalent | Parity | Status |
|-----------|--------|--------------|--------|--------|
| CLI | typer | **Cobra** (43K stars, Apache 2.0) | High | Industry standard |
| MCP SDK | mcp | **modelcontextprotocol/go-sdk** v1.0 (Apache 2.0) | Full | Official, Google co-maintained |
| LLM SDK | anthropic | **anthropic-sdk-go** v1.26 (MIT) | Full | Official, all features |
| Local LLM | N/A | **sashabaranov/go-openai** (10.6K stars) + **ollama/api** | Superior | Ollama is Go-native (164K stars); universal OpenAI-compatible client |
| Agent SDK | claude-agent-sdk | **severity1/claude-agent-sdk-go** v0.6.12 (MIT) | Medium | 101 stars, zero deps, single maintainer |
| Web search | ddgs (8K stars) | **No mature equivalent** | Gap | Must build thin adapter |
| Testing | pytest | **testing + testify** (26K stars) | High | Fixtures less powerful |
| Linting | ruff + mypy | **Go compiler + golangci-lint** | Superior | Compiler does 80% for free |
| RLM pattern | Custom Python (iterative search+reason) | **Native Go** | High | No code-exec sandbox needed; goroutines can parallelize |

**Tool-agnostic by design:** alty is NOT locked to Claude Code. It generates configs for multiple AI coding tools — all use JSON-based MCP config:

| Tool | Written in | Config | MCP Support |
|------|-----------|--------|-------------|
| Claude Code | TypeScript | `.claude/` | Yes |
| Cursor | Electron | `.cursor/mcp.json` | Yes |
| Roo Code | TypeScript (VS Code ext) | `.roo/mcp.json` | Yes |
| OpenCode/Crush | **Go** (11K/21K stars) | `opencode.json` / `.crush.json` | Yes |

OpenCode being written in Go reinforces the ecosystem fit. Ollama (164K stars) also Go-native.

See full details: [docs/research/20260306_4_go_ecosystem_parity_evaluation.md](20260306_4_go_ecosystem_parity_evaluation.md) and [docs/research/20260307_1_go_ai_tool_ecosystem_local_llm.md](20260307_1_go_ai_tool_ecosystem_local_llm.md)

### 5. Type safety: compile-time vs analysis-time

**Finding: Go is strictly superior for DDD boundary enforcement.**

| Enforcement | Python (mypy) | Go (compiler) |
|-------------|--------------|---------------|
| Port compliance | mypy analysis (optional, can skip) | Compiler error (cannot skip) |
| Unused imports | ruff warning | Compiler error |
| Missing error handling | No enforcement | `errcheck` linter |
| Layer boundary | Convention only | `internal/` package (compiler-enforced) |
| Value object immutability | `frozen=True` (runtime) | Unexported fields (compile-time) |
| Interface satisfaction | Protocol check at analysis time | Structural typing at compile time |

**What we lose:**
- `@dataclass` convenience (Go structs need explicit constructors)
- pytest fixtures composability (Go setup/teardown is less elegant)
- Dynamic Protocol dispatch (minor — Go interfaces are structural)
- Rapid prototyping speed (Go requires more boilerplate upfront)

**What we gain:**
- Compilation catches all type errors before runtime
- `internal/` prevents domain layer imports from infrastructure at compiler level
- No `# type: ignore` escape hatches
- Binary ships with zero runtime dependencies

### 6. DDD event system and messaging

**Finding: Watermill (ThreeDotsLabs) with GoChannel backend. NATS when MCP concurrency arrives.**

Evaluated 7 options. Watermill wins:

| Criterion | Winner | Why |
|-----------|--------|-----|
| DDD events | Watermill | Typed event bus, multiple handlers per event |
| CQRS | Watermill | First-class CommandBus + EventBus |
| Single binary | Watermill + GoChannel | Pure in-process, zero deps |
| Upgrade path | Watermill | Swap GoChannel → embedded NATS, zero app code changes |
| Middleware | Watermill | Retry, recovery, correlation ID, throttle built-in |

Migration path for event infrastructure:
```
Phase 1 (CLI):   Watermill + GoChannel (in-process, ~0 MB overhead)
Phase 2 (MCP):   Watermill + Embedded NATS (in-process, ~10-20 MB, concurrent)
Phase 3 (Scale):  Watermill + External NATS (if needed)
```

Your Tachikoma NATS JetStream pattern is the Phase 2 target — but Watermill wraps it so the application code never changes.

See full details: [docs/research/20260306_2_go_ddd_event_systems_messaging.md](20260306_2_go_ddd_event_systems_messaging.md)

---

## Go Architecture Design

### Package Layout

```
alty/
├── cmd/
│   ├── alty/                          # CLI binary entry point
│   │   └── main.go                    # Cobra root, DI wiring
│   └── alty-mcp/                      # MCP server binary entry point
│       └── main.go                    # MCP server with Go SDK
├── internal/
│   ├── domain/                        # ZERO external deps (compiler-enforced)
│   │   ├── bootstrap/                 # Bootstrap aggregate + events
│   │   │   ├── session.go             # BootstrapSession aggregate root
│   │   │   ├── events.go              # BootstrapCompleted, etc.
│   │   │   └── session_test.go
│   │   ├── discovery/                 # Discovery aggregate + events
│   │   │   ├── session.go
│   │   │   ├── events.go
│   │   │   └── session_test.go
│   │   ├── ddd/                       # DDD artifact value objects
│   │   │   ├── domain_model.go        # DomainModel aggregate root
│   │   │   ├── bounded_context.go     # Value object
│   │   │   ├── domain_story.go        # Value object
│   │   │   ├── aggregate_design.go    # Value object
│   │   │   └── domain_model_test.go
│   │   ├── ticket/                    # Ticket value objects + freshness
│   │   │   ├── ticket.go
│   │   │   ├── freshness.go
│   │   │   ├── implementability.go
│   │   │   └── ticket_test.go
│   │   ├── fitness/                   # Fitness function value objects
│   │   ├── quality/                   # Quality gate value objects
│   │   ├── knowledge/                 # Knowledge entry value objects
│   │   └── errors/                    # Domain error types
│   │       └── errors.go             # InvariantViolationError, etc.
│   ├── application/                   # Depends on domain + ports only
│   │   ├── ports/                     # Go interfaces (was Python Protocols)
│   │   │   ├── bootstrap.go           # BootstrapPort interface
│   │   │   ├── llm.go                 # LLMClient interface
│   │   │   ├── search.go              # WebSearchPort interface
│   │   │   ├── ticket.go              # TicketGenerationPort interface
│   │   │   └── ...                    # 23 port interfaces total
│   │   ├── commands/                  # Command handlers
│   │   │   ├── bootstrap.go           # BootstrapHandler
│   │   │   ├── discovery.go
│   │   │   ├── artifact_generation.go
│   │   │   └── ...
│   │   └── queries/                   # Query handlers
│   │       ├── doc_health.go
│   │       ├── ticket_health.go
│   │       └── knowledge_lookup.go
│   ├── infrastructure/                # Implements ports, depends on external libs
│   │   ├── anthropic/                 # anthropic-sdk-go adapter
│   │   │   ├── llm_client.go
│   │   │   └── challenger.go
│   │   ├── search/                    # DDG search adapter (custom)
│   │   │   └── ddg_adapter.go
│   │   ├── persistence/               # File I/O adapters
│   │   │   ├── file_writer.go
│   │   │   ├── doc_scanner.go
│   │   │   └── markdown_renderer.go
│   │   ├── git/                       # Git operations adapter
│   │   │   └── git_ops.go
│   │   ├── beads/                     # Beads integration adapter
│   │   │   └── ticket_reader.go
│   │   ├── subprocess/                # Quality gate runner
│   │   │   └── gate_runner.go
│   │   ├── ollama/                    # Ollama LLM adapter (ollama/api)
│   │   │   └── llm_client.go
│   │   ├── openai/                    # Universal OpenAI-compatible adapter
│   │   │   └── llm_client.go          # sashabaranov/go-openai (works w/ Ollama too)
│   │   └── eventbus/                  # Watermill event infrastructure
│   │       ├── setup.go               # GoChannel or NATS backend init
│   │       └── handlers.go            # Cross-context event wiring
│   └── composition/                   # DI wiring (constructor injection)
│       └── app.go                     # AppContext struct, wire everything
├── go.mod
├── go.sum
├── Makefile                           # Build, test, lint, release targets
└── .golangci.yml                      # Linter config
```

### DDD Boundary Enforcement

Go's `internal/` package provides what Python cannot — **compiler-enforced layer boundaries**:

```
internal/domain/     → Cannot be imported by anything outside internal/
internal/domain/ddd/ → Cannot import from internal/infrastructure/
                       The compiler rejects it. No linter needed.
```

### Event Architecture with Watermill

```go
// internal/application/ports/eventbus.go
type EventBus interface {
    Publish(ctx context.Context, topic string, event any) error
}

type CommandBus interface {
    Send(ctx context.Context, cmd any) error
}

// internal/infrastructure/eventbus/setup.go
func NewInProcessEventBus(logger watermill.LoggerAdapter) (*cqrs.Facade, error) {
    pubSub := gochannel.NewGoChannel(gochannel.Config{}, logger)
    // Configure CQRS with typed handlers...
}

// Future: swap to NATS with zero app code changes
func NewNATSEventBus(natsURL string, logger watermill.LoggerAdapter) (*cqrs.Facade, error) {
    publisher, _ := jetstream.NewPublisher(...)
    subscriber, _ := jetstream.NewSubscriber(...)
    // Same CQRS configuration, different backend
}
```

---

## Migration Plan

### Strategy: Test-Driven Migration

The 30K+ lines of Python tests are the **executable specification**. Migration order:

1. **Translate test expectations first** (what behavior must be preserved)
2. **Write Go tests matching Python test assertions** (table-driven, parallel)
3. **Implement Go code to pass those tests** (Red → Green → Refactor)

This preserves TDD discipline and ensures feature parity.

### Phase 1: Domain Layer (Week 1-2)

**Goal:** All domain models, services, events, errors in Go with 100% test parity.

| Task | Python Source | Go Target | Test Count (approx) |
|------|-------------|-----------|-------------------|
| Value objects | `domain/models/domain_values.py` | `domain/ddd/*.go` | ~80 |
| DomainModel aggregate | `domain/models/domain_model.py` | `domain/ddd/domain_model.go` | ~30 |
| Bootstrap session | `domain/models/bootstrap_session.py` | `domain/bootstrap/session.go` | ~25 |
| Discovery session | `domain/models/discovery_session.py` | `domain/discovery/session.go` | ~20 |
| Ticket values | `domain/models/ticket_values.py` | `domain/ticket/*.go` | ~40 |
| Fitness values | `domain/models/fitness_values.py` | `domain/fitness/*.go` | ~30 |
| Domain services (8) | `domain/services/*.py` | `domain/*/service.go` | ~100 |
| Domain events (9 types) | `domain/events/*.py` | `domain/*/events.go` | ~15 |
| Error types | `domain/models/errors.py` | `domain/errors/errors.go` | ~10 |

**Milestone:** `go test ./internal/domain/...` — all pass, zero external deps.

### Phase 2: Application Layer (Week 3)

**Goal:** All 23 ports as Go interfaces, all handlers as Go functions.

| Task | Python Source | Go Target | Test Count |
|------|-------------|-----------|-----------|
| 23 port interfaces | `application/ports/*.py` | `application/ports/*.go` | 0 (interfaces) |
| 14 command handlers | `application/commands/*.py` | `application/commands/*.go` | ~150 |
| 3 query handlers | `application/queries/*.py` | `application/queries/*.go` | ~30 |

**Milestone:** `go test ./internal/application/...` — all pass with mock adapters.

### Phase 3: Infrastructure Adapters (Week 4-5)

**Goal:** All adapters implementing ports with integration tests.

| Task | Complexity | Notes |
|------|-----------|-------|
| File I/O adapters | Easy | `os` package, straightforward |
| Markdown renderer | Easy | `text/template` or string building |
| Git adapter | Easy | `os/exec` or go-git |
| Beads adapter | Easy | `os/exec` calling `bd` CLI |
| Quality gate runner | Easy | `os/exec` calling ruff/mypy/pytest |
| Anthropic LLM client | Medium | anthropic-sdk-go, different API shape |
| DDG search adapter | Medium | Custom HTTP scraper (no library) |
| Challenger/Simulator | Medium | Anthropic SDK prompt construction |
| Session store | Easy | `sync.Map` or struct with mutex |

### Phase 4: CLI + MCP (Week 6-7)

**Goal:** Fully functional CLI and MCP server.

| Task | Complexity | Notes |
|------|-----------|-------|
| Cobra CLI setup | Easy | Root command + 9 subcommands |
| CLI argument parsing | Easy | Cobra flags → handler params |
| MCP server | Medium | Go MCP SDK v1.0, 11+ tools |
| DI composition | Easy | Constructor injection in main.go |

### Phase 5: Event Infrastructure + Polish (Week 8)

**Goal:** Watermill event bus, cross-compilation, CI.

| Task | Notes |
|------|-------|
| Watermill + GoChannel setup | EventBus and CommandBus ports |
| Domain event wiring | Connect bounded context events |
| Makefile with cross-compile | 5 platform targets |
| golangci-lint config | Match ruff strictness |
| CI pipeline | Build, test, lint, release |
| README update | Installation: download binary |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| DDG search — no Go library | Certain | Low | Build thin HTTP scraper (~200 LoC); DDG's API is simple |
| MCP server rewrite complexity | Medium | Medium | Go MCP SDK v1.0 is well-documented; 11 tools manageable |
| Test migration tedium (30K LoC) | Certain | Medium | Translate domain tests first (highest value); infrastructure tests last |
| Watermill maintenance slows | Low | Low | MIT license; GoChannel backend is simple enough to fork/maintain |
| Go verbose boilerplate | Certain | Low | Accepted tradeoff for compile-time safety and single binary |
| Feature velocity during migration | High | Medium | Run Python and Go in parallel; migrate incrementally by layer |

---

## Recommendation

**Proceed with Go rewrite.** The architecture is ready for it:

1. **Strict DDD layering** means each layer can be migrated independently
2. **803+ tests** serve as executable specification — translate them, and features follow
3. **23 Protocol-based ports** become 23 Go interfaces — the seam for test doubles
4. **Zero async in domain** — 70% of the codebase translates mechanically
5. **Watermill + GoChannel** provides DDD-native event system with NATS upgrade path
6. **Single binary** eliminates the #1 adoption friction (Python 3.12 + uv requirement)
7. **Tool-agnostic** — alty supports Claude Code, Cursor, Roo Code, OpenCode/Crush. All use JSON MCP configs. Go rewrite does not change this; OpenCode itself is Go-native.
8. **Local LLM ecosystem** — Ollama (164K stars, Go-native) + sashabaranov/go-openai (universal client) gives alty local LLM support without Python dependency

**Approach:** Domain-first, test-driven, layer-by-layer migration over 6-8 weeks.

---

## Related Reports

- [Go Ecosystem Parity Evaluation](20260306_4_go_ecosystem_parity_evaluation.md) — dependency comparison table
- [Go DDD Event Systems & Messaging](20260306_2_go_ddd_event_systems_messaging.md) — Watermill vs NATS vs alternatives
- [DuckDuckGo Search Evaluation](20260306_1_duckduckgo_search_evaluation.md) — search library gap analysis
- [AI Tool Ecosystem & Local LLM](20260307_1_go_ai_tool_ecosystem_local_llm.md) — OpenCode/Crush, Roo Code, Ollama, llama.cpp
- [severity1/claude-agent-sdk-go](20260306_3_severity1_claude_agent_sdk_go.md) — Agent SDK evaluation

## References

- [Cobra CLI](https://github.com/spf13/cobra) — 43K stars, Apache 2.0
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) — v1.0, Apache 2.0, Google co-maintained
- [Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go) — v1.26, MIT, official
- [severity1/claude-agent-sdk-go](https://github.com/severity1/claude-agent-sdk-go) — v0.6.12, MIT, zero deps
- [Ollama](https://github.com/ollama/ollama) — 164K stars, MIT, Go-native
- [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) — 10.6K stars, Apache 2.0, universal OpenAI client
- [Watermill](https://github.com/ThreeDotsLabs/watermill) — 9.4K stars, MIT, DDD/CQRS
- [Wild Workouts DDD Example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example) — reference architecture
- [testify](https://github.com/stretchr/testify) — 26K stars, MIT
- [golangci-lint](https://github.com/golangci/golangci-lint) — 18.6K stars
- [Embedding NATS in Go](https://gosuda.org/blog/posts/how-embedded-nats-communicate-with-go-application-z36089af0)
- [OpenCode](https://github.com/opencode-ai/opencode) — 11.3K stars, MIT, Go, archived
- [Crush (Charmbracelet)](https://github.com/charmbracelet/crush) — 21K stars, FSL-1.1-MIT, Go
- [hybridgroup/yzma](https://github.com/hybridgroup/yzma) — llama.cpp Go binding, no CGo
