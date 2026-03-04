"""Domain models for the Ticket Freshness bounded context (CORE subdomain).

Value objects for tracking ticket staleness, ripple review flags,
and health reporting across the project backlog.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass

from src.domain.models.errors import InvariantViolationError


@dataclass(frozen=True)
class ContextDiff:
    """Captures what changed when a ticket was closed.

    Attributes:
        summary: Human-readable description of the change.
        triggering_ticket_id: The ticket that was closed.
        produced_at: ISO date string when the change was produced.

    Invariant:
        summary must be non-empty and non-whitespace.
    """

    summary: str
    triggering_ticket_id: str
    produced_at: str

    def __post_init__(self) -> None:
        if not self.summary or not self.summary.strip():
            msg = "ContextDiff summary must not be empty or whitespace-only"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class FreshnessFlag:
    """A single flag indicating a ticket needs review due to upstream change.

    Attributes:
        context_diff: The change that triggered this flag.
        flagged_at: ISO datetime string when the flag was set.
    """

    context_diff: ContextDiff
    flagged_at: str


class TicketFreshnessStatus(enum.Enum):
    """Freshness status of a ticket in the backlog."""

    FRESH = "fresh"
    REVIEW_NEEDED = "review_needed"
    NEVER_REVIEWED = "never_reviewed"


@dataclass(frozen=True)
class FlaggedTicket:
    """A ticket that has one or more freshness flags pending review.

    Invariant 3: flags is a tuple allowing multiple entries (stacking).

    Attributes:
        ticket_id: The flagged ticket's identifier.
        title: Human-readable ticket title.
        flags: Stacked freshness flags (one per upstream change).
        status: Current freshness status.
    """

    ticket_id: str
    title: str
    flags: tuple[FreshnessFlag, ...]
    status: TicketFreshnessStatus

    @property
    def flag_count(self) -> int:
        """Number of pending freshness flags."""
        return len(self.flags)


@dataclass(frozen=True)
class TicketHealthReport:
    """Aggregate report on ticket freshness across the backlog.

    Attributes:
        flagged_tickets: Tickets that need review.
        total_open: Total number of open tickets in the backlog.
        oldest_last_reviewed: ISO date of the oldest last-reviewed ticket,
            or None if no tickets have been reviewed.
    """

    flagged_tickets: tuple[FlaggedTicket, ...]
    total_open: int
    oldest_last_reviewed: str | None = None

    @property
    def review_needed_count(self) -> int:
        """Number of tickets needing review."""
        return len(self.flagged_tickets)

    @property
    def has_issues(self) -> bool:
        """Whether any tickets need attention."""
        return self.review_needed_count > 0

    @property
    def freshness_pct(self) -> float:
        """Percentage of open tickets that are fresh (not needing review).

        Returns 100.0 when there are no open tickets (avoids division by zero).
        Formula: (total_open - review_needed_count) / total_open * 100
        """
        if self.total_open == 0:
            return 100.0
        return (self.total_open - self.review_needed_count) / self.total_open * 100


@dataclass(frozen=True)
class EpicHealthSummary:
    """Epic-level breakdown of ticket freshness.

    Provides a per-epic view of how many tickets are fresh vs stale,
    used when reporting health grouped by epic rather than globally.

    Attributes:
        epic_id: Identifier of the epic.
        total_tickets: Total tickets in this epic.
        fresh_count: Number of fresh (non-stale) tickets.
        stale_count: Number of tickets needing review.
    """

    epic_id: str
    total_tickets: int
    fresh_count: int
    stale_count: int

    def __post_init__(self) -> None:
        if self.total_tickets < 0 or self.fresh_count < 0 or self.stale_count < 0:
            msg = "Counts must be non-negative"
            raise InvariantViolationError(msg)
        if self.fresh_count + self.stale_count != self.total_tickets:
            msg = (
                f"fresh_count ({self.fresh_count}) + stale_count ({self.stale_count})"
                f" must equal total_tickets ({self.total_tickets})"
            )
            raise InvariantViolationError(msg)

    @property
    def freshness_pct(self) -> float:
        """Percentage of tickets that are fresh.

        Returns 100.0 when there are no tickets (avoids division by zero).
        """
        if self.total_tickets == 0:
            return 100.0
        return self.fresh_count / self.total_tickets * 100


@dataclass(frozen=True)
class OpenTicketData:
    """Raw data for an open ticket, used by the reader protocol.

    This is an ACL-facing VO that infrastructure adapters populate.

    Attributes:
        ticket_id: The ticket identifier.
        title: Human-readable ticket title.
        labels: Labels attached to the ticket.
        last_reviewed: ISO date of last review, or None.
    """

    ticket_id: str
    title: str
    labels: tuple[str, ...]
    last_reviewed: str | None = None
