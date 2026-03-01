"""Port for writing beads tickets and epics to the issue tracker.

Defines the interface for beads output so that the application layer
remains decoupled from concrete beads CLI or JSONL I/O.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.ticket_values import GeneratedEpic, GeneratedTicket


@runtime_checkable
class BeadsWriterPort(Protocol):
    """Interface for writing beads tickets and epics.

    Adapters implement this to create epics, tickets, and dependency
    relationships in the beads issue tracker.
    """

    def write_epic(self, epic: GeneratedEpic) -> str:
        """Write an epic to the issue tracker.

        Args:
            epic: The generated epic to write.

        Returns:
            The beads issue ID assigned to the epic.
        """
        ...

    def write_ticket(self, ticket: GeneratedTicket) -> str:
        """Write a ticket to the issue tracker.

        Args:
            ticket: The generated ticket to write.

        Returns:
            The beads issue ID assigned to the ticket.
        """
        ...

    def set_dependency(self, ticket_id: str, depends_on_id: str) -> None:
        """Set a dependency between two tickets.

        Args:
            ticket_id: The ticket that depends on another.
            depends_on_id: The ticket that must be completed first.
        """
        ...
