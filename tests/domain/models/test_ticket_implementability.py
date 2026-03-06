"""Tests for implementability validation value objects (Ticket Freshness BC).

Covers FindingSeverity enum, ImplementabilityFinding, InterfaceMismatch,
UnresolvedDependency, TicketSection, TicketStructure, and DesignTraceResult
— all frozen dataclass VOs.
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import InvariantViolationError
from src.domain.models.ticket_implementability import (
    DesignTraceResult,
    FindingSeverity,
    ImplementabilityFinding,
    InterfaceMismatch,
    TicketSection,
    TicketStructure,
    UnresolvedDependency,
)


class TestFindingSeverity:
    def test_enum_values(self) -> None:
        assert FindingSeverity.CRITICAL.value == "critical"
        assert FindingSeverity.MAJOR.value == "major"
        assert FindingSeverity.MINOR.value == "minor"


class TestImplementabilityFinding:
    def test_valid_finding(self) -> None:
        finding = ImplementabilityFinding(
            severity=FindingSeverity.CRITICAL,
            location="Design",
            description="Missing port reference for web search",
        )
        assert finding.severity == FindingSeverity.CRITICAL
        assert finding.location == "Design"
        assert "Missing port" in finding.description

    def test_is_frozen(self) -> None:
        finding = ImplementabilityFinding(
            severity=FindingSeverity.MAJOR,
            location="SOLID",
            description="Signature mismatch",
        )
        with pytest.raises(AttributeError):
            finding.description = "changed"  # type: ignore[misc]

    def test_rejects_empty_description(self) -> None:
        with pytest.raises(InvariantViolationError):
            ImplementabilityFinding(
                severity=FindingSeverity.MINOR,
                location="AC",
                description="",
            )

    def test_rejects_whitespace_only_description(self) -> None:
        with pytest.raises(InvariantViolationError):
            ImplementabilityFinding(
                severity=FindingSeverity.MINOR,
                location="AC",
                description="   ",
            )


class TestInterfaceMismatch:
    def test_mismatch_vo(self) -> None:
        mismatch = InterfaceMismatch(
            section_a="ISP",
            section_b="Sequence Diagram",
            description="research() signature differs",
        )
        assert mismatch.section_a == "ISP"
        assert mismatch.section_b == "Sequence Diagram"
        assert "signature" in mismatch.description

    def test_is_frozen(self) -> None:
        mismatch = InterfaceMismatch(
            section_a="ISP",
            section_b="Diagram",
            description="differs",
        )
        with pytest.raises(AttributeError):
            mismatch.section_a = "other"  # type: ignore[misc]


class TestUnresolvedDependency:
    def test_valid_unresolved(self) -> None:
        dep = UnresolvedDependency(
            port_name="WebSearchPort",
            location="Design",
            description="No such port exists in codebase",
        )
        assert dep.port_name == "WebSearchPort"
        assert dep.location == "Design"

    def test_rejects_empty_port_name(self) -> None:
        with pytest.raises(InvariantViolationError):
            UnresolvedDependency(
                port_name="",
                location="Design",
                description="Missing port",
            )


class TestTicketSection:
    def test_section_vo(self) -> None:
        section = TicketSection(heading="## Design", content="Some design text")
        assert section.heading == "## Design"
        assert section.content == "Some design text"

    def test_is_frozen(self) -> None:
        section = TicketSection(heading="## Goal", content="text")
        with pytest.raises(AttributeError):
            section.heading = "## Other"  # type: ignore[misc]


class TestTicketStructure:
    def test_get_section_found(self) -> None:
        sections = (
            TicketSection(heading="## Goal", content="goal text"),
            TicketSection(heading="## Design", content="design text"),
        )
        structure = TicketStructure(sections=sections)
        result = structure.get_section("## Design")
        assert result is not None
        assert result.content == "design text"

    def test_get_section_missing_returns_none(self) -> None:
        structure = TicketStructure(sections=())
        assert structure.get_section("## Missing") is None

    def test_is_frozen(self) -> None:
        structure = TicketStructure(sections=())
        with pytest.raises(AttributeError):
            structure.sections = ()  # type: ignore[misc]


class TestDesignTraceResult:
    def test_is_frozen(self) -> None:
        result = DesignTraceResult(ticket_id="t1", findings=())
        with pytest.raises(AttributeError):
            result.findings = ()  # type: ignore[misc]

    def test_has_findings(self) -> None:
        findings = (
            ImplementabilityFinding(
                severity=FindingSeverity.CRITICAL,
                location="Design",
                description="Missing port",
            ),
            ImplementabilityFinding(
                severity=FindingSeverity.MAJOR,
                location="AC",
                description="No checkboxes",
            ),
        )
        result = DesignTraceResult(ticket_id="t1", findings=findings)
        assert len(result.findings) == 2

    def test_is_valid_when_no_findings(self) -> None:
        result = DesignTraceResult(ticket_id="t1", findings=())
        assert result.is_valid is True

    def test_is_invalid_with_findings(self) -> None:
        findings = (
            ImplementabilityFinding(
                severity=FindingSeverity.MINOR,
                location="Goal",
                description="Vague goal",
            ),
        )
        result = DesignTraceResult(ticket_id="t1", findings=findings)
        assert result.is_valid is False

    def test_critical_count(self) -> None:
        findings = (
            ImplementabilityFinding(
                severity=FindingSeverity.CRITICAL,
                location="Design",
                description="Missing port",
            ),
            ImplementabilityFinding(
                severity=FindingSeverity.MAJOR,
                location="AC",
                description="No checkboxes",
            ),
            ImplementabilityFinding(
                severity=FindingSeverity.CRITICAL,
                location="SOLID",
                description="Mismatch",
            ),
        )
        result = DesignTraceResult(ticket_id="t1", findings=findings)
        assert result.critical_count == 2
