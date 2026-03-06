"""Tests for ChallengeHandler — orchestrates the challenge round lifecycle.

Verifies generate → record response → complete cycle, convergence delta
calculation, and delegation to ChallengerPort.
"""

from __future__ import annotations

from unittest.mock import AsyncMock

import pytest

from src.application.commands.challenge_handler import ChallengeHandler
from src.domain.models.challenge import (
    Challenge,
    ChallengeIteration,
    ChallengeResponse,
    ChallengeType,
)
from src.domain.models.domain_model import DomainModel


def _make_challenge(q: str = "Is this right?", ctx: str = "Sales") -> Challenge:
    return Challenge(
        challenge_type=ChallengeType.LANGUAGE,
        question_text=q,
        context_name=ctx,
        source_reference="test reference",
    )


class TestChallengeHandlerGenerate:
    @pytest.mark.asyncio
    async def test_generate_challenges_delegates_to_port(self) -> None:
        expected = (_make_challenge(),)
        port = AsyncMock()
        port.generate_challenges.return_value = expected
        handler = ChallengeHandler(challenger=port)
        model = DomainModel()

        result = await handler.generate_challenges(model)

        port.generate_challenges.assert_awaited_once_with(model, 5)
        assert result == expected

    @pytest.mark.asyncio
    async def test_generate_challenges_custom_max(self) -> None:
        port = AsyncMock()
        port.generate_challenges.return_value = ()
        handler = ChallengeHandler(challenger=port)
        model = DomainModel()

        await handler.generate_challenges(model, max_per_type=3)

        port.generate_challenges.assert_awaited_once_with(model, 3)

    @pytest.mark.asyncio
    async def test_generate_stores_challenges_internally(self) -> None:
        challenges = (_make_challenge("Q1"), _make_challenge("Q2"))
        port = AsyncMock()
        port.generate_challenges.return_value = challenges
        handler = ChallengeHandler(challenger=port)

        await handler.generate_challenges(DomainModel())
        iteration = handler.complete()

        assert iteration.challenges == challenges


class TestChallengeHandlerRecordResponse:
    @pytest.mark.asyncio
    async def test_record_response_stores_response(self) -> None:
        port = AsyncMock()
        port.generate_challenges.return_value = (_make_challenge(),)
        handler = ChallengeHandler(challenger=port)
        await handler.generate_challenges(DomainModel())

        response = ChallengeResponse(
            challenge_id="c1",
            user_response="Good point",
            accepted=True,
        )
        handler.record_response(response)
        iteration = handler.complete()

        assert len(iteration.responses) == 1
        assert iteration.responses[0].accepted is True

    @pytest.mark.asyncio
    async def test_multiple_responses_recorded(self) -> None:
        port = AsyncMock()
        port.generate_challenges.return_value = (_make_challenge(),)
        handler = ChallengeHandler(challenger=port)
        await handler.generate_challenges(DomainModel())

        handler.record_response(
            ChallengeResponse(challenge_id="c1", user_response="Yes", accepted=True)
        )
        handler.record_response(
            ChallengeResponse(challenge_id="c2", user_response="No", accepted=False)
        )
        iteration = handler.complete()

        assert len(iteration.responses) == 2


class TestChallengeHandlerComplete:
    @pytest.mark.asyncio
    async def test_complete_returns_challenge_iteration(self) -> None:
        port = AsyncMock()
        port.generate_challenges.return_value = (_make_challenge(),)
        handler = ChallengeHandler(challenger=port)
        await handler.generate_challenges(DomainModel())

        iteration = handler.complete()

        assert isinstance(iteration, ChallengeIteration)

    @pytest.mark.asyncio
    async def test_convergence_delta_counts_accepted_updates(self) -> None:
        port = AsyncMock()
        port.generate_challenges.return_value = (_make_challenge(),)
        handler = ChallengeHandler(challenger=port)
        await handler.generate_challenges(DomainModel())

        handler.record_response(
            ChallengeResponse(
                challenge_id="c1",
                user_response="Yes",
                accepted=True,
                artifact_updates=("Add invariant", "Add term"),
            )
        )
        handler.record_response(
            ChallengeResponse(
                challenge_id="c2",
                user_response="No",
                accepted=False,
            )
        )
        handler.record_response(
            ChallengeResponse(
                challenge_id="c3",
                user_response="Yes",
                accepted=True,
                artifact_updates=("Add story",),
            )
        )

        iteration = handler.complete()

        # 2 updates from c1 + 0 from c2 + 1 from c3 = 3
        assert iteration.convergence_delta == 3

    @pytest.mark.asyncio
    async def test_convergence_delta_zero_when_all_rejected(self) -> None:
        port = AsyncMock()
        port.generate_challenges.return_value = (_make_challenge(),)
        handler = ChallengeHandler(challenger=port)
        await handler.generate_challenges(DomainModel())

        handler.record_response(
            ChallengeResponse(
                challenge_id="c1", user_response="No", accepted=False
            )
        )

        iteration = handler.complete()
        assert iteration.convergence_delta == 0

    @pytest.mark.asyncio
    async def test_complete_with_no_responses(self) -> None:
        port = AsyncMock()
        port.generate_challenges.return_value = (_make_challenge(),)
        handler = ChallengeHandler(challenger=port)
        await handler.generate_challenges(DomainModel())

        iteration = handler.complete()

        assert iteration.responses == ()
        assert iteration.convergence_delta == 0
