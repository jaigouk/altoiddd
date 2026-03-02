"""Tests for InMemoryDiscoveryAdapter.

Verifies that the adapter correctly bridges the DiscoveryPort protocol
to the DiscoverySession aggregate via SessionStore, delegating all
state transitions to the domain model.
"""

from __future__ import annotations

import pytest

from src.application.ports.discovery_port import DiscoveryPort
from src.domain.models.discovery_session import DiscoveryStatus
from src.domain.models.errors import InvariantViolationError, SessionNotFoundError
from src.infrastructure.session.in_memory_discovery_adapter import InMemoryDiscoveryAdapter
from src.infrastructure.session.session_store import SessionStore

# -- Helpers ------------------------------------------------------------------

_README = "A test project idea in 4-5 sentences."

# All 10 question IDs in phase order (ACTORS -> STORY -> EVENTS -> BOUNDARIES).
_ALL_QUESTION_IDS = ["Q1", "Q2", "Q3", "Q4", "Q5", "Q6", "Q7", "Q8", "Q9", "Q10"]


def _make_adapter() -> InMemoryDiscoveryAdapter:
    """Create a fresh adapter with a new SessionStore."""
    return InMemoryDiscoveryAdapter(store=SessionStore())


def _start_with_persona(adapter: InMemoryDiscoveryAdapter, choice: str = "1") -> str:
    """Start a session and detect persona, returning session_id."""
    session = adapter.start_session(_README)
    adapter.detect_persona(session.session_id, choice)
    return session.session_id


def _answer_all_questions(adapter: InMemoryDiscoveryAdapter, session_id: str) -> None:
    """Answer all 10 questions, confirming playbacks as they arise.

    Playback triggers after every 3 answers:
      Q1, Q2, Q3 -> playback -> confirm
      Q4, Q5, Q6 -> playback -> confirm
      Q7, Q8, Q9 -> playback -> confirm
      Q10 -> no playback (only 1 answer since last confirmation)
    """
    for qid in _ALL_QUESTION_IDS:
        session = adapter.answer_question(session_id, qid, f"Answer for {qid}")
        if session.status == DiscoveryStatus.PLAYBACK_PENDING:
            adapter.confirm_playback(session_id, True)


# -- Happy Path Tests --------------------------------------------------------


class TestHappyPath:
    def test_start_session_returns_created_session(self) -> None:
        adapter = _make_adapter()
        session = adapter.start_session(_README)
        assert session.session_id
        assert session.status == DiscoveryStatus.CREATED
        assert session.readme_content == _README

    def test_detect_persona_transitions(self) -> None:
        adapter = _make_adapter()
        session = adapter.start_session(_README)
        updated = adapter.detect_persona(session.session_id, "1")
        assert updated.status == DiscoveryStatus.PERSONA_DETECTED

    def test_answer_question_records_answer(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        updated = adapter.answer_question(sid, "Q1", "Users and admins")
        assert len(updated.answers) == 1
        assert updated.answers[0].question_id == "Q1"
        assert updated.answers[0].response_text == "Users and admins"

    def test_skip_question_records_skip(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        # Skip Q1 (non-MVP for the skip test, but Q1 is MVP -- skip Q2 instead)
        adapter.answer_question(sid, "Q1", "Actors")
        updated = adapter.skip_question(sid, "Q2", "Not relevant for this project")
        # Q2 should not appear in answers
        answered_ids = {a.question_id for a in updated.answers}
        assert "Q2" not in answered_ids

    def test_confirm_playback_transitions(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        # Answer 3 questions to trigger playback
        adapter.answer_question(sid, "Q1", "Actors")
        adapter.answer_question(sid, "Q2", "Entities")
        session = adapter.answer_question(sid, "Q3", "Use case")
        assert session.status == DiscoveryStatus.PLAYBACK_PENDING
        updated = adapter.confirm_playback(sid, True)
        assert updated.status == DiscoveryStatus.ANSWERING

    def test_confirm_playback_with_corrections(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        adapter.answer_question(sid, "Q1", "Actors")
        adapter.answer_question(sid, "Q2", "Entities")
        adapter.answer_question(sid, "Q3", "Use case")
        updated = adapter.confirm_playback(sid, True, "fix actors")
        assert updated.playback_confirmations[-1].corrections == "fix actors"

    def test_complete_transitions(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        _answer_all_questions(adapter, sid)
        completed = adapter.complete(sid)
        assert completed.status == DiscoveryStatus.COMPLETED
        assert len(completed.events) == 1

    def test_get_session_returns_stored(self) -> None:
        adapter = _make_adapter()
        session = adapter.start_session(_README)
        retrieved = adapter.get_session(session.session_id)
        assert retrieved.session_id == session.session_id
        assert retrieved.readme_content == _README

    def test_full_discovery_flow(self) -> None:
        """Complete end-to-end flow: start -> detect -> answer all -> complete."""
        adapter = _make_adapter()
        session = adapter.start_session(_README)
        sid = session.session_id

        # Detect persona
        adapter.detect_persona(sid, "1")

        # Answer all 10 questions with playback confirmations
        _answer_all_questions(adapter, sid)

        # Complete
        completed = adapter.complete(sid)
        assert completed.status == DiscoveryStatus.COMPLETED
        assert len(completed.answers) == 10
        assert len(completed.events) == 1
        # 3 playback confirmations (after Q3, Q6, Q9)
        assert len(completed.playback_confirmations) == 3

    def test_satisfies_discovery_port_protocol(self) -> None:
        adapter = _make_adapter()
        assert isinstance(adapter, DiscoveryPort)

    def test_multiple_concurrent_sessions(self) -> None:
        adapter = _make_adapter()
        s1 = adapter.start_session("Project A idea.")
        s2 = adapter.start_session("Project B idea.")
        assert s1.session_id != s2.session_id

        adapter.detect_persona(s1.session_id, "1")
        adapter.detect_persona(s2.session_id, "2")

        r1 = adapter.get_session(s1.session_id)
        r2 = adapter.get_session(s2.session_id)
        assert r1.status == DiscoveryStatus.PERSONA_DETECTED
        assert r2.status == DiscoveryStatus.PERSONA_DETECTED
        assert r1.readme_content == "Project A idea."
        assert r2.readme_content == "Project B idea."


# -- Error Propagation Tests --------------------------------------------------


class TestErrorPropagation:
    def test_get_session_not_found(self) -> None:
        adapter = _make_adapter()
        with pytest.raises(SessionNotFoundError):
            adapter.get_session("nonexistent-id")

    def test_detect_persona_invalid_choice(self) -> None:
        adapter = _make_adapter()
        session = adapter.start_session(_README)
        with pytest.raises(ValueError, match="Invalid persona choice"):
            adapter.detect_persona(session.session_id, "5")

    def test_detect_persona_already_detected(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        with pytest.raises(InvariantViolationError, match="CREATED"):
            adapter.detect_persona(sid, "2")

    def test_answer_empty_response(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        with pytest.raises(ValueError, match="empty"):
            adapter.answer_question(sid, "Q1", "")

    def test_answer_unknown_question(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        with pytest.raises(ValueError, match="Unknown question"):
            adapter.answer_question(sid, "Q99", "Some answer")

    def test_answer_duplicate(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        adapter.answer_question(sid, "Q1", "Actors")
        with pytest.raises(InvariantViolationError, match="already answered"):
            adapter.answer_question(sid, "Q1", "Different answer")

    def test_answer_during_playback(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        adapter.answer_question(sid, "Q1", "Actors")
        adapter.answer_question(sid, "Q2", "Entities")
        adapter.answer_question(sid, "Q3", "Use case")
        with pytest.raises(InvariantViolationError, match="playback"):
            adapter.answer_question(sid, "Q4", "Failure")

    def test_skip_empty_reason(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        with pytest.raises(ValueError, match="reason"):
            adapter.skip_question(sid, "Q1", "")

    def test_skip_during_playback(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        adapter.answer_question(sid, "Q1", "Actors")
        adapter.answer_question(sid, "Q2", "Entities")
        adapter.answer_question(sid, "Q3", "Use case")
        with pytest.raises(InvariantViolationError, match="playback"):
            adapter.skip_question(sid, "Q4", "Not needed")

    def test_complete_without_enough_answers(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        adapter.answer_question(sid, "Q1", "Actors")
        adapter.answer_question(sid, "Q2", "Entities")
        adapter.answer_question(sid, "Q3", "Use case")
        adapter.confirm_playback(sid, True)
        with pytest.raises(InvariantViolationError, match="MVP"):
            adapter.complete(sid)

    def test_complete_during_playback(self) -> None:
        adapter = _make_adapter()
        sid = _start_with_persona(adapter)
        adapter.answer_question(sid, "Q1", "Actors")
        adapter.answer_question(sid, "Q2", "Entities")
        adapter.answer_question(sid, "Q3", "Use case")
        with pytest.raises(InvariantViolationError, match="ANSWERING"):
            adapter.complete(sid)
