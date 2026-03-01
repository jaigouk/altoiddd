"""Port for the Domain Model bounded context (artifact generation).

Defines the interface for rendering DDD artifacts (PRD, DDD.md,
ARCHITECTURE.md) from a finalized DomainModel aggregate.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel


@runtime_checkable
class ArtifactRendererPort(Protocol):
    """Interface for rendering a DomainModel into markdown documents.

    Adapters implement this to produce PRD, DDD.md, and ARCHITECTURE.md
    strings from a finalized DomainModel.
    """

    def render_prd(self, model: DomainModel) -> str:
        """Render the PRD markdown from a domain model.

        Args:
            model: A finalized DomainModel aggregate.

        Returns:
            PRD markdown string matching PRD_TEMPLATE.md structure.
        """
        ...

    def render_ddd(self, model: DomainModel) -> str:
        """Render the DDD.md markdown from a domain model.

        Args:
            model: A finalized DomainModel aggregate.

        Returns:
            DDD markdown string matching DDD_STORY_TEMPLATE.md structure.
        """
        ...

    def render_architecture(self, model: DomainModel) -> str:
        """Render the ARCHITECTURE.md markdown from a domain model.

        Args:
            model: A finalized DomainModel aggregate.

        Returns:
            Architecture markdown string matching ARCHITECTURE_TEMPLATE.md.
        """
        ...
