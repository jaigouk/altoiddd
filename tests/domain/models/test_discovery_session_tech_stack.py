"""Tests for TechStack wiring into DiscoverySession and DiscoveryCompleted."""

from __future__ import annotations

import pytest

from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.errors import InvariantViolationError
from src.domain.models.tech_stack import TechStack


def _make_session(readme: str = "# Test") -> DiscoverySession:
    """Create a fresh DiscoverySession."""
    return DiscoverySession(readme_content=readme)


def _session_in_persona_detected(readme: str = "# Test") -> DiscoverySession:
    """Create a session with persona detected."""
    session = _make_session(readme)
    session.detect_persona("1")
    return session


def _make_answering_session() -> DiscoverySession:
    """Create a session in ANSWERING state (persona detected + 1 answer)."""
    session = _session_in_persona_detected()
    session.answer_question("Q1", "Some actors")
    return session


class TestSessionTechStackDefault:
    """New DiscoverySession has tech_stack=None."""

    def test_default_none(self) -> None:
        session = _make_session()
        assert session.tech_stack is None


class TestSessionSetTechStack:
    """Session.set_tech_stack() stores the value."""

    def test_set_in_created_state(self) -> None:
        session = _make_session()
        ts = TechStack(language="python", package_manager="uv")
        session.set_tech_stack(ts)
        assert session.tech_stack == ts

    def test_set_in_persona_detected_state(self) -> None:
        session = _session_in_persona_detected()
        ts = TechStack(language="python", package_manager="uv")
        session.set_tech_stack(ts)
        assert session.tech_stack == ts

    def test_set_twice_overwrites(self) -> None:
        session = _make_session()
        ts1 = TechStack(language="python", package_manager="uv")
        ts2 = TechStack(language="rust", package_manager="cargo")
        session.set_tech_stack(ts1)
        session.set_tech_stack(ts2)
        assert session.tech_stack == ts2

    def test_set_in_answering_state_raises(self) -> None:
        session = _make_answering_session()
        ts = TechStack(language="python", package_manager="uv")
        with pytest.raises(InvariantViolationError, match="CREATED or PERSONA_DETECTED"):
            session.set_tech_stack(ts)

    def test_set_in_completed_state_raises(self) -> None:
        session = _session_in_persona_detected()
        # Answer all MVP questions to allow completion
        from src.domain.models.question import Question

        for i, q in enumerate(Question.CATALOG):
            session.answer_question(q.id, f"Answer {i}")
            if session.status == DiscoveryStatus.PLAYBACK_PENDING:
                session.confirm_playback(confirmed=True)
        session.complete()

        ts = TechStack(language="python", package_manager="uv")
        with pytest.raises(InvariantViolationError, match="CREATED or PERSONA_DETECTED"):
            session.set_tech_stack(ts)


class TestSessionSnapshotWithTechStack:
    """to_snapshot() serializes tech_stack correctly."""

    def test_snapshot_includes_tech_stack(self) -> None:
        session = _make_session()
        ts = TechStack(language="python", package_manager="uv")
        session.set_tech_stack(ts)
        snapshot = session.to_snapshot()
        assert snapshot["tech_stack"] == {"language": "python", "package_manager": "uv"}

    def test_snapshot_tech_stack_none(self) -> None:
        session = _make_session()
        snapshot = session.to_snapshot()
        assert snapshot["tech_stack"] is None


class TestSessionFromSnapshotTechStack:
    """from_snapshot() restores tech_stack correctly."""

    def test_from_snapshot_with_tech_stack(self) -> None:
        session = _make_session()
        ts = TechStack(language="python", package_manager="uv")
        session.set_tech_stack(ts)
        snapshot = session.to_snapshot()

        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.tech_stack == ts

    def test_from_snapshot_with_null_tech_stack(self) -> None:
        session = _make_session()
        snapshot = session.to_snapshot()
        assert snapshot["tech_stack"] is None

        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.tech_stack is None

    def test_from_snapshot_without_tech_stack_key(self) -> None:
        """Old snapshots missing tech_stack key default to None."""
        session = _make_session()
        snapshot = session.to_snapshot()
        # Simulate old snapshot by removing the key
        del snapshot["tech_stack"]

        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.tech_stack is None


class TestDiscoveryCompletedCarriesTechStack:
    """DiscoveryCompleted event includes tech_stack field."""

    def _complete_session(
        self, tech_stack: TechStack | None = None
    ) -> DiscoverySession:
        """Helper: build a completed session with optional tech_stack."""
        from src.domain.models.question import Question

        session = _session_in_persona_detected()
        if tech_stack is not None:
            # Must set before answering (PERSONA_DETECTED state)
            session.set_tech_stack(tech_stack)

        for i, q in enumerate(Question.CATALOG):
            session.answer_question(q.id, f"Answer {i}")
            if session.status == DiscoveryStatus.PLAYBACK_PENDING:
                session.confirm_playback(confirmed=True)
        session.complete()
        return session

    def test_event_carries_tech_stack(self) -> None:
        ts = TechStack(language="python", package_manager="uv")
        session = self._complete_session(tech_stack=ts)
        events = session.events
        assert len(events) == 1
        assert events[0].tech_stack == ts

    def test_event_tech_stack_none_when_not_set(self) -> None:
        session = self._complete_session(tech_stack=None)
        events = session.events
        assert len(events) == 1
        assert events[0].tech_stack is None
