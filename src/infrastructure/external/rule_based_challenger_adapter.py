"""Rule-based challenger adapter — local fallback using ChallengerService.

Implements ChallengerPort by delegating to the stateless domain service.
No LLM required — pure heuristic challenge generation.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.services.challenger_service import ChallengerService

if TYPE_CHECKING:
    from src.domain.models.challenge import Challenge
    from src.domain.models.domain_model import DomainModel


class RuleBasedChallengerAdapter:
    """ChallengerPort adapter that uses rule-based heuristics (no LLM)."""

    async def generate_challenges(
        self,
        model: DomainModel,
        max_per_type: int = 5,
    ) -> tuple[Challenge, ...]:
        """Generate challenges using ChallengerService domain service.

        Args:
            model: The DomainModel to inspect.
            max_per_type: Maximum challenges per type.

        Returns:
            Tuple of Challenge VOs.
        """
        return ChallengerService.generate(model, max_per_type)
