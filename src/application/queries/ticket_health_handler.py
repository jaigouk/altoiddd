"""Application query handler for ticket health reporting.

TicketHealthHandler reads open tickets via a TicketReaderProtocol
and builds a TicketHealthReport covering freshness flags, staleness,
and review status.
"""

from __future__ import annotations

from typing import Protocol, runtime_checkable

from src.domain.models.ticket_freshness import (
    FlaggedTicket,
    FreshnessFlag,
    OpenTicketData,
    TicketFreshnessStatus,
    TicketHealthReport,
)


@runtime_checkable
class TicketReaderProtocol(Protocol):
    """Port interface for reading ticket data from the issue tracker."""

    def read_open_tickets(self) -> tuple[OpenTicketData, ...]: ...

    def read_flags(self, ticket_id: str) -> tuple[FreshnessFlag, ...]: ...


class TicketHealthHandler:
    """Query handler that builds a TicketHealthReport from the backlog.

    Reads all open tickets, identifies those needing review, gathers
    their freshness flags, and computes the oldest last-reviewed date.
    """

    def __init__(self, reader: TicketReaderProtocol) -> None:
        self._reader = reader

    def report(self) -> TicketHealthReport:
        """Build a ticket health report.

        Returns:
            TicketHealthReport with flagged tickets, totals, and staleness info.
        """
        open_tickets = self._reader.read_open_tickets()

        flagged: list[FlaggedTicket] = []
        for ticket in open_tickets:
            if "review_needed" in ticket.labels:
                flags = self._reader.read_flags(ticket.ticket_id)
                flagged.append(
                    FlaggedTicket(
                        ticket_id=ticket.ticket_id,
                        title=ticket.title,
                        flags=flags,
                        status=TicketFreshnessStatus.REVIEW_NEEDED,
                    )
                )

        # Find the oldest last_reviewed across ALL open tickets
        reviewed_dates = [t.last_reviewed for t in open_tickets if t.last_reviewed is not None]
        oldest = min(reviewed_dates) if reviewed_dates else None

        return TicketHealthReport(
            flagged_tickets=tuple(flagged),
            total_open=len(open_tickets),
            oldest_last_reviewed=oldest,
        )
