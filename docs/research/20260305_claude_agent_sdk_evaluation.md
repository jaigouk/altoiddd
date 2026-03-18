# Claude Agent SDK for Python -- Evaluation Report

**Date:** 2026-03-05
**Researcher:** researcher agent
**Status:** Complete

## Research Questions

1. Can the Claude Agent SDK be used programmatically (not just as CLI)?
2. Can it serve as an LLM backend for generating text/analysis within another Python application?
3. What is the API surface -- can we send prompts and get structured responses?
4. Does it require API keys or does it use the bundled Claude Code CLI auth?
5. How would we use it as a "port adapter" in hexagonal architecture (e.g., `ChallengerPort`)?
6. Is it possible to use it for non-interactive (programmatic) LLM calls?

---

## 1. Package Identity

| Attribute        | Value                                                                                       |
| ---------------- | ------------------------------------------------------------------------------------------- |
| **Package name** | `claude-agent-sdk` (successor to deprecated `claude-code-sdk`)                              |
| **PyPI**         | https://pypi.org/project/claude-agent-sdk/                                                  |
| **GitHub**       | https://github.com/anthropics/claude-agent-sdk-python                                       |
| **Version**      | 0.1.46 (released 2026-03-05)                                                                |
| **License**      | MIT (permissive)                                                                            |
| **Python**       | >=3.10 (supports 3.10, 3.11, 3.12, 3.13)                                                   |
| **Status**       | Alpha (3 - Alpha on PyPI), but very actively maintained (~60 releases since initial)        |
| **Owner**        | Anthropic, PBC                                                                              |

### Deprecation Note

The original `claude-code-sdk` (v0.0.25, 2025-09-29) is **deprecated and no longer maintained**. All users must migrate to `claude-agent-sdk`. The rename reflects Anthropic's recognition that the SDK is not limited to coding tasks.

Source: https://pypi.org/project/claude-code-sdk/

---

## 2. Architecture: How It Works Under the Hood

**Critical architectural fact:** The Claude Agent SDK does NOT call the Anthropic Messages API directly. It spawns a Claude Code CLI subprocess and communicates with it via JSON-over-stdin/stdout. The Claude Code CLI is **bundled** inside the pip package.

```
Your Python app
    |
    v
claude-agent-sdk (Python)
    |
    v  (subprocess spawn + JSON IPC)
Claude Code CLI (bundled Node.js binary)
    |
    v  (HTTPS API calls)
Anthropic API / Bedrock / Vertex AI
```

This is fundamentally different from the `anthropic` Python SDK, which makes direct HTTPS calls to the API.

### Implications

- **Heavier than a direct API call** -- spawns a subprocess for each `query()` call
- **Includes the full Claude Code agent loop** -- tool execution, permission checks, file operations
- **Bundles tools** -- Read, Write, Edit, Bash, Glob, Grep, WebSearch, WebFetch are built-in
- **Cannot be used as a thin LLM wrapper** -- it is an agent framework, not an API client

---

## 3. Answers to Research Questions

### Q1: Can it be used programmatically?

**Yes, fully programmatic.** Two APIs are available:

**`query()` -- One-shot tasks (no session state)**

```python
import asyncio
from claude_agent_sdk import query, ClaudeAgentOptions, ResultMessage

async def main():
    async for message in query(
        prompt="Analyze this code for security issues",
        options=ClaudeAgentOptions(
            system_prompt="You are a security auditor",
            max_turns=1,  # Single turn = one LLM call
            allowed_tools=["Read", "Glob", "Grep"],
            cwd="/path/to/project",
        ),
    ):
        if isinstance(message, ResultMessage):
            print(message.result)

asyncio.run(main())
```

**`ClaudeSDKClient` -- Multi-turn conversations with session state**

```python
import asyncio
from claude_agent_sdk import ClaudeSDKClient, ClaudeAgentOptions, AssistantMessage, TextBlock

async def main():
    async with ClaudeSDKClient(options=ClaudeAgentOptions(
        system_prompt="You are a domain expert",
        max_turns=3,
    )) as client:
        await client.query("What bounded contexts exist in this codebase?")
        async for message in client.receive_response():
            if isinstance(message, AssistantMessage):
                for block in message.content:
                    if isinstance(block, TextBlock):
                        print(block.text)

        # Follow-up in same session -- Claude remembers context
        await client.query("Now evaluate the coupling between them")
        async for message in client.receive_response():
            ...

asyncio.run(main())
```

Source: https://platform.claude.com/docs/en/agent-sdk/python

### Q2: Can it serve as an LLM backend for text generation/analysis?

**Yes, but with caveats.** It is designed as an agent framework, not a simple completion API.

**What works well:**
- Sending a prompt and getting a text response
- Getting structured JSON output via Pydantic schemas
- Using `max_turns=1` for single-turn LLM calls
- Disabling tools entirely for pure text generation

**What is overkill:**
- If you just need `messages.create()` with a prompt and response, use the `anthropic` SDK directly
- The Agent SDK spawns a CLI subprocess, adding ~1-3s startup overhead per `query()` call
- It includes the full agent loop (tool execution, permissions, etc.) even when you don't need it

**For alto's use case (e.g., ChallengerPort that asks probing questions):**
The Agent SDK is appropriate because alto needs more than simple text generation -- it needs structured output, system prompts, and potentially multi-turn conversations. But the `anthropic` SDK would also work and be lighter.

### Q3: API surface -- structured responses?

**Yes, first-class structured output support via JSON Schema / Pydantic.**

```python
import asyncio
from pydantic import BaseModel
from claude_agent_sdk import query, ClaudeAgentOptions, ResultMessage


class ChallengeResult(BaseModel):
    """Structured output from a domain model challenge."""
    challenges: list[str]
    severity: str  # "low", "medium", "high"
    suggestions: list[str]
    confidence: float


async def challenge_domain_model(model_description: str) -> ChallengeResult | None:
    async for message in query(
        prompt=f"Challenge this domain model:\n\n{model_description}",
        options=ClaudeAgentOptions(
            system_prompt="You are a DDD expert. Challenge the domain model.",
            max_turns=1,
            output_format={
                "type": "json_schema",
                "schema": ChallengeResult.model_json_schema(),
            },
        ),
    ):
        if isinstance(message, ResultMessage) and message.structured_output:
            return ChallengeResult.model_validate(message.structured_output)
    return None
```

The `ResultMessage` includes:
- `result: str | None` -- plain text result
- `structured_output: Any` -- validated JSON matching the schema
- `subtype: str` -- "success" or "error_max_structured_output_retries"
- `total_cost_usd: float | None` -- cost tracking
- `usage: dict` -- token usage (input_tokens, output_tokens, cache tokens)

Source: https://platform.claude.com/docs/en/agent-sdk/structured-outputs

### Q4: Authentication -- API keys vs CLI auth?

**API key is required.** Set as environment variable:

```bash
export ANTHROPIC_API_KEY=your-api-key
```

Alternative providers supported:
- **Amazon Bedrock**: `CLAUDE_CODE_USE_BEDROCK=1` + AWS credentials
- **Google Vertex AI**: `CLAUDE_CODE_USE_VERTEX=1` + GCP credentials
- **Microsoft Azure**: `CLAUDE_CODE_USE_FOUNDRY=1` + Azure credentials

**Important policy note from Anthropic:**
> Unless previously approved, Anthropic does not allow third party developers to offer claude.ai login or rate limits for their products, including agents built on the Claude Agent SDK. Please use the API key authentication methods described in this document instead.

This means alto **cannot** piggyback on the user's Claude Code subscription. Users must have their own API key (or use Bedrock/Vertex).

Source: https://platform.claude.com/docs/en/agent-sdk/overview

### Q5: How to use as a port adapter in hexagonal architecture?

Here is a concrete design for implementing a `ChallengerPort` using the Claude Agent SDK:

```python
# src/application/ports/challenger.py
from __future__ import annotations

from typing import Protocol

from src.domain.models.challenge import ChallengeResult


class ChallengerPort(Protocol):
    """Port for challenging domain models with probing questions."""

    async def challenge(
        self, model_description: str, context: str
    ) -> ChallengeResult:
        """Challenge a domain model and return structured feedback."""
        ...
```

```python
# src/infrastructure/external/claude_challenger.py
from __future__ import annotations

from claude_agent_sdk import query, ClaudeAgentOptions, ResultMessage

from src.application.ports.challenger import ChallengerPort
from src.domain.models.challenge import ChallengeResult


class ClaudeChallengerAdapter:
    """Infrastructure adapter implementing ChallengerPort via Claude Agent SDK."""

    def __init__(
        self,
        system_prompt: str | None = None,
        max_turns: int = 1,
        model: str | None = None,
    ) -> None:
        self._system_prompt = system_prompt or self._default_system_prompt()
        self._max_turns = max_turns
        self._model = model

    async def challenge(
        self, model_description: str, context: str
    ) -> ChallengeResult:
        options = ClaudeAgentOptions(
            system_prompt=self._system_prompt,
            max_turns=self._max_turns,
            model=self._model,
            output_format={
                "type": "json_schema",
                "schema": ChallengeResult.model_json_schema(),
            },
            # No tools needed -- pure text analysis
            allowed_tools=[],
        )

        prompt = (
            f"Context:\n{context}\n\n"
            f"Domain Model:\n{model_description}\n\n"
            "Challenge this model. Identify weaknesses, missing concepts, "
            "and incorrect boundaries."
        )

        async for message in query(prompt=prompt, options=options):
            if isinstance(message, ResultMessage):
                if message.subtype == "success" and message.structured_output:
                    return ChallengeResult.model_validate(
                        message.structured_output
                    )
                raise RuntimeError(
                    f"Claude challenge failed: {message.subtype}"
                )

        raise RuntimeError("No result message received from Claude")

    @staticmethod
    def _default_system_prompt() -> str:
        return (
            "You are a senior DDD practitioner. Your job is to challenge "
            "domain models by asking probing questions, identifying missing "
            "aggregates, incorrect boundaries, and anemic models."
        )
```

```python
# For testing -- a fake adapter that doesn't call Claude
class FakeChallengerAdapter:
    """Test double for ChallengerPort."""

    def __init__(self, result: ChallengeResult) -> None:
        self._result = result

    async def challenge(
        self, model_description: str, context: str
    ) -> ChallengeResult:
        return self._result
```

This pattern follows DDD + hexagonal architecture:
- **Port** (`ChallengerPort`): lives in `application/ports/`, no external deps
- **Adapter** (`ClaudeChallengerAdapter`): lives in `infrastructure/external/`, depends on `claude_agent_sdk`
- **Domain model** (`ChallengeResult`): lives in `domain/models/`, is a Pydantic BaseModel
- **Test double** (`FakeChallengerAdapter`): no external deps, used in unit tests

### Q6: Non-interactive (programmatic) LLM calls?

**Yes, fully supported.** The `query()` function is designed for non-interactive use:

```python
# Non-interactive, single-turn, no tools, no user interaction
async for message in query(
    prompt="Summarize this text: ...",
    options=ClaudeAgentOptions(
        max_turns=1,
        allowed_tools=[],  # No file/shell access
        permission_mode="bypassPermissions",  # No interactive prompts
    ),
):
    if isinstance(message, ResultMessage):
        return message.result
```

Key options for non-interactive use:
- `max_turns=1` -- prevents multi-turn loops
- `allowed_tools=[]` -- disables all tool use (pure text generation)
- `permission_mode="bypassPermissions"` -- never prompts for user input
- `output_format={...}` -- forces structured JSON output

---

## 4. Comparison: Claude Agent SDK vs Anthropic SDK

| Dimension                     | `claude-agent-sdk`                      | `anthropic` SDK                         |
| ----------------------------- | --------------------------------------- | --------------------------------------- |
| **Package**                   | `pip install claude-agent-sdk`          | `pip install anthropic`                 |
| **PyPI version**              | 0.1.46 (2026-03-05)                    | 0.49.x+ (2026-03)                      |
| **License**                   | MIT                                     | MIT                                     |
| **Architecture**              | Subprocess (spawns CLI)                 | Direct HTTPS to API                     |
| **Startup overhead**          | ~1-3s (CLI process spawn)              | ~0s (HTTP connection)                   |
| **Built-in tools**            | Read, Write, Edit, Bash, Glob, Grep... | None (you implement tools)             |
| **Agent loop**                | Built-in (multi-turn, tool execution)  | You implement it yourself              |
| **Structured output**         | `output_format` + Pydantic             | `response_format` + JSON mode          |
| **Streaming**                 | AsyncIterator over messages            | SSE streaming, async                    |
| **Subagents**                 | Built-in via `AgentDefinition`         | You implement it yourself              |
| **MCP servers**               | Built-in support                       | Not applicable                          |
| **Session management**        | Built-in (resume, fork)                | You manage conversation history         |
| **Cost tracking**             | `ResultMessage.total_cost_usd`         | Token counts in response                |
| **Best for**                  | Agent workflows, code tasks            | Simple LLM calls, direct API access     |

### Recommendation for alto

**For simple LLM calls** (e.g., "challenge this domain model", "generate DDD questions"):
Use the `anthropic` SDK directly. It is lighter, faster (no subprocess), and gives you more control.

**For agent workflows** (e.g., "scan this codebase and generate a gap analysis"):
Use the Claude Agent SDK. The built-in tools (file reading, grep, bash) and agent loop save significant implementation effort.

**alto likely needs both:**
- `anthropic` SDK for `ChallengerPort`, `QuestionGeneratorPort`, `AnalysisPort` (simple text/structured output)
- `claude-agent-sdk` if alto wants to offer an "AI-assisted" mode that reads the user's codebase and generates artifacts

---

## 5. Resource & Cost Considerations

### Subprocess Overhead

Each `query()` call spawns a Claude Code CLI process (Node.js):
- **Memory**: ~100-200MB per subprocess (Node.js runtime)
- **Startup latency**: ~1-3 seconds for process spawn + initialization
- **For frequent calls**: Use `ClaudeSDKClient` to maintain a session and avoid repeated startup

### API Costs

The SDK uses the same Anthropic API pricing. The `ResultMessage` includes:
- `total_cost_usd` -- total cost of the session
- `usage.input_tokens` / `usage.output_tokens` -- token consumption
- `max_budget_usd` -- can set a budget cap per session

### Compared to Direct API

Using the `anthropic` SDK directly avoids the subprocess overhead entirely. For alto's use case where we might make 5-10 LLM calls during a bootstrap session, the subprocess overhead adds ~10-30 seconds total vs near-instant with direct API calls.

---

## 6. Risk Assessment

| Risk                                        | Severity | Mitigation                                         |
| ------------------------------------------- | -------- | -------------------------------------------------- |
| Alpha status (0.1.x)                        | Medium   | Very active releases; Anthropic-maintained          |
| Breaking API changes                        | Medium   | Pin version; port adapter isolates changes          |
| Subprocess overhead for simple calls        | Medium   | Use `anthropic` SDK for simple prompts              |
| Requires API key (cannot use CLI auth)      | Low      | Users already need API key for any programmatic use |
| Node.js bundled binary size                 | Low      | ~100MB disk; one-time install cost                  |
| No sync API (async only)                    | Low      | alto already uses async patterns                    |

---

## 7. Recommendation

### For alto's LLM integration needs, use a dual-SDK approach:

1. **`anthropic` SDK** (direct API) for all Port implementations that need simple LLM calls:
   - `ChallengerPort` -- challenge domain models
   - `QuestionGeneratorPort` -- generate DDD discovery questions
   - `AnalysisPort` -- analyze code/text and return structured results
   - Lower latency, no subprocess overhead, lighter dependency

2. **`claude-agent-sdk`** (Agent SDK) for advanced agent workflows if/when needed:
   - Codebase scanning in rescue mode (`alto init --existing`)
   - AI-assisted gap analysis that needs to read files and run commands
   - Multi-step workflows where Claude needs tool access

3. **Port adapter pattern** isolates both from domain logic:
   - `application/ports/` defines Protocols with no external deps
   - `infrastructure/external/` implements adapters for either SDK
   - Tests use fake adapters -- zero LLM dependency in test suite

### Key Finding

The Claude Agent SDK is an **agent framework** (subprocess-based, with built-in tools and agent loop), not a simple LLM API client. For alto's port adapters that just need structured text generation, the direct `anthropic` SDK is the better fit. The Agent SDK is the right choice only when you need Claude to autonomously use tools (read files, run commands, etc.).

### Next Steps

- Create follow-up ticket for `ChallengerPort` implementation using `anthropic` SDK
- Create follow-up ticket for evaluating whether rescue mode (`--existing`) should use `claude-agent-sdk` for codebase scanning
- Add `anthropic` as an optional dependency in `pyproject.toml` (users who don't need AI features can skip it)
- Define the `LLMPort` Protocol family in `application/ports/`

---

## Sources

- [Claude Agent SDK - PyPI](https://pypi.org/project/claude-agent-sdk/) -- Version 0.1.46, release date 2026-03-05
- [Claude Code SDK (deprecated) - PyPI](https://pypi.org/project/claude-code-sdk/) -- Deprecation notice
- [Claude Agent SDK - GitHub](https://github.com/anthropics/claude-agent-sdk-python) -- README, examples, license
- [Agent SDK Overview - Anthropic Docs](https://platform.claude.com/docs/en/agent-sdk/overview) -- Authentication, capabilities, comparison
- [Agent SDK Python Reference - Anthropic Docs](https://platform.claude.com/docs/en/agent-sdk/python) -- Full API reference
- [Structured Outputs - Anthropic Docs](https://platform.claude.com/docs/en/agent-sdk/structured-outputs) -- Pydantic integration, JSON schema
- [Anthropic Python SDK - GitHub](https://github.com/anthropics/anthropic-sdk-python) -- Direct API client
