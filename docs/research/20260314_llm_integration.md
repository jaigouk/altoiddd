---
last_reviewed: 2026-03-14
owner: researcher
status: complete
type: spike
ticket: alty-cli-f0g.1
---

# LLM Provider Detection and Integration Options

## Decision Context

alty needs to detect LLM credentials at `alty init` time so the guided discovery flow can use LLM-powered features (natural language understanding, gap inference, DDD question enrichment). The codebase already has a provider-agnostic LLM layer that this spike must build on — not redesign.

### Existing Infrastructure (from `internal/shared/infrastructure/llm/`)

| File | Purpose | Key Details |
|------|---------|-------------|
| `client.go` | `Client` interface, `Provider` enum, `Config` VO, `Response` VO | Two methods: `StructuredOutput`, `TextCompletion` |
| `factory.go` | `Factory.Create(Config) Client` | Switch on `Provider`, graceful degradation to `NoopClient` |
| `anthropic_client.go` | Direct HTTP adapter for Anthropic Messages API | No SDK — uses `net/http`, `encoding/json` |
| `noop_client.go` | Always returns `ErrLLMUnavailable` | Enables local-first degradation |
| `errors.go` | `ErrLLMUnavailable` sentinel | Used by `NoopClient` and `AnthropicClient` |

**Existing Provider enum:** `ProviderAnthropic`, `ProviderOllama`, `ProviderVertexAI`, `ProviderNone`

**Existing Config fields:** `provider`, `model`, `apiKey`, `timeout`

### Project Constraints (from `docs/PRD.md`)

| Constraint | Value |
|-----------|-------|
| Language | Go 1.26+ |
| No cloud dependencies for core | Core functionality runs without API keys |
| LLM is enhancement | `NoopClient` degradation path must be preserved |
| Minimal dependencies | High bar for adding SDK deps; existing HTTP client works |
| Cross-compilation | Prefer pure Go (no CGO) |

---

## 1. Credential Detection Locations and Order

### Recommended Detection Order

The detection order follows the principle of **closest scope wins** — project-level config overrides user-level, which overrides environment.

| Priority | Source | Credentials Found | Rationale |
|----------|--------|-------------------|-----------|
| 1 (highest) | **CLI flags** | `--llm-provider`, `--llm-api-key`, `--llm-model` | Explicit user intent for this invocation |
| 2 | **Project config** | `.alty/config.toml` `[llm]` section | Project-specific provider choice |
| 3 | **Environment variables** | `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `OPENAI_BASE_URL`, `OLLAMA_HOST` | Standard provider env vars |
| 4 | **User config** | `~/.config/alty/config.toml` `[llm]` section | User-level default |
| 5 | **Claude Code config** | `~/.claude.json` or `~/.claude/claude.json` | Opportunistic — user already has Anthropic key |
| 6 (lowest) | **None detected** | → `ProviderNone` → `NoopClient` | Graceful degradation |

### Environment Variables to Detect

| Variable | Provider | Notes |
|----------|----------|-------|
| `ANTHROPIC_API_KEY` | `ProviderAnthropic` | Standard. Used by anthropic-sdk-go, aider, langchaingo ([Source: Anthropic SDK Go README](https://github.com/anthropics/anthropic-sdk-go)) |
| `OPENAI_API_KEY` | `ProviderOpenAI` | Standard. Used by openai-go, aider, mods, langchaingo ([Source: OpenAI Go README](https://github.com/openai/openai-go)) |
| `OPENAI_BASE_URL` | `ProviderOpenAICompatible` | Signals OpenAI-compatible endpoint (LM Studio, vLLM, etc.) |
| `OLLAMA_HOST` | `ProviderOllama` | Default: `http://localhost:11434`. No API key needed. ([Source: Ollama docs](https://docs.ollama.com/api/authentication)) |
| `GOOGLE_APPLICATION_CREDENTIALS` | `ProviderVertexAI` | Service account JSON path |

### Claude Code Config Detection

When alty detects it's running inside a Claude Code session (or `~/.claude.json` exists), it can opportunistically read the Anthropic API key. This is a **read-only, best-effort** detection — alty never writes to Claude config files.

**Detection logic:**
```
1. Check if ~/.claude.json exists
2. Parse JSON, look for api_key or auth token
3. If found → suggest ProviderAnthropic with detected key
4. User confirms or overrides during `alty init`
```

Source: [Claude Code settings docs](https://code.claude.com/docs/en/settings), [Managing API Keys in Claude Code](https://support.claude.com/en/articles/12304248-managing-api-key-environment-variables-in-claude-code)

### Config File Format

```toml
# .alty/config.toml or ~/.config/alty/config.toml
[llm]
provider = "anthropic"          # anthropic | openai | openai-compatible | ollama | vertexai | none
model = "claude-sonnet-4-20250514"
api_key_env = "ANTHROPIC_API_KEY"  # reference env var, never store raw key
base_url = ""                   # only for openai-compatible
timeout = 30.0
```

**Security rule:** Config files store `api_key_env` (env var name reference), never raw API keys. This prevents accidental commits of credentials.

---

## 2. SDK Comparison

### Evaluation Criteria

For each SDK: what does it add over the existing direct HTTP client at `anthropic_client.go`?

The existing HTTP client is ~140 lines, handles one endpoint (Messages API), and has zero external dependencies beyond `net/http`. The bar for adding an SDK dependency is: **does it solve a problem we actually have?**

### anthropic-sdk-go

| Attribute | Value | Source |
|-----------|-------|--------|
| Module | `github.com/anthropics/anthropic-sdk-go` | [pkg.go.dev](https://pkg.go.dev/github.com/anthropics/anthropic-sdk-go) |
| Version | v1.26.0 (Feb 19, 2026) | [Releases](https://github.com/anthropics/anthropic-sdk-go/releases) |
| License | MIT | [pkg.go.dev](https://pkg.go.dev/github.com/anthropics/anthropic-sdk-go) |
| Min Go | 1.22+ | go.mod |
| Imports | 33 | pkg.go.dev |
| CGO | No | Pure Go |
| Maintained by | Anthropic (official) | GitHub org |

**What it adds over existing HTTP client:**

| Feature | Existing Client Has? | SDK Adds? | Value to alty |
|---------|---------------------|-----------|---------------|
| Messages API | Yes (hardcoded) | Yes (typed) | Low — we only use one endpoint |
| Streaming | No | Yes | Low — alty doesn't stream |
| Tool use / function calling | No | Yes (typed structs) | Medium — future DDD question enrichment |
| Error types (`*anthropic.Error`) | Basic (`ErrLLMUnavailable`) | Rich (status, request/response dump) | Low — we wrap all errors anyway |
| Auto-retry with backoff | No | Yes | Medium — production resilience |
| Structured output helpers | Manual (JSON schema in system prompt) | Better (native tool_use) | Medium |
| Model constants | Hardcoded string | Typed constants | Low — cosmetic |
| API versioning | Hardcoded header | Automatic | Low |

**Verdict:** The SDK adds tool use support and retry logic, but the existing 140-line HTTP client covers alty's current needs (text completion + JSON-schema structured output). **Do not add yet.** Revisit when tool use becomes a requirement (f0g.4 conversational flow).

### openai-go (official)

| Attribute | Value | Source |
|-----------|-------|--------|
| Module | `github.com/openai/openai-go/v3` | [pkg.go.dev](https://pkg.go.dev/github.com/openai/openai-go/v3) |
| Version | v3.28.0 (Mar 14, 2026) | [Releases](https://github.com/openai/openai-go/releases) |
| License | Apache-2.0 | [pkg.go.dev](https://pkg.go.dev/github.com/openai/openai-go/v3) |
| Min Go | 1.22+ | go.mod |
| Imports | 30 | pkg.go.dev |
| CGO | No | Pure Go |
| Maintained by | OpenAI (official) | GitHub org |

**What it adds over existing HTTP client:**

| Feature | Value to alty |
|---------|---------------|
| Structured outputs (native JSON schema) | Medium — typed schema enforcement |
| Azure OpenAI support | Low — not a target provider |
| Base URL override | High — enables OpenAI-compatible providers |
| Typed model constants | Low |
| `option.WithBaseURL()` | Key for LM Studio, vLLM, Groq, etc. |

**Verdict:** If alty adds OpenAI/compatible provider support, we should write a direct HTTP client (like `AnthropicClient`) rather than pull in the SDK. The OpenAI Chat Completions API is simpler than Anthropic's Messages API — a direct client would be ~100 lines. **Do not add.** Write `OpenAIClient` adapter with `net/http`.

### langchaingo

| Attribute | Value | Source |
|-----------|-------|--------|
| Module | `github.com/tmc/langchaingo` | [pkg.go.dev](https://pkg.go.dev/github.com/tmc/langchaingo) |
| Version | v0.1.13+ (2026) | [Releases](https://github.com/tmc/langchaingo/releases) |
| License | MIT | GitHub |
| Min Go | 1.22+ | go.mod |
| Dependencies | **Heavy** — pulls in drivers for 15+ providers, vector stores, agents | go.sum |
| CGO | Some transitive deps may require CGO | Depends on enabled features |
| Maintained by | Community (tmc) | GitHub |

**What it adds over existing HTTP client:**

| Feature | Value to alty |
|---------|---------------|
| Multi-provider abstraction | We already have this (`Client` interface + `Factory`) |
| Chain/Agent patterns | Not needed — alty is not an agent framework |
| Vector store integration | Not needed |
| Memory management | Not needed |

**Verdict:** **Do not add.** langchaingo is a framework, not a library. It would replace alty's clean `Client` interface with a much heavier abstraction. The dependency tree is massive for what alty needs. alty's existing `Client` + `Factory` pattern is the right level of abstraction.

### Summary Decision Matrix

| SDK | License | Pure Go | Dependencies | Value-Add | Recommendation |
|-----|---------|---------|-------------|-----------|----------------|
| anthropic-sdk-go | MIT | Yes | 33 | Tool use, retry | **Skip** — revisit at f0g.4 |
| openai-go v3 | Apache-2.0 | Yes | 30 | Base URL, structured output | **Skip** — write direct HTTP client |
| langchaingo | MIT | Partial | Heavy (100+) | Multi-provider framework | **Hard no** — wrong abstraction level |

**Overall recommendation:** Continue with direct HTTP clients. Write an `OpenAIClient` adapter (~100 lines) following the same pattern as `AnthropicClient`. This keeps the dependency count at zero for the LLM layer.

---

## 3. Provider Enum Recommendation

### Current State

```go
type Provider string
const (
    ProviderAnthropic Provider = "anthropic"
    ProviderOllama    Provider = "ollama"
    ProviderVertexAI  Provider = "vertexai"
    ProviderNone      Provider = "none"
)
```

### Recommended Changes

Add two new providers:

```go
const (
    ProviderAnthropic        Provider = "anthropic"
    ProviderOpenAI           Provider = "openai"            // NEW
    ProviderOpenAICompatible Provider = "openai-compatible"  // NEW
    ProviderOllama           Provider = "ollama"
    ProviderVertexAI         Provider = "vertexai"
    ProviderNone             Provider = "none"
)
```

**Rationale:**

| Provider | Why |
|----------|-----|
| `ProviderOpenAI` | OpenAI is the second-most common LLM provider. Users with `OPENAI_API_KEY` should have first-class support. |
| `ProviderOpenAICompatible` | Many local/cloud providers expose OpenAI-compatible endpoints (LM Studio, vLLM, Groq, Together AI, Fireworks). A single provider type with `base_url` covers all of them. |

**Why not just `ProviderOpenAI` with base URL override?** Separating them makes credential detection cleaner:
- `OPENAI_API_KEY` alone → `ProviderOpenAI` (api.openai.com)
- `OPENAI_API_KEY` + `OPENAI_BASE_URL` → `ProviderOpenAICompatible` (custom endpoint)
- This distinction matters for model selection defaults and error messages.

### Config Changes

`Config` needs a new field:

```go
type Config struct {
    provider Provider
    model    string
    apiKey   string
    baseURL  string   // NEW — used by OpenAI, OpenAICompatible, Ollama
    timeout  float64
}
```

The `baseURL` field has defaults per provider:
- `ProviderOpenAI`: `https://api.openai.com/v1`
- `ProviderOpenAICompatible`: required (no default)
- `ProviderOllama`: `http://localhost:11434`
- Others: not used

---

## 4. Config/Factory Integration Design

### How Detected Credentials Flow into Existing Infrastructure

```
┌─────────────────┐
│   CLI flags      │──┐
│   Project config │──┤
│   Env vars       │──┤──▶ CredentialDetector ──▶ llm.Config ──▶ Factory.Create() ──▶ Client
│   User config    │──┤         (NEW)              (existing)      (existing)         (existing)
│   Claude config  │──┤
│   No creds       │──┘                             ▼
│                  │                           ProviderNone
│                  │                                 ▼
│                  │                           NoopClient
└─────────────────┘                           (existing)
```

### New Component: `CredentialDetector`

Lives in `internal/shared/infrastructure/llm/` as a new file.

```go
// credential_detector.go

// DetectedCredentials holds the result of credential detection.
type DetectedCredentials struct {
    Provider Provider
    APIKey   string
    BaseURL  string
    Model    string
    Source   string // "cli-flag", "project-config", "env:ANTHROPIC_API_KEY", etc.
}

// CredentialDetector scans configured sources for LLM credentials.
type CredentialDetector struct {
    envReader    func(string) string          // os.Getenv or test stub
    fileReader   func(string) ([]byte, error) // os.ReadFile or test stub
    projectDir   string
}

// Detect returns the highest-priority credentials found, or a
// DetectedCredentials with Provider=ProviderNone if nothing detected.
func (d *CredentialDetector) Detect() DetectedCredentials {
    // Check sources in priority order (see Section 1)
    // Return first match with Source field set for observability
}
```

### Factory Changes

`Factory.Create` remains unchanged. The new flow is:

```go
// In composition root (internal/composition/app.go):
detector := llm.NewCredentialDetector(os.Getenv, os.ReadFile, projectDir)
creds := detector.Detect()
config := llm.NewConfig(creds.Provider, creds.Model, creds.APIKey, 30.0)
// When baseURL field is added:
// config := llm.NewConfigWithBaseURL(creds.Provider, creds.Model, creds.APIKey, creds.BaseURL, 30.0)
client := llm.Factory{}.Create(config)
```

### Factory Extension for New Providers

```go
func (f Factory) Create(config Config) Client {
    switch config.Provider() {
    case ProviderAnthropic:
        if config.APIKey() == "" {
            return &NoopClient{}
        }
        return NewAnthropicClient(config.APIKey(), config.Model(), config.Timeout())
    case ProviderOpenAI:
        if config.APIKey() == "" {
            return &NoopClient{}
        }
        return NewOpenAIClient(config.APIKey(), config.Model(), "", config.Timeout())
    case ProviderOpenAICompatible:
        if config.APIKey() == "" || config.BaseURL() == "" {
            return &NoopClient{}
        }
        return NewOpenAIClient(config.APIKey(), config.Model(), config.BaseURL(), config.Timeout())
    case ProviderOllama:
        baseURL := config.BaseURL()
        if baseURL == "" {
            baseURL = "http://localhost:11434"
        }
        return NewOllamaClient(baseURL, config.Model(), config.Timeout())
    case ProviderVertexAI, ProviderNone:
        return &NoopClient{}
    }
    return &NoopClient{}
}
```

### OpenAIClient Adapter (Sketch)

A direct HTTP client following the `AnthropicClient` pattern. Uses the OpenAI Chat Completions API (`/v1/chat/completions`). ~100 lines.

```go
// openai_client.go
type OpenAIClient struct {
    apiKey     string
    model      string
    baseURL    string
    httpClient *http.Client
}

func NewOpenAIClient(apiKey, model, baseURL string, timeout float64) *OpenAIClient {
    if model == "" {
        model = "gpt-4o"
    }
    if baseURL == "" {
        baseURL = "https://api.openai.com/v1"
    }
    // ... similar to AnthropicClient
}
```

The OpenAI Chat Completions API uses `Authorization: Bearer <key>` header and a simpler request/response format than Anthropic's Messages API. The same client works for OpenAI-compatible providers (LM Studio, vLLM, etc.) by changing `baseURL`.

---

## 5. CLI Tool Survey: LLM Credential Detection Patterns

### Go-Native Tools

#### Charm Mods (Go, MIT) — ARCHIVED March 2026

| Attribute | Detail |
|-----------|--------|
| Language | Go |
| Detection | Env vars per provider: `OPENAI_API_KEY`, `AZURE_OPENAI_KEY`, `COHERE_API_KEY`, `GROQ_API_KEY`, `GOOGLE_API_KEY` |
| Config file | `~/.config/mods/mods.yml` (XDG) |
| Priority | Config file → env vars (env override for API keys) |
| Default provider | OpenAI (GPT-4, falls back to GPT-3.5 Turbo) |
| Multi-provider | Yes — config file defines named "apis" with base URL + models |
| Status | **Archived** March 9, 2026. Succeeded by Crush. |

Source: [GitHub charmbracelet/mods](https://github.com/charmbracelet/mods)

**Pattern takeaway:** Named API configurations in YAML/TOML with env var overrides for keys. Clean separation of "which provider" (config) from "what key" (env var).

#### Ollama CLI (Go, MIT)

| Attribute | Detail |
|-----------|--------|
| Language | Go |
| Detection | `OLLAMA_HOST` env var, defaults to `http://localhost:11434` |
| API key | Not required for local. `OLLAMA_API_KEY` for cloud (ollama.com). Bearer token auth. |
| Config file | None — env vars only |
| Multi-provider | No — Ollama only |

Source: [Ollama Authentication docs](https://docs.ollama.com/api/authentication)

**Pattern takeaway:** Local-first with zero config. Env var for non-default host. API key only needed for cloud endpoint.

### Non-Go Tools (Pattern Inspiration Only)

#### Aider (Python, Apache-2.0)

| Attribute | Detail |
|-----------|--------|
| Language | **Python** |
| Detection order | 1. CLI flags (`--openai-api-key`) → 2. Env vars (`OPENAI_API_KEY`) → 3. `.env` file → 4. `.aider.conf.yml` |
| Special support | OpenAI and Anthropic have dedicated CLI flags; other providers use generic `--api-key provider=<key>` |
| Auto-model selection | Detects which keys are available, selects best model from available providers |
| Multi-provider | Yes — via LiteLLM abstraction layer |
| Config locations | `~/.aider.conf.yml` (user), `.aider.conf.yml` (project), `.env` (project) |

Source: [Aider API Keys docs](https://aider.chat/docs/config/api-keys.html), [Aider .env config](https://aider.chat/docs/config/dotenv.html)

**Pattern takeaways:**
- **Auto-model selection based on available keys** — excellent UX. If user has `ANTHROPIC_API_KEY`, use Claude. If they have `OPENAI_API_KEY`, use GPT-4. If both, prefer Claude (user can override).
- **CLI flags > env vars > config files** — standard precedence.
- **Dedicated flags for top-2 providers** — reduces friction for the common case.

#### Continue.dev (TypeScript, Apache-2.0)

| Attribute | Detail |
|-----------|--------|
| Language | **TypeScript** |
| Detection | `~/.continue/config.yaml` with secret interpolation: `${{ secrets.API_KEY_NAME }}` |
| Multi-provider | Yes — 40+ providers via unified `ILLM` interface |
| Auto-detection | Minimal — credentials must be explicitly configured |
| Capability detection | Automatic (tool use, streaming) based on model |

Source: [Continue Model Providers docs](https://docs.continue.dev/customize/model-providers/overview), [Continue Configuration docs](https://docs.continue.dev/customize/deep-dives/configuration)

**Pattern takeaway:** Secret interpolation (`${{ secrets.X }}`) is interesting but overkill for alty. The capability detection pattern (auto-detecting what a model supports) is worth noting for future work.

### Cross-Tool Pattern Summary

| Pattern | Used By | Adopt for alty? |
|---------|---------|-----------------|
| CLI flags > env vars > config file priority | Aider, Mods | **Yes** — standard and unsurprising |
| Env var per provider (`PROVIDER_API_KEY`) | All tools | **Yes** — universal convention |
| Auto-select provider from available keys | Aider | **Yes** — great UX for `alty init` |
| Config file never stores raw keys | Continue (via interpolation) | **Yes** — store env var name reference |
| Named API configurations | Mods | **No** — overkill for alty's use case |
| LiteLLM/framework abstraction | Aider | **No** — alty's `Client` interface is sufficient |
| Local-first with zero config | Ollama | **Yes** — aligns with `NoopClient` pattern |

---

## 6. Follow-Up Ticket Recommendations for f0g.2

### f0g.2: Implement LLM Credential Detection at Init

Based on this spike's findings, f0g.2 should implement:

1. **`CredentialDetector`** in `internal/shared/infrastructure/llm/credential_detector.go`
   - Priority order: CLI flags → project config → env vars → user config → Claude config → none
   - Returns `DetectedCredentials` with `Source` field for observability
   - Pure Go, testable with injected `envReader` and `fileReader`

2. **Provider enum additions** in `client.go`
   - Add `ProviderOpenAI` and `ProviderOpenAICompatible`
   - Add `baseURL` field to `Config`

3. **`OpenAIClient`** adapter in `openai_client.go`
   - Direct HTTP client (~100 lines), same pattern as `AnthropicClient`
   - Supports base URL override for OpenAI-compatible providers

4. **Factory updates** in `factory.go`
   - Handle new providers in `Create` switch

5. **Config file support** in composition root
   - Read `.alty/config.toml` `[llm]` section
   - Store `api_key_env` reference, never raw keys

### Additional Follow-Up Tickets

| Ticket | Description | Depends On |
|--------|-------------|------------|
| **f0g.2.1** (if f0g.2 is split) | `CredentialDetector` + provider enum changes | f0g.1 (this spike) |
| **f0g.2.2** (if f0g.2 is split) | `OpenAIClient` HTTP adapter | f0g.2.1 |
| **New: OllamaClient adapter** | Direct HTTP client for Ollama `/api/generate` endpoint | f0g.2 |
| **New: SDK evaluation revisit** | Re-evaluate anthropic-sdk-go when tool use is needed | f0g.4 |

---

## Research Questions — Answers

### Q1: What LLM credential locations should alty detect?

**Answer:** Six sources in priority order: CLI flags → project `.alty/config.toml` → env vars (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `OPENAI_BASE_URL`, `OLLAMA_HOST`) → user `~/.config/alty/config.toml` → Claude Code config (`~/.claude.json`) → none (→ `NoopClient`). See Section 1 for full details.

### Q2: Which Go SDKs add value over the existing direct HTTP client?

**Answer:** None add sufficient value to justify the dependency cost today. The existing `AnthropicClient` pattern (direct HTTP, ~140 lines) should be replicated for OpenAI. anthropic-sdk-go becomes worth revisiting when tool use is needed (f0g.4). langchaingo is a hard no — wrong abstraction level. See Section 2 for full comparison.

### Q3: Should alty support OpenAI/compatible providers?

**Answer:** Yes. Add `ProviderOpenAI` and `ProviderOpenAICompatible` to the enum. Write a direct `OpenAIClient` HTTP adapter (~100 lines). The `ProviderOpenAICompatible` + `baseURL` covers LM Studio, vLLM, Groq, Together AI, and similar providers with one adapter. See Section 3.

### Q4: How do other CLI tools handle LLM credential detection?

**Answer:** The universal pattern is CLI flags > env vars > config files, with env var names following `PROVIDER_API_KEY` convention. Aider's auto-model-selection (pick best model from available keys) is the standout UX pattern alty should adopt. All tools use local-first degradation. See Section 5.
