"""Tests for BootstrapSession aggregate root.

Verifies the state machine (CREATED -> PREVIEWED -> CONFIRMED -> EXECUTING ->
COMPLETED / CANCELLED) and the preview-before-action invariant.
"""

from __future__ import annotations

from pathlib import Path

import pytest

from src.domain.events.bootstrap_events import BootstrapCompleted
from src.domain.models.bootstrap_session import (
    BootstrapSession,
    InvariantViolationError,
    SessionStatus,
)
from src.domain.models.preview import FileAction, FileActionType, Preview

# ── Helpers ──────────────────────────────────────────────────


def _make_preview() -> Preview:
    """Build a minimal valid Preview for testing."""
    return Preview(
        file_actions=(FileAction(path=Path("docs/PRD.md"), action_type=FileActionType.CREATE),),
    )


# ── Creation ─────────────────────────────────────────────────


class TestBootstrapSessionCreation:
    def test_new_session_starts_in_created_state(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        assert session.status == SessionStatus.CREATED

    def test_new_session_has_unique_id(self):
        s1 = BootstrapSession(project_dir=Path("/tmp/a"))
        s2 = BootstrapSession(project_dir=Path("/tmp/b"))
        assert s1.session_id != s2.session_id

    def test_new_session_has_project_dir(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        assert session.project_dir == Path("/tmp/proj")


# ── State Transitions ────────────────────────────────────────


class TestBootstrapSessionStateTransitions:
    def test_set_preview_transitions_to_previewed(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        assert session.status == SessionStatus.PREVIEWED

    def test_confirm_requires_preview_first(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        with pytest.raises(InvariantViolationError, match="Cannot confirm without preview"):
            session.confirm()

    def test_confirm_transitions_to_confirmed(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.confirm()
        assert session.status == SessionStatus.CONFIRMED

    def test_cancel_transitions_to_cancelled(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.cancel()
        assert session.status == SessionStatus.CANCELLED

    def test_cancel_from_created_raises(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        with pytest.raises(InvariantViolationError, match="Can only cancel from previewed state"):
            session.cancel()

    def test_begin_execution_requires_confirmation(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        with pytest.raises(InvariantViolationError, match="Cannot execute without confirmation"):
            session.begin_execution()

    def test_begin_execution_transitions_to_executing(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.confirm()
        session.begin_execution()
        assert session.status == SessionStatus.EXECUTING

    def test_complete_transitions_to_completed(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.confirm()
        session.begin_execution()
        session.complete()
        assert session.status == SessionStatus.COMPLETED

    def test_complete_produces_bootstrap_completed_event(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.confirm()
        session.begin_execution()
        session.complete()

        assert len(session.events) == 1
        event = session.events[0]
        assert isinstance(event, BootstrapCompleted)
        assert event.session_id == session.session_id
        assert event.project_dir == Path("/tmp/proj")


# ── Invariants ───────────────────────────────────────────────


class TestBootstrapSessionInvariants:
    def test_cannot_preview_twice_replaces_preview(self):
        """Setting preview again from PREVIEWED replaces it (idempotent)."""
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        p1 = _make_preview()
        p2 = Preview(
            file_actions=(
                FileAction(path=Path("docs/DDD.md"), action_type=FileActionType.CREATE),
            ),
        )
        session.set_preview(p1)
        session.set_preview(p2)
        assert session.preview == p2
        assert session.status == SessionStatus.PREVIEWED

    def test_cannot_confirm_cancelled_session(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.cancel()
        with pytest.raises(InvariantViolationError, match="Cannot confirm without preview"):
            session.confirm()

    def test_cannot_execute_cancelled_session(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.cancel()
        with pytest.raises(InvariantViolationError, match="Cannot execute without confirmation"):
            session.begin_execution()

    def test_cannot_preview_from_confirmed_state(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.confirm()
        with pytest.raises(InvariantViolationError, match="Cannot preview in confirmed state"):
            session.set_preview(_make_preview())

    def test_cannot_preview_from_executing_state(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.confirm()
        session.begin_execution()
        with pytest.raises(InvariantViolationError, match="Cannot preview in executing state"):
            session.set_preview(_make_preview())

    def test_cannot_preview_from_completed_state(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.confirm()
        session.begin_execution()
        session.complete()
        with pytest.raises(InvariantViolationError, match="Cannot preview in completed state"):
            session.set_preview(_make_preview())

    def test_double_complete_raises(self):
        """Calling complete() twice raises InvariantViolationError."""
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.confirm()
        session.begin_execution()
        session.complete()
        with pytest.raises(InvariantViolationError, match="Cannot complete unless executing"):
            session.complete()

    def test_events_returns_defensive_copy(self):
        """Mutating the returned events list does not affect the session."""
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_preview(_make_preview())
        session.confirm()
        session.begin_execution()
        session.complete()
        events = session.events
        events.clear()
        assert len(session.events) == 1


# ── Detected Tools ──────────────────────────────────────────


class TestBootstrapSessionDetectedTools:
    def test_new_session_has_empty_detected_tools(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        assert session.detected_tools == ()

    def test_set_detected_tools_stores_tools(self):
        session = BootstrapSession(project_dir=Path("/tmp/proj"))
        session.set_detected_tools(["claude", "cursor"])
        assert session.detected_tools == ("claude", "cursor")
