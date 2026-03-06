"""Tests for Challenge domain value objects.

Verifies ChallengeType enum variants, Challenge/ChallengeResponse/ChallengeIteration
immutability, and validation rules.
"""

from __future__ import annotations

import dataclasses

import pytest

from src.domain.models.challenge import (
    Challenge,
    ChallengeIteration,
    ChallengeResponse,
    ChallengeType,
)
from src.domain.models.errors import InvariantViolationError


class TestChallengeTypeEnum:
    def test_challenge_type_enum_has_six_members(self) -> None:
        assert len(ChallengeType) == 6

    def test_language_value(self) -> None:
        assert ChallengeType.LANGUAGE.value == "language"

    def test_invariant_value(self) -> None:
        assert ChallengeType.INVARIANT.value == "invariant"

    def test_failure_mode_value(self) -> None:
        assert ChallengeType.FAILURE_MODE.value == "failure_mode"

    def test_boundary_value(self) -> None:
        assert ChallengeType.BOUNDARY.value == "boundary"

    def test_aggregate_value(self) -> None:
        assert ChallengeType.AGGREGATE.value == "aggregate"

    def test_communication_value(self) -> None:
        assert ChallengeType.COMMUNICATION.value == "communication"


class TestChallengeVO:
    def test_challenge_is_frozen(self) -> None:
        c = Challenge(
            challenge_type=ChallengeType.LANGUAGE,
            question_text="Is 'Order' the same in Sales and Shipping?",
            context_name="Sales",
            source_reference="UL term 'Order'",
        )
        with pytest.raises(dataclasses.FrozenInstanceError):
            c.question_text = "changed"  # type: ignore[misc]

    def test_challenge_requires_question_text(self) -> None:
        with pytest.raises(InvariantViolationError, match="question_text"):
            Challenge(
                challenge_type=ChallengeType.LANGUAGE,
                question_text="",
                context_name="Sales",
                source_reference="UL term",
            )

    def test_challenge_requires_context_name(self) -> None:
        with pytest.raises(InvariantViolationError, match="context_name"):
            Challenge(
                challenge_type=ChallengeType.LANGUAGE,
                question_text="Is this ambiguous?",
                context_name="",
                source_reference="UL term",
            )

    def test_challenge_requires_source_reference(self) -> None:
        with pytest.raises(InvariantViolationError, match="source_reference"):
            Challenge(
                challenge_type=ChallengeType.LANGUAGE,
                question_text="Is this ambiguous?",
                context_name="Sales",
                source_reference="",
            )

    def test_challenge_has_source_reference(self) -> None:
        c = Challenge(
            challenge_type=ChallengeType.BOUNDARY,
            question_text="Should Shipping own delivery tracking?",
            context_name="Shipping",
            source_reference="Context map: Sales→Shipping",
        )
        assert c.source_reference == "Context map: Sales→Shipping"

    def test_challenge_default_evidence_empty(self) -> None:
        c = Challenge(
            challenge_type=ChallengeType.INVARIANT,
            question_text="What prevents negative totals?",
            context_name="Sales",
            source_reference="OrderAggregate",
        )
        assert c.evidence == ""

    def test_challenge_with_evidence(self) -> None:
        c = Challenge(
            challenge_type=ChallengeType.FAILURE_MODE,
            question_text="What happens if payment fails mid-checkout?",
            context_name="Sales",
            source_reference="Checkout Flow story",
            evidence="Step 2 says 'System validates payment'",
        )
        assert c.evidence == "Step 2 says 'System validates payment'"


class TestChallengeResponseVO:
    def test_challenge_response_is_frozen(self) -> None:
        r = ChallengeResponse(
            challenge_id="abc-123",
            user_response="Yes, we need separate definitions",
            accepted=True,
        )
        with pytest.raises(dataclasses.FrozenInstanceError):
            r.accepted = False  # type: ignore[misc]

    def test_challenge_response_captures_acceptance(self) -> None:
        r = ChallengeResponse(
            challenge_id="abc-123",
            user_response="Good point",
            accepted=True,
        )
        assert r.accepted is True

    def test_challenge_response_captures_rejection(self) -> None:
        r = ChallengeResponse(
            challenge_id="abc-123",
            user_response="Not relevant",
            accepted=False,
        )
        assert r.accepted is False

    def test_challenge_response_default_artifact_updates_empty(self) -> None:
        r = ChallengeResponse(
            challenge_id="abc-123",
            user_response="Yes",
            accepted=True,
        )
        assert r.artifact_updates == ()

    def test_challenge_response_with_artifact_updates(self) -> None:
        r = ChallengeResponse(
            challenge_id="abc-123",
            user_response="Yes, add invariant",
            accepted=True,
            artifact_updates=("Add invariant: Total must be positive",),
        )
        assert len(r.artifact_updates) == 1


class TestChallengeIterationVO:
    def test_challenge_iteration_is_frozen(self) -> None:
        it = ChallengeIteration(
            challenges=(),
            responses=(),
            convergence_delta=0,
        )
        with pytest.raises(dataclasses.FrozenInstanceError):
            it.convergence_delta = 5  # type: ignore[misc]

    def test_challenge_iteration_convergence_delta(self) -> None:
        it = ChallengeIteration(
            challenges=(),
            responses=(),
            convergence_delta=3,
        )
        assert it.convergence_delta == 3

    def test_challenge_iteration_holds_challenges_and_responses(self) -> None:
        challenge = Challenge(
            challenge_type=ChallengeType.LANGUAGE,
            question_text="Is 'Order' ambiguous?",
            context_name="Sales",
            source_reference="UL glossary",
        )
        response = ChallengeResponse(
            challenge_id="r1",
            user_response="Yes",
            accepted=True,
        )
        it = ChallengeIteration(
            challenges=(challenge,),
            responses=(response,),
            convergence_delta=1,
        )
        assert len(it.challenges) == 1
        assert len(it.responses) == 1
