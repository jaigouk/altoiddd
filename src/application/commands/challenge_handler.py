"""ChallengeHandler — orchestrates the Round 2 challenge lifecycle.

Thin orchestrator that delegates challenge generation to a ChallengerPort,
collects user responses, and produces a ChallengeIteration summary.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.challenge import ChallengeIteration, ChallengeResponse

if TYPE_CHECKING:
    from src.application.ports.challenger_port import ChallengerPort
    from src.domain.models.challenge import Challenge
    from src.domain.models.domain_model import DomainModel


class ChallengeHandler:
    """Orchestrates generate → respond → complete cycle for Round 2."""

    def __init__(self, challenger: ChallengerPort) -> None:
        self._challenger = challenger
        self._challenges: tuple[Challenge, ...] = ()
        self._responses: list[ChallengeResponse] = []

    async def generate_challenges(
        self,
        model: DomainModel,
        max_per_type: int = 5,
    ) -> tuple[Challenge, ...]:
        """Delegate challenge generation to the port.

        Args:
            model: The DomainModel to analyze.
            max_per_type: Maximum challenges per type.

        Returns:
            Tuple of generated Challenge VOs.
        """
        self._challenges = await self._challenger.generate_challenges(
            model, max_per_type
        )
        return self._challenges

    def record_response(self, response: ChallengeResponse) -> None:
        """Record a user response to a challenge.

        Args:
            response: The user's response to a specific challenge.
        """
        self._responses.append(response)

    def complete(self) -> ChallengeIteration:
        """Finalize the challenge round and produce a summary.

        Returns:
            ChallengeIteration with all challenges, responses, and
            convergence_delta (count of artifact updates from accepted responses).
        """
        delta = sum(
            len(r.artifact_updates)
            for r in self._responses
            if r.accepted
        )
        return ChallengeIteration(
            challenges=self._challenges,
            responses=tuple(self._responses),
            convergence_delta=delta,
        )
