"""Adapter implementing TicketHealthPort via BeadsTicketReader.

Bridges the TicketHealthPort interface to the BeadsTicketReader ACL
and TicketHealthHandler query handler.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.ticket_freshness import TicketHealthReport


class BeadsTicketHealthAdapter:
    """TicketHealthPort implementation backed by the beads issue tracker.

    Constructs a BeadsTicketReader for the given project directory and
    delegates report generation to TicketHealthHandler.
    """

    def report(self, project_dir: Path) -> TicketHealthReport:
        """Generate a ticket health report for the project.

        Args:
            project_dir: The project directory containing .beads/.

        Returns:
            A TicketHealthReport covering staleness and review flags.
        """
        from src.application.queries.ticket_health_handler import TicketHealthHandler
        from src.infrastructure.external.beads_ticket_reader import BeadsTicketReader

        beads_dir = project_dir / ".beads"
        reader = BeadsTicketReader(beads_dir=beads_dir)
        handler = TicketHealthHandler(reader=reader)
        return handler.report()
