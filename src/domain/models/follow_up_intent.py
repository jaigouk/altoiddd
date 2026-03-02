"""Domain models for spike follow-up auditing (Ticket Freshness bounded context).

Value objects representing follow-up intentions extracted from spike
research reports and the audit result comparing them to created tickets.
"""

from __future__ import annotations

from dataclasses import dataclass

from src.domain.models.errors import InvariantViolationError


@dataclass(frozen=True)
class FollowUpIntent:
    """A concrete ticket idea discovered during a spike.

    Format-agnostic: the domain does not know whether this came from
    a Markdown heading, YAML, or any other report format.

    Attributes:
        title: Short description of the intended ticket.
        description: Optional detail about the ticket's scope.

    Invariant:
        title must be non-empty and non-whitespace.
    """

    title: str
    description: str

    def __post_init__(self) -> None:
        if not self.title or not self.title.strip():
            msg = "FollowUpIntent title must not be empty or whitespace-only"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class FollowUpAuditResult:
    """Result of auditing spike follow-ups against created tickets.

    Attributes:
        spike_id: The spike ticket identifier.
        report_path: Path to the research report that was scanned.
        defined_intents: Follow-up intents extracted from the report.
        matched_ticket_ids: Beads ticket IDs that matched defined intents.
        orphaned_intents: Intents with no matching ticket in Beads.
    """

    spike_id: str
    report_path: str
    defined_intents: tuple[FollowUpIntent, ...]
    matched_ticket_ids: tuple[str, ...]
    orphaned_intents: tuple[FollowUpIntent, ...]

    @property
    def defined_count(self) -> int:
        """Number of follow-up intents defined in the spike report."""
        return len(self.defined_intents)

    @property
    def orphaned_count(self) -> int:
        """Number of intents with no corresponding ticket."""
        return len(self.orphaned_intents)

    @property
    def has_orphans(self) -> bool:
        """Whether any defined intents are missing corresponding tickets."""
        return self.orphaned_count > 0
