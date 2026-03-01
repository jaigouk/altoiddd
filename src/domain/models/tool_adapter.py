"""Tool adapter protocol and concrete adapters for the Tool Translation context.

Each adapter translates a DomainModel into tool-native configuration files.
Adapters are pure domain logic -- they produce content strings without I/O.

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


@runtime_checkable
class ToolAdapterProtocol(Protocol):
    """Interface for translating a DomainModel into tool-native config sections."""

    def translate(self, model: DomainModel) -> tuple[ConfigSection, ...]:
        """Translate a DomainModel into config sections for a specific tool.

        Args:
            model: A finalized DomainModel with bounded contexts and UL.

        Returns:
            Tuple of ConfigSection value objects ready for output.
        """
        ...


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


def _build_quality_gates() -> str:
    """Render quality gate commands as markdown."""
    return """## Quality Gates

```bash
uv run ruff check .              # Lint
uv run mypy .                    # Type check
uv run pytest                    # Tests
```
"""


def _build_agents_md(model: DomainModel) -> str:
    """Build a generic AGENTS.md with project conventions from the domain model."""
    parts: list[str] = [
        "# Project Conventions",
        "",
        _build_ubiquitous_language_section(model),
        _build_bounded_context_section(model),
        _build_ddd_layer_rules(),
        _build_quality_gates(),
    ]
    return "\n".join(parts)


# ---------------------------------------------------------------------------
# Concrete adapters
# ---------------------------------------------------------------------------


class ClaudeCodeAdapter:
    """Generates .claude/CLAUDE.md for Claude Code."""

    def translate(self, model: DomainModel) -> tuple[ConfigSection, ...]:
        """Translate DomainModel into Claude Code configuration."""
        parts: list[str] = [
            "# CLAUDE.md",
            "",
            _build_ubiquitous_language_section(model),
            _build_bounded_context_section(model),
            _build_ddd_layer_rules(),
            _build_quality_gates(),
        ]
        content = "\n".join(parts)
        return (
            ConfigSection(
                file_path=".claude/CLAUDE.md",
                content=content,
                section_name="Claude Code config",
            ),
        )


class CursorAdapter:
    """Generates AGENTS.md and .cursor/rules/project-conventions.mdc for Cursor."""

    def translate(self, model: DomainModel) -> tuple[ConfigSection, ...]:
        """Translate DomainModel into Cursor configuration."""
        agents_content = _build_agents_md(model)

        mdc_parts: list[str] = [
            "---",
            "description: Project conventions derived from domain model",
            "globs: **/*.py",
            "---",
            "",
            _build_ddd_layer_rules(),
            _build_quality_gates(),
        ]
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

    def translate(self, model: DomainModel) -> tuple[ConfigSection, ...]:
        """Translate DomainModel into Roo Code configuration."""
        agents_content = _build_agents_md(model)

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
            _build_quality_gates(),
        ]
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

    def translate(self, model: DomainModel) -> tuple[ConfigSection, ...]:
        """Translate DomainModel into OpenCode configuration."""
        agents_content = _build_agents_md(model)

        rules_parts: list[str] = [
            "# Project Conventions",
            "",
            _build_ddd_layer_rules(),
            _build_quality_gates(),
        ]
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
