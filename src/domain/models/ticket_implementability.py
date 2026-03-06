"""Domain models for implementability validation (Ticket Freshness bounded context).

Value objects for detecting whether a generated ticket is actually implementable:
unresolved port references, signature mismatches, unspecified dependencies.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass

from src.domain.models.errors import InvariantViolationError


class FindingSeverity(enum.Enum):
    """Severity of an implementability finding."""

    CRITICAL = "critical"
    MAJOR = "major"
    MINOR = "minor"


@dataclass(frozen=True)
class ImplementabilityFinding:
    """A single implementability issue found during validation.

    Attributes:
        severity: How critical this finding is.
        location: Which ticket section contains the issue.
        description: Human-readable description of the finding.

    Invariant:
        description must be non-empty and non-whitespace.
    """

    severity: FindingSeverity
    location: str
    description: str

    def __post_init__(self) -> None:
        if not self.description or not self.description.strip():
            msg = "ImplementabilityFinding description must not be empty"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class InterfaceMismatch:
    """Port/method signature mismatch between ticket sections.

    Attributes:
        section_a: First section name (e.g., "ISP").
        section_b: Second section name (e.g., "Sequence Diagram").
        description: What differs.
    """

    section_a: str
    section_b: str
    description: str


@dataclass(frozen=True)
class UnresolvedDependency:
    """Reference to a port/library that does not exist.

    Attributes:
        port_name: The name of the unresolved port or library.
        location: Which ticket section references it.
        description: Human-readable context.

    Invariant:
        port_name must be non-empty and non-whitespace.
    """

    port_name: str
    location: str
    description: str

    def __post_init__(self) -> None:
        if not self.port_name or not self.port_name.strip():
            msg = "UnresolvedDependency port_name must not be empty"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class TicketSection:
    """Parsed section of a ticket description.

    Attributes:
        heading: The markdown heading (e.g., "## Design").
        content: The text content under this heading.
    """

    heading: str
    content: str


@dataclass(frozen=True)
class TicketStructure:
    """Full parsed ticket as a collection of TicketSections.

    Attributes:
        sections: All sections found in the ticket description.
    """

    sections: tuple[TicketSection, ...]

    def get_section(self, heading: str) -> TicketSection | None:
        """Look up a section by heading text. Returns None if not found."""
        for section in self.sections:
            if section.heading == heading:
                return section
        return None


@dataclass(frozen=True)
class DesignTraceResult:
    """Structured validation result with findings list.

    Attributes:
        ticket_id: Which ticket was validated.
        findings: All implementability findings discovered.
    """

    ticket_id: str
    findings: tuple[ImplementabilityFinding, ...]

    @property
    def is_valid(self) -> bool:
        """Whether the ticket passed all implementability checks."""
        return len(self.findings) == 0

    @property
    def critical_count(self) -> int:
        """Number of CRITICAL findings."""
        return sum(
            1 for f in self.findings if f.severity == FindingSeverity.CRITICAL
        )
