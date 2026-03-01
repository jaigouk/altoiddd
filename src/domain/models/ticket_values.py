"""Value objects for the Ticket Pipeline bounded context.

TicketDetailLevel maps SubdomainClassification to generation depth.
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


@dataclass(frozen=True)
class DependencyOrder:
    """Topologically sorted ticket execution order.

    Attributes:
        ordered_ids: Ticket IDs in dependency-safe execution order
                     (foundation first, dependents later).
    """

    ordered_ids: tuple[str, ...]
