"""Tests for TicketDetailRenderer domain service."""

from __future__ import annotations

from src.domain.models.domain_values import AggregateDesign
from src.domain.models.ticket_values import TicketDetailLevel
from src.domain.services.ticket_detail_renderer import TicketDetailRenderer

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_aggregate() -> AggregateDesign:
    """Build a sample aggregate design with all fields populated."""
    return AggregateDesign(
        name="OrderAggregate",
        context_name="Orders",
        root_entity="Order",
        contained_objects=("OrderLine", "OrderStatus"),
        invariants=("total must be positive", "at least one line item"),
        commands=("PlaceOrder", "CancelOrder"),
        domain_events=("OrderPlaced", "OrderCancelled"),
    )


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


class TestFullDetail:
    def test_full_has_all_sections(self):
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL)

        assert "## Goal" in result
        assert "## DDD Alignment" in result
        assert "## Design" in result
        assert "### Invariants" in result
        assert "### Commands" in result
        assert "### Domain Events" in result
        assert "## SOLID Mapping" in result
        assert "## TDD Workflow" in result
        assert "## Steps" in result
        assert "## Acceptance Criteria" in result
        assert "## Edge Cases" in result
        assert "## Quality Gates" in result

    def test_full_includes_aggregate_name(self):
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL)

        assert "OrderAggregate" in result
        assert "Orders" in result

    def test_full_includes_invariants(self):
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL)

        assert "total must be positive" in result
        assert "at least one line item" in result

    def test_full_includes_commands_and_events(self):
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL)

        assert "PlaceOrder" in result
        assert "CancelOrder" in result
        assert "OrderPlaced" in result
        assert "OrderCancelled" in result


class TestStandardDetail:
    def test_standard_has_core_sections(self):
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.STANDARD)

        assert "## Goal" in result
        assert "## DDD Alignment" in result
        assert "## Steps" in result
        assert "## Acceptance Criteria" in result
        assert "## Quality Gates" in result

    def test_standard_omits_full_sections(self):
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.STANDARD)

        assert "## Design" not in result
        assert "## SOLID Mapping" not in result
        assert "## TDD Workflow" not in result
        assert "## Edge Cases" not in result


class TestStubDetail:
    def test_stub_is_minimal(self):
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.STUB)

        assert "## Goal" in result
        assert "Integrate" in result
        assert "## Acceptance Criteria" in result
        assert "Boundary test passes" in result

    def test_stub_has_no_extra_sections(self):
        agg = _make_aggregate()
        result = TicketDetailRenderer.render(agg, TicketDetailLevel.STUB)

        assert "## DDD Alignment" not in result
        assert "## Design" not in result
        assert "## Steps" not in result
        assert "## SOLID Mapping" not in result
