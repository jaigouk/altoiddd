"""Tests for discovery value objects.

Verifies Persona, Register, QuestionPhase enums and Answer/Playback
frozen dataclass immutability.
"""

from __future__ import annotations

import dataclasses

import pytest

from src.domain.models.discovery_values import (
    Answer,
    Persona,
    Playback,
    QuestionPhase,
    Register,
)

# -- Persona enum -----------------------------------------------------------


class TestPersona:
    def test_has_four_members(self):
        assert len(Persona) == 4

    def test_developer_value(self):
        assert Persona.DEVELOPER.value == "developer"

    def test_product_owner_value(self):
        assert Persona.PRODUCT_OWNER.value == "product_owner"

    def test_domain_expert_value(self):
        assert Persona.DOMAIN_EXPERT.value == "domain_expert"

    def test_mixed_value(self):
        assert Persona.MIXED.value == "mixed"


# -- Register enum -----------------------------------------------------------


class TestRegister:
    def test_has_two_members(self):
        assert len(Register) == 2

    def test_technical_value(self):
        assert Register.TECHNICAL.value == "technical"

    def test_non_technical_value(self):
        assert Register.NON_TECHNICAL.value == "non_technical"


# -- QuestionPhase enum -----------------------------------------------------


class TestQuestionPhase:
    def test_has_five_phases(self):
        assert len(QuestionPhase) == 5

    def test_phase_values(self):
        assert QuestionPhase.SEED.value == "seed"
        assert QuestionPhase.ACTORS.value == "actors"
        assert QuestionPhase.STORY.value == "story"
        assert QuestionPhase.EVENTS.value == "events"
        assert QuestionPhase.BOUNDARIES.value == "boundaries"

    def test_phase_ordering(self):
        """Phases are ordered by their position in the discovery flow."""
        phases = list(QuestionPhase)
        assert phases == [
            QuestionPhase.SEED,
            QuestionPhase.ACTORS,
            QuestionPhase.STORY,
            QuestionPhase.EVENTS,
            QuestionPhase.BOUNDARIES,
        ]


# -- Answer value object -----------------------------------------------------


class TestAnswer:
    def test_answer_creation(self):
        answer = Answer(question_id="Q1", response_text="Users and admins")
        assert answer.question_id == "Q1"
        assert answer.response_text == "Users and admins"

    def test_answer_is_frozen(self):
        answer = Answer(question_id="Q1", response_text="Users and admins")
        with pytest.raises(dataclasses.FrozenInstanceError):
            answer.question_id = "Q2"  # type: ignore[misc]

    def test_answer_equality(self):
        a1 = Answer(question_id="Q1", response_text="Users and admins")
        a2 = Answer(question_id="Q1", response_text="Users and admins")
        assert a1 == a2

    def test_answer_inequality(self):
        a1 = Answer(question_id="Q1", response_text="Users and admins")
        a2 = Answer(question_id="Q1", response_text="Just admins")
        assert a1 != a2


# -- Playback value object ---------------------------------------------------


class TestPlayback:
    def test_playback_creation_defaults(self):
        pb = Playback(summary_text="Summary here")
        assert pb.summary_text == "Summary here"
        assert pb.confirmed is False
        assert pb.corrections == ""

    def test_playback_creation_with_all_fields(self):
        pb = Playback(summary_text="Sum", confirmed=True, corrections="Fix actors")
        assert pb.confirmed is True
        assert pb.corrections == "Fix actors"

    def test_playback_is_frozen(self):
        pb = Playback(summary_text="Sum")
        with pytest.raises(dataclasses.FrozenInstanceError):
            pb.confirmed = True  # type: ignore[misc]

    def test_playback_equality(self):
        p1 = Playback(summary_text="Sum", confirmed=True, corrections="")
        p2 = Playback(summary_text="Sum", confirmed=True, corrections="")
        assert p1 == p2
