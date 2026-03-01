"""Port for persona management in the Tool Translation bounded context.

Defines the interface for listing and generating AI agent persona
configurations for supported coding tools.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.persona import PersonaDefinition


@runtime_checkable
class PersonaPort(Protocol):
    """Interface for AI agent persona management.

    Adapters implement this to list available agent personas and generate
    tool-native persona configurations (e.g., .claude/agents/*.md).
    """

    def list_personas(self) -> tuple[PersonaDefinition, ...]:
        """List all available agent persona definitions.

        Returns:
            Tuple of PersonaDefinition values from the persona registry.
        """
        ...

    def generate(self, persona_name: str, tools: tuple[str, ...], output_dir: Path) -> None:
        """Generate persona configuration files for specified tools.

        Args:
            persona_name: The persona to generate configs for.
            tools: AI coding tool identifiers.
            output_dir: Directory where generated persona files will be written.
        """
        ...
