"""Tests for stack-aware TDD section in TicketDetailRenderer.

Verifies that the TDD Workflow section uses StackProfile to render
stack-specific test runner examples while keeping TDD principles universal.
"""

from __future__ import annotations

from src.domain.models.domain_values import AggregateDesign
from src.domain.models.stack_profile import GenericProfile, PythonUvProfile
from src.domain.models.ticket_values import TicketDetailLevel
from src.domain.services.ticket_detail_renderer import TicketDetailRenderer


def _make_aggregate() -> AggregateDesign:
    """Build a sample aggregate design."""
    return AggregateDesign(
        name="OrderAggregate",
        context_name="Orders",
        root_entity="Order",
        contained_objects=("OrderLine",),
        invariants=("total must be positive",),
        commands=("PlaceOrder",),
        domain_events=("OrderPlaced",),
    )


class TestTddSectionGenericProfile:
    """GenericProfile TDD section contains no Python-specific content."""

    def test_generic_tdd_has_no_pytest(self) -> None:
        """GenericProfile TDD section must not contain 'pytest'."""
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, GenericProfile())

        tdd_section = _extract_tdd_section(result)
        assert "pytest" not in tdd_section

    def test_generic_tdd_has_no_uv_run(self) -> None:
        """GenericProfile TDD section must not contain 'uv run'."""
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, GenericProfile())

        tdd_section = _extract_tdd_section(result)
        assert "uv run" not in tdd_section

    def test_generic_tdd_has_no_src_domain(self) -> None:
        """GenericProfile TDD section must not contain 'src/domain/'."""
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, GenericProfile())

        tdd_section = _extract_tdd_section(result)
        assert "src/domain/" not in tdd_section

    def test_generic_tdd_has_universal_principles(self) -> None:
        """GenericProfile TDD section still contains RED/GREEN/REFACTOR principles."""
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, GenericProfile())

        tdd_section = _extract_tdd_section(result)
        assert "RED" in tdd_section
        assert "GREEN" in tdd_section
        assert "REFACTOR" in tdd_section

    def test_generic_tdd_uses_generic_test_runner_placeholder(self) -> None:
        """GenericProfile TDD section uses generic placeholder for test commands."""
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, GenericProfile())

        tdd_section = _extract_tdd_section(result)
        assert "<test-runner>" in tdd_section


class TestTddSectionPythonProfile:
    """PythonUvProfile TDD section includes Python-specific test runner examples."""

    def test_python_tdd_includes_test_runner(self) -> None:
        """PythonUvProfile TDD section includes the test runner command."""
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, PythonUvProfile())

        tdd_section = _extract_tdd_section(result)
        assert "uv run pytest" in tdd_section

    def test_python_tdd_has_universal_principles(self) -> None:
        """PythonUvProfile TDD section still has universal RED/GREEN/REFACTOR."""
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, PythonUvProfile())

        tdd_section = _extract_tdd_section(result)
        assert "RED" in tdd_section
        assert "GREEN" in tdd_section
        assert "REFACTOR" in tdd_section


def _extract_tdd_section(rendered: str) -> str:
    """Extract the TDD Workflow section from rendered ticket output."""
    start = rendered.find("## TDD Workflow")
    if start == -1:
        return ""
    # Find the next ## heading after the TDD section
    next_section = rendered.find("\n## ", start + 1)
    if next_section == -1:
        return rendered[start:]
    return rendered[start:next_section]
