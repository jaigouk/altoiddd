"""Port for the AI Challenger bounded context.

Defines the interface for generating typed challenges that probe a
DomainModel for gaps: ambiguous language, missing invariants,
unexamined failure modes, and questionable boundaries.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.challenge import Challenge
    from src.domain.models.domain_model import DomainModel


@runtime_checkable
class ChallengerPort(Protocol):
    """Interface for challenge generation in Round 2 discovery.

    Adapters implement this to generate challenges from a DomainModel,
    either via rule-based heuristics (local) or LLM-powered analysis.
    """

    async def generate_challenges(
        self,
        model: DomainModel,
        max_per_type: int = 5,
    ) -> tuple[Challenge, ...]:
        """Generate typed challenges by analyzing the domain model.

        Args:
            model: The DomainModel aggregate to inspect.
            max_per_type: Maximum challenges per ChallengeType.

        Returns:
            Tuple of Challenge VOs.
        """
        ...
