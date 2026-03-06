"""Tests for ImplementabilityValidator domain service.

Verifies that the stateless service detects implementability gaps:
missing sections, unspecified dependencies, empty invariants/AC,
and internal inconsistencies within generated tickets.
"""

from __future__ import annotations

import src.domain.services.implementability_validator as validator_mod
from src.domain.models.ticket_implementability import (
    DesignTraceResult,
    FindingSeverity,
)
from src.domain.models.ticket_values import GeneratedTicket, TicketDetailLevel
from src.domain.services.implementability_validator import ImplementabilityValidator


def _make_full_description(
    *,
    invariants: str = "### Invariants\n- Order total must be positive\n",
    ac_items: str = "- [ ] Aggregate created\n- [ ] All tests pass\n",
    design_extra: str = "",
) -> str:
    """Build a FULL-detail ticket description with all 9 sections."""
    return (
        "## Goal\nImplement the `Order` aggregate in the `Sales` bounded context.\n\n"
        "## DDD Alignment\n- **Bounded Context:** Sales\n- **Aggregate Root:** Order\n\n"
        f"## Design\n{invariants}{design_extra}\n"
        "## SOLID Mapping\n- **S:** Order owns Sales logic only\n\n"
        "## TDD Workflow\n1. RED: Write failing tests\n\n"
        f"## Steps\n1. Create Order aggregate\n\n"
        f"## Acceptance Criteria\n{ac_items}\n"
        "## Edge Cases\n- Empty inputs raise InvariantViolationError\n\n"
        "## Quality Gates\n- `uv run pytest` -- all pass\n"
    )


def _make_ticket(
    *,
    detail_level: TicketDetailLevel = TicketDetailLevel.FULL,
    description: str | None = None,
    ticket_id: str = "t-001",
) -> GeneratedTicket:
    """Build a GeneratedTicket with given detail level and description."""
    if description is None:
        description = _make_full_description()
    return GeneratedTicket(
        ticket_id=ticket_id,
        title="Implement Order aggregate",
        description=description,
        detail_level=detail_level,
        epic_id="e-001",
        bounded_context_name="Sales",
        aggregate_name="Order",
    )


class TestImplementabilityValidator:
    def test_passes_well_formed_ticket(self) -> None:
        """FULL ticket with all sections present → is_valid."""
        ticket = _make_ticket()
        result = ImplementabilityValidator.validate(ticket)
        assert result.is_valid is True

    def test_detects_unspecified_dependency(self) -> None:
        """Description says 'adapter performs web search' with no port → CRITICAL."""
        desc = _make_full_description(
            design_extra="The adapter performs iterative web search to gather results.\n",
        )
        ticket = _make_ticket(description=desc)
        result = ImplementabilityValidator.validate(ticket)
        assert not result.is_valid
        critical = [f for f in result.findings if f.severity == FindingSeverity.CRITICAL]
        assert len(critical) >= 1
        assert any("unspecified" in f.description.lower() for f in critical)

    def test_detects_empty_invariants_on_full(self) -> None:
        """FULL ticket with empty Design/Invariants → MAJOR finding."""
        desc = _make_full_description(invariants="")
        ticket = _make_ticket(description=desc)
        result = ImplementabilityValidator.validate(ticket)
        major = [f for f in result.findings if f.severity == FindingSeverity.MAJOR]
        assert len(major) >= 1
        assert any("invariant" in f.description.lower() for f in major)

    def test_detects_empty_acceptance_criteria(self) -> None:
        """Ticket with no checkbox items → MAJOR finding."""
        desc = _make_full_description(ac_items="TBD\n")
        ticket = _make_ticket(description=desc)
        result = ImplementabilityValidator.validate(ticket)
        major = [f for f in result.findings if f.severity == FindingSeverity.MAJOR]
        assert len(major) >= 1
        assert any("acceptance criteria" in f.description.lower() for f in major)

    def test_passes_stub_ticket_skips_deep_validation(self) -> None:
        """STUB tickets skip deep checks → is_valid."""
        stub_desc = (
            "> **Stub ticket.**\n\n"
            "## Goal / Problem\nIntegrate boundary.\n\n"
            "## DDD Alignment\n| Aspect | Detail |\n"
        )
        ticket = _make_ticket(
            detail_level=TicketDetailLevel.STUB,
            description=stub_desc,
        )
        result = ImplementabilityValidator.validate(ticket)
        assert result.is_valid is True

    def test_returns_structured_result(self) -> None:
        ticket = _make_ticket()
        result = ImplementabilityValidator.validate(ticket)
        assert isinstance(result, DesignTraceResult)
        assert result.ticket_id == "t-001"

    def test_multiple_findings_accumulated(self) -> None:
        """Both empty invariants AND empty AC → 2+ findings."""
        desc = _make_full_description(invariants="", ac_items="TBD\n")
        ticket = _make_ticket(description=desc)
        result = ImplementabilityValidator.validate(ticket)
        assert len(result.findings) >= 2


class TestValidatePlan:
    def test_validate_plan_returns_per_ticket_results(self) -> None:
        """validate_plan returns one DesignTraceResult per ticket."""
        tickets = (
            _make_ticket(ticket_id="t-001"),
            _make_ticket(ticket_id="t-002"),
        )
        results = ImplementabilityValidator.validate_plan(tickets)
        assert len(results) == 2
        assert results[0].ticket_id == "t-001"
        assert results[1].ticket_id == "t-002"


class TestModuleHygiene:
    """Ensure no dead code or unused symbols in the validator module."""

    def test_no_unused_module_level_regex(self) -> None:
        """Every module-level compiled regex is used in at least one method."""
        import re

        # Collect all module-level compiled regexes
        module_regexes = {
            name: obj
            for name, obj in vars(validator_mod).items()
            if isinstance(obj, re.Pattern)
        }

        # Collect all regexes referenced in Load context (usage, not assignment)
        import ast
        import inspect

        source = inspect.getsource(validator_mod)
        tree = ast.parse(source)

        referenced: set[str] = set()
        for node in ast.walk(tree):
            if (
                isinstance(node, ast.Name)
                and node.id in module_regexes
                and isinstance(node.ctx, ast.Load)
            ):
                referenced.add(node.id)

        unused = set(module_regexes) - referenced
        assert unused == set(), f"Unused compiled regex constants: {unused}"

    def test_no_missing_section_for_detail_level(self) -> None:
        """STANDARD sections must be a subset of FULL sections."""
        assert set(validator_mod._STANDARD_SECTIONS).issubset(
            set(validator_mod._FULL_SECTIONS)
        )
