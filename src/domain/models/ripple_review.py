"""RippleReview aggregate root for the Ticket Freshness bounded context.

When a ticket closes, a RippleReview is created to flag open dependents
and siblings that may be affected by the change.

Invariants:
1. Non-empty context diff (enforced by ContextDiff VO __post_init__).
2. Only open tickets can be flagged.
3. Stacking allowed — multiple tickets can be flagged per review.
4. Explicit review required to clear a flag (ticket must be flagged first).
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.events.ticket_freshness_events import FlagCleared, TicketFlagged
from src.domain.models.errors import InvariantViolationError

if TYPE_CHECKING:
    from src.domain.models.ticket_freshness import ContextDiff


class RippleReview:
    """Aggregate root: manages flagging and clearing of freshness flags.

    Attributes:
        review_id: Unique identifier for this review.
        closed_ticket_id: The ticket whose closure triggered this review.
        context_diff: Description of what the closed ticket produced.
    """

    def __init__(
        self,
        review_id: str,
        closed_ticket_id: str,
        context_diff: ContextDiff,
    ) -> None:
        self.review_id = review_id
        self.closed_ticket_id = closed_ticket_id
        self.context_diff = context_diff
        self._flagged_ticket_ids: list[str] = []
        self._events: list[TicketFlagged | FlagCleared] = []

    # -- Properties -----------------------------------------------------------

    @property
    def flagged_tickets(self) -> tuple[str, ...]:
        """Currently flagged ticket IDs (defensive copy)."""
        return tuple(self._flagged_ticket_ids)

    @property
    def events(self) -> tuple[TicketFlagged | FlagCleared, ...]:
        """Domain events produced by this aggregate (defensive copy)."""
        return tuple(self._events)

    # -- Commands -------------------------------------------------------------

    def flag_ticket(
        self, ticket_id: str, *, is_open: bool, flagged_at: str = ""
    ) -> None:
        """Flag a ticket for review.

        Args:
            ticket_id: The ticket to flag.
            is_open: Whether the ticket is currently open.
            flagged_at: ISO timestamp when the flag was created.

        Raises:
            InvariantViolationError: If the ticket is not open (invariant 2).
        """
        if not is_open:
            msg = f"Only open tickets can be flagged; '{ticket_id}' is not open"
            raise InvariantViolationError(msg)

        self._flagged_ticket_ids.append(ticket_id)
        self._events.append(
            TicketFlagged(
                review_id=self.review_id,
                ticket_id=ticket_id,
                context_diff=self.context_diff,
                flagged_at=flagged_at,
            )
        )

    def clear_flag(self, ticket_id: str, *, cleared_at: str = "") -> None:
        """Clear a freshness flag after explicit review.

        Args:
            ticket_id: The ticket whose flag to clear.
            cleared_at: ISO timestamp when the flag was cleared.

        Raises:
            InvariantViolationError: If the ticket is not flagged (invariant 4).
        """
        if ticket_id not in self._flagged_ticket_ids:
            msg = f"Ticket '{ticket_id}' is not flagged and cannot be cleared"
            raise InvariantViolationError(msg)

        self._flagged_ticket_ids.remove(ticket_id)
        self._events.append(
            FlagCleared(
                review_id=self.review_id,
                ticket_id=ticket_id,
                cleared_at=cleared_at,
            )
        )
