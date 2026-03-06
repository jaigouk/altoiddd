"""Port for the Domain Research bounded context.

Defines the interface for researching domain-specific knowledge
from external sources. Findings carry source attribution and
trust levels — the user confirms before facts enter the model.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.research import ResearchBriefing


@runtime_checkable
class DomainResearchPort(Protocol):
    """Interface for domain research in Round 2 discovery.

    Adapters implement this to search for domain-specific knowledge
    (competitive landscape, industry patterns, failure modes) and
    return findings with provenance metadata.
    """

    async def research(
        self,
        model: DomainModel,
        max_areas: int = 5,
    ) -> ResearchBriefing:
        """Research domain areas using external sources.

        Args:
            model: The DomainModel aggregate to inform search queries.
            max_areas: Maximum domain areas to research.

        Returns:
            ResearchBriefing with findings and no-data areas.
        """
        ...
