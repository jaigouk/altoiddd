"""Domain events for the Ticket Freshness bounded context."""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from src.domain.models.ticket_freshness import ContextDiff


@dataclass(frozen=True)
class TicketFlagged:
    """Emitted when a ticket is flagged for review due to an upstream change.

    Attributes:
        review_id: Unique ID of the RippleReview that produced this flag.
        ticket_id: The ticket that was flagged.
        context_diff: The upstream change that triggered the flag.
        flagged_at: ISO datetime string when the flag was set.
    """

    review_id: str
    ticket_id: str
    context_diff: ContextDiff
    flagged_at: str


@dataclass(frozen=True)
class FlagCleared:
    """Emitted when a freshness flag is cleared after explicit review.

    Attributes:
        review_id: Unique ID of the RippleReview that produced this clear.
        ticket_id: The ticket whose flag was cleared.
        cleared_at: ISO datetime string when the flag was cleared.
    """

    review_id: str
    ticket_id: str
    cleared_at: str
