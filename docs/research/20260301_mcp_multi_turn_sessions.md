---
last_reviewed: 2026-03-01
owner: architecture
status: complete
type: spike
ticket: alty-k7m.29
---

# MCP Multi-Turn Sessions: Guided DDD Flow over MCP

> **Spike:** k7m.29 — Guided DDD flow over MCP (multi-turn sessions)
> **Timebox:** 4 hours
> **Decision:** Server-side session store with TTL, multiple focused MCP tools

## 1. Problem Statement

The guided DDD discovery flow is a 10-question, multi-turn, stateful conversation:

- 6 states: CREATED -> PERSONA_DETECTED -> ANSWERING <-> PLAYBACK_PENDING -> COMPLETED
- 5 question phases (SEED, ACTORS, STORY, EVENTS, BOUNDARIES) with strict ordering
- Playback confirmation every 3 answers
- Persona detection determines question register (technical vs plain language)
- 5 invariants enforced by the DiscoverySession aggregate

MCP tools are request/response (stateless). Each tool call is independent — there is
no built-in session concept in the MCP protocol. How should alty bridge this gap?

## 2. Current Architecture

### DiscoverySession Aggregate (`src/domain/models/discovery_session.py`)

State machine with 6 states, 10 questions, playback loops. Key methods:

- `detect_persona(choice)` — CREATED -> PERSONA_DETECTED
- `answer_question(question_id, response)` — records answer, auto-triggers playback every 3
- `confirm_playback(confirmed)` — PLAYBACK_PENDING -> ANSWERING
- `complete()` — validates MVP questions, emits DiscoveryCompleted event

### DiscoveryHandler (`src/application/commands/discovery_handler.py`)

In-memory session store: `dict[str, DiscoverySession]`. Methods mirror the aggregate.
Already supports multi-session concurrency via session_id lookup.

### MCP Server (`src/infrastructure/mcp/server.py`)

Single `guide_ddd(readme_content)` tool stub. Calls `app.discovery.start_session()`.
AppContext shared via FastMCP lifespan pattern.

### DiscoveryPort (`src/application/ports/discovery_port.py`)

5 methods returning `str`: `start_session`, `detect_persona`, `answer_question`,
`confirm_playback`, `complete`. All use `session_id` for lookup.

## 3. Options Investigated

### Option 1: Server-Side Session Store (In-Memory Dict with TTL)

AppContext holds a session store backed by `dict[str, DiscoverySession]`. Each MCP
tool call passes `session_id`. Server loads session, processes action, stores update.

```python
@dataclass
class SessionStore:
    """In-memory session store with TTL-based cleanup."""
    _sessions: dict[str, tuple[DiscoverySession, float]] = field(default_factory=dict)
    ttl_seconds: float = 1800.0  # 30 minutes

    def get(self, session_id: str) -> DiscoverySession:
        entry = self._sessions.get(session_id)
        if entry is None:
            raise SessionNotFoundError(session_id)
        session, created_at = entry
        if time.monotonic() - created_at > self.ttl_seconds:
            del self._sessions[session_id]
            raise SessionNotFoundError(session_id)
        return session

    def put(self, session: DiscoverySession) -> None:
        self._sessions[session.session_id] = (session, time.monotonic())

    def cleanup_expired(self) -> int:
        now = time.monotonic()
        expired = [k for k, (_, t) in self._sessions.items() if now - t > self.ttl_seconds]
        for k in expired:
            del self._sessions[k]
        return len(expired)
```

| Criterion | Rating | Notes |
|-----------|--------|-------|
| MCP compatibility | High | Works with any MCP client — only session_id in params |
| State reliability | High | Server manages state, no client-side requirements |
| Implementation simplicity | High | DiscoveryHandler already uses this exact pattern |
| Payload size | High | Only session_id (36 chars) passes over wire |
| Concurrency | Medium | Dict is not thread-safe; use asyncio lock if needed |

**Pros:**
- Clean, focused tool APIs with small payloads
- DiscoveryHandler already proves this pattern works
- Works with Claude Code, Cursor, any MCP client without modification
- Session state is encapsulated — clients never see internals

**Cons:**
- State lost on server restart
- Memory grows with concurrent sessions (mitigated by TTL)
- No crash recovery

### Option 2: Client-Side Context Passing

Entire DiscoverySession state serialized to JSON. Returned in tool response.
Client passes full state back in next tool call.

```python
@mcp.tool()
async def guide_answer(
    session_state: str,  # JSON-encoded full session state
    question_id: str,
    answer: str,
) -> str:
    session = DiscoverySession.from_json(session_state)
    session.answer_question(question_id, answer)
    return json.dumps({"state": session.to_json(), "next": ...})
```

| Criterion | Rating | Notes |
|-----------|--------|-------|
| MCP compatibility | Medium | Works but clients must store and replay state |
| State reliability | Low | Client could lose/corrupt state |
| Implementation simplicity | Low | Requires serialization of entire aggregate |
| Payload size | Low | 5-10KB per call (10 answers + playbacks + metadata) |
| Concurrency | High | Truly stateless, no server-side contention |

**Pros:**
- Server is purely stateless — no session management
- No TTL or cleanup needed
- Survives server restarts (state is on client)

**Cons:**
- Breaks aggregate encapsulation — must expose internal state as JSON
- Large payloads (5-10KB per tool call with full session state)
- Client must faithfully store and replay state — error-prone
- Serialization/deserialization adds complexity to domain model
- MCP clients (Claude Code, Cursor) don't expect to manage opaque state blobs

### Option 3: MCP Prompts / Sampling

**Prompts feature:** Defines reusable prompt templates with variables. Clients fill
variables and use them to guide LLM inference. Templates are static — no state
management, no multi-turn session concept. **Not applicable.**

**Sampling feature:** Allows servers to request LLM completions from clients via
`sampling/createMessage`. Supports multi-turn message arrays. However:
- Designed for AI reasoning, not application state management
- "Not yet supported in the Claude Desktop client" (per MCP docs)
- Would require alty's server to request LLM completions — wrong abstraction
  (alty needs to manage question flow, not generate AI completions)

| Criterion | Rating | Notes |
|-----------|--------|-------|
| MCP compatibility | Low | Sampling not supported in Claude Desktop; prompts are templates only |
| State reliability | N/A | Neither feature manages application state |
| Implementation simplicity | N/A | Would require misusing the feature |
| Payload size | N/A | Not applicable |
| Concurrency | N/A | Not applicable |

**Verdict:** Neither MCP prompts nor sampling solve the multi-turn session problem.
They serve different purposes (LLM guidance and AI reasoning delegation).

### Option 4: Hybrid — Server-Side Session + Serializable State

Combines Option 1 (server-side store) with optional serialization for recovery.

- Primary: In-memory session store with TTL (exactly like Option 1)
- Optional: `to_snapshot()` / `from_snapshot()` on DiscoverySession for persistence
- SessionStore lifecycle managed by AppContext lifespan
- Cleanup runs on a background asyncio task

```python
@dataclass
class AppContext:
    bootstrap: BootstrapPort
    discovery: DiscoveryPort
    session_store: SessionStore  # NEW: shared session store
    # ... other ports
```

| Criterion | Rating | Notes |
|-----------|--------|-------|
| MCP compatibility | High | Same as Option 1 — session_id only |
| State reliability | High | In-memory + optional persistence for crash recovery |
| Implementation simplicity | Medium | Option 1 core + optional serialization |
| Payload size | High | session_id only |
| Concurrency | Medium | Same as Option 1 |

**Pros:**
- All benefits of Option 1
- Serialization enables future persistence (filesystem, SQLite)
- Crash recovery possible without changing MCP API
- Incremental: start with in-memory only, add persistence later

**Cons:**
- Serialization methods on aggregate add complexity (but optional)
- Slightly more infrastructure than Option 1

## 4. Recommendation: Option 4 (Hybrid)

**Use server-side session store with optional serialization.**

Rationale:

1. **DiscoveryHandler already uses this pattern** — `dict[str, DiscoverySession]` with
   session_id lookup. Option 4 extracts this into a reusable SessionStore.

2. **AppContext lifespan fits perfectly** — SessionStore initializes at server startup,
   shared across all tool calls, cleaned up on shutdown.

3. **Minimal MCP API** — Clients only pass `session_id` (36 chars). No state blobs,
   no client-side requirements, works with any MCP client.

4. **DiscoverySession aggregate needs NO changes** for the basic implementation.
   Serialization is optional future work.

5. **MCP prompts/sampling don't solve this** — They serve different purposes.
   Server-side state management is the correct approach.

## 5. MCP Tool API Design

Replace the single `guide_ddd` stub with 6 focused tools:

```python
@mcp.tool()
async def guide_start(readme_content: str, ctx: McpContext) -> str:
    """Start a new guided DDD discovery session.

    Returns the session_id to use in subsequent guide_* calls.
    """
    app = _get_app(ctx)
    return app.discovery.start_session(readme_content)


@mcp.tool()
async def guide_detect_persona(session_id: str, choice: str, ctx: McpContext) -> str:
    """Detect user persona for the discovery session.

    Args:
        session_id: The session ID from guide_start.
        choice: "1" (Developer), "2" (Product Owner),
                "3" (Domain Expert), "4" (Mixed).

    Returns confirmation of detected persona and register.
    """
    app = _get_app(ctx)
    return app.discovery.detect_persona(session_id, choice)


@mcp.tool()
async def guide_answer(
    session_id: str,
    question_id: str,
    answer: str,
    ctx: McpContext,
) -> str:
    """Answer a question in the guided DDD discovery flow.

    Args:
        session_id: The session ID from guide_start.
        question_id: The question ID (Q1-Q10).
        answer: The user's answer.

    Returns the next question, or a playback summary if 3 answers
    have been given since the last playback.
    """
    app = _get_app(ctx)
    return app.discovery.answer_question(session_id, question_id, answer)


@mcp.tool()
async def guide_confirm_playback(
    session_id: str,
    confirmed: bool,
    ctx: McpContext,
) -> str:
    """Confirm or reject a playback summary.

    Args:
        session_id: The session ID from guide_start.
        confirmed: True to confirm, False to request corrections.

    Returns the next question or a correction prompt.
    """
    app = _get_app(ctx)
    return app.discovery.confirm_playback(session_id, confirmed)


@mcp.tool()
async def guide_complete(session_id: str, ctx: McpContext) -> str:
    """Complete the discovery session and generate domain artifacts.

    Requires all MVP questions (Q1, Q3, Q4, Q9, Q10) to be answered.
    Returns a summary of generated artifacts.
    """
    app = _get_app(ctx)
    return app.discovery.complete(session_id)


@mcp.tool()
async def guide_status(session_id: str, ctx: McpContext) -> str:
    """Get the current state of a discovery session.

    Returns: current phase, answered questions, pending questions,
    session status, and persona/register.
    """
    app = _get_app(ctx)
    # DiscoveryAdapter lookup + format status
    session = app.discovery.get_session(session_id)
    return json.dumps({
        "session_id": session.session_id,
        "status": session.status.value,
        "phase": session.current_phase.value,
        "answered": [a.question_id for a in session.answers],
        "persona": session.persona.value if session.persona else None,
    })
```

**Why multiple tools instead of one:**

- Each tool has a clear, focused schema with specific parameters
- MCP clients show tool descriptions — separate tools are more discoverable
- No dynamic parameter validation needed
- Follows Single Responsibility Principle
- AI coding tools (Claude Code, Cursor) can reason about specific tools better

## 6. DiscoverySession Impact Assessment

### No Changes Needed (Basic Implementation)

The DiscoverySession aggregate requires **zero modifications** for Option 4:

- SessionStore wraps it externally — no internal state exposure
- DiscoveryHandler already manages `dict[str, DiscoverySession]`
- SessionStore extracts this pattern into a reusable component

### Optional Future Changes (Persistence / Crash Recovery)

If crash recovery is desired later:

```python
# Add to DiscoverySession (optional, future work)
def to_snapshot(self) -> dict[str, Any]:
    """Serialize session state for persistence."""
    return {
        "session_id": self.session_id,
        "readme_content": self.readme_content,
        "status": self._status.value,
        "persona": self._persona.value if self._persona else None,
        "register": self._register.value if self._register else None,
        "answers": [(a.question_id, a.response_text) for a in self._answers],
        "skipped": list(self._skipped),
        "playbacks": [(p.summary_text, p.confirmed, p.corrections)
                      for p in self._playback_confirmations],
    }

@classmethod
def from_snapshot(cls, data: dict[str, Any]) -> DiscoverySession:
    """Restore session from serialized state."""
    # Reconstruct without re-running invariant checks
    ...
```

This is **not needed for initial implementation** — just a path for future evolution.

## 7. Implementation Plan

### SessionStore Component

```python
# src/infrastructure/session/session_store.py

from __future__ import annotations

import time
from dataclasses import dataclass, field
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from src.domain.models.discovery_session import DiscoverySession

from src.domain.models.errors import SessionNotFoundError


@dataclass
class SessionStore:
    """In-memory session store with TTL-based cleanup.

    Stores DiscoverySession instances keyed by session_id. Expired
    sessions are cleaned up on access or via explicit cleanup.
    """

    ttl_seconds: float = 1800.0  # 30 minutes
    _sessions: dict[str, tuple[DiscoverySession, float]] = field(
        default_factory=dict
    )

    def get(self, session_id: str) -> DiscoverySession:
        """Retrieve a session by ID. Raises SessionNotFoundError if missing or expired."""
        entry = self._sessions.get(session_id)
        if entry is None:
            raise SessionNotFoundError(session_id)
        session, created_at = entry
        if time.monotonic() - created_at > self.ttl_seconds:
            del self._sessions[session_id]
            raise SessionNotFoundError(session_id)
        return session

    def put(self, session: DiscoverySession) -> None:
        """Store or update a session."""
        self._sessions[session.session_id] = (session, time.monotonic())

    def cleanup_expired(self) -> int:
        """Remove expired sessions. Returns count of removed sessions."""
        now = time.monotonic()
        expired = [
            k for k, (_, t) in self._sessions.items()
            if now - t > self.ttl_seconds
        ]
        for k in expired:
            del self._sessions[k]
        return len(expired)

    @property
    def active_count(self) -> int:
        """Number of sessions currently stored (including potentially expired)."""
        return len(self._sessions)
```

### AppContext Update

```python
@dataclass
class AppContext:
    bootstrap: BootstrapPort
    discovery: DiscoveryPort  # DiscoveryAdapter wraps SessionStore internally
    tool_detection: ToolDetectionPort
    # ... rest unchanged
```

Note: SessionStore is an internal implementation detail of `DiscoveryAdapter`, not
exposed on `AppContext`. MCP tools access sessions via `app.discovery` (the port).

### MCP Server Update

Replace single `guide_ddd` tool with 6 tools per Section 5.

## 8. Follow-Up Implementation Tickets

### Ticket 1: Implement SessionStore and MCP Discovery Tools

**Type:** Task
**Priority:** P2
**Bounded Context:** MCP Server Framework (Generic) + Knowledge Base (Supporting)

**Steps:**
1. Create `src/infrastructure/session/session_store.py` with SessionStore
2. Update `src/infrastructure/composition.py` — add SessionStore to AppContext
3. Create `src/infrastructure/mcp/discovery_adapter.py` — implements DiscoveryPort, wraps DiscoveryHandler + SessionStore
4. Update `src/infrastructure/mcp/server.py` — replace `guide_ddd` with 6 focused tools
5. Write tests for SessionStore (TTL, get, put, cleanup, expired)
6. Write tests for MCP discovery tools (start, persona, answer, playback, complete, status)

**Depends on:** k7m.27 (MCP server), k7m.29 (this spike)

### Ticket 2: (Optional) Add Serialization to DiscoverySession

**Type:** Task
**Priority:** P3
**Bounded Context:** Guided Discovery (Core)

**Steps:**
1. Add `to_snapshot()` method to DiscoverySession
2. Add `from_snapshot()` classmethod to DiscoverySession
3. Update SessionStore to optionally persist to filesystem
4. Write round-trip serialization tests

**Depends on:** Ticket 1

## 9. References

- MCP Python SDK: https://github.com/modelcontextprotocol/python-sdk
- FastMCP lifespan docs: Context7 `/modelcontextprotocol/python-sdk`
- MCP Specification (2025-11-25): https://modelcontextprotocol.io/specification/2025-11-25
- MCP Sampling: https://modelcontextprotocol.info/docs/concepts/sampling/
- MCP Prompts: https://modelcontextprotocol.info/docs/concepts/prompts/
- MCP Features Guide: https://workos.com/blog/mcp-features-guide
- FastMCP GitHub: https://github.com/jlowin/fastmcp
- alty MCP server: `src/infrastructure/mcp/server.py`
- DiscoverySession: `src/domain/models/discovery_session.py`
- DiscoveryHandler: `src/application/commands/discovery_handler.py`
- CLI/MCP design spike (k7m.4): `docs/research/20260222_cli_mcp_design.md`
