"""Domain service for rendering ticket descriptions at varying detail levels.

TicketDetailRenderer is a stateless domain service that produces ticket body
text from aggregate metadata. Detail level is driven by subdomain classification:
Core gets full DDD/TDD/SOLID sections, Supporting gets core sections, Generic
gets a minimal stub.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.ticket_values import TicketDetailLevel

if TYPE_CHECKING:
    from src.domain.models.domain_values import AggregateDesign


class TicketDetailRenderer:
    """Renders ticket description body at the appropriate detail level.

    This is a stateless domain service -- all methods are static.
    """

    @staticmethod
    def render(aggregate: AggregateDesign, detail_level: TicketDetailLevel) -> str:
        """Render a ticket description for the given aggregate and detail level.

        Args:
            aggregate: The aggregate design providing domain metadata.
            detail_level: How much detail to include.

        Returns:
            Rendered ticket description as a multi-line string.
        """
        if detail_level == TicketDetailLevel.FULL:
            return TicketDetailRenderer._render_full(aggregate)
        if detail_level == TicketDetailLevel.STANDARD:
            return TicketDetailRenderer._render_standard(aggregate)
        return TicketDetailRenderer._render_stub(aggregate)

    @staticmethod
    def _render_full(aggregate: AggregateDesign) -> str:
        """Render FULL detail: all DDD/TDD/SOLID sections."""
        sections = [
            _render_goal_section(aggregate),
            _render_ddd_alignment_section(aggregate),
            _render_design_section(aggregate),
            _render_solid_section(aggregate),
            _render_tdd_section(),
            _render_steps_section(aggregate),
            _render_acceptance_criteria_section(aggregate),
            _render_edge_cases_section(),
            _render_quality_gates_section(),
        ]
        return "\n".join(sections)

    @staticmethod
    def _render_standard(aggregate: AggregateDesign) -> str:
        """Render STANDARD detail: core sections only."""
        lines: list[str] = [
            "## Goal",
            (
                f"Implement the `{aggregate.name}` aggregate in the "
                f"`{aggregate.context_name}` bounded context."
            ),
            "",
            "## DDD Alignment",
            f"- **Bounded Context:** {aggregate.context_name}",
            f"- **Aggregate Root:** {aggregate.root_entity}",
            "",
            "## Steps",
            f"1. Create `{aggregate.name}` aggregate with core logic",
            "2. Add repository port interface",
            "3. Write unit tests",
            "",
            "## Acceptance Criteria",
            f"- [ ] `{aggregate.name}` aggregate root created",
            "- [ ] All tests pass",
            "- [ ] Coverage >= 80%",
            "",
            "## Quality Gates",
            "- `uv run ruff check .` -- zero errors",
            "- `uv run mypy .` -- zero errors",
            "- `uv run pytest` -- all pass",
        ]
        return "\n".join(lines)

    @staticmethod
    def _render_stub(aggregate: AggregateDesign) -> str:
        """Render STUB detail: minimal placeholder for Generic subdomains."""
        lines: list[str] = [
            "## Goal",
            (f"Integrate `{aggregate.context_name}` boundary for `{aggregate.name}`."),
            "",
            "## Acceptance Criteria",
            "- [ ] Boundary test passes",
        ]
        return "\n".join(lines)


# -- Private section renderers (used by _render_full) -------------------------


def _render_goal_section(aggregate: AggregateDesign) -> str:
    """Render the Goal section."""
    return (
        "## Goal\n"
        f"Implement the `{aggregate.name}` aggregate in the "
        f"`{aggregate.context_name}` bounded context.\n"
    )


def _render_ddd_alignment_section(aggregate: AggregateDesign) -> str:
    """Render the DDD Alignment section."""
    lines = [
        "## DDD Alignment",
        f"- **Bounded Context:** {aggregate.context_name}",
        f"- **Aggregate Root:** {aggregate.root_entity}",
    ]
    if aggregate.contained_objects:
        lines.append(f"- **Contained Objects:** {', '.join(aggregate.contained_objects)}")
    lines.append("")
    return "\n".join(lines)


def _render_design_section(aggregate: AggregateDesign) -> str:
    """Render the Design section with invariants, commands, and events."""
    lines: list[str] = ["## Design"]
    if aggregate.invariants:
        lines.append("### Invariants")
        lines.extend(f"- {inv}" for inv in aggregate.invariants)
        lines.append("")
    if aggregate.commands:
        lines.append("### Commands")
        lines.extend(f"- {cmd}" for cmd in aggregate.commands)
        lines.append("")
    if aggregate.domain_events:
        lines.append("### Domain Events")
        lines.extend(f"- {evt}" for evt in aggregate.domain_events)
        lines.append("")
    return "\n".join(lines)


def _render_solid_section(aggregate: AggregateDesign) -> str:
    """Render the SOLID Mapping section."""
    return "\n".join(
        [
            "## SOLID Mapping",
            f"- **S:** `{aggregate.name}` owns {aggregate.context_name} logic only",
            "- **O:** Extend via new commands/events, not modification",
            "- **L:** Subtypes honor aggregate contract",
            "- **I:** Focused repository interface for this aggregate",
            "- **D:** Depend on ports, not infrastructure",
            "",
        ]
    )


def _render_tdd_section() -> str:
    """Render the TDD Workflow section."""
    return (
        "## TDD Workflow\n"
        "1. **RED:** Write failing tests for each invariant\n"
        "2. **GREEN:** Implement minimal code to pass\n"
        "3. **REFACTOR:** Clean up while tests stay green\n"
    )


def _render_steps_section(aggregate: AggregateDesign) -> str:
    """Render the Steps section."""
    return "\n".join(
        [
            "## Steps",
            f"1. Create `{aggregate.name}` aggregate with invariant enforcement",
            "2. Implement commands and domain events",
            "3. Add repository port interface",
            "4. Write unit tests for all invariants",
            "",
        ]
    )


def _render_acceptance_criteria_section(aggregate: AggregateDesign) -> str:
    """Render the Acceptance Criteria section."""
    lines = [
        "## Acceptance Criteria",
        f"- [ ] `{aggregate.name}` aggregate root created",
    ]
    lines.extend(f"- [ ] Invariant enforced: {inv}" for inv in aggregate.invariants)
    lines.extend(f"- [ ] Command implemented: {cmd}" for cmd in aggregate.commands)
    lines.extend(
        [
            "- [ ] All tests pass",
            "- [ ] Coverage >= 80%",
            "",
        ]
    )
    return "\n".join(lines)


def _render_edge_cases_section() -> str:
    """Render the Edge Cases section."""
    return (
        "## Edge Cases\n"
        "- Empty or invalid inputs raise `InvariantViolationError`\n"
        "- Duplicate operations are idempotent or raise\n"
    )


def _render_quality_gates_section() -> str:
    """Render the Quality Gates section."""
    return (
        "## Quality Gates\n"
        "- `uv run ruff check .` -- zero errors\n"
        "- `uv run mypy .` -- zero errors\n"
        "- `uv run pytest` -- all pass"
    )
