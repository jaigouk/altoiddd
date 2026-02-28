"""Tests for DiscoveryHandler application command.

Verifies handler orchestration, session lookup, and SessionNotFoundError.
"""

from __future__ import annotations

import pytest

from src.application.commands.discovery_handler import DiscoveryHandler
from src.domain.models.bootstrap_session import SessionNotFoundError
from src.domain.models.discovery_session import DiscoveryStatus
from src.domain.models.discovery_values import Persona, Register


class TestDiscoveryHandlerStartSession:
    def test_start_session_returns_session(self):
        handler = DiscoveryHandler()
        session = handler.start_session("A project idea in 4-5 sentences.")
        assert session.status == DiscoveryStatus.CREATED
        assert session.readme_content == "A project idea in 4-5 sentences."

    def test_start_session_creates_unique_ids(self):
        handler = DiscoveryHandler()
        s1 = handler.start_session("Idea A")
        s2 = handler.start_session("Idea B")
        assert s1.session_id != s2.session_id


class TestDiscoveryHandlerDetectPersona:
    def test_detect_persona_returns_updated_session(self):
        handler = DiscoveryHandler()
        session = handler.start_session("Idea")
        result = handler.detect_persona(session.session_id, "1")
        assert result.persona == Persona.DEVELOPER
        assert result.register == Register.TECHNICAL
        assert result.status == DiscoveryStatus.PERSONA_DETECTED

    def test_detect_persona_not_found_raises(self):
        handler = DiscoveryHandler()
        with pytest.raises(SessionNotFoundError):
            handler.detect_persona("nonexistent-id", "1")


class TestDiscoveryHandlerAnswerQuestion:
    def test_answer_question_returns_updated_session(self):
        handler = DiscoveryHandler()
        session = handler.start_session("Idea")
        handler.detect_persona(session.session_id, "1")
        result = handler.answer_question(session.session_id, "Q1", "Users and admins")
        assert len(result.answers) == 1
        assert result.answers[0].question_id == "Q1"

    def test_answer_question_not_found_raises(self):
        handler = DiscoveryHandler()
        with pytest.raises(SessionNotFoundError):
            handler.answer_question("nonexistent-id", "Q1", "Answer")


class TestDiscoveryHandlerConfirmPlayback:
    def test_confirm_playback_returns_updated_session(self):
        handler = DiscoveryHandler()
        session = handler.start_session("Idea")
        handler.detect_persona(session.session_id, "1")
        handler.answer_question(session.session_id, "Q1", "Users")
        handler.answer_question(session.session_id, "Q2", "Entities")
        handler.answer_question(session.session_id, "Q3", "Use case")
        result = handler.confirm_playback(session.session_id, confirmed=True)
        assert result.status == DiscoveryStatus.ANSWERING

    def test_confirm_playback_not_found_raises(self):
        handler = DiscoveryHandler()
        with pytest.raises(SessionNotFoundError):
            handler.confirm_playback("nonexistent-id", confirmed=True)


class TestDiscoveryHandlerComplete:
    def test_complete_returns_completed_session(self):
        handler = DiscoveryHandler()
        session = handler.start_session("Idea")
        handler.detect_persona(session.session_id, "1")
        # Answer all 10 questions, confirming playbacks
        for qid in ["Q1", "Q2", "Q3"]:
            handler.answer_question(session.session_id, qid, f"Answer {qid}")
        handler.confirm_playback(session.session_id, confirmed=True)
        for qid in ["Q4", "Q5", "Q6"]:
            handler.answer_question(session.session_id, qid, f"Answer {qid}")
        handler.confirm_playback(session.session_id, confirmed=True)
        for qid in ["Q7", "Q8", "Q9"]:
            handler.answer_question(session.session_id, qid, f"Answer {qid}")
        handler.confirm_playback(session.session_id, confirmed=True)
        handler.answer_question(session.session_id, "Q10", "Answer Q10")

        result = handler.complete(session.session_id)
        assert result.status == DiscoveryStatus.COMPLETED
        assert len(result.events) == 1

    def test_complete_not_found_raises(self):
        handler = DiscoveryHandler()
        with pytest.raises(SessionNotFoundError):
            handler.complete("nonexistent-id")
