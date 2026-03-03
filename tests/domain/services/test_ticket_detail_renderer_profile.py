"""Tests for TicketDetailRenderer profile integration.

Verifies that the renderer derives quality gate bullet-point format
from profile.quality_gate_commands, and omits the section entirely
for GenericProfile (empty commands dict).
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


class TestFullDetailProfile:
    """FULL detail renders quality gates from profile.quality_gate_commands."""

    def test_python_profile_includes_quality_gates(self) -> None:
        """FULL detail with PythonUvProfile has Quality Gates section."""
        agg = _make_aggregate()
        profile = PythonUvProfile()

        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, profile)

        assert "## Quality Gates" in result

    def test_python_profile_uses_bullet_format(self) -> None:
        """FULL detail renders bullet-point format, NOT code block."""
        agg = _make_aggregate()
        profile = PythonUvProfile()

        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, profile)

        # Bullet format
        assert "- `uv run ruff check .` -- zero errors" in result
        assert "- `uv run mypy .` -- zero errors" in result
        assert "- `uv run pytest` -- all pass" in result
        # NOT code block format
        assert "```bash" not in result

    def test_generic_profile_omits_quality_gates(self) -> None:
        """FULL detail with GenericProfile omits Quality Gates section."""
        agg = _make_aggregate()
        profile = GenericProfile()

        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, profile)

        assert "## Quality Gates" not in result
        assert "uv run" not in result

    def test_python_profile_excludes_fitness_gate(self) -> None:
        """FULL detail excludes FITNESS gate from quality gates section."""
        agg = _make_aggregate()
        profile = PythonUvProfile()

        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, profile)

        assert "fitness" not in result.lower().split("## quality gates")[-1].split("##")[0]

    def test_python_output_identical_to_current(self) -> None:
        """PythonUvProfile output matches current hardcoded format exactly."""
        agg = _make_aggregate()
        profile = PythonUvProfile()

        result = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, profile)

        # Must contain exact bullet-point lines
        assert "- `uv run ruff check .` -- zero errors" in result
        assert "- `uv run mypy .` -- zero errors" in result
        assert "- `uv run pytest` -- all pass" in result


class TestStandardDetailProfile:
    """STANDARD detail derives inline quality gates from profile."""

    def test_python_profile_includes_quality_gates(self) -> None:
        """STANDARD detail with PythonUvProfile has Quality Gates section."""
        agg = _make_aggregate()
        profile = PythonUvProfile()

        result = TicketDetailRenderer.render(agg, TicketDetailLevel.STANDARD, profile)

        assert "## Quality Gates" in result
        assert "uv run ruff check ." in result

    def test_generic_profile_omits_quality_gates(self) -> None:
        """STANDARD detail with GenericProfile omits Quality Gates entirely."""
        agg = _make_aggregate()
        profile = GenericProfile()

        result = TicketDetailRenderer.render(agg, TicketDetailLevel.STANDARD, profile)

        assert "## Quality Gates" not in result
        assert "uv run" not in result

    def test_standard_python_output_matches_current(self) -> None:
        """STANDARD PythonUvProfile output matches current hardcoded format."""
        agg = _make_aggregate()
        profile = PythonUvProfile()

        result = TicketDetailRenderer.render(agg, TicketDetailLevel.STANDARD, profile)

        assert "- `uv run ruff check .` -- zero errors" in result
        assert "- `uv run mypy .` -- zero errors" in result
        assert "- `uv run pytest` -- all pass" in result


class TestStubDetailProfile:
    """STUB detail is unaffected by profile (no quality gates)."""

    def test_stub_has_no_quality_gates_regardless(self) -> None:
        """STUB detail never has quality gates, regardless of profile."""
        agg = _make_aggregate()
        profile = PythonUvProfile()

        result = TicketDetailRenderer.render(agg, TicketDetailLevel.STUB, profile)

        assert "## Quality Gates" not in result
