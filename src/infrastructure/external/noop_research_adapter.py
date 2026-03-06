"""NoOp research adapter — returns an empty briefing.

Used as the default when no research infrastructure is configured.
Lists all bounded context names as no-data areas.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.research import ResearchBriefing

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel


class NoOpResearchAdapter:
    """Returns an empty ResearchBriefing with context names as no_data_areas."""

    async def research(
        self,
        model: DomainModel,
        max_areas: int = 5,
    ) -> ResearchBriefing:
        """Return empty briefing — all areas listed as no-data."""
        no_data = tuple(ctx.name for ctx in model.bounded_contexts)
        return ResearchBriefing(findings=(), no_data_areas=no_data, summary="")
