"""Tests for DiscoveryMode + DiscoveryRound enums and mode-aware state machine.

RED phase: these tests define the contract for Express/Deep mode selection
on the DiscoverySession aggregate root (alty-20c.3).

Tests cover:
- DiscoveryMode and DiscoveryRound enum values
- set_mode() preconditions (CREATED state, single call)
- Mode-aware complete() behavior (EXPRESS vs DEEP)
- Deep mode state transitions (ROUND_1_COMPLETE -> CHALLENGING -> ...)
- Snapshot round-trip with new mode/round fields
- Backward compatibility with old snapshots (no mode field)
"""

from __future__ import annotations

import pytest

from src.domain.events.discovery_events import DiscoveryCompleted
from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.errors import InvariantViolationError
from src.domain.models.question import Question

# -- Helpers ------------------------------------------------------------------


def _session_created() -> DiscoverySession:
    """Create a session in CREATED state."""
    return DiscoverySession(readme_content="A test project idea.")


def _session_with_persona(choice: str = "1") -> DiscoverySession:
    """Create a session with persona already detected (default mode = EXPRESS)."""
    session = _session_created()
    session.detect_persona(choice)
    return session


def _answer_all_questions(
    session: DiscoverySession,
) -> None:
    """Answer all 10 questions, confirming playbacks as they arise."""
    for q in Question.CATALOG:
        if session.status == DiscoveryStatus.PLAYBACK_PENDING:
            session.confirm_playback(confirmed=True)
        session.answer_question(q.id, f"Answer for {q.id}")
    if session.status == DiscoveryStatus.PLAYBACK_PENDING:
        session.confirm_playback(confirmed=True)


def _session_deep_answering() -> DiscoverySession:
    """Create a DEEP mode session with all questions answered (ANSWERING state)."""
    from src.domain.models.discovery_values import DiscoveryMode

    session = _session_created()
    session.set_mode(DiscoveryMode.DEEP)
    session.detect_persona("1")
    _answer_all_questions(session)
    return session


# -- Enum Values --------------------------------------------------------------


class TestDiscoveryModeEnum:
    def test_express_value(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        assert DiscoveryMode.EXPRESS.value == "express"

    def test_deep_value(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        assert DiscoveryMode.DEEP.value == "deep"

    def test_enum_has_exactly_two_members(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        assert len(DiscoveryMode) == 2


class TestDiscoveryRoundEnum:
    def test_discovery_value(self) -> None:
        from src.domain.models.discovery_values import DiscoveryRound

        assert DiscoveryRound.DISCOVERY.value == "discovery"

    def test_challenge_value(self) -> None:
        from src.domain.models.discovery_values import DiscoveryRound

        assert DiscoveryRound.CHALLENGE.value == "challenge"

    def test_simulate_value(self) -> None:
        from src.domain.models.discovery_values import DiscoveryRound

        assert DiscoveryRound.SIMULATE.value == "simulate"

    def test_enum_has_exactly_three_members(self) -> None:
        from src.domain.models.discovery_values import DiscoveryRound

        assert len(DiscoveryRound) == 3


# -- set_mode() ---------------------------------------------------------------


class TestSetMode:
    def test_set_mode_deep_in_created_state(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.DEEP)
        assert session.mode == DiscoveryMode.DEEP

    def test_set_mode_express_in_created_state(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.EXPRESS)
        assert session.mode == DiscoveryMode.EXPRESS

    def test_set_mode_not_in_created_state_raises(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_with_persona()  # PERSONA_DETECTED state
        with pytest.raises(InvariantViolationError, match="CREATED"):
            session.set_mode(DiscoveryMode.DEEP)

    def test_set_mode_in_answering_state_raises(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_with_persona()
        session.answer_question("Q1", "Users")
        with pytest.raises(InvariantViolationError, match="CREATED"):
            session.set_mode(DiscoveryMode.DEEP)

    def test_set_mode_twice_raises(self) -> None:
        """Once mode is set, it cannot be changed (even in CREATED state)."""
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.DEEP)
        with pytest.raises(InvariantViolationError):
            session.set_mode(DiscoveryMode.EXPRESS)

    def test_set_mode_twice_same_value_raises(self) -> None:
        """Even setting the same mode twice is rejected."""
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.DEEP)
        with pytest.raises(InvariantViolationError):
            session.set_mode(DiscoveryMode.DEEP)


# -- Default mode property ----------------------------------------------------


class TestDefaultMode:
    def test_default_mode_is_express(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        assert session.mode == DiscoveryMode.EXPRESS

    def test_default_mode_after_persona_detection(self) -> None:
        """Mode defaults to EXPRESS even after persona detection."""
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_with_persona()
        assert session.mode == DiscoveryMode.EXPRESS


# -- Express mode complete() (no regression) -----------------------------------


class TestExpressModeComplete:
    def test_express_complete_sets_completed_status(self) -> None:
        session = _session_with_persona()
        _answer_all_questions(session)
        session.complete()
        assert session.status == DiscoveryStatus.COMPLETED

    def test_express_complete_emits_event(self) -> None:
        session = _session_with_persona()
        _answer_all_questions(session)
        session.complete()
        assert len(session.events) == 1
        assert isinstance(session.events[0], DiscoveryCompleted)

    def test_explicit_express_mode_complete_sets_completed(self) -> None:
        """Explicitly setting EXPRESS mode has same behavior as default."""
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.EXPRESS)
        session.detect_persona("1")
        _answer_all_questions(session)
        session.complete()
        assert session.status == DiscoveryStatus.COMPLETED
        assert len(session.events) == 1


# -- Deep mode complete() → ROUND_1_COMPLETE ----------------------------------


class TestDeepModeComplete:
    def test_deep_complete_sets_round_1_complete(self) -> None:
        session = _session_deep_answering()
        session.complete()
        assert session.status == DiscoveryStatus.ROUND_1_COMPLETE

    def test_deep_complete_does_not_emit_event(self) -> None:
        """Deep mode round 1 completion does NOT emit DiscoveryCompleted."""
        session = _session_deep_answering()
        session.complete()
        assert len(session.events) == 0


# -- Deep mode state transitions -----------------------------------------------


class TestDeepModeTransitions:
    def test_start_challenge_from_round_1_complete(self) -> None:
        session = _session_deep_answering()
        session.complete()  # -> ROUND_1_COMPLETE
        session.start_challenge()
        assert session.status == DiscoveryStatus.CHALLENGING

    def test_start_challenge_in_express_mode_raises(self) -> None:
        """Express mode cannot start a challenge round."""
        session = _session_with_persona()
        _answer_all_questions(session)
        session.complete()
        with pytest.raises(InvariantViolationError):
            session.start_challenge()

    def test_start_challenge_from_wrong_state_raises(self) -> None:
        session = _session_deep_answering()
        # Still in ANSWERING, not ROUND_1_COMPLETE
        with pytest.raises(InvariantViolationError):
            session.start_challenge()

    def test_complete_challenge_from_challenging(self) -> None:
        session = _session_deep_answering()
        session.complete()
        session.start_challenge()
        session.complete_challenge()
        assert session.status == DiscoveryStatus.ROUND_2_COMPLETE

    def test_complete_challenge_from_wrong_state_raises(self) -> None:
        session = _session_deep_answering()
        session.complete()  # ROUND_1_COMPLETE
        with pytest.raises(InvariantViolationError):
            session.complete_challenge()

    def test_start_simulate_from_round_2_complete(self) -> None:
        session = _session_deep_answering()
        session.complete()
        session.start_challenge()
        session.complete_challenge()
        session.start_simulate()
        assert session.status == DiscoveryStatus.SIMULATING

    def test_start_simulate_before_round_2_complete_raises(self) -> None:
        session = _session_deep_answering()
        session.complete()  # ROUND_1_COMPLETE
        with pytest.raises(InvariantViolationError):
            session.start_simulate()

    def test_start_simulate_in_express_mode_raises(self) -> None:
        """Express mode cannot simulate."""
        session = _session_with_persona()
        _answer_all_questions(session)
        session.complete()
        with pytest.raises(InvariantViolationError):
            session.start_simulate()

    def test_complete_simulation_from_simulating(self) -> None:
        session = _session_deep_answering()
        session.complete()
        session.start_challenge()
        session.complete_challenge()
        session.start_simulate()
        session.complete_simulation()
        assert session.status == DiscoveryStatus.COMPLETED

    def test_complete_simulation_from_wrong_state_raises(self) -> None:
        session = _session_deep_answering()
        session.complete()
        session.start_challenge()
        session.complete_challenge()
        # ROUND_2_COMPLETE, not SIMULATING
        with pytest.raises(InvariantViolationError):
            session.complete_simulation()

    def test_full_deep_flow_emits_event_at_end(self) -> None:
        """DiscoveryCompleted event emitted only after complete_simulation()."""
        session = _session_deep_answering()
        session.complete()
        assert len(session.events) == 0
        session.start_challenge()
        assert len(session.events) == 0
        session.complete_challenge()
        assert len(session.events) == 0
        session.start_simulate()
        assert len(session.events) == 0
        session.complete_simulation()
        assert len(session.events) == 1
        assert isinstance(session.events[0], DiscoveryCompleted)


# -- Snapshot round-trip with mode/round fields --------------------------------


class TestModeSnapshotRoundTrip:
    def test_snapshot_includes_mode_and_round(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.DEEP)
        snapshot = session.to_snapshot()
        assert "mode" in snapshot
        assert snapshot["mode"] == "deep"

    def test_snapshot_none_mode_when_not_set(self) -> None:
        session = _session_created()
        snapshot = session.to_snapshot()
        assert snapshot["mode"] is None
        assert snapshot["round"] is None

    def test_round_trip_deep_mode_preserves_mode(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.DEEP)
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.mode == DiscoveryMode.DEEP

    def test_round_trip_express_default_preserves_express(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.mode == DiscoveryMode.EXPRESS

    def test_round_trip_deep_round_1_complete(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_deep_answering()
        session.complete()
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.mode == DiscoveryMode.DEEP
        assert restored.status == DiscoveryStatus.ROUND_1_COMPLETE

    def test_old_snapshot_without_mode_defaults_to_express(self) -> None:
        """Backward compatibility: old snapshots missing 'mode' key."""
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        snapshot = session.to_snapshot()
        # Simulate old snapshot by removing mode/round keys
        snapshot.pop("mode", None)
        snapshot.pop("round", None)
        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.mode == DiscoveryMode.EXPRESS

    def test_snapshot_round_trip_through_full_deep_flow(self) -> None:
        """Serialize after each deep state transition, verify restoration."""
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_deep_answering()
        session.complete()
        session.start_challenge()

        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.status == DiscoveryStatus.CHALLENGING
        assert restored.mode == DiscoveryMode.DEEP


# -- Event emission edge cases ------------------------------------------------


class TestEmitCompletedEvent:
    def test_express_complete_event_has_all_fields(self) -> None:
        """DiscoveryCompleted event from express mode includes all data."""
        session = _session_with_persona()
        _answer_all_questions(session)
        session.complete()
        assert len(session.events) == 1
        event = session.events[0]
        assert event.session_id == session.session_id
        assert event.persona is not None
        assert event.register is not None
        assert len(event.answers) > 0

    def test_deep_complete_simulation_event_has_all_fields(self) -> None:
        """DiscoveryCompleted event from deep mode includes all data."""
        session = _session_deep_answering()
        session.complete()
        session.start_challenge()
        session.complete_challenge()
        session.start_simulate()
        session.complete_simulation()
        assert len(session.events) == 1
        event = session.events[0]
        assert event.session_id == session.session_id
        assert event.persona is not None
        assert event.register is not None
        assert len(event.answers) > 0

    def test_express_and_deep_events_have_same_shape(self) -> None:
        """Both modes produce events with the same structure."""
        # Express
        express = _session_with_persona()
        _answer_all_questions(express)
        express.complete()
        e_event = express.events[0]

        # Deep
        deep = _session_deep_answering()
        deep.complete()
        deep.start_challenge()
        deep.complete_challenge()
        deep.start_simulate()
        deep.complete_simulation()
        d_event = deep.events[0]

        # Same attributes present
        assert type(e_event) is type(d_event)
        assert hasattr(e_event, "session_id")
        assert hasattr(d_event, "session_id")


# -- Mode property edge cases -------------------------------------------------


class TestModeProperty:
    def test_mode_property_defaults_express_without_set(self) -> None:
        """Accessing mode before set_mode returns EXPRESS."""
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        assert session.mode == DiscoveryMode.EXPRESS

    def test_mode_property_after_explicit_deep(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.DEEP)
        assert session.mode == DiscoveryMode.DEEP

    def test_mode_survives_persona_detection(self) -> None:
        """Mode is preserved across state transitions."""
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.DEEP)
        session.detect_persona("1")
        assert session.mode == DiscoveryMode.DEEP

    def test_mode_survives_answering(self) -> None:
        from src.domain.models.discovery_values import DiscoveryMode

        session = _session_created()
        session.set_mode(DiscoveryMode.DEEP)
        session.detect_persona("1")
        session.answer_question("Q1", "Users")
        assert session.mode == DiscoveryMode.DEEP


# -- Double complete guard ----------------------------------------------------


class TestDoubleComplete:
    def test_express_cannot_complete_twice(self) -> None:
        """Once COMPLETED, calling complete() again fails."""
        session = _session_with_persona()
        _answer_all_questions(session)
        session.complete()
        with pytest.raises(InvariantViolationError):
            session.complete()

    def test_deep_round1_cannot_complete_twice(self) -> None:
        """Once ROUND_1_COMPLETE, calling complete() again fails."""
        session = _session_deep_answering()
        session.complete()
        with pytest.raises(InvariantViolationError):
            session.complete()
