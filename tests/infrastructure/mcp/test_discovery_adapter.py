"""Tests for DiscoveryAdapter — session lifecycle through SessionStore.

Covers:
- start_session creates and persists a session
- set_mode persists mode change to store
- set_tech_stack persists tech stack
- detect_persona persists persona
- answer_question persists answer
- skip_question persists skip
- confirm_playback persists confirmation
- complete persists completion
- get_session raises on unknown ID
- full lifecycle round-trip through adapter
"""

from __future__ import annotations

import pytest

from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.discovery_values import DiscoveryMode
from src.domain.models.errors import SessionNotFoundError
from src.domain.models.question import Question
from src.infrastructure.mcp.discovery_adapter import DiscoveryAdapter
from src.infrastructure.session.session_store import SessionStore

# -- Helpers ------------------------------------------------------------------


def _adapter_with_store() -> tuple[DiscoveryAdapter, SessionStore]:
    store = SessionStore()
    adapter = DiscoveryAdapter(store=store)
    return adapter, store


def _start_session(adapter: DiscoveryAdapter) -> DiscoverySession:
    return adapter.start_session(readme_content="A test project.")


def _session_through_persona(adapter: DiscoveryAdapter) -> DiscoverySession:
    session = _start_session(adapter)
    return adapter.detect_persona(session.session_id, "1")


def _answer_all(adapter: DiscoveryAdapter, session_id: str) -> None:
    """Answer all 10 questions, confirming playbacks as they arise."""
    for q in Question.CATALOG:
        s = adapter.get_session(session_id)
        if s.status == DiscoveryStatus.PLAYBACK_PENDING:
            adapter.confirm_playback(session_id, confirmed=True)
        adapter.answer_question(session_id, q.id, f"Answer for {q.id}")
    s = adapter.get_session(session_id)
    if s.status == DiscoveryStatus.PLAYBACK_PENDING:
        adapter.confirm_playback(session_id, confirmed=True)


# -- start_session ------------------------------------------------------------


class TestStartSession:
    def test_creates_session_in_store(self) -> None:
        adapter, store = _adapter_with_store()
        session = _start_session(adapter)
        assert store.active_count() == 1
        assert isinstance(store.get(session.session_id), DiscoverySession)

    def test_session_starts_in_created_state(self) -> None:
        adapter, _ = _adapter_with_store()
        session = _start_session(adapter)
        assert session.status == DiscoveryStatus.CREATED


# -- get_session --------------------------------------------------------------


class TestGetSession:
    def test_retrieves_existing_session(self) -> None:
        adapter, _ = _adapter_with_store()
        session = _start_session(adapter)
        retrieved = adapter.get_session(session.session_id)
        assert retrieved.session_id == session.session_id

    def test_raises_on_unknown_id(self) -> None:
        adapter, _ = _adapter_with_store()
        with pytest.raises(SessionNotFoundError):
            adapter.get_session("nonexistent-id")


# -- set_mode persistence -----------------------------------------------------


class TestSetModePersistence:
    def test_set_mode_persists_to_store(self) -> None:
        adapter, store = _adapter_with_store()
        session = _start_session(adapter)
        adapter.set_mode(session.session_id, DiscoveryMode.DEEP)
        retrieved = store.get(session.session_id)
        assert isinstance(retrieved, DiscoverySession)
        assert retrieved.mode == DiscoveryMode.DEEP

    def test_set_mode_express_persists(self) -> None:
        adapter, store = _adapter_with_store()
        session = _start_session(adapter)
        adapter.set_mode(session.session_id, DiscoveryMode.EXPRESS)
        retrieved = store.get(session.session_id)
        assert isinstance(retrieved, DiscoverySession)
        assert retrieved.mode == DiscoveryMode.EXPRESS

    def test_set_mode_returns_updated_session(self) -> None:
        adapter, _ = _adapter_with_store()
        session = _start_session(adapter)
        result = adapter.set_mode(session.session_id, DiscoveryMode.DEEP)
        assert result.mode == DiscoveryMode.DEEP


# -- set_tech_stack persistence -----------------------------------------------


class TestSetTechStackPersistence:
    def test_set_tech_stack_returns_session(self) -> None:
        from src.domain.models.tech_stack import TechStack

        adapter, _ = _adapter_with_store()
        session = _start_session(adapter)
        ts = TechStack(language="python", package_manager="uv")
        result = adapter.set_tech_stack(session.session_id, ts)
        assert result.tech_stack is not None
        assert result.tech_stack.language == "python"


# -- detect_persona persistence -----------------------------------------------


class TestDetectPersonaPersistence:
    def test_detect_persona_returns_session_with_persona(self) -> None:
        adapter, _ = _adapter_with_store()
        session = _start_session(adapter)
        result = adapter.detect_persona(session.session_id, "1")
        assert result.persona is not None
        assert result.status == DiscoveryStatus.PERSONA_DETECTED


# -- answer_question persistence ----------------------------------------------


class TestAnswerQuestionPersistence:
    def test_answer_persists_through_adapter(self) -> None:
        adapter, _ = _adapter_with_store()
        session = _session_through_persona(adapter)
        result = adapter.answer_question(session.session_id, "Q1", "Users and admins")
        assert len(result.answers) == 1
        assert result.answers[0].question_id == "Q1"


# -- skip_question persistence ------------------------------------------------


class TestSkipQuestionPersistence:
    def test_skip_persists_through_adapter(self) -> None:
        adapter, _ = _adapter_with_store()
        session = _session_through_persona(adapter)
        # Answer Q1 first to enter ANSWERING state, then skip Q2
        adapter.answer_question(session.session_id, "Q1", "Users")
        result = adapter.skip_question(session.session_id, "Q2", "Not relevant")
        assert result.status in (DiscoveryStatus.ANSWERING, DiscoveryStatus.PLAYBACK_PENDING)


# -- confirm_playback persistence ---------------------------------------------


class TestConfirmPlaybackPersistence:
    def test_confirm_playback_transitions_state(self) -> None:
        adapter, _ = _adapter_with_store()
        session = _session_through_persona(adapter)
        # Answer 3 questions to trigger playback
        adapter.answer_question(session.session_id, "Q1", "Users")
        adapter.answer_question(session.session_id, "Q2", "Orders")
        adapter.answer_question(session.session_id, "Q3", "Create order")
        s = adapter.get_session(session.session_id)
        assert s.status == DiscoveryStatus.PLAYBACK_PENDING
        result = adapter.confirm_playback(session.session_id, confirmed=True)
        assert result.status == DiscoveryStatus.ANSWERING


# -- complete persistence -----------------------------------------------------


class TestCompletePersistence:
    def test_express_complete_persists(self) -> None:
        adapter, _ = _adapter_with_store()
        session = _session_through_persona(adapter)
        _answer_all(adapter, session.session_id)
        result = adapter.complete(session.session_id)
        assert result.status == DiscoveryStatus.COMPLETED

    def test_deep_complete_persists_round_1(self) -> None:
        adapter, _ = _adapter_with_store()
        session = _start_session(adapter)
        adapter.set_mode(session.session_id, DiscoveryMode.DEEP)
        adapter.detect_persona(session.session_id, "1")
        _answer_all(adapter, session.session_id)
        result = adapter.complete(session.session_id)
        assert result.status == DiscoveryStatus.ROUND_1_COMPLETE


# -- Full lifecycle round-trip ------------------------------------------------


class TestFullLifecycleRoundTrip:
    def test_express_full_flow(self) -> None:
        """Express mode: start → persona → answer all → complete."""
        adapter, store = _adapter_with_store()
        session = _start_session(adapter)
        sid = session.session_id
        adapter.detect_persona(sid, "2")
        _answer_all(adapter, sid)
        result = adapter.complete(sid)
        assert result.status == DiscoveryStatus.COMPLETED
        assert len(result.events) == 1
        # Session persisted in store
        assert store.active_count() == 1

    def test_deep_full_flow(self) -> None:
        """Deep mode: start → mode → persona → answer all → complete → challenge → simulate."""
        adapter, _store = _adapter_with_store()
        session = _start_session(adapter)
        sid = session.session_id

        adapter.set_mode(sid, DiscoveryMode.DEEP)
        adapter.detect_persona(sid, "1")
        _answer_all(adapter, sid)

        result = adapter.complete(sid)
        assert result.status == DiscoveryStatus.ROUND_1_COMPLETE
        assert len(result.events) == 0

        # Challenge and simulate happen on the session directly.
        # Re-fetch session between transitions to avoid mypy narrowing issues.
        s = adapter.get_session(sid)
        s.start_challenge()
        assert adapter.get_session(sid).status == DiscoveryStatus.CHALLENGING
        s.complete_challenge()
        assert adapter.get_session(sid).status == DiscoveryStatus.ROUND_2_COMPLETE
        s.start_simulate()
        assert adapter.get_session(sid).status == DiscoveryStatus.SIMULATING
        s.complete_simulation()
        assert adapter.get_session(sid).status == DiscoveryStatus.COMPLETED
        assert len(s.events) == 1
