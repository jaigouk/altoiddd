"""Port for the Ticket Pipeline bounded context (ticket generation).

Defines the interface for generating dependency-ordered beads tickets
from DDD artifacts with complexity-budget-driven detail levels.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class TicketGenerationPort(Protocol):
    """Interface for generating beads tickets from a domain model.

    Adapters implement this to produce dependency-ordered tickets using
    two-tier generation: full-detail for near-term work (Core subdomains)
    and stubs for far-term work (Generic subdomains).
    """

    def generate(self, domain_model: str, output_dir: Path) -> str:
        """Generate beads tickets from a domain model.

        Args:
            domain_model: Serialized domain model with bounded contexts,
                aggregates, and subdomain classification.
            output_dir: Directory where generated ticket files will be written.

        Returns:
            Summary of the generated tickets.
        """
        ...
