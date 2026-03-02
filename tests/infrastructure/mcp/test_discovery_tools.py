"""Tests for MCP discovery tools -- 6 focused tools replacing guide_ddd.

RED phase: these tests define the contract for guide_start, guide_detect_persona,
guide_answer, guide_skip_question, guide_confirm_playback, guide_complete, and
guide_status.

The tools operate against a shared SessionStore that persists DiscoverySessions
across MCP request/response cycles.
"""

from __future__ import annotations

import asyncio
from typing import Any
from unittest.mock import MagicMock

import pytest

from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.discovery_values import Persona, Register
from src.infrastructure.session.session_store import SessionStore

# ── Test helpers ─────────────────────────────────────────────────────


def _make_ctx_with_store(store: SessionStore | None = None) -> Any:
    """Create a minimal MCP context with a real DiscoveryHandler + SessionStore.

    Returns a mock context whose .request_context.lifespan_context is an
    AppContext with a real DiscoveryAdapter wired to the given store.
    """
    from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

    if store is None:
        store = SessionStore(ttl_seconds=60)
    adapter = DiscoveryAdapter(store=store)

    # Build a real AppContext with the adapter as the discovery port
    from src.infrastructure.composition import create_app

    app = create_app()
    # Replace the stub discovery with our adapter
    object.__setattr__(app, "discovery", adapter)

    ctx = MagicMock()
    ctx.request_context.lifespan_context = app
    return ctx, store


# ── guide_start ──────────────────────────────────────────────────────


class TestGuideStart:
    def test_returns_session_id(self) -> None:
        from src.infrastructure.mcp.server import guide_start

        ctx, store = _make_ctx_with_store()
        result = asyncio.run(guide_start("A project idea in 4-5 sentences.", ctx))
        assert "session_id" in result or "session" in result.lower()
        assert store.active_count() == 1

    def test_session_is_retrievable(self) -> None:
        from src.infrastructure.mcp.server import guide_start

        ctx, store = _make_ctx_with_store()
        result = asyncio.run(guide_start("My idea", ctx))
        # Extract session_id from response
        session_id = _extract_session_id(result)
        session = store.get(session_id)
        assert isinstance(session, DiscoverySession)
        assert session.status == DiscoveryStatus.CREATED

    def test_multiple_sessions_are_independent(self) -> None:
        from src.infrastructure.mcp.server import guide_start

        ctx, store = _make_ctx_with_store()
        r1 = asyncio.run(guide_start("Idea A", ctx))
        r2 = asyncio.run(guide_start("Idea B", ctx))
        id1 = _extract_session_id(r1)
        id2 = _extract_session_id(r2)
        assert id1 != id2
        assert store.active_count() == 2


# ── guide_detect_persona ────────────────────────────────────────────


class TestGuideDetectPersona:
    def test_sets_persona_and_register(self) -> None:
        from src.infrastructure.mcp.server import guide_detect_persona, guide_start

        ctx, store = _make_ctx_with_store()
        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)

        result = asyncio.run(guide_detect_persona(session_id, "1", ctx))
        session = store.get(session_id)
        assert session.persona == Persona.DEVELOPER
        assert session.register == Register.TECHNICAL
        assert session.status == DiscoveryStatus.PERSONA_DETECTED
        assert "persona" in result.lower() or "developer" in result.lower()

    def test_invalid_session_id_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_detect_persona

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_detect_persona("nonexistent-id", "1", ctx))
        assert "error" in result.lower() or "not found" in result.lower()

    def test_invalid_choice_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_detect_persona, guide_start

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)

        result = asyncio.run(guide_detect_persona(session_id, "99", ctx))
        assert "error" in result.lower() or "invalid" in result.lower()


# ── guide_answer ─────────────────────────────────────────────────────


class TestGuideAnswer:
    def _setup_answerable_session(self, ctx: Any) -> str:
        """Create a session ready to answer questions (persona detected)."""
        from src.infrastructure.mcp.server import guide_detect_persona, guide_start

        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)
        asyncio.run(guide_detect_persona(session_id, "1", ctx))
        return session_id

    def test_answer_records_response(self) -> None:
        from src.infrastructure.mcp.server import guide_answer

        ctx, store = _make_ctx_with_store()
        session_id = self._setup_answerable_session(ctx)

        result = asyncio.run(guide_answer(session_id, "Q1", "Users and admins", ctx))
        session = store.get(session_id)
        assert len(session.answers) == 1
        assert session.answers[0].question_id == "Q1"
        assert "Q1" in result or "answer" in result.lower()

    def test_triggers_playback_after_3_answers(self) -> None:
        from src.infrastructure.mcp.server import guide_answer

        ctx, store = _make_ctx_with_store()
        session_id = self._setup_answerable_session(ctx)

        asyncio.run(guide_answer(session_id, "Q1", "Users", ctx))
        asyncio.run(guide_answer(session_id, "Q2", "Entities", ctx))
        result = asyncio.run(guide_answer(session_id, "Q3", "Use case", ctx))

        session = store.get(session_id)
        assert session.status == DiscoveryStatus.PLAYBACK_PENDING
        assert "playback" in result.lower()

    def test_invalid_session_id_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_answer

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_answer("bad-id", "Q1", "Answer", ctx))
        assert "error" in result.lower() or "not found" in result.lower()

    def test_answer_before_persona_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_answer, guide_start

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)

        result = asyncio.run(guide_answer(session_id, "Q1", "Answer", ctx))
        assert "error" in result.lower()

    def test_empty_answer_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_answer

        ctx, _ = _make_ctx_with_store()
        session_id = self._setup_answerable_session(ctx)

        result = asyncio.run(guide_answer(session_id, "Q1", "  ", ctx))
        assert "error" in result.lower()

    def test_duplicate_answer_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_answer

        ctx, _ = _make_ctx_with_store()
        session_id = self._setup_answerable_session(ctx)

        asyncio.run(guide_answer(session_id, "Q1", "Users", ctx))
        result = asyncio.run(guide_answer(session_id, "Q1", "Users again", ctx))
        assert "error" in result.lower() or "already" in result.lower()


# ── guide_skip_question ──────────────────────────────────────────────


class TestGuideSkipQuestion:
    def _setup_answerable_session(self, ctx: Any) -> str:
        from src.infrastructure.mcp.server import guide_detect_persona, guide_start

        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)
        asyncio.run(guide_detect_persona(session_id, "1", ctx))
        return session_id

    def test_skip_records_skip(self) -> None:
        from src.infrastructure.mcp.server import guide_skip_question

        ctx, _store = _make_ctx_with_store()
        session_id = self._setup_answerable_session(ctx)

        result = asyncio.run(guide_skip_question(session_id, "Q1", "Not relevant", ctx))
        assert "skip" in result.lower() or "Q1" in result

    def test_invalid_session_id_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_skip_question

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_skip_question("bad-id", "Q1", "Reason", ctx))
        assert "error" in result.lower() or "not found" in result.lower()

    def test_empty_reason_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_skip_question

        ctx, _ = _make_ctx_with_store()
        session_id = self._setup_answerable_session(ctx)

        result = asyncio.run(guide_skip_question(session_id, "Q1", "", ctx))
        assert "error" in result.lower()


# ── guide_confirm_playback ───────────────────────────────────────────


class TestGuideConfirmPlayback:
    def _setup_playback_pending(self, ctx: Any, store: SessionStore) -> str:
        """Create a session in PLAYBACK_PENDING state."""
        from src.infrastructure.mcp.server import (
            guide_answer,
            guide_detect_persona,
            guide_start,
        )

        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)
        asyncio.run(guide_detect_persona(session_id, "1", ctx))
        asyncio.run(guide_answer(session_id, "Q1", "Users", ctx))
        asyncio.run(guide_answer(session_id, "Q2", "Entities", ctx))
        asyncio.run(guide_answer(session_id, "Q3", "Use case", ctx))

        session = store.get(session_id)
        assert isinstance(session, DiscoverySession)
        assert session.status == DiscoveryStatus.PLAYBACK_PENDING
        return session_id

    def test_confirm_resumes_answering(self) -> None:
        from src.infrastructure.mcp.server import guide_confirm_playback

        ctx, store = _make_ctx_with_store()
        session_id = self._setup_playback_pending(ctx, store)

        result = asyncio.run(guide_confirm_playback(session_id, True, ctx))
        session = store.get(session_id)
        assert session.status == DiscoveryStatus.ANSWERING
        assert "confirmed" in result.lower() or "playback" in result.lower()

    def test_invalid_session_id_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_confirm_playback

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_confirm_playback("bad-id", True, ctx))
        assert "error" in result.lower() or "not found" in result.lower()

    def test_confirm_when_not_pending_returns_error(self) -> None:
        from src.infrastructure.mcp.server import (
            guide_confirm_playback,
            guide_detect_persona,
            guide_start,
        )

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)
        asyncio.run(guide_detect_persona(session_id, "1", ctx))
        # Session is in PERSONA_DETECTED, not PLAYBACK_PENDING
        result = asyncio.run(guide_confirm_playback(session_id, True, ctx))
        assert "error" in result.lower()


# ── guide_complete ───────────────────────────────────────────────────


class TestGuideComplete:
    def _setup_completable_session(self, ctx: Any, store: SessionStore) -> str:
        """Create a session with all 10 questions answered, ready to complete."""
        from src.infrastructure.mcp.server import (
            guide_answer,
            guide_confirm_playback,
            guide_detect_persona,
            guide_start,
        )

        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)
        asyncio.run(guide_detect_persona(session_id, "1", ctx))

        # Answer Q1-Q3, confirm playback
        for qid in ["Q1", "Q2", "Q3"]:
            asyncio.run(guide_answer(session_id, qid, f"Answer {qid}", ctx))
        asyncio.run(guide_confirm_playback(session_id, True, ctx))

        # Answer Q4-Q6, confirm playback
        for qid in ["Q4", "Q5", "Q6"]:
            asyncio.run(guide_answer(session_id, qid, f"Answer {qid}", ctx))
        asyncio.run(guide_confirm_playback(session_id, True, ctx))

        # Answer Q7-Q9, confirm playback
        for qid in ["Q7", "Q8", "Q9"]:
            asyncio.run(guide_answer(session_id, qid, f"Answer {qid}", ctx))
        asyncio.run(guide_confirm_playback(session_id, True, ctx))

        # Answer Q10
        asyncio.run(guide_answer(session_id, "Q10", "Answer Q10", ctx))

        return session_id

    def test_complete_marks_completed(self) -> None:
        from src.infrastructure.mcp.server import guide_complete

        ctx, store = _make_ctx_with_store()
        session_id = self._setup_completable_session(ctx, store)

        result = asyncio.run(guide_complete(session_id, ctx))
        session = store.get(session_id)
        assert session.status == DiscoveryStatus.COMPLETED
        assert "complete" in result.lower()

    def test_complete_emits_event(self) -> None:
        from src.infrastructure.mcp.server import guide_complete

        ctx, store = _make_ctx_with_store()
        session_id = self._setup_completable_session(ctx, store)

        asyncio.run(guide_complete(session_id, ctx))
        session = store.get(session_id)
        assert len(session.events) == 1

    def test_invalid_session_id_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_complete

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_complete("bad-id", ctx))
        assert "error" in result.lower() or "not found" in result.lower()

    def test_complete_without_mvp_returns_error(self) -> None:
        from src.infrastructure.mcp.server import (
            guide_answer,
            guide_complete,
            guide_detect_persona,
            guide_start,
        )

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)
        asyncio.run(guide_detect_persona(session_id, "1", ctx))
        asyncio.run(guide_answer(session_id, "Q1", "Answer", ctx))
        # Only 1 question answered, not enough for MVP
        result = asyncio.run(guide_complete(session_id, ctx))
        assert "error" in result.lower()


# ── guide_status ─────────────────────────────────────────────────────


class TestGuideStatus:
    def test_status_shows_created_state(self) -> None:
        from src.infrastructure.mcp.server import guide_start, guide_status

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)

        result = asyncio.run(guide_status(session_id, ctx))
        assert "created" in result.lower()

    def test_status_shows_persona_after_detection(self) -> None:
        from src.infrastructure.mcp.server import (
            guide_detect_persona,
            guide_start,
            guide_status,
        )

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)
        asyncio.run(guide_detect_persona(session_id, "1", ctx))

        result = asyncio.run(guide_status(session_id, ctx))
        assert "developer" in result.lower()

    def test_status_shows_answered_count(self) -> None:
        from src.infrastructure.mcp.server import (
            guide_answer,
            guide_detect_persona,
            guide_start,
            guide_status,
        )

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_start("Idea", ctx))
        session_id = _extract_session_id(result)
        asyncio.run(guide_detect_persona(session_id, "1", ctx))
        asyncio.run(guide_answer(session_id, "Q1", "Users", ctx))

        result = asyncio.run(guide_status(session_id, ctx))
        assert "1" in result  # at least 1 answer

    def test_invalid_session_id_returns_error(self) -> None:
        from src.infrastructure.mcp.server import guide_status

        ctx, _ = _make_ctx_with_store()
        result = asyncio.run(guide_status("bad-id", ctx))
        assert "error" in result.lower() or "not found" in result.lower()


# ── Full flow integration ────────────────────────────────────────────


class TestDiscoveryFullFlow:
    """End-to-end test of the complete discovery flow via MCP tools."""

    def test_full_discovery_flow(self) -> None:
        from src.infrastructure.mcp.server import (
            guide_answer,
            guide_complete,
            guide_confirm_playback,
            guide_detect_persona,
            guide_start,
            guide_status,
        )

        ctx, store = _make_ctx_with_store()

        # 1. Start session
        result = asyncio.run(guide_start("A project idea for a todo app.", ctx))
        session_id = _extract_session_id(result)
        assert store.active_count() == 1

        # 2. Detect persona
        asyncio.run(guide_detect_persona(session_id, "1", ctx))

        # 3. Answer Q1-Q3, confirm playback
        for qid in ["Q1", "Q2", "Q3"]:
            asyncio.run(guide_answer(session_id, qid, f"Answer for {qid}", ctx))
        asyncio.run(guide_confirm_playback(session_id, True, ctx))

        # 4. Answer Q4-Q6, confirm playback
        for qid in ["Q4", "Q5", "Q6"]:
            asyncio.run(guide_answer(session_id, qid, f"Answer for {qid}", ctx))
        asyncio.run(guide_confirm_playback(session_id, True, ctx))

        # 5. Answer Q7-Q9, confirm playback
        for qid in ["Q7", "Q8", "Q9"]:
            asyncio.run(guide_answer(session_id, qid, f"Answer for {qid}", ctx))
        asyncio.run(guide_confirm_playback(session_id, True, ctx))

        # 6. Answer Q10
        asyncio.run(guide_answer(session_id, "Q10", "Answer Q10", ctx))

        # 7. Check status before completing
        status = asyncio.run(guide_status(session_id, ctx))
        assert "answering" in status.lower()

        # 8. Complete
        result = asyncio.run(guide_complete(session_id, ctx))
        assert "complete" in result.lower()

        # 9. Verify final state
        session = store.get(session_id)
        assert session.status == DiscoveryStatus.COMPLETED
        assert len(session.events) == 1
        assert len(session.answers) == 10


# ── Tool registration ────────────────────────────────────────────────


class TestDiscoveryToolRegistration:
    """Verify all 7 discovery tools are registered in the MCP server."""

    @pytest.fixture
    def tool_names(self) -> set[str]:
        from src.infrastructure.mcp.server import mcp

        return set(mcp._tool_manager._tools.keys())

    def test_guide_start_registered(self, tool_names: set[str]) -> None:
        assert "guide_start" in tool_names

    def test_guide_detect_persona_registered(self, tool_names: set[str]) -> None:
        assert "guide_detect_persona" in tool_names

    def test_guide_answer_registered(self, tool_names: set[str]) -> None:
        assert "guide_answer" in tool_names

    def test_guide_skip_question_registered(self, tool_names: set[str]) -> None:
        assert "guide_skip_question" in tool_names

    def test_guide_confirm_playback_registered(self, tool_names: set[str]) -> None:
        assert "guide_confirm_playback" in tool_names

    def test_guide_complete_registered(self, tool_names: set[str]) -> None:
        assert "guide_complete" in tool_names

    def test_guide_status_registered(self, tool_names: set[str]) -> None:
        assert "guide_status" in tool_names

    def test_guide_ddd_removed(self, tool_names: set[str]) -> None:
        """The monolithic guide_ddd tool should be replaced by focused tools."""
        assert "guide_ddd" not in tool_names


# ── DiscoveryAdapter unit tests ──────────────────────────────────────


class TestDiscoveryAdapter:
    """Tests for the DiscoveryAdapter that bridges SessionStore + DiscoveryHandler."""

    def test_implements_discovery_port(self) -> None:
        from src.application.ports.discovery_port import DiscoveryPort
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        adapter = DiscoveryAdapter(store=SessionStore())
        assert isinstance(adapter, DiscoveryPort)

    def test_start_session_stores_in_store(self) -> None:
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        store = SessionStore()
        adapter = DiscoveryAdapter(store=store)
        session = adapter.start_session("Idea")
        assert store.get(session.session_id) is session

    def test_detect_persona_updates_session(self) -> None:
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        store = SessionStore()
        adapter = DiscoveryAdapter(store=store)
        session = adapter.start_session("Idea")
        updated = adapter.detect_persona(session.session_id, "1")
        assert updated.persona == Persona.DEVELOPER

    def test_answer_question_delegates_to_session(self) -> None:
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        store = SessionStore()
        adapter = DiscoveryAdapter(store=store)
        session = adapter.start_session("Idea")
        adapter.detect_persona(session.session_id, "1")
        updated = adapter.answer_question(session.session_id, "Q1", "Users")
        assert len(updated.answers) == 1

    def test_skip_question_delegates_to_session(self) -> None:
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        store = SessionStore()
        adapter = DiscoveryAdapter(store=store)
        session = adapter.start_session("Idea")
        adapter.detect_persona(session.session_id, "1")
        updated = adapter.skip_question(session.session_id, "Q1", "Not needed")
        assert updated.status == DiscoveryStatus.PERSONA_DETECTED

    def test_confirm_playback_delegates_to_session(self) -> None:
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        store = SessionStore()
        adapter = DiscoveryAdapter(store=store)
        session = adapter.start_session("Idea")
        adapter.detect_persona(session.session_id, "1")
        adapter.answer_question(session.session_id, "Q1", "Users")
        adapter.answer_question(session.session_id, "Q2", "Entities")
        adapter.answer_question(session.session_id, "Q3", "Use case")
        updated = adapter.confirm_playback(session.session_id, True)
        assert updated.status == DiscoveryStatus.ANSWERING

    def test_complete_delegates_to_session(self) -> None:
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        store = SessionStore()
        adapter = DiscoveryAdapter(store=store)
        session = adapter.start_session("Idea")
        adapter.detect_persona(session.session_id, "1")
        for qid in ["Q1", "Q2", "Q3"]:
            adapter.answer_question(session.session_id, qid, f"Answer {qid}")
        adapter.confirm_playback(session.session_id, True)
        for qid in ["Q4", "Q5", "Q6"]:
            adapter.answer_question(session.session_id, qid, f"Answer {qid}")
        adapter.confirm_playback(session.session_id, True)
        for qid in ["Q7", "Q8", "Q9"]:
            adapter.answer_question(session.session_id, qid, f"Answer {qid}")
        adapter.confirm_playback(session.session_id, True)
        adapter.answer_question(session.session_id, "Q10", "Answer Q10")
        updated = adapter.complete(session.session_id)
        assert updated.status == DiscoveryStatus.COMPLETED

    def test_get_session_returns_session(self) -> None:
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        store = SessionStore()
        adapter = DiscoveryAdapter(store=store)
        session = adapter.start_session("Idea")
        retrieved = adapter.get_session(session.session_id)
        assert retrieved is session

    def test_get_session_not_found_raises(self) -> None:
        from src.domain.models.errors import SessionNotFoundError
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        store = SessionStore()
        adapter = DiscoveryAdapter(store=store)
        with pytest.raises(SessionNotFoundError):
            adapter.get_session("nonexistent")

    def test_not_found_raises_session_not_found(self) -> None:
        from src.domain.models.errors import SessionNotFoundError
        from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter

        store = SessionStore()
        adapter = DiscoveryAdapter(store=store)
        with pytest.raises(SessionNotFoundError):
            adapter.detect_persona("nonexistent", "1")


# ── Helpers ──────────────────────────────────────────────────────────


def _extract_session_id(result: str) -> str:
    """Extract a UUID session_id from a tool response string.

    Expects the session_id to be a UUID-4 format string in the response.
    """
    import re

    match = re.search(r"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", result)
    if match:
        return match.group()
    msg = f"Could not extract session_id from: {result}"
    raise ValueError(msg)
