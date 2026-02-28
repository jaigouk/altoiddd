"""Port for the Ticket Freshness bounded context (ticket health).

Defines the interface for reporting on ticket staleness and
ripple review status across the project backlog.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class TicketHealthPort(Protocol):
    """Interface for ticket health reporting.

    Adapters implement this to analyze ticket freshness, flag stale
    tickets, and report on ripple review status.
    """

    def report(self, project_dir: Path) -> str:
        """Generate a ticket health report.

        Args:
            project_dir: The project directory containing the ticket store.

        Returns:
            A health report covering staleness, review_needed flags,
            and dependency graph status.
        """
        ...
