"""Command handler for persona management in the Tool Translation bounded context.

PersonaHandler orchestrates listing, previewing, and writing agent persona
configurations. Follows the preview-before-action pattern per
ARCHITECTURE.md Design Principle 3.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

from src.domain.models.errors import InvariantViolationError
from src.domain.models.persona import (
    PERSONA_REGISTRY,
    SUPPORTED_TOOLS,
    TOOL_TARGET_PATHS,
    PersonaType,
)

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.ports.file_writer_port import FileWriterPort
    from src.domain.models.persona import PersonaDefinition


@dataclass
class PersonaPreview:
    """Rendered persona content ready for user review before writing.

    Attributes:
        persona: The PersonaDefinition being previewed.
        tool: Target AI coding tool identifier.
        content: Rendered persona instructions content.
        target_path: Relative file path where this persona will be written.
        summary: Human-readable preview summary.
    """

    persona: PersonaDefinition
    tool: str
    content: str
    target_path: str
    summary: str


class PersonaHandler:
    """Orchestrates persona listing, preview, and file writing.

    Reads persona definitions from PERSONA_REGISTRY, validates inputs,
    renders previews, and writes via the FileWriterPort.
    """

    def __init__(self, writer: FileWriterPort) -> None:
        self._writer = writer

    def list_personas(self) -> tuple[PersonaDefinition, ...]:
        """Return all registered persona definitions.

        Returns:
            Tuple of all PersonaDefinition values from PERSONA_REGISTRY.
        """
        return tuple(PERSONA_REGISTRY.values())

    def build_preview(self, persona_name: str, tool: str) -> PersonaPreview:
        """Build a preview for the given persona and tool without writing.

        Looks up the persona by name (case-insensitive) or by PersonaType value.

        Args:
            persona_name: Display name or PersonaType value to match.
            tool: Target AI coding tool identifier (e.g. "claude-code").

        Returns:
            PersonaPreview with rendered content and target path.

        Raises:
            InvariantViolationError: If persona not found or tool not supported.
        """
        persona = self._resolve_persona(persona_name)
        self._validate_tool(tool)

        # Derive a filename-safe slug from the persona name
        slug = persona.name.lower().replace(" ", "-")
        target_path = TOOL_TARGET_PATHS[tool].format(name=slug)

        content = persona.instructions_template

        summary = (
            f"Persona: {persona.name} ({persona.register.value})\n"
            f"Tool: {tool}\n"
            f"Target: {target_path}"
        )

        return PersonaPreview(
            persona=persona,
            tool=tool,
            content=content,
            target_path=target_path,
            summary=summary,
        )

    def approve_and_write(self, preview: PersonaPreview, output_dir: Path) -> None:
        """Write a previously previewed persona configuration to disk.

        Args:
            preview: The PersonaPreview from build_preview().
            output_dir: Project root directory.
        """
        self._writer.write_file(
            output_dir / preview.target_path,
            preview.content,
        )

    # -- Private helpers -------------------------------------------------------

    @staticmethod
    def _resolve_persona(persona_name: str) -> PersonaDefinition:
        """Find a persona by display name (case-insensitive) or type value.

        Raises:
            InvariantViolationError: If no matching persona found.
        """
        lower_name = persona_name.lower().strip()

        # Match by display name (case-insensitive)
        for defn in PERSONA_REGISTRY.values():
            if defn.name.lower() == lower_name:
                return defn

        # Match by PersonaType value
        for ptype in PersonaType:
            if ptype.value == lower_name:
                return PERSONA_REGISTRY[ptype]

        valid_names = ", ".join(f"'{d.name}'" for d in PERSONA_REGISTRY.values())
        msg = f"Unknown persona '{persona_name}'. Valid personas: {valid_names}"
        raise InvariantViolationError(msg)

    @staticmethod
    def _validate_tool(tool: str) -> None:
        """Validate the tool is in SUPPORTED_TOOLS.

        Raises:
            InvariantViolationError: If tool is not supported.
        """
        if tool not in SUPPORTED_TOOLS:
            valid_tools = ", ".join(f"'{t}'" for t in SUPPORTED_TOOLS)
            msg = f"Unsupported tool '{tool}'. Valid tools: {valid_tools}"
            raise InvariantViolationError(msg)
