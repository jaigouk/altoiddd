"""Domain events for the Ticket Pipeline bounded context."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class TicketPlanApproved:
    """Emitted when a TicketPlan is approved and ready for output.

    Attributes:
        plan_id: Unique ID of the approved ticket plan.
        approved_ticket_ids: IDs of tickets approved for generation.
        dismissed_ticket_ids: IDs of tickets excluded from this approval.
    """

    plan_id: str
    approved_ticket_ids: tuple[str, ...]
    dismissed_ticket_ids: tuple[str, ...]
