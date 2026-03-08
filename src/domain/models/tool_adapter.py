"""Tool adapter protocol and concrete adapters for the Tool Translation context.

Each adapter translates a DomainModel into tool-native configuration files.
Adapters are pure domain logic -- they produce content strings without I/O.

Quality gate display and file globs are read from a StackProfile so that
output is stack-aware. GenericProfile produces configs with no quality gates.

Supported tools:
- Claude Code: .claude/CLAUDE.md
- Cursor: AGENTS.md + .cursor/rules/project-conventions.mdc
- Roo Code: AGENTS.md + .roomodes + .roo/rules/project-conventions.md
- OpenCode: AGENTS.md + .opencode/rules/project-conventions.md + opencode.json
"""

from __future__ import annotations

import json
from typing import TYPE_CHECKING, Protocol, runtime_checkable

from src.domain.models.tool_config import ConfigSection

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.stack_profile import StackProfile


@runtime_checkable
class ToolAdapterProtocol(Protocol):
    """Interface for translating a DomainModel into tool-native config sections."""

    def translate(self, model: DomainModel, profile: StackProfile) -> tuple[ConfigSection, ...]:
        """Translate a DomainModel into config sections for a specific tool.

        Args:
            model: A finalized DomainModel with bounded contexts and UL.
            profile: Stack profile providing quality gate display and file globs.

        Returns:
            Tuple of ConfigSection value objects ready for output.
        """
        ...


# ---------------------------------------------------------------------------
# Shared constants
# ---------------------------------------------------------------------------

_AFTER_CLOSE_PROTOCOL_TEXT = """## After-Close Protocol

After every `bd close <id>`, run these steps:

1. **Ripple review** -- `bin/bd-ripple <id> "<what this ticket produced>"`
2. **Review flagged tickets** -- `bd query label=review_needed`, read ripple comments,
   draft updates, present to user for approval.
   For **dependent tickets**: run a compatibility check -- read the delivered source files
   and trace the interfaces the flagged ticket assumes. Cite file:line for every claim.
3. **Follow-up tickets** -- create using beads templates, set dependencies
4. **Groom next ticket** -- `bd ready`, run grooming checklist on top pick
"""

# ---------------------------------------------------------------------------
# Shared content builders
# ---------------------------------------------------------------------------


def _build_ubiquitous_language_section(model: DomainModel) -> str:
    """Render the ubiquitous language glossary as markdown."""
    lines: list[str] = ["## Ubiquitous Language", ""]
    lines.append("| Term | Definition | Context |")
    lines.append("|------|-----------|---------|")
    lines.extend(
        f"| {entry.term} | {entry.definition} | {entry.context_name} |"
        for entry in model.ubiquitous_language.terms
    )
    lines.append("")
    return "\n".join(lines)


def _build_bounded_context_section(model: DomainModel) -> str:
    """Render bounded contexts and their classifications as markdown."""
    lines: list[str] = ["## Bounded Contexts", ""]
    for ctx in model.bounded_contexts:
        classification = ctx.classification.value if ctx.classification else "unclassified"
        lines.append(f"- **{ctx.name}** ({classification}): {ctx.responsibility}")
    lines.append("")
    return "\n".join(lines)


def _build_ddd_layer_rules() -> str:
    """Render DDD layer dependency rules as markdown."""
    return """## DDD Layer Rules

- `domain/` has ZERO external dependencies (no frameworks, no DB, no HTTP)
- `application/` depends on `domain/` and `ports/` (interfaces only)
- `infrastructure/` implements `ports/` and depends on external libraries
- Dependencies flow inward: infrastructure -> application -> domain
"""


def _build_after_close_section() -> str:
    """Render the after-close protocol section for CLAUDE.md / AGENTS.md.

    Reuses the same protocol text as MEMORY.md for consistency.
    """
    return _AFTER_CLOSE_PROTOCOL_TEXT


def _build_quality_gates(profile: StackProfile) -> str:
    """Render quality gate commands from the stack profile.

    Returns profile.quality_gate_display (code block format) for stacks
    that have quality gates, or empty string for GenericProfile.
    """
    return profile.quality_gate_display


def _build_memory_md(model: DomainModel, profile: StackProfile) -> str:
    """Build MEMORY.md encoding the DDD agile work process.

    Content is kept under 200 lines (Claude Code's critical window).
    Quality gates are included only when profile provides them.
    """
    parts: list[str] = [
        "# Project Memory",
        "",
        _build_memory_beads_workflow(),
        _build_memory_after_close_protocol(),
        _build_memory_grooming_checklist(),
        _build_memory_bounded_contexts(model),
        _build_memory_ubiquitous_language(model),
    ]
    gates = _build_quality_gates(profile)
    if gates:
        parts.append(gates)
    return "\n".join(parts)


def _build_memory_beads_workflow() -> str:
    """Render beads workflow commands for MEMORY.md."""
    return """## Beads Workflow

```bash
bd ready                         # Find available work
bd show <id>                     # View ticket details
bd update <id> --status in_progress  # Claim a ticket
bd close <id>                    # Close completed ticket
bin/bd-ripple <id> "<summary>"   # Flag dependents (ripple review)
bd query label=review_needed     # See tickets needing review
bd label remove <id> review_needed   # Clear flag after review
```
"""


def _build_memory_after_close_protocol() -> str:
    """Render the after-close protocol for MEMORY.md."""
    return _AFTER_CLOSE_PROTOCOL_TEXT


def _build_memory_grooming_checklist() -> str:
    """Render the grooming checklist for MEMORY.md."""
    return """## Grooming Checklist

Before claiming a ticket:

1. **Template compliance** -- description follows beads template
2. **Freshness check** -- `bd label list <id>` for `review_needed`
3. **PRD traceability** -- `/prd-traceability <id>` to verify capability coverage
4. **DDD alignment** -- bounded context boundaries respected
5. **Ubiquitous language** -- terms match DDD.md glossary
6. **TDD & SOLID** -- RED/GREEN/REFACTOR phases documented
7. **Acceptance criteria** -- testable checkboxes, edge cases, coverage >= 80%
"""


def _build_memory_bounded_contexts(model: DomainModel) -> str:
    """Render bounded contexts summary for MEMORY.md."""
    lines: list[str] = ["## Bounded Contexts", ""]
    for ctx in model.bounded_contexts:
        classification = ctx.classification.value if ctx.classification else "unclassified"
        lines.append(f"- **{ctx.name}** ({classification}): {ctx.responsibility}")
    lines.append("")
    return "\n".join(lines)


def _build_memory_ubiquitous_language(model: DomainModel) -> str:
    """Render ubiquitous language summary for MEMORY.md."""
    lines: list[str] = ["## Ubiquitous Language", ""]
    lines.append("| Term | Definition | Context |")
    lines.append("|------|-----------|---------|")
    lines.extend(
        f"| {entry.term} | {entry.definition} | {entry.context_name} |"
        for entry in model.ubiquitous_language.terms
    )
    lines.append("")
    return "\n".join(lines)


def _build_agents_md(model: DomainModel, profile: StackProfile) -> str:
    """Build a generic AGENTS.md with project conventions from the domain model."""
    parts: list[str] = [
        "# Project Conventions",
        "",
        _build_ubiquitous_language_section(model),
        _build_bounded_context_section(model),
        _build_ddd_layer_rules(),
        _build_after_close_section(),
    ]
    gates = _build_quality_gates(profile)
    if gates:
        parts.append(gates)
    return "\n".join(parts)


# ---------------------------------------------------------------------------
# Concrete adapters
# ---------------------------------------------------------------------------


class ClaudeCodeAdapter:
    """Generates .claude/CLAUDE.md and .claude/memory/MEMORY.md for Claude Code."""

    def translate(self, model: DomainModel, profile: StackProfile) -> tuple[ConfigSection, ...]:
        """Translate DomainModel into Claude Code configuration."""
        parts: list[str] = [
            "# CLAUDE.md",
            "",
            _build_ubiquitous_language_section(model),
            _build_bounded_context_section(model),
            _build_ddd_layer_rules(),
            _build_after_close_section(),
        ]
        gates = _build_quality_gates(profile)
        if gates:
            parts.append(gates)
        content = "\n".join(parts)
        memory_content = _build_memory_md(model, profile)
        return (
            ConfigSection(
                file_path=".claude/CLAUDE.md",
                content=content,
                section_name="Claude Code config",
            ),
            ConfigSection(
                file_path=".claude/memory/MEMORY.md",
                content=memory_content,
                section_name="Claude Code memory",
            ),
        )


class CursorAdapter:
    """Generates AGENTS.md and .cursor/rules/project-conventions.mdc for Cursor."""

    def translate(self, model: DomainModel, profile: StackProfile) -> tuple[ConfigSection, ...]:
        """Translate DomainModel into Cursor configuration."""
        agents_content = _build_agents_md(model, profile)

        mdc_parts: list[str] = [
            "---",
            "description: Project conventions derived from domain model",
            f"globs: {profile.file_glob}",
            "---",
            "",
            _build_ddd_layer_rules(),
        ]
        gates = _build_quality_gates(profile)
        if gates:
            mdc_parts.append(gates)
        mdc_content = "\n".join(mdc_parts)

        return (
            ConfigSection(
                file_path="AGENTS.md",
                content=agents_content,
                section_name="Cursor agents",
            ),
            ConfigSection(
                file_path=".cursor/rules/project-conventions.mdc",
                content=mdc_content,
                section_name="Cursor rules",
            ),
        )


class RooCodeAdapter:
    """Generates AGENTS.md, .roomodes, and .roo/rules/project-conventions.md."""

    def translate(self, model: DomainModel, profile: StackProfile) -> tuple[ConfigSection, ...]:
        """Translate DomainModel into Roo Code configuration."""
        agents_content = _build_agents_md(model, profile)

        roomodes = json.dumps(
            {
                "customModes": [
                    {
                        "slug": "ddd-developer",
                        "name": "DDD Developer",
                        "description": "Follows domain-driven design conventions",
                    }
                ]
            },
            indent=2,
        )

        rules_parts: list[str] = [
            "# Project Conventions",
            "",
            _build_ddd_layer_rules(),
        ]
        gates = _build_quality_gates(profile)
        if gates:
            rules_parts.append(gates)
        rules_content = "\n".join(rules_parts)

        return (
            ConfigSection(
                file_path="AGENTS.md",
                content=agents_content,
                section_name="Roo Code agents",
            ),
            ConfigSection(
                file_path=".roomodes",
                content=roomodes,
                section_name="Roo Code modes",
            ),
            ConfigSection(
                file_path=".roo/rules/project-conventions.md",
                content=rules_content,
                section_name="Roo Code rules",
            ),
        )


class OpenCodeAdapter:
    """Generates AGENTS.md, .opencode/rules/project-conventions.md, and opencode.json."""

    def translate(self, model: DomainModel, profile: StackProfile) -> tuple[ConfigSection, ...]:
        """Translate DomainModel into OpenCode configuration."""
        agents_content = _build_agents_md(model, profile)

        rules_parts: list[str] = [
            "# Project Conventions",
            "",
            _build_ddd_layer_rules(),
        ]
        gates = _build_quality_gates(profile)
        if gates:
            rules_parts.append(gates)
        rules_content = "\n".join(rules_parts)

        opencode_json = json.dumps(
            {
                "context": {
                    "include": ["AGENTS.md", ".opencode/rules/*.md"],
                }
            },
            indent=2,
        )

        return (
            ConfigSection(
                file_path="AGENTS.md",
                content=agents_content,
                section_name="OpenCode agents",
            ),
            ConfigSection(
                file_path=".opencode/rules/project-conventions.md",
                content=rules_content,
                section_name="OpenCode rules",
            ),
            ConfigSection(
                file_path="opencode.json",
                content=opencode_json,
                section_name="OpenCode config",
            ),
        )
