# severity1/claude-agent-sdk-go -- Detailed Evaluation

**Date:** 2026-03-06
**Type:** Spike Research
**Status:** Complete

## Research Question

What does the community Go port `severity1/claude-agent-sdk-go` provide? How mature is it? How does it compare to the official Python `claude-agent-sdk`? Is it viable for production use in a Go-based CLI+MCP project?

---

## 1. Package Identity

| Attribute | Value |
|-----------|-------|
| **Repository** | [github.com/severity1/claude-agent-sdk-go](https://github.com/severity1/claude-agent-sdk-go) |
| **Module path** | `github.com/severity1/claude-agent-sdk-go` |
| **pkg.go.dev** | [pkg.go.dev/github.com/severity1/claude-agent-sdk-go](https://pkg.go.dev/github.com/severity1/claude-agent-sdk-go) |
| **Latest version** | v0.6.12 (January 24, 2026) |
| **License** | MIT |
| **Go requirement** | Go 1.18+ |
| **External deps** | Zero (stdlib only) |
| **GitHub stars** | ~101 |
| **Forks** | 14 |
| **Open issues** | 5 |
| **Author** | John Reilly Pospos (severity1) |
| **Renamed from** | `claude-code-sdk-go` |

Source: [pkg.go.dev](https://pkg.go.dev/github.com/severity1/claude-agent-sdk-go), [GitHub](https://github.com/severity1/claude-agent-sdk-go)

---

## 2. Architecture: CLI Subprocess Wrapper

Like the official Python SDK, this Go SDK does **NOT** call the Anthropic Messages API directly. It wraps the **Claude Code CLI** (`@anthropic-ai/claude-code`) as a subprocess and communicates via JSON-over-stdin/stdout.

```
Your Go application
    |
    v
severity1/claude-agent-sdk-go (Go)
    |
    v  (subprocess spawn + JSON lines IPC)
Claude Code CLI (Node.js, installed via npm)
    |
    v  (HTTPS API calls)
Anthropic API / Bedrock / Vertex AI
```

### Prerequisites

- Go 1.18+
- Node.js (for the CLI runtime)
- Claude Code CLI: `npm install -g @anthropic-ai/claude-code`
- `ANTHROPIC_API_KEY` environment variable (or Bedrock/Vertex credentials)

### Internal Package Structure

| Package | Purpose |
|---------|---------|
| `claudecode` (root) | Public API: `Query()`, `WithClient()`, `Client` interface |
| `internal/shared` | Shared types and error helpers |
| `internal/control` | Control protocol management |
| `internal/parser` | JSON message parser |
| `internal/query` | Query orchestration |
| `internal/client` | Internal client implementation |

The architecture uses a `Transport` interface abstraction over the subprocess communication, allowing custom test transports.

Source: [GitHub ARCHITECTURE.md](https://github.com/severity1/claude-agent-sdk-go/blob/main/ARCHITECTURE.md), [pkg.go.dev](https://pkg.go.dev/github.com/severity1/claude-agent-sdk-go)

---

## 3. Feature Inventory

### 3.1 Two Main APIs

**Query API** -- One-shot operations with automatic cleanup:
```go
iterator, err := claudecode.Query(ctx, "What is 2+2?",
    claudecode.WithSystemPrompt("You are a helpful assistant"),
    claudecode.WithMaxTurns(1),
)
```

**Client API** -- Interactive streaming conversations with session management:
```go
err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
    return client.Query(ctx, "Your question here")
})
```

### 3.2 Full Feature List

| Feature | Supported | Notes |
|---------|-----------|-------|
| One-shot queries (`Query()`) | Yes | Automatic resource cleanup |
| Interactive sessions (`WithClient`) | Yes | Go-idiomatic context manager pattern |
| Bidirectional streaming | Yes | Channels + goroutines (not async/await) |
| Tool access control | Yes | `WithAllowedTools()` / `WithDisallowedTools()` |
| Permission callbacks | Yes | `WithCanUseTool()` for custom policies |
| Lifecycle hooks | Yes | `WithHook()` -- PreToolUse, PostToolUse, etc. |
| MCP server integration | Yes | External MCP servers via `WithMcpServers()` |
| Custom SDK MCP servers | Yes | In-process tool servers |
| Session management | Yes | Isolated sessions, custom session IDs |
| File checkpointing | Yes | Checkpoint/rewind file state |
| Structured output | Yes | JSON schema constraints |
| System prompts | Yes | `WithSystemPrompt()` |
| Model selection | Yes | `WithModel()` -- dynamic switching via `SetModel()` |
| Environment variables | Yes | `WithEnvVar()` |
| Working directory | Yes | `WithCwd()` |
| Proxy settings | Yes | Supported |
| Sandbox/security | Yes | Command isolation (Linux/macOS) |
| Programmatic sub-agents | Yes | Specialized agent definitions |
| Plugin system | Yes | Plugin integration |
| Stream diagnostics | Yes | `GetStreamIssues()`, `GetStreamStats()` |
| Partial streaming | Yes | Progressive real-time updates |
| Debug output | Yes | `WithDebugStderr()` |
| CLI version checking | Yes | Warnings + skip via env var |

### 3.3 Configuration Options

Over 80 functional options (`With*` pattern), including:

- `WithSystemPrompt()`, `WithModel()`, `WithMaxTurns()`
- `WithAllowedTools()`, `WithDisallowedTools()`
- `WithPermissionMode()`, `WithCanUseTool()`
- `WithMcpServers()`, `WithHook()`
- `WithFileCheckpointing()`, `WithCwd()`
- `WithEnvVar()`, `WithDebugStderr()`

### 3.4 Message Types

| Type | Purpose |
|------|---------|
| `UserMessage` | User input (includes `tool_use_result` field) |
| `AssistantMessage` | Claude's response |
| `SystemMessage` | System-level messages |
| `ResultMessage` | Final result with cost/usage data |
| `StreamEvent` | Streaming partial updates |

### 3.5 Content Block Types

- `TextBlock` -- Plain text
- `ThinkingBlock` -- Chain-of-thought
- `ToolUseBlock` -- Tool invocations
- `ToolResultBlock` -- Tool execution results

### 3.6 Error Handling

Idiomatic Go error types with helper functions:
- `AsCLINotFoundError()` -- CLI not installed
- `AsConnectionError()` -- Subprocess communication failure
- `AsProcessError()` -- CLI process errors
- `IsConnectionError()` -- Boolean check variants
- `ErrNoMoreMessages` -- Sentinel for graceful stream termination

Source: [pkg.go.dev](https://pkg.go.dev/github.com/severity1/claude-agent-sdk-go), [GitHub README](https://github.com/severity1/claude-agent-sdk-go)

---

## 4. Development Activity

### Release History (last 10 releases)

| Version | Date | Highlight |
|---------|------|-----------|
| v0.6.12 | Jan 24, 2026 | `tool_use_result` field on UserMessage (Python SDK v0.1.22 parity) |
| v0.6.11 | Jan 22, 2026 | Comprehensive fuzz testing (2,343 seed corpus files) |
| v0.6.10 | Jan 6, 2026 | Benchmark tests for performance-critical paths |
| v0.6.9 | Jan 6, 2026 | Refactored transport.go into focused modules |
| v0.6.8 | Jan 6, 2026 | Split protocol.go into four concern files |
| v0.6.7 | Jan 5, 2026 | ARCHITECTURE.md and CONTRIBUTING.md |
| v0.6.6 | Jan 5, 2026 | Error type helper functions (Is*/As* patterns) |
| v0.6.5 | Jan 5, 2026 | CLI version checking with warnings |
| v0.6.4 | Jan 4, 2026 | Reference docs and Python parity docs (~1800 lines) |
| v0.6.3 | Jan 4, 2026 | Session terminology clarification |

### Activity Assessment

- **12 releases in January 2026** alone -- very active
- **Last commit**: Late January 2026 (about 6 weeks ago from today)
- **Commit velocity**: High during Jan 2026, appears to have slowed since
- **Contributors**: Appears to be primarily a single-author project (severity1)
- **Forks**: 14 -- indicates community interest
- **Issue count**: 5 open -- manageable backlog
- **Quality investment**: Fuzz testing, benchmarks, architecture docs, comprehensive examples

Source: [GitHub tags](https://github.com/severity1/claude-agent-sdk-go/tags)

---

## 5. Documentation and Examples

### Documentation

| Document | Description |
|----------|-------------|
| README.md | Comprehensive usage guide with code examples |
| ARCHITECTURE.md | Internal architecture and package structure |
| CONTRIBUTING.md | Contribution guidelines |
| docs/reference.md | API reference documentation |
| docs/parity.md | Python SDK parity comparison (~1800 lines) |

### Examples (20 total)

Organized by difficulty level:

**Beginner:**
1. `01_quickstart` -- Basic Query API
2. `02_client_streaming` -- Real-time streaming
3. `03_client_multi_turn` -- Multi-turn conversations

**Intermediate:**
4. `04_query_with_tools` -- File operations
5. `05_client_with_tools` -- Interactive file workflows
6. `06_query_with_mcp` -- MCP tools (timezone queries)
7. `07_client_with_mcp` -- Multi-turn MCP workflows

**Advanced:**
8. `08_client_advanced` -- Dynamic model switching
9. `09_context_manager` -- WithClient vs manual connection
10. `10_session_management` -- Session isolation
11. `11_permission_callback` -- Tool access control
12. `12_hooks` -- Lifecycle event interception
13. `13_file_checkpointing` -- Checkpoint/rewind
14. `14_sdk_mcp_server` -- Custom in-process MCP servers

**Expert:**
15. `15_programmatic_subagents` -- Specialized agents
16. `16_structured_output` -- Type-safe JSON schema
17. `17_plugins` -- Plugin system
18. `18_sandbox_security` -- Command isolation
19. `19_partial_streaming` -- Progressive updates
20. `20_debugging_and_diagnostics` -- Debug and health monitoring

This is significantly more extensive than most community ports.

Source: [GitHub examples/](https://github.com/severity1/claude-agent-sdk-go/tree/main/examples), [pkg.go.dev examples](https://pkg.go.dev/github.com/severity1/claude-agent-sdk-go/examples)

---

## 6. Python SDK Parity Comparison

The project claims "100% Feature Parity Achieved" with the Python SDK. Based on the parity documentation:

### Features at Parity

| Category | Python SDK | Go SDK | Status |
|----------|-----------|--------|--------|
| Core functions | `query()`, tools, MCP | `Query()`, tools, MCP | Parity |
| Client methods | 7 base methods | All 7 implemented | Parity |
| Message types | 5 types | 5 types | Parity |
| Content block types | 4 types | 4 types | Parity |
| Configuration options | 30+ | 30+ (80+ with helpers) | Parity+ |
| Streaming | AsyncIterator | Channels + goroutines | Parity |
| Session management | Built-in | Built-in | Parity |
| Structured output | Pydantic JSON schema | JSON schema | Parity |
| Hooks | Lifecycle hooks | `WithHook()` | Parity |
| MCP servers | Built-in | Built-in | Parity |
| File checkpointing | Built-in | Built-in | Parity |
| Sub-agents | `AgentDefinition` | Programmatic agents | Parity |
| Plugins | Supported | Supported | Parity |

### Go-Specific Additions (~40+ helpers beyond Python)

- `WithEnvVar()`, `WithDebugStderr()` -- convenience config
- `IsConnectionError()`, `AsProcessError()` -- idiomatic Go error checking
- `GetStreamIssues()`, `GetStreamStats()` -- stream diagnostics
- `GetServerInfo(ctx)` -- runtime diagnostic info
- `SetModel(ctx, model)`, `SetPermissionMode(ctx, mode)` -- dynamic runtime changes
- `Transport` interface for custom/test transports

### Key Architectural Differences

| Aspect | Python SDK | Go SDK |
|--------|-----------|--------|
| **Pattern** | Class-based, async/await | Interface-based, functional options |
| **Concurrency** | asyncio | goroutines + channels |
| **Resource mgmt** | `async with` context manager | `WithClient()` callback pattern |
| **Config** | `ClaudeAgentOptions` dataclass | Functional options (`With*`) |
| **Error handling** | Exception hierarchy | Error types with `Is*/As*` helpers |
| **Cancellation** | asyncio cancellation | `context.Context` (first parameter) |

### Version Lag

The Go SDK at v0.6.12 tracks Python SDK v0.1.22 parity (per the v0.6.12 release note). The Python SDK is currently at v0.1.47 (March 6, 2026). This means the Go SDK is approximately **25 Python releases behind** as of today. The Go SDK was last updated January 24, 2026 -- about 6 weeks ago.

Source: [docs/parity.md](https://github.com/severity1/claude-agent-sdk-go/blob/main/docs/parity.md), [PyPI claude-agent-sdk](https://pypi.org/project/claude-agent-sdk/)

---

## 7. Comparison with Other Community Go Ports

| Attribute | severity1 | schlunsen | dotcommander |
|-----------|-----------|-----------|--------------|
| **Latest version** | v0.6.12 (Jan 24) | v1.0.0+ (first stable) | Published Jan 19 |
| **Stars** | ~101 | Lower | Lower |
| **Examples** | 20 comprehensive | Fewer | Several |
| **Fuzz testing** | Yes (2,343 corpus) | Unknown | Unknown |
| **Benchmarks** | Yes | Unknown | Unknown |
| **Architecture docs** | Yes | No | No |
| **Parity docs** | Yes (~1800 lines) | No | No |
| **Zero deps** | Yes (stdlib only) | Unknown | Unknown |
| **Python SDK parity** | Claims 100% | Partial | Partial |
| **License** | MIT | MIT | Unknown |

The severity1 fork is the most mature, best-documented, and most actively maintained of the community ports.

Source: [GitHub severity1](https://github.com/severity1/claude-agent-sdk-go), [GitHub schlunsen](https://github.com/schlunsen/claude-agent-sdk-go), [pkg.go.dev dotcommander](https://pkg.go.dev/github.com/dotcommander/agent-sdk-go)

---

## 8. Risk Assessment

| Risk | Severity | Details |
|------|----------|---------|
| **Single maintainer** | High | Primarily one author (severity1); bus factor = 1 |
| **Version lag** | Medium | 25 Python releases behind; last update 6 weeks ago |
| **Not official** | Medium | Unofficial community port; Anthropic could release official Go SDK at any time |
| **Pre-v1.0** | Medium | v0.6.x signals API instability; breaking changes possible |
| **CLI subprocess dependency** | Medium | Requires Node.js + Claude Code CLI installed; same as Python SDK |
| **Low star count** | Low | 101 stars -- limited community validation |
| **No guarantee of continued maintenance** | Medium | Single maintainer could abandon; 14 forks exist as fallback |
| **Feature request #498** | Informational | Anthropic has an open feature request for official Go SDK support |

Source: [GitHub issues](https://github.com/anthropics/claude-agent-sdk-python/issues/498)

---

## 9. Suitability Assessment

### What it IS good for

- **Prototyping** Go applications that need Claude agent capabilities
- **Agent workflows** where Claude needs to read files, run commands, use tools
- **Multi-turn conversations** with session management
- **Projects already committed to the CLI subprocess model** (same architecture as Python SDK)

### What it is NOT good for

- **Simple LLM calls** (structured text generation, analysis) -- use `anthropic-sdk-go` directly instead
- **Production systems requiring SLA guarantees** -- single maintainer, pre-v1.0
- **Low-latency applications** -- subprocess spawn adds 1-3s overhead per call
- **Environments without Node.js** -- requires npm-installed CLI

### For alty's Go migration specifically

The same dual-SDK recommendation from the Python evaluation applies to Go:

1. **For simple LLM port adapters** (ChallengerPort, QuestionGeneratorPort): Use `anthropic-sdk-go` (MIT, v1.26.0, 868 stars, official Anthropic SDK). Direct API calls, no subprocess overhead, production-ready.

2. **For agent workflows** (codebase scanning, rescue mode): `severity1/claude-agent-sdk-go` is the best available option, but carries single-maintainer risk. Consider wrapping the Claude Code CLI directly as a fallback strategy.

---

## 10. Summary

| Dimension | Assessment |
|-----------|------------|
| **Feature completeness** | Excellent -- claims 100% Python parity with Go-specific additions |
| **Code quality** | Good -- fuzz testing, benchmarks, architecture docs, 20 examples |
| **Documentation** | Excellent -- reference docs, parity docs, examples at 4 difficulty levels |
| **Maintenance** | Concerning -- single author, last release 6 weeks ago, 25 Python releases behind |
| **Production readiness** | Low-Medium -- pre-v1.0, single maintainer, low community validation |
| **License** | MIT -- fully permissive |
| **Dependencies** | Zero external (stdlib only) -- excellent |
| **Architecture** | CLI subprocess wrapper -- same model as official Python SDK |

---

## Sources

- [severity1/claude-agent-sdk-go -- GitHub](https://github.com/severity1/claude-agent-sdk-go) -- Repository, README, stars, issues
- [severity1/claude-agent-sdk-go -- pkg.go.dev](https://pkg.go.dev/github.com/severity1/claude-agent-sdk-go) -- v0.6.12, Jan 24 2026, API docs
- [severity1/claude-agent-sdk-go -- Tags](https://github.com/severity1/claude-agent-sdk-go/tags) -- Release history
- [severity1/claude-agent-sdk-go -- Parity doc](https://github.com/severity1/claude-agent-sdk-go/blob/main/docs/parity.md) -- Python SDK comparison
- [severity1/claude-agent-sdk-go -- Examples](https://github.com/severity1/claude-agent-sdk-go/tree/main/examples) -- 20 examples
- [claude-agent-sdk -- PyPI](https://pypi.org/project/claude-agent-sdk/) -- Python SDK v0.1.47 (current)
- [anthropics/claude-agent-sdk-python -- GitHub](https://github.com/anthropics/claude-agent-sdk-python) -- Official Python SDK
- [Feature Request: Go SDK -- Issue #498](https://github.com/anthropics/claude-agent-sdk-python/issues/498) -- Official Go SDK request
- [schlunsen/claude-agent-sdk-go -- GitHub](https://github.com/schlunsen/claude-agent-sdk-go) -- Alternative community port
- [dotcommander/agent-sdk-go -- pkg.go.dev](https://pkg.go.dev/github.com/dotcommander/agent-sdk-go) -- Alternative community port
