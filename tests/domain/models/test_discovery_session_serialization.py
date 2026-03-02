"""Tests for DiscoverySession snapshot serialization (alty-rhm).

RED phase: these tests define the contract for to_snapshot() / from_snapshot()
round-trip serialization on the DiscoverySession aggregate root.
"""

from __future__ import annotations

import json

import pytest

from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.discovery_values import (
    Answer,
    Persona,
    Register,
)
from src.domain.models.errors import InvariantViolationError
from src.domain.models.question import Question

# ── Helpers ──────────────────────────────────────────────────────────


def _build_session_created() -> DiscoverySession:
    """Build a session in CREATED state."""
    return DiscoverySession(readme_content="# My Project\nA cool tool.")


def _build_session_persona_detected() -> DiscoverySession:
    """Build a session in PERSONA_DETECTED state."""
    session = _build_session_created()
    session.detect_persona("1")  # DEVELOPER / TECHNICAL
    return session


def _build_session_answering(answer_count: int = 1) -> DiscoverySession:
    """Build a session in ANSWERING state with `answer_count` answers.

    Answers questions in order: Q1, Q2, Q3, ... up to answer_count.
    Confirms playback automatically when triggered (every 3 answers).
    """
    session = _build_session_persona_detected()
    questions = list(Question.CATALOG)
    for i in range(answer_count):
        if session.status == DiscoveryStatus.PLAYBACK_PENDING:
            session.confirm_playback(confirmed=True)
        session.answer_question(questions[i].id, f"Answer for {questions[i].id}")
    return session


def _build_session_playback_pending() -> DiscoverySession:
    """Build a session in PLAYBACK_PENDING state (3 answers, no confirm)."""
    session = _build_session_persona_detected()
    questions = list(Question.CATALOG)
    for i in range(3):
        session.answer_question(questions[i].id, f"Answer for {questions[i].id}")
    assert session.status == DiscoveryStatus.PLAYBACK_PENDING
    return session


def _build_session_completed() -> DiscoverySession:
    """Build a fully completed session with all 10 answers."""
    session = _build_session_answering(answer_count=10)
    if session.status == DiscoveryStatus.PLAYBACK_PENDING:
        session.confirm_playback(confirmed=True)
    session.complete()
    return session


# ── Round-trip tests: all reachable states ───────────────────────────


class TestSnapshotRoundTripStates:
    """Each reachable DiscoveryStatus can be serialized and restored."""

    def test_round_trip_created_state(self) -> None:
        session = _build_session_created()
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)

        assert restored.session_id == session.session_id
        assert restored.readme_content == session.readme_content
        assert restored.status == DiscoveryStatus.CREATED
        assert restored.persona is None
        assert restored.register is None
        assert restored.answers == ()
        assert restored.playback_confirmations == ()

    def test_round_trip_persona_detected_state(self) -> None:
        session = _build_session_persona_detected()
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)

        assert restored.status == DiscoveryStatus.PERSONA_DETECTED
        assert restored.persona == Persona.DEVELOPER
        assert restored.register == Register.TECHNICAL

    def test_round_trip_answering_state(self) -> None:
        session = _build_session_answering(answer_count=2)
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)

        assert restored.status == DiscoveryStatus.ANSWERING
        assert len(restored.answers) == 2
        assert restored.answers[0].question_id == "Q1"
        assert restored.answers[1].question_id == "Q2"

    def test_round_trip_playback_pending_state(self) -> None:
        session = _build_session_playback_pending()
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)

        assert restored.status == DiscoveryStatus.PLAYBACK_PENDING

    def test_round_trip_completed_state(self) -> None:
        session = _build_session_completed()
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)

        assert restored.status == DiscoveryStatus.COMPLETED
        assert len(restored.answers) == 10


# ── Field preservation ───────────────────────────────────────────────


class TestSnapshotFieldPreservation:
    """All internal state fields survive the round-trip."""

    def test_snapshot_contains_all_required_fields(self) -> None:
        session = _build_session_answering(answer_count=2)
        snapshot = session.to_snapshot()

        required_keys = {
            "session_id",
            "readme_content",
            "status",
            "persona",
            "register",
            "answers",
            "skipped",
            "playback_confirmations",
            "answers_since_last_playback",
        }
        assert required_keys <= set(snapshot.keys())

    def test_preserves_session_id(self) -> None:
        session = _build_session_created()
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.session_id == session.session_id

    def test_preserves_readme_content(self) -> None:
        session = DiscoverySession(readme_content="# Long README\n\nWith **markdown** content.")
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.readme_content == "# Long README\n\nWith **markdown** content."

    def test_preserves_answers_with_content(self) -> None:
        session = _build_session_answering(answer_count=2)
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)

        assert restored.answers[0] == Answer(question_id="Q1", response_text="Answer for Q1")
        assert restored.answers[1] == Answer(question_id="Q2", response_text="Answer for Q2")

    def test_preserves_playback_confirmations(self) -> None:
        session = _build_session_playback_pending()
        session.confirm_playback(confirmed=True, corrections="Minor fix")
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)

        assert len(restored.playback_confirmations) == 1
        pb = restored.playback_confirmations[0]
        assert pb.confirmed is True
        assert pb.corrections == "Minor fix"

    def test_preserves_playback_counter(self) -> None:
        """_answers_since_last_playback survives the round-trip."""
        session = _build_session_answering(answer_count=2)
        snapshot = session.to_snapshot()
        assert snapshot["answers_since_last_playback"] == 2

        restored = DiscoverySession.from_snapshot(snapshot)
        # After restoring with counter=2, one more answer should trigger playback
        restored.answer_question("Q3", "Third answer")
        assert restored.status == DiscoveryStatus.PLAYBACK_PENDING

    def test_preserves_skipped_questions(self) -> None:
        session = _build_session_persona_detected()
        session.skip_question("Q1", "Not relevant")
        session.skip_question("Q2", "Ditto")
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)

        # Restored session should allow answering Q3 (STORY phase) since
        # ACTORS phase Q1/Q2 were skipped
        restored.answer_question("Q3", "Story answer")
        assert len(restored.answers) == 1

    def test_preserves_persona_product_owner(self) -> None:
        session = _build_session_created()
        session.detect_persona("2")  # PRODUCT_OWNER / NON_TECHNICAL
        snapshot = session.to_snapshot()
        restored = DiscoverySession.from_snapshot(snapshot)

        assert restored.persona == Persona.PRODUCT_OWNER
        assert restored.register == Register.NON_TECHNICAL


# ── JSON serialization ───────────────────────────────────────────────


class TestSnapshotJsonSerializable:
    """Snapshot dict is fully JSON-serializable."""

    def test_snapshot_is_json_serializable(self) -> None:
        session = _build_session_completed()
        snapshot = session.to_snapshot()
        json_str = json.dumps(snapshot)
        assert isinstance(json_str, str)

    def test_round_trip_through_json(self) -> None:
        """Serialize to JSON string and back — full fidelity."""
        session = _build_session_answering(answer_count=5)
        if session.status == DiscoveryStatus.PLAYBACK_PENDING:
            session.confirm_playback(confirmed=True)
        snapshot = session.to_snapshot()
        json_str = json.dumps(snapshot)
        loaded = json.loads(json_str)
        restored = DiscoverySession.from_snapshot(loaded)

        assert restored.session_id == session.session_id
        assert restored.status == session.status
        assert restored.answers == session.answers
        assert restored.playback_confirmations == session.playback_confirmations


# ── Restored session enforces invariants ─────────────────────────────


class TestRestoredSessionInvariants:
    """A restored session continues to enforce all aggregate invariants."""

    def test_restored_answering_can_continue_answering(self) -> None:
        session = _build_session_answering(answer_count=2)
        restored = DiscoverySession.from_snapshot(session.to_snapshot())
        restored.answer_question("Q3", "Continued answer")
        assert len(restored.answers) == 3

    def test_restored_answering_rejects_duplicate_answer(self) -> None:
        session = _build_session_answering(answer_count=2)
        restored = DiscoverySession.from_snapshot(session.to_snapshot())
        with pytest.raises(Exception, match="already answered"):
            restored.answer_question("Q1", "Duplicate")

    def test_restored_playback_pending_blocks_answers(self) -> None:
        session = _build_session_playback_pending()
        restored = DiscoverySession.from_snapshot(session.to_snapshot())
        with pytest.raises(Exception, match="playback"):
            restored.answer_question("Q4", "Not allowed")

    def test_restored_playback_pending_can_confirm(self) -> None:
        session = _build_session_playback_pending()
        restored = DiscoverySession.from_snapshot(session.to_snapshot())
        restored.confirm_playback(confirmed=True)
        assert restored.status == DiscoveryStatus.ANSWERING

    def test_restored_completed_rejects_answers(self) -> None:
        session = _build_session_completed()
        restored = DiscoverySession.from_snapshot(session.to_snapshot())
        with pytest.raises(InvariantViolationError, match="Cannot answer"):
            restored.answer_question("Q1", "Too late")

    def test_restored_session_can_complete(self) -> None:
        """An ANSWERING session with enough MVP answers can complete after restore."""
        session = _build_session_answering(answer_count=10)
        if session.status == DiscoveryStatus.PLAYBACK_PENDING:
            session.confirm_playback(confirmed=True)
        restored = DiscoverySession.from_snapshot(session.to_snapshot())
        restored.complete()
        assert restored.status == DiscoveryStatus.COMPLETED

    def test_restored_session_enforces_phase_order(self) -> None:
        """Cannot skip ahead to BOUNDARIES phase from ACTORS."""
        session = _build_session_persona_detected()
        restored = DiscoverySession.from_snapshot(session.to_snapshot())
        with pytest.raises(Exception, match="phase"):
            restored.answer_question("Q9", "Jumping ahead")


# ── Edge cases: invalid snapshots ────────────────────────────────────


class TestFromSnapshotEdgeCases:
    """Edge cases and error handling for from_snapshot()."""

    def test_missing_required_field_raises(self) -> None:
        session = _build_session_created()
        snapshot = session.to_snapshot()
        del snapshot["session_id"]
        with pytest.raises((ValueError, KeyError)):
            DiscoverySession.from_snapshot(snapshot)

    def test_invalid_status_raises(self) -> None:
        session = _build_session_created()
        snapshot = session.to_snapshot()
        snapshot["status"] = "bogus_state"
        with pytest.raises((ValueError, KeyError)):
            DiscoverySession.from_snapshot(snapshot)

    def test_invalid_persona_raises(self) -> None:
        session = _build_session_persona_detected()
        snapshot = session.to_snapshot()
        snapshot["persona"] = "alien"
        with pytest.raises((ValueError, KeyError)):
            DiscoverySession.from_snapshot(snapshot)

    def test_invalid_register_raises(self) -> None:
        session = _build_session_persona_detected()
        snapshot = session.to_snapshot()
        snapshot["register"] = "casual"
        with pytest.raises((ValueError, KeyError)):
            DiscoverySession.from_snapshot(snapshot)

    def test_extra_fields_ignored(self) -> None:
        session = _build_session_created()
        snapshot = session.to_snapshot()
        snapshot["unknown_field"] = "should be ignored"
        snapshot["another_extra"] = 42
        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.session_id == session.session_id

    def test_corrupted_answers_raises(self) -> None:
        session = _build_session_answering(answer_count=2)
        snapshot = session.to_snapshot()
        snapshot["answers"] = "not a list"
        with pytest.raises((ValueError, TypeError, AttributeError)):
            DiscoverySession.from_snapshot(snapshot)

    def test_empty_dict_raises(self) -> None:
        with pytest.raises((ValueError, KeyError)):
            DiscoverySession.from_snapshot({})

    def test_none_persona_in_created_state_ok(self) -> None:
        """CREATED state has None persona — should serialize and restore fine."""
        session = _build_session_created()
        snapshot = session.to_snapshot()
        assert snapshot["persona"] is None
        restored = DiscoverySession.from_snapshot(snapshot)
        assert restored.persona is None

    def test_negative_playback_counter_raises(self) -> None:
        session = _build_session_answering(answer_count=1)
        snapshot = session.to_snapshot()
        snapshot["answers_since_last_playback"] = -1
        with pytest.raises((ValueError, TypeError)):
            DiscoverySession.from_snapshot(snapshot)

    def test_playback_counter_exceeding_interval_raises(self) -> None:
        """Counter should never exceed the playback interval (3)."""
        session = _build_session_answering(answer_count=1)
        snapshot = session.to_snapshot()
        snapshot["answers_since_last_playback"] = 99
        with pytest.raises((ValueError, TypeError)):
            DiscoverySession.from_snapshot(snapshot)

    def test_corrupted_skipped_raises(self) -> None:
        session = _build_session_created()
        snapshot = session.to_snapshot()
        snapshot["skipped"] = "not a list"
        with pytest.raises(ValueError, match="skipped must be a list"):
            DiscoverySession.from_snapshot(snapshot)

    def test_corrupted_playback_confirmations_raises(self) -> None:
        session = _build_session_created()
        snapshot = session.to_snapshot()
        snapshot["playback_confirmations"] = "not a list"
        with pytest.raises(ValueError, match="playback_confirmations must be a list"):
            DiscoverySession.from_snapshot(snapshot)


# ── Cross-validation: status vs persona/counter consistency ──────────


class TestFromSnapshotCrossValidation:
    """Status must be consistent with persona and playback counter."""

    def test_created_with_persona_set_raises(self) -> None:
        """CREATED state must have persona=None."""
        session = _build_session_created()
        snapshot = session.to_snapshot()
        snapshot["persona"] = "developer"
        snapshot["register"] = "technical"
        with pytest.raises(ValueError, match="CREATED state must have persona=None"):
            DiscoverySession.from_snapshot(snapshot)

    def test_answering_with_persona_none_raises(self) -> None:
        """ANSWERING state requires a persona."""
        session = _build_session_answering(answer_count=1)
        snapshot = session.to_snapshot()
        snapshot["persona"] = None
        snapshot["register"] = None
        with pytest.raises(ValueError, match="requires a persona"):
            DiscoverySession.from_snapshot(snapshot)

    def test_playback_pending_counter_mismatch_raises(self) -> None:
        """PLAYBACK_PENDING state requires counter == 3."""
        session = _build_session_playback_pending()
        snapshot = session.to_snapshot()
        snapshot["answers_since_last_playback"] = 1
        with pytest.raises(ValueError, match="PLAYBACK_PENDING state requires counter=3"):
            DiscoverySession.from_snapshot(snapshot)

    def test_answering_counter_at_interval_raises(self) -> None:
        """ANSWERING state requires counter < 3."""
        session = _build_session_answering(answer_count=1)
        snapshot = session.to_snapshot()
        snapshot["answers_since_last_playback"] = 3
        with pytest.raises(ValueError, match="ANSWERING state requires counter < 3"):
            DiscoverySession.from_snapshot(snapshot)

    # Note: CANCELLED is defined in DiscoveryStatus but currently unreachable —
    # no cancel() command exists on the aggregate. Cross-validation and
    # round-trip tests for CANCELLED are deferred until a cancel flow is added.
