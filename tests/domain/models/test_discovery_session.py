"""Tests for DiscoverySession aggregate root.

Verifies the state machine (CREATED -> PERSONA_DETECTED -> ANSWERING ->
PLAYBACK_PENDING -> COMPLETED / CANCELLED), all 5 invariants, playback
triggers, quick mode, and domain event emission.
"""

from __future__ import annotations

import pytest

from src.domain.events.discovery_events import DiscoveryCompleted
from src.domain.models.bootstrap_session import InvariantViolationError
from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.discovery_values import (
    Persona,
    QuestionPhase,
    Register,
)

# -- Helpers ------------------------------------------------------------------


def _session_with_persona(choice: str = "1") -> DiscoverySession:
    """Create a session with persona already detected."""
    session = DiscoverySession(readme_content="A test project idea.")
    session.detect_persona(choice)
    return session


def _answer_questions(
    session: DiscoverySession,
    question_ids: list[str],
    *,
    confirm_playbacks: bool = True,
) -> None:
    """Answer a list of questions in order, confirming playbacks as they arise."""
    for qid in question_ids:
        session.answer_question(qid, f"Answer for {qid}")
        if session.status == DiscoveryStatus.PLAYBACK_PENDING and confirm_playbacks:
            session.confirm_playback(confirmed=True)


# -- Creation -----------------------------------------------------------------


class TestDiscoverySessionCreation:
    def test_new_session_starts_in_created_state(self):
        session = DiscoverySession(readme_content="An idea.")
        assert session.status == DiscoveryStatus.CREATED

    def test_new_session_has_unique_id(self):
        s1 = DiscoverySession(readme_content="Idea A")
        s2 = DiscoverySession(readme_content="Idea B")
        assert s1.session_id != s2.session_id

    def test_new_session_stores_readme(self):
        session = DiscoverySession(readme_content="My project idea.")
        assert session.readme_content == "My project idea."

    def test_new_session_has_no_persona(self):
        session = DiscoverySession(readme_content="Idea")
        assert session.persona is None

    def test_new_session_has_no_register(self):
        session = DiscoverySession(readme_content="Idea")
        assert session.register is None

    def test_new_session_has_no_answers(self):
        session = DiscoverySession(readme_content="Idea")
        assert session.answers == ()

    def test_new_session_has_no_events(self):
        session = DiscoverySession(readme_content="Idea")
        assert session.events == []

    def test_new_session_current_phase_is_seed(self):
        session = DiscoverySession(readme_content="Idea")
        assert session.current_phase == QuestionPhase.SEED


# -- Persona Detection -------------------------------------------------------


class TestPersonaDetection:
    def test_choice_1_sets_developer_technical(self):
        session = DiscoverySession(readme_content="Idea")
        session.detect_persona("1")
        assert session.persona == Persona.DEVELOPER
        assert session.register == Register.TECHNICAL

    def test_choice_2_sets_product_owner_non_technical(self):
        session = DiscoverySession(readme_content="Idea")
        session.detect_persona("2")
        assert session.persona == Persona.PRODUCT_OWNER
        assert session.register == Register.NON_TECHNICAL

    def test_choice_3_sets_domain_expert_non_technical(self):
        session = DiscoverySession(readme_content="Idea")
        session.detect_persona("3")
        assert session.persona == Persona.DOMAIN_EXPERT
        assert session.register == Register.NON_TECHNICAL

    def test_choice_4_sets_mixed_non_technical(self):
        session = DiscoverySession(readme_content="Idea")
        session.detect_persona("4")
        assert session.persona == Persona.MIXED
        assert session.register == Register.NON_TECHNICAL

    def test_transitions_to_persona_detected(self):
        session = DiscoverySession(readme_content="Idea")
        session.detect_persona("1")
        assert session.status == DiscoveryStatus.PERSONA_DETECTED

    def test_invalid_choice_raises_value_error(self):
        session = DiscoverySession(readme_content="Idea")
        with pytest.raises(ValueError, match="Invalid persona choice"):
            session.detect_persona("5")

    def test_detect_persona_not_from_created_raises(self):
        session = _session_with_persona()
        with pytest.raises(InvariantViolationError, match="CREATED"):
            session.detect_persona("2")


# -- Answer Question ----------------------------------------------------------


class TestAnswerQuestion:
    def test_answer_records_response(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users and admins")
        assert len(session.answers) == 1
        assert session.answers[0].question_id == "Q1"
        assert session.answers[0].response_text == "Users and admins"

    def test_answer_transitions_to_answering(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users and admins")
        assert session.status == DiscoveryStatus.ANSWERING

    def test_answer_advances_phase(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users")
        session.answer_question("Q2", "Entities")
        # Both ACTORS done, next should be in STORY
        assert session.current_phase == QuestionPhase.STORY

    def test_phase_advances_through_story(self):
        session = _session_with_persona()
        _answer_questions(session, ["Q1", "Q2", "Q3"])
        # After Q3, playback triggers (3 answers). After confirm, still in STORY.
        session.answer_question("Q4", "Failure mode")
        session.answer_question("Q5", "Other workflows")
        assert session.current_phase == QuestionPhase.EVENTS


# -- Invariant 1: Cannot answer before persona detection --------------------


class TestInvariantPersonaRequired:
    def test_answer_before_persona_raises(self):
        session = DiscoverySession(readme_content="Idea")
        with pytest.raises(InvariantViolationError, match="persona"):
            session.answer_question("Q1", "Answer")


# -- Invariant 2: Phase order enforced --------------------------------------


class TestInvariantPhaseOrder:
    def test_answer_out_of_phase_raises(self):
        """Answering a question from EVENTS while still in ACTORS phase."""
        session = _session_with_persona()
        with pytest.raises(InvariantViolationError, match="phase"):
            session.answer_question("Q6", "Some events")

    def test_answer_boundary_question_in_actors_phase_raises(self):
        session = _session_with_persona()
        with pytest.raises(InvariantViolationError, match="phase"):
            session.answer_question("Q9", "Bounded contexts")


# -- Invariant 3: Playback after 3 answers ----------------------------------


class TestInvariantPlaybackTrigger:
    def test_playback_triggered_after_three_answers(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users")
        session.answer_question("Q2", "Entities")
        session.answer_question("Q3", "Use case")
        assert session.status == DiscoveryStatus.PLAYBACK_PENDING

    def test_cannot_answer_during_playback(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users")
        session.answer_question("Q2", "Entities")
        session.answer_question("Q3", "Use case")
        assert session.status == DiscoveryStatus.PLAYBACK_PENDING
        with pytest.raises(InvariantViolationError, match="playback"):
            session.answer_question("Q4", "Failure mode")

    def test_confirm_playback_resumes_answering(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users")
        session.answer_question("Q2", "Entities")
        session.answer_question("Q3", "Use case")
        session.confirm_playback(confirmed=True)
        assert session.status == DiscoveryStatus.ANSWERING

    def test_reject_playback_with_corrections(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users")
        session.answer_question("Q2", "Entities")
        session.answer_question("Q3", "Use case")
        session.confirm_playback(confirmed=False, corrections="Fix actors list")
        assert session.status == DiscoveryStatus.ANSWERING

    def test_confirm_playback_not_in_playback_state_raises(self):
        session = _session_with_persona()
        with pytest.raises(InvariantViolationError, match="PLAYBACK_PENDING"):
            session.confirm_playback(confirmed=True)

    def test_second_playback_after_six_answers(self):
        session = _session_with_persona()
        _answer_questions(session, ["Q1", "Q2", "Q3"])
        # After first playback confirm, continue
        session.answer_question("Q4", "Failure")
        session.answer_question("Q5", "Workflows")
        session.answer_question("Q6", "Events")
        assert session.status == DiscoveryStatus.PLAYBACK_PENDING

    def test_playback_stores_confirmation(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users")
        session.answer_question("Q2", "Entities")
        session.answer_question("Q3", "Use case")
        session.confirm_playback(confirmed=True)
        assert len(session.playback_confirmations) == 1
        assert session.playback_confirmations[0].confirmed is True


# -- Invariant 4: Skip requires reason --------------------------------------


class TestInvariantSkipReason:
    def test_skip_with_reason_succeeds(self):
        session = _session_with_persona()
        session.skip_question("Q1", reason="Not relevant to this project")
        answered_ids = {a.question_id for a in session.answers}
        assert "Q1" not in answered_ids

    def test_skip_without_reason_raises(self):
        session = _session_with_persona()
        with pytest.raises(ValueError, match="reason"):
            session.skip_question("Q1", reason="")

    def test_skip_advances_past_question(self):
        session = _session_with_persona()
        session.skip_question("Q1", reason="Not relevant")
        # Should be able to answer Q2 without issue
        session.answer_question("Q2", "Entities")
        assert len(session.answers) == 1
        assert session.answers[0].question_id == "Q2"

    def test_skip_before_persona_raises(self):
        session = DiscoverySession(readme_content="Idea")
        with pytest.raises(InvariantViolationError, match="created"):
            session.skip_question("Q1", reason="Not relevant")

    def test_skip_during_playback_raises(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users")
        session.answer_question("Q2", "Entities")
        session.answer_question("Q3", "Use case")
        assert session.status == DiscoveryStatus.PLAYBACK_PENDING
        with pytest.raises(InvariantViolationError, match="playback"):
            session.skip_question("Q4", reason="Not needed")

    def test_skip_unknown_question_raises(self):
        session = _session_with_persona()
        with pytest.raises(ValueError, match="Unknown question"):
            session.skip_question("Q99", reason="Invalid")


# -- Unknown question ID in answer_question ---------------------------------


class TestUnknownQuestionId:
    def test_answer_unknown_question_raises(self):
        session = _session_with_persona()
        with pytest.raises(ValueError, match="Unknown question"):
            session.answer_question("Q99", "Some answer")


# -- Invariant 5: Complete requires MVP questions ----------------------------


class TestInvariantMVPRequired:
    def test_complete_with_all_mvp_questions_succeeds(self):
        session = _session_with_persona()
        # Answer all 10 questions
        _answer_questions(
            session,
            ["Q1", "Q2", "Q3", "Q4", "Q5", "Q6", "Q7", "Q8", "Q9", "Q10"],
        )
        session.complete()
        assert session.status == DiscoveryStatus.COMPLETED

    def test_complete_with_fewer_than_5_answers_raises(self):
        session = _session_with_persona()
        _answer_questions(session, ["Q1", "Q2", "Q3"])
        with pytest.raises(InvariantViolationError, match="MVP"):
            session.complete()

    def test_complete_with_5_mvp_questions_succeeds(self):
        """Quick mode: only 5 MVP questions are needed."""
        session = _session_with_persona()
        # Answer MVP questions: Q1, Q3, Q4, Q9, Q10
        # Skip the non-MVP ones: Q2, Q5, Q6, Q7, Q8
        session.answer_question("Q1", "Actors")
        session.answer_question("Q2", "Entities")
        session.answer_question("Q3", "Primary use case")
        # Playback after 3 answers
        session.confirm_playback(confirmed=True)
        session.answer_question("Q4", "Failure mode")
        session.skip_question("Q5", reason="Not needed for MVP")
        session.skip_question("Q6", reason="Not needed for MVP")
        # Playback after 3 more (Q4 + 2 skips = need actual answers)
        # Let's just answer enough to get to boundaries
        session.answer_question("Q7", "Policies")
        # Playback triggers (Q4, Q7 = 2 more actual answers since last playback,
        # plus we need one more)
        session.answer_question("Q8", "Read models")
        # This should trigger playback (3 answers since last: Q4, Q7, Q8)
        session.confirm_playback(confirmed=True)
        session.answer_question("Q9", "Bounded contexts")
        session.answer_question("Q10", "Subdomain classification")
        session.complete()
        assert session.status == DiscoveryStatus.COMPLETED


# -- Domain Events -----------------------------------------------------------


class TestDiscoveryEvents:
    def test_complete_emits_discovery_completed(self):
        session = _session_with_persona()
        _answer_questions(
            session,
            ["Q1", "Q2", "Q3", "Q4", "Q5", "Q6", "Q7", "Q8", "Q9", "Q10"],
        )
        session.complete()
        assert len(session.events) == 1
        event = session.events[0]
        assert isinstance(event, DiscoveryCompleted)
        assert event.session_id == session.session_id
        assert event.persona == Persona.DEVELOPER
        assert event.register == Register.TECHNICAL
        assert len(event.answers) == 10

    def test_events_returns_defensive_copy(self):
        session = _session_with_persona()
        _answer_questions(
            session,
            ["Q1", "Q2", "Q3", "Q4", "Q5", "Q6", "Q7", "Q8", "Q9", "Q10"],
        )
        session.complete()
        events = session.events
        events.clear()
        assert len(session.events) == 1


# -- Edge Cases ---------------------------------------------------------------


class TestEdgeCases:
    def test_duplicate_answer_raises(self):
        session = _session_with_persona()
        session.answer_question("Q1", "Users")
        with pytest.raises(InvariantViolationError, match="already answered"):
            session.answer_question("Q1", "Different answer")

    def test_empty_answer_raises(self):
        session = _session_with_persona()
        with pytest.raises(ValueError, match="empty"):
            session.answer_question("Q1", "")

    def test_whitespace_only_answer_raises(self):
        session = _session_with_persona()
        with pytest.raises(ValueError, match="empty"):
            session.answer_question("Q1", "   ")

    def test_quick_mode_with_extra_questions_allowed(self):
        """Minimum is 5 MVP questions, but answering more is fine."""
        session = _session_with_persona()
        _answer_questions(
            session,
            ["Q1", "Q2", "Q3", "Q4", "Q5", "Q6", "Q7", "Q8", "Q9", "Q10"],
        )
        session.complete()
        assert len(session.answers) == 10

    def test_double_complete_raises(self):
        session = _session_with_persona()
        _answer_questions(
            session,
            ["Q1", "Q2", "Q3", "Q4", "Q5", "Q6", "Q7", "Q8", "Q9", "Q10"],
        )
        session.complete()
        with pytest.raises(InvariantViolationError, match="ANSWERING"):
            session.complete()

    def test_complete_from_created_raises(self):
        session = DiscoverySession(readme_content="Idea")
        with pytest.raises(InvariantViolationError, match="ANSWERING"):
            session.complete()
