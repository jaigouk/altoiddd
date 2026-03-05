"""Value objects for the Ticket Pipeline bounded context.

TicketDetailLevel maps SubdomainClassification to generation depth.
Tier and TierClassification implement two-tier generation: near-term
tickets (depth ≤2) keep classification-based detail, far-term (depth >2)
are downgraded to STUB. Core subdomain overrides depth (always FULL).

GeneratedEpic, GeneratedTicket, and DependencyOrder are frozen dataclasses
representing ticket pipeline artifacts.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from src.domain.models.domain_values import SubdomainClassification


class TicketDetailLevel(enum.Enum):
    """Generation depth for a ticket, driven by subdomain classification.

    FULL (Core):       All sections -- Goal, DDD Alignment, Design, SOLID Mapping,
                       TDD Workflow, Steps, AC, Edge Cases, Quality Gates.
    STANDARD (Supporting): Core sections -- Goal, DDD Alignment, Steps, AC, Quality Gates.
    STUB (Generic):    Minimal -- Goal + single AC.
    """

    FULL = "full"
    STANDARD = "standard"
    STUB = "stub"

    @staticmethod
    def from_classification(classification: SubdomainClassification) -> TicketDetailLevel:
        """Map a SubdomainClassification to its TicketDetailLevel."""
        from src.domain.models.domain_values import SubdomainClassification

        mapping = {
            SubdomainClassification.CORE: TicketDetailLevel.FULL,
            SubdomainClassification.SUPPORTING: TicketDetailLevel.STANDARD,
            SubdomainClassification.GENERIC: TicketDetailLevel.STUB,
        }
        return mapping[classification]


# ---------------------------------------------------------------------------
# Two-tier generation (alty-2j7.11)
# ---------------------------------------------------------------------------

_NEAR_TERM_MAX_DEPTH = 2


class Tier(enum.Enum):
    """Near-term vs far-term classification for two-tier generation."""

    NEAR_TERM = "near_term"
    FAR_TERM = "far_term"


@dataclass(frozen=True)
class TierClassification:
    """Depth-based tier with human-readable reason.

    Attributes:
        tier: Whether this ticket is near-term or far-term.
        reason: Why this classification was assigned.
    """

    tier: Tier
    reason: str


def classify_tier(
    depth: int,
    classification: SubdomainClassification,
) -> TierClassification:
    """Classify a ticket as near-term or far-term based on depth and subdomain.

    Rules:
        1. Core subdomain → always near-term (regardless of depth).
        2. depth ≤ 2 → near-term.
        3. depth > 2 → far-term.

    Args:
        depth: Hop count from dependency root (0 = no dependencies).
        classification: Subdomain classification of the ticket's bounded context.

    Returns:
        TierClassification with tier and reason.
    """
    from src.domain.models.domain_values import SubdomainClassification

    if classification == SubdomainClassification.CORE:
        return TierClassification(Tier.NEAR_TERM, "Core subdomain: always near-term")
    if depth <= _NEAR_TERM_MAX_DEPTH:
        return TierClassification(Tier.NEAR_TERM, f"Depth {depth} <= {_NEAR_TERM_MAX_DEPTH}")
    return TierClassification(Tier.FAR_TERM, f"Depth {depth} > {_NEAR_TERM_MAX_DEPTH}")


def tier_to_detail_level(
    tier: TierClassification,
    classification: SubdomainClassification,
) -> TicketDetailLevel:
    """Map a TierClassification to a TicketDetailLevel.

    Core subdomain is always FULL (invariant: Core tickets must have full
    AC, TDD phases, and SOLID mapping). Far-term non-Core tickets are
    always STUB. Near-term tickets use classification-based mapping.

    Args:
        tier: The tier classification.
        classification: Subdomain classification of the bounded context.

    Returns:
        The appropriate TicketDetailLevel.
    """
    from src.domain.models.domain_values import SubdomainClassification

    if classification == SubdomainClassification.CORE:
        return TicketDetailLevel.FULL
    if tier.tier == Tier.FAR_TERM:
        return TicketDetailLevel.STUB
    return TicketDetailLevel.from_classification(classification)


@dataclass(frozen=True)
class GeneratedEpic:
    """An epic grouping tickets for one bounded context.

    Attributes:
        epic_id: Unique identifier for this epic.
        title: Human-readable epic title.
        description: Epic description summarizing the bounded context scope.
        bounded_context_name: Which bounded context this epic covers.
        classification: Subdomain classification driving detail level.
    """

    epic_id: str
    title: str
    description: str
    bounded_context_name: str
    classification: SubdomainClassification


@dataclass(frozen=True)
class GeneratedTicket:
    """A generated beads ticket for an aggregate within a bounded context.

    Attributes:
        ticket_id: Unique identifier for this ticket.
        title: Human-readable ticket title.
        description: Rendered ticket body (detail depends on level).
        detail_level: Generation depth (FULL/STANDARD/STUB).
        epic_id: Parent epic this ticket belongs to.
        bounded_context_name: Which bounded context this ticket targets.
        aggregate_name: Which aggregate this ticket implements.
        dependencies: IDs of tickets that must be completed first.
    """

    ticket_id: str
    title: str
    description: str
    detail_level: TicketDetailLevel
    epic_id: str
    bounded_context_name: str
    aggregate_name: str
    dependencies: tuple[str, ...] = ()
    depth: int = 0


@dataclass(frozen=True)
class DependencyOrder:
    """Topologically sorted ticket execution order.

    Attributes:
        ordered_ids: Ticket IDs in dependency-safe execution order
                     (foundation first, dependents later).
    """

    ordered_ids: tuple[str, ...]
