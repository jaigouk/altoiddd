"""Tests for enriched stub rendering (alty-2j7.11).

Verifies _render_stub() matches the beads-stub-template.md format:
notice, Goal / Problem, DDD Alignment table, Risks / Dependencies.
"""

from __future__ import annotations

from src.domain.models.domain_values import AggregateDesign
from src.domain.models.ticket_values import TicketDetailLevel
from src.domain.services.ticket_detail_renderer import TicketDetailRenderer


def _render_stub(context: str = "Shipping", aggregate: str = "ShipmentRoot") -> str:
    """Render a stub for testing."""
    agg = AggregateDesign(
        name=aggregate,
        context_name=context,
        root_entity=aggregate,
    )
    return TicketDetailRenderer.render(agg, TicketDetailLevel.STUB)


class TestStubContent:
    """Enriched stub matches beads-stub-template format."""

    def test_stub_has_notice(self):
        output = _render_stub()
        assert "Stub ticket" in output

    def test_stub_has_goal_section(self):
        output = _render_stub()
        assert "## Goal / Problem" in output

    def test_stub_has_context_name_in_goal(self):
        output = _render_stub("Payments", "PaymentRoot")
        assert "Payments" in output
        assert "PaymentRoot" in output

    def test_stub_has_ddd_alignment_table(self):
        output = _render_stub("Shipping")
        assert "## DDD Alignment" in output
        assert "| Bounded Context | Shipping |" in output

    def test_stub_has_risks_section(self):
        output = _render_stub()
        assert "## Risks / Dependencies" in output

    def test_stub_no_solid_section(self):
        output = _render_stub()
        assert "## SOLID" not in output

    def test_stub_no_tdd_section(self):
        output = _render_stub()
        assert "## TDD" not in output

    def test_stub_no_quality_gates_section(self):
        output = _render_stub()
        assert "## Quality Gates" not in output
