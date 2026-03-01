"""Port for the Ticket Pipeline bounded context (ticket generation).

Defines the interface for generating dependency-ordered beads tickets
from DDD artifacts with complexity-budget-driven detail levels.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.domain_model import DomainModel


@runtime_checkable
class TicketGenerationPort(Protocol):
    """Interface for generating beads tickets from a domain model.

    Adapters implement this to produce dependency-ordered tickets using
    two-tier generation: full-detail for near-term work (Core subdomains)
    and stubs for far-term work (Generic subdomains).

    Handlers using this port implement the preview-before-action pattern:
    build_preview() renders content, approve_and_write() commits it.
    """

    def generate(self, model: DomainModel, output_dir: Path) -> None:
        """Generate beads tickets from a domain model.

        Args:
            model: DomainModel with classified bounded contexts and aggregates.
            output_dir: Directory where generated ticket files will be written.
        """
        ...
