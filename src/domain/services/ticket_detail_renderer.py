"""Domain service for rendering ticket descriptions at varying detail levels.

TicketDetailRenderer is a stateless domain service that produces ticket body
text from aggregate metadata. Detail level is driven by subdomain classification:
Core gets full DDD/TDD/SOLID sections, Supporting gets core sections, Generic
gets a minimal stub.

Quality gate sections are derived from a StackProfile's quality_gate_commands
dict in bullet-point format. GenericProfile (empty commands) omits the section.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.quality_gate import QualityGate
from src.domain.models.ticket_values import TicketDetailLevel

if TYPE_CHECKING:
    from src.domain.models.domain_values import AggregateDesign
    from src.domain.models.stack_profile import StackProfile


# Gate labels for bullet-point rendering in tickets
_GATE_LABELS: dict[QualityGate, str] = {
    QualityGate.LINT: "zero errors",
    QualityGate.TYPES: "zero errors",
    QualityGate.TESTS: "all pass",
    QualityGate.FITNESS: "all pass",
}

# Gates excluded from ticket quality gates section
_EXCLUDED_GATES: frozenset[QualityGate] = frozenset({QualityGate.FITNESS})


class TicketDetailRenderer:
    """Renders ticket description body at the appropriate detail level.

    This is a stateless domain service -- all methods are static.
    """

    @staticmethod
    def render(
        aggregate: AggregateDesign,
        detail_level: TicketDetailLevel,
        profile: StackProfile | None = None,
    ) -> str:
        """Render a ticket description for the given aggregate and detail level.

        Args:
            aggregate: The aggregate design providing domain metadata.
            detail_level: How much detail to include.
            profile: Stack profile for quality gate commands. Defaults to
                GenericProfile when not provided.

        Returns:
            Rendered ticket description as a multi-line string.
        """
        if profile is None:
            from src.domain.models.stack_profile import GenericProfile

            profile = GenericProfile()

        if detail_level == TicketDetailLevel.FULL:
            return TicketDetailRenderer._render_full(aggregate, profile)
        if detail_level == TicketDetailLevel.STANDARD:
            return TicketDetailRenderer._render_standard(aggregate, profile)
        return TicketDetailRenderer._render_stub(aggregate)

    @staticmethod
    def _render_full(aggregate: AggregateDesign, profile: StackProfile) -> str:
        """Render FULL detail: all DDD/TDD/SOLID sections."""
        sections = [
            _render_goal_section(aggregate),
            _render_ddd_alignment_section(aggregate),
            _render_design_section(aggregate),
            _render_solid_section(aggregate),
            _render_tdd_section(profile),
            _render_steps_section(aggregate),
            _render_acceptance_criteria_section(aggregate),
            _render_edge_cases_section(),
        ]
        gates = _render_quality_gates_section(profile)
        if gates:
            sections.append(gates)
        return "\n".join(sections)

    @staticmethod
    def _render_standard(aggregate: AggregateDesign, profile: StackProfile) -> str:
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
        ]
        gate_lines = _render_quality_gates_lines(profile)
        if gate_lines:
            lines.append("")
            lines.extend(gate_lines)
        return "\n".join(lines)

    @staticmethod
    def _render_stub(aggregate: AggregateDesign) -> str:
        """Render STUB detail: matches beads-stub-template.md format.

        Includes stub notice, Goal / Problem, DDD Alignment table,
        and Risks / Dependencies section.
        """
        lines: list[str] = [
            "> **Stub ticket.** Full specification will be added when blockers are resolved.",
            "",
            "## Goal / Problem",
            "",
            f"Integrate `{aggregate.context_name}` boundary for `{aggregate.name}`.",
            "",
            "## DDD Alignment",
            "",
            "| Aspect | Detail |",
            "|--------|--------|",
            f"| Bounded Context | {aggregate.context_name} |",
            "| Layer | domain |",
            "",
            "## Risks / Dependencies",
            "",
            "- Blocked by: (see formal dependencies)",
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


def _render_tdd_section(profile: StackProfile | None = None) -> str:
    """Render the TDD Workflow section with stack-specific test runner examples.

    Args:
        profile: Stack profile for deriving the test runner command.
            When None or GenericProfile (no TESTS gate), uses ``<test-runner>``
            placeholder. When a concrete profile, uses the actual command.
    """
    test_cmd = "<test-runner>"
    if profile is not None:
        cmds = profile.quality_gate_commands
        if QualityGate.TESTS in cmds:
            test_cmd = " ".join(cmds[QualityGate.TESTS])

    return (
        "## TDD Workflow\n"
        "1. **RED:** Write failing tests for each invariant\n"
        f"   - Run: `{test_cmd}` → should FAIL\n"
        "2. **GREEN:** Implement minimal code to pass\n"
        f"   - Run: `{test_cmd}` → should PASS\n"
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


def _render_quality_gates_section(profile: StackProfile) -> str:
    """Render the Quality Gates section from profile in bullet-point format.

    Returns empty string for GenericProfile (empty commands dict).
    """
    lines = _render_quality_gates_lines(profile)
    if not lines:
        return ""
    return "\n".join(lines)


def _render_quality_gates_lines(profile: StackProfile) -> list[str]:
    """Build quality gate bullet-point lines from profile.quality_gate_commands.

    Returns empty list for GenericProfile (empty commands dict).
    """
    commands = profile.quality_gate_commands
    if not commands:
        return []

    lines: list[str] = ["## Quality Gates"]
    for gate, cmd_parts in commands.items():
        if gate in _EXCLUDED_GATES:
            continue
        cmd_str = " ".join(cmd_parts)
        label = _GATE_LABELS.get(gate, "passes")
        lines.append(f"- `{cmd_str}` -- {label}")
    return lines
