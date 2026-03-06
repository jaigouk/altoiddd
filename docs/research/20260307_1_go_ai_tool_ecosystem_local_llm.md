# Go AI Tool Ecosystem & Local LLM Integration Research

**Date:** 2026-03-07
**Type:** Spike Research
**Status:** Complete

## Research Questions

1. What is OpenCode and does it support MCP? Config format? Written in Go?
2. Does Roo Code support MCP? Config format? Go-specific integrations?
3. What Go bindings exist for llama.cpp? Maturity?
4. Is Ollama written in Go? Official Go client? OpenAI API compatibility?
5. For a Go rewrite of alty: what is the best local LLM integration strategy?

## Executive Summary

The Go ecosystem for AI coding tools and local LLM integration is remarkably strong as of March 2026. OpenCode (now Crush) and Roo Code both support MCP servers via JSON config. Ollama (164k stars, MIT, written in Go) provides the simplest local LLM path with an official Go API client and OpenAI-compatible endpoints. For direct llama.cpp integration without a server, yzma (Apache 2.0, no CGo required) is the most promising option. The recommended strategy for alty's Go rewrite is: connect to Ollama via its Go API client for local LLMs, and use sashabaranov/go-openai (Apache 2.0, 10.6k stars) as the universal OpenAI-compatible client that works with Ollama, OpenAI, Anthropic, and other providers.

---

## 1. OpenCode (now Crush)

### OpenCode (Archived)

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/opencode-ai/opencode](https://github.com/opencode-ai/opencode) |
| **Stars** | 11.3k |
| **Language** | Go |
| **License** | MIT (Copyright 2025 Kujtim Hoxha) |
| **Status** | **Archived** September 18, 2025 |
| **MCP Support** | Yes (stdio and remote transport) |
| **Config Format** | JSON (`opencode.json` / `opencode.jsonc`) |

**Config file locations (priority order):**
1. Project-local: `./opencode.json`
2. XDG config: `$XDG_CONFIG_HOME/opencode/opencode.json`
3. Home: `$HOME/opencode.json`

**Config structure (key fields):**
```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4-5",
  "small_model": "anthropic/claude-haiku-4-5",
  "provider": { "anthropic": { "options": { "timeout": 600000 } } },
  "mcp": {
    "server-name": {
      "type": "local",
      "command": ["npx", "-y", "@upstash/context7-mcp@latest"],
      "environment": { "VAR": "value" },
      "enabled": true
    }
  },
  "tools": {},
  "agent": {},
  "permission": {},
  "instructions": []
}
```

**Source:** [OpenCode Config Docs](https://opencode.ai/docs/config/), [OpenCode MCP Docs](https://opencode.ai/docs/mcp-servers/)

### Crush (Active Successor)

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/charmbracelet/crush](https://github.com/charmbracelet/crush) |
| **Stars** | 21k |
| **Language** | Go (Bubble Tea TUI framework) |
| **License** | **FSL-1.1-MIT** (NOT permissive -- see note below) |
| **Latest Release** | v0.47.2 (March 5, 2026) |
| **MCP Support** | Yes (stdio, HTTP, SSE transports) |
| **Config Format** | JSON (`.crush.json` or `crush.json`) |
| **Maintainer** | Charm (charmbracelet), original OpenCode author Kujtim Hoxha |

**License warning:** FSL-1.1-MIT is a Functional Source License created by Sentry. It is **not** permissive open source. It prohibits commercial competitive use for the first 2 years, then converts to MIT. This means alty cannot fork or embed Crush code, but can generate configs for it.

**Config file locations (priority order):**
1. Project: `.crush.json` or `crush.json`
2. Global: `$HOME/.config/crush/crush.json`

**Config structure:** Similar to OpenCode (Crush is the continuation), with `mcp` section for MCP servers. Schema at `https://charm.land/crush.json`.

**Key insight for alty:** The OpenCode/Crush config format is very close to Claude Code's approach -- JSON with provider, model, MCP, and tool sections. alty already generates Claude Code and Cursor configs; adding OpenCode/Crush config generation is straightforward.

**Source:** [Crush GitHub](https://github.com/charmbracelet/crush), [Crush Blog](https://charm.land/blog/crush-comes-home/), [The New Stack Review](https://thenewstack.io/terminal-user-interfaces-review-of-crush-ex-opencode-al/)

---

## 2. Roo Code

| Attribute | Value |
|-----------|-------|
| **Type** | VS Code Extension |
| **GitHub** | [github.com/RooCodeInc/Roo-Code](https://github.com/RooCodeInc/Roo-Code) |
| **MCP Support** | Yes (stdio, Streamable HTTP, SSE) |
| **Config Format** | JSON |

### MCP Configuration

**Global:** `mcp_settings.json` (VS Code settings area)
**Project:** `.roo/mcp.json` (committed to version control)
**Precedence:** Project overrides global for same server names.

```json
{
  "mcpServers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp@latest"]
    },
    "remote-server": {
      "type": "streamable-http",
      "url": "https://your-server.com/mcp",
      "headers": { "X-API-Key": "key" }
    }
  }
}
```

### Custom Modes & Rules

Roo Code uses a `.roo/` directory for project-level configuration:

```
.roo/
  mcp.json                    # MCP server config
  rules/                      # General rules (markdown/txt files)
    01-general.md
    02-coding-style.md
  rules-{modeSlug}/           # Mode-specific rules
    01-js-style.md
.roomodes                     # Custom mode definitions (YAML or JSON)
```

**.roomodes format (YAML):**
```yaml
customModes:
  - slug: docs-writer
    name: Documentation Writer
    roleDefinition: "You are a technical writer..."
    groups:
      - read
      - ["edit", { fileRegex: "\\.(md|mdx)$" }]
```

**Global rules:** `~/.roo/rules/` and `~/.roo/rules-{modeSlug}/`

### Go-Specific Features

Roo Code is language-agnostic -- no Go-specific integrations beyond what any VS Code extension provides. It can execute Go commands via terminal, read Go files, and use Go-related MCP servers. No special Go support is needed for alty to generate Roo Code configs.

**Key insight for alty:** Roo Code config generation requires:
1. `.roo/mcp.json` -- MCP server config (same `mcpServers` format as Claude Code)
2. `.roo/rules/*.md` -- Agent persona rules (maps to alty's agent personas)
3. `.roomodes` -- Custom mode definitions (maps to alty's agent roles)

**Source:** [Roo Code MCP Docs](https://docs.roocode.com/features/mcp/using-mcp-in-roo), [Roo Code Custom Modes](https://docs.roocode.com/features/custom-modes), [Roo Code Custom Instructions](https://docs.roocode.com/features/custom-instructions)

---

## 3. llama.cpp Go Bindings

### Option A: go-skynet/go-llama.cpp (STALE)

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/go-skynet/go-llama.cpp](https://github.com/go-skynet/go-llama.cpp) |
| **License** | MIT |
| **Status** | **Unmaintained since October 2023** |
| **Approach** | CGo with static library linking |
| **Verdict** | **Do not use.** Over 2 years behind llama.cpp upstream. |

**Source:** [go-skynet/go-llama.cpp](https://github.com/go-skynet/go-llama.cpp), [pkg.go.dev](https://pkg.go.dev/github.com/go-skynet/go-llama.cpp)

### Option B: hybridgroup/yzma (RECOMMENDED for direct integration)

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/hybridgroup/yzma](https://github.com/hybridgroup/yzma) |
| **Stars** | 341 |
| **License** | Apache 2.0 |
| **Latest Release** | v1.10.0 (February 23, 2026) |
| **First Release** | September 21, 2025 |
| **Commits** | 397 |
| **CGo Required** | **No** -- uses purego + libffi |
| **llama.cpp Sync** | Version-synchronized with llama.cpp releases |
| **Hardware Acceleration** | CUDA, Metal, Vulkan, ROCm, SYCL |
| **Platforms** | macOS, Linux, Windows |

**Key advantages:**
- No C compiler needed -- `go build` / `go run` just works
- Automatically downloads pre-built llama.cpp shared libraries
- Works with 164,000+ GGUF models from Hugging Face
- Supports VLMs (vision language models), not just text
- Tests run automatically with each new llama.cpp release
- Backed by HybridGroup (the TinyGo / GoBot team -- credible maintainers)

**Source:** [yzma GitHub](https://github.com/hybridgroup/yzma), [GoLab Talk October 2025](https://social.tinygo.org/@deadprogram/115328423310484842)

### Option C: dianlight/gollama.cpp (NICHE)

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/dianlight/gollama.cpp](https://github.com/dianlight/gollama.cpp) |
| **Stars** | 20 |
| **License** | MIT |
| **Approach** | purego (similar to yzma) |
| **CGo Required** | No |
| **Verdict** | Functional but yzma is better maintained and more widely adopted. |

**Source:** [gollama.cpp GitHub](https://github.com/dianlight/gollama.cpp)

### Option D: ardanlabs/kronk (HIGH-LEVEL WRAPPER)

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/ardanlabs/kronk](https://github.com/ardanlabs/kronk) |
| **Stars** | 215 |
| **License** | Apache 2.0 |
| **Built on** | yzma |
| **Commits** | 415 |
| **Key Feature** | OpenAI-compatible API + model server |

Kronk wraps yzma with a high-level API that feels like using the OpenAI API. Includes a model server for chat completions, embeddings, and reranking. Covers 94% of llama.cpp functionality via yzma.

**Relevance:** If alty ever needs to ship a self-contained model server, kronk shows it is feasible. But for alty's use case (connecting to local LLMs, not running them), this is overkill.

**Source:** [kronk GitHub](https://github.com/ardanlabs/kronk)

---

## 4. Ollama

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/ollama/ollama](https://github.com/ollama/ollama) |
| **Stars** | 164k |
| **Language** | Go (with llama.cpp C/C++ for inference) |
| **License** | MIT (CLI and server; desktop app has separate license) |
| **Latest Release** | v0.17.7 (March 6, 2026) |
| **API** | REST API on localhost:11434 |

### Official Go Client

| Attribute | Value |
|-----------|-------|
| **Import** | `github.com/ollama/ollama/api` |
| **Latest** | v0.17.7 (March 6, 2026) |
| **License** | MIT |
| **Importers** | 339 known packages |

```go
import "github.com/ollama/ollama/api"

client, err := api.ClientFromEnvironment()  // reads OLLAMA_HOST
// or
client := api.NewClient(baseURL, httpClient)

// Chat completion
resp, err := client.Chat(ctx, &api.ChatRequest{
    Model:    "llama3",
    Messages: []api.Message{{Role: "user", Content: "Hello"}},
})
```

**Key methods:** `Generate()`, `Chat()`, `Embed()`, `Pull()`, `List()`, `Show()`, `Create()`, `Delete()`, `ListRunning()`, `Version()`, `Heartbeat()`

**Tool calling:** Supported via `api.Tool` and `api.ToolCall` types.

**Security note:** As of March 2026, the package has 9 known vulnerabilities (DoS, auth bypass). Review CVEs before production use.

### OpenAI Compatibility

Ollama provides an OpenAI-compatible REST API at `http://localhost:11434/v1/`:

- `/v1/chat/completions` -- Chat (streaming and non-streaming)
- `/v1/embeddings` -- Embeddings
- `/v1/models` -- Model listing

This means any OpenAI client (including sashabaranov/go-openai) works with Ollama by changing `BaseURL`.

```go
config := openai.DefaultConfig("ollama")
config.BaseURL = "http://localhost:11434/v1"
client := openai.NewClientWithConfig(config)
```

### Can Ollama Be Embedded In-Process?

**No.** GitHub issue [#7450](https://github.com/ollama/ollama/issues/7450) requested this but it remains open with no resolution. The `ollama/api` package is an HTTP client -- it requires a running Ollama server. Community attempts to call internal `cmd.RunServer` failed due to missing initialization.

For in-process llama.cpp integration without a server, use **yzma** instead.

**Source:** [Ollama GitHub](https://github.com/ollama/ollama), [Ollama API Docs](https://docs.ollama.com/api/openai-compatibility), [Ollama Go API pkg.go.dev](https://pkg.go.dev/github.com/ollama/ollama/api), [Issue #7450](https://github.com/ollama/ollama/issues/7450)

---

## 5. OpenAI-Compatible Go Client: sashabaranov/go-openai

| Attribute | Value |
|-----------|-------|
| **GitHub** | [github.com/sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) |
| **Stars** | 10.6k |
| **License** | Apache 2.0 |
| **Latest Release** | v1.41.2 |
| **Go Requirement** | Go 1.18+ |
| **Published** | August 29, 2025 (pkg.go.dev) |

**Features:**
- Chat completions (streaming and non-streaming)
- Function calling / tool use
- Text embeddings
- Image generation (DALL-E)
- Audio transcription (Whisper)
- Multiple API types: OpenAI, Azure, Azure AD, Cloudflare, **Anthropic**

**Universal client capability:** Works with any OpenAI-compatible API by changing `BaseURL`:
- OpenAI: default
- Ollama: `http://localhost:11434/v1`
- Groq: `https://api.groq.com/openai/v1`
- Any other OpenAI-compatible provider

**Source:** [go-openai GitHub](https://github.com/sashabaranov/go-openai), [pkg.go.dev](https://pkg.go.dev/github.com/sashabaranov/go-openai)

---

## 6. AI Tool Config Format Comparison

| Tool | Config File | MCP Key | Format | Committed? |
|------|------------|---------|--------|------------|
| Claude Code | `.claude/settings.json` | `mcpServers` | JSON | Yes |
| Cursor | `.cursor/mcp.json` | `mcpServers` | JSON | Yes |
| Roo Code | `.roo/mcp.json` | `mcpServers` | JSON | Yes |
| OpenCode | `opencode.json` | `mcp` | JSON | Yes |
| Crush | `.crush.json` | `mcp` | JSON | Yes |

**Key observation:** All five tools use JSON for MCP config. Four of five use the identical `mcpServers` key structure. OpenCode/Crush use a slightly different `mcp` key with `type: "local"` wrapper, but the content (command, args, env) is nearly identical.

### Agent/Rules Config Comparison

| Tool | Agent Config | Format | Notes |
|------|-------------|--------|-------|
| Claude Code | `.claude/agents/*.md` + `CLAUDE.md` | Markdown | Agent personas in markdown files |
| Cursor | `.cursor/rules/*.mdc` | MDC (Markdown) | Rules with metadata frontmatter |
| Roo Code | `.roo/rules/*.md` + `.roomodes` | Markdown + YAML/JSON | Mode definitions + rule files |
| OpenCode | `opencode.json` `agent` key | JSON | Inline agent definitions |
| Crush | `.crush.json` `agent` key | JSON | Inline agent definitions |

**Key insight for alty:** All tools share the same general pattern:
1. MCP server config (JSON, nearly identical across tools)
2. Agent/persona rules (markdown files or JSON config)
3. Project-level settings that override global defaults

alty's config generator can use a shared domain model (MCP servers, agent personas, quality gates) and translate to each tool's native format.

---

## 7. Decision: Local LLM Strategy for Go-based alty

### Option Analysis

| Strategy | Complexity | Binary Size | Startup | Dependencies | Offline? |
|----------|-----------|-------------|---------|-------------|----------|
| **A. Ollama HTTP API** | Low | +0 MB | <100ms (client only) | Ollama installed separately | Yes (with Ollama running) |
| **B. yzma (direct llama.cpp)** | Medium | +shared lib (~50-200MB) | 1-5s (model load) | llama.cpp shared libs | Yes (fully self-contained) |
| **C. Ollama CLI subprocess** | Low | +0 MB | 200-500ms | Ollama installed | Yes (with Ollama running) |
| **D. go-openai universal client** | Low | +0 MB | <100ms | Any OpenAI-compatible server | Depends on provider |

### Recommendation: Tiered Strategy

**Tier 1 (implement first): Ollama HTTP API via official Go client**
- Import `github.com/ollama/ollama/api` for Ollama-native API
- Use `sashabaranov/go-openai` as universal OpenAI-compatible client
- Works with Ollama, OpenAI, Anthropic, Groq, Azure, any compatible provider
- Zero binary size overhead; user installs Ollama separately
- alty detects Ollama availability via `api.Heartbeat()`
- This covers 95% of local LLM use cases

**Tier 2 (future, if needed): Direct llama.cpp via yzma**
- Only if users demand single-binary with embedded model inference
- No CGo required (purego), but adds shared library dependency
- Significantly more complex: model management, GPU detection, memory management
- Use ardanlabs/kronk as reference implementation

**NOT recommended:**
- Embedding Ollama in-process: not supported (issue #7450 open, no resolution)
- go-skynet/go-llama.cpp: unmaintained since October 2023
- Subprocess to ollama CLI: fragile, harder to parse output than HTTP API

### Architecture Implication

```
                      +-------------------+
                      |   alty CLI (Go)    |
                      +-------------------+
                              |
                    +--------------------+
                    |  LLM Port (interface) |
                    +--------------------+
                    /         |          \
            +--------+  +--------+  +--------+
            |Anthropic|  | OpenAI |  | Ollama |
            |Adapter  |  |Adapter |  |Adapter |
            +--------+  +--------+  +--------+
                |            |           |
          anthropic-   go-openai    ollama/api
          sdk-go     (universal)    (native)
```

The Port interface should expose:
- `Chat(ctx, messages, options) -> Response`
- `Embed(ctx, text, model) -> []float64`
- `ListModels(ctx) -> []Model`
- `IsAvailable(ctx) -> bool`

The go-openai adapter can serve as the "universal" adapter for any OpenAI-compatible provider, including Ollama's OpenAI endpoint. The native Ollama adapter provides access to Ollama-specific features (model pulling, server management).

---

## 8. Config Generation Impact for alty

Adding support for OpenCode/Crush and Roo Code requires these new config generators:

### OpenCode/Crush Config Generator
- Output: `opencode.json` or `.crush.json`
- Content: model selection, MCP servers, agent definitions, tool permissions
- Complexity: Low -- JSON format, similar structure to existing Claude Code generator

### Roo Code Config Generator
- Output: `.roo/mcp.json`, `.roo/rules/*.md`, `.roomodes`
- Content: MCP servers, agent persona rules as markdown, custom mode definitions
- Complexity: Medium -- multiple files, markdown rule files, YAML/JSON mode defs
- Note: `.roo/rules/` maps directly to alty's existing agent persona templates

### Shared MCP Config Model
All tools share enough structure that a single domain model can represent MCP config:
```
MCPServerConfig {
  name: str
  command: str
  args: list[str]
  env: dict[str, str]
  enabled: bool
  transport: "stdio" | "http" | "sse"
  url: str (for remote)
}
```

This maps cleanly to all five tool formats with minimal per-tool translation.

---

## Summary of Findings

| Question | Answer |
|----------|--------|
| Is OpenCode written in Go? | Yes. Archived Sep 2025. Continued as Crush (charmbracelet). |
| Does OpenCode/Crush support MCP? | Yes. JSON config. stdio + HTTP + SSE transports. |
| Does Roo Code support MCP? | Yes. `.roo/mcp.json`. stdio + HTTP + SSE transports. |
| Best Go llama.cpp binding? | yzma (Apache 2.0, 341 stars, no CGo, actively maintained). |
| Is Ollama written in Go? | Yes. 164k stars, MIT license. |
| Ollama Go client? | Official: `ollama/ollama/api` (MIT, v0.17.7). |
| Ollama OpenAI compatibility? | Yes. `/v1/` endpoint works with any OpenAI client. |
| Can Go binary embed Ollama? | No. Server must run separately. |
| Best universal Go LLM client? | sashabaranov/go-openai (Apache 2.0, 10.6k stars). |
| Local LLM strategy? | Connect to Ollama via HTTP API (Tier 1). Direct yzma (Tier 2). |

---

## License Summary

| Library | License | Permissive? | OK for alty? |
|---------|---------|-------------|-------------|
| OpenCode (archived) | MIT | Yes | N/A (archived) |
| Crush | FSL-1.1-MIT | **No** (2-year restriction) | Config gen only, no code reuse |
| Roo Code | Apache 2.0 | Yes | Config gen only (VS Code ext) |
| ollama/ollama | MIT | Yes | Yes |
| ollama/ollama/api | MIT | Yes | Yes |
| sashabaranov/go-openai | Apache 2.0 | Yes | Yes |
| hybridgroup/yzma | Apache 2.0 | Yes | Yes |
| ardanlabs/kronk | Apache 2.0 | Yes | Yes (reference only) |
| go-skynet/go-llama.cpp | MIT | Yes | No (unmaintained) |
| dianlight/gollama.cpp | MIT | Yes | Functional but prefer yzma |

---

## Follow-Up Work

1. **Config generator tickets** -- Create implementation tickets for OpenCode/Crush and Roo Code config generators
2. **LLM Port design** -- Design the `LLMPort` interface in alty's application layer
3. **Ollama adapter** -- Implement Ollama adapter using `ollama/api` package
4. **Universal OpenAI adapter** -- Implement adapter using `sashabaranov/go-openai`
5. **Tool detection** -- Add OpenCode/Crush and Roo Code to alty's tool detection logic
