"""Domain service for implementability validation (Ticket Freshness BC).

Stateless validator that checks whether generated tickets are actually
implementable: section completeness, unspecified dependencies,
empty invariants/AC, and internal consistency.
"""

from __future__ import annotations

import re
from typing import TYPE_CHECKING

from src.domain.models.ticket_implementability import (
    DesignTraceResult,
    FindingSeverity,
    ImplementabilityFinding,
)
from src.domain.models.ticket_values import TicketDetailLevel

if TYPE_CHECKING:
    from src.domain.models.ticket_values import GeneratedTicket

# Section headings expected per detail level
_FULL_SECTIONS = (
    "## Goal",
    "## DDD Alignment",
    "## Design",
    "## SOLID Mapping",
    "## TDD Workflow",
    "## Steps",
    "## Acceptance Criteria",
    "## Edge Cases",
)
_STANDARD_SECTIONS = (
    "## Goal",
    "## DDD Alignment",
    "## Steps",
    "## Acceptance Criteria",
)

# Patterns that suggest unspecified dependencies ("magic happens here")
_VAGUE_ACTION_RE = re.compile(
    r"\b(?:adapter|service|handler)\s+"
    r"(?:performs?|does|executes?|runs?|handles?)\s+"
    r"(?:iterative\s+)?(?:web\s+)?(?:search|lookup|query|fetch|scan|call)",
    re.IGNORECASE,
)

# Acceptance criteria checkbox pattern
_AC_CHECKBOX_RE = re.compile(r"^- \[ \]", re.MULTILINE)


class ImplementabilityValidator:
    """Stateless domain service: validates ticket implementability."""

    @staticmethod
    def validate(ticket: GeneratedTicket) -> DesignTraceResult:
        """Validate a single ticket's implementability.

        STUB tickets skip deep validation (they are intentionally minimal).
        FULL and STANDARD tickets are checked for section completeness,
        unspecified dependencies, empty invariants, and empty AC.
        """
        if ticket.detail_level == TicketDetailLevel.STUB:
            return DesignTraceResult(ticket_id=ticket.ticket_id, findings=())

        findings: list[ImplementabilityFinding] = []
        desc = ticket.description

        findings.extend(
            ImplementabilityValidator._check_section_presence(ticket)
        )
        findings.extend(
            ImplementabilityValidator._check_unspecified_dependencies(desc)
        )
        if ticket.detail_level == TicketDetailLevel.FULL:
            findings.extend(
                ImplementabilityValidator._check_empty_invariants(desc)
            )
        findings.extend(
            ImplementabilityValidator._check_empty_acceptance_criteria(desc)
        )

        return DesignTraceResult(
            ticket_id=ticket.ticket_id,
            findings=tuple(findings),
        )

    @staticmethod
    def validate_plan(
        tickets: tuple[GeneratedTicket, ...],
    ) -> tuple[DesignTraceResult, ...]:
        """Validate all tickets. Returns one result per ticket."""
        return tuple(
            ImplementabilityValidator.validate(t) for t in tickets
        )

    @staticmethod
    def _check_section_presence(
        ticket: GeneratedTicket,
    ) -> list[ImplementabilityFinding]:
        """Check that expected sections are present for the detail level."""
        expected = (
            _FULL_SECTIONS
            if ticket.detail_level == TicketDetailLevel.FULL
            else _STANDARD_SECTIONS
        )
        return [
            ImplementabilityFinding(
                severity=FindingSeverity.MAJOR,
                location=heading,
                description=f"Missing section: {heading}",
            )
            for heading in expected
            if heading not in ticket.description
        ]

    @staticmethod
    def _check_unspecified_dependencies(
        description: str,
    ) -> list[ImplementabilityFinding]:
        """Detect vague action phrases without concrete port/library names."""
        return [
            ImplementabilityFinding(
                severity=FindingSeverity.CRITICAL,
                location="Design",
                description=(
                    f"Unspecified dependency: '{match.group()}' — "
                    f"which port or library implements this?"
                ),
            )
            for match in _VAGUE_ACTION_RE.finditer(description)
        ]

    @staticmethod
    def _check_empty_invariants(
        description: str,
    ) -> list[ImplementabilityFinding]:
        """FULL tickets should have invariants in the Design section."""
        if "### Invariants" not in description:
            return [
                ImplementabilityFinding(
                    severity=FindingSeverity.MAJOR,
                    location="## Design",
                    description=(
                        "No invariant subsection found in Design — "
                        "FULL tickets should specify domain invariants"
                    ),
                )
            ]
        return []

    @staticmethod
    def _check_empty_acceptance_criteria(
        description: str,
    ) -> list[ImplementabilityFinding]:
        """Every non-STUB ticket needs at least one checkbox item."""
        if not _AC_CHECKBOX_RE.search(description):
            return [
                ImplementabilityFinding(
                    severity=FindingSeverity.MAJOR,
                    location="## Acceptance Criteria",
                    description=(
                        "No acceptance criteria checkboxes found — "
                        "tickets must have testable AC items (- [ ] ...)"
                    ),
                )
            ]
        return []
