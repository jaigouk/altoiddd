"""ToolConfig aggregate root for the Tool Translation bounded context.

Generates tool-native configuration sections from a DomainModel using
pluggable ToolAdapterProtocol adapters. Enforces preview-before-approve
workflow per ARCHITECTURE.md Design Principle 3.

Invariants:
1. Cannot regenerate sections after approval.
2. Cannot approve without generated sections.
3. Cannot approve twice.
4. Cannot preview without sections.
"""

from __future__ import annotations

import enum
import uuid
from dataclasses import dataclass
from typing import TYPE_CHECKING

from src.domain.models.errors import InvariantViolationError

if TYPE_CHECKING:
    from src.domain.events.config_events import ConfigsGenerated
    from src.domain.models.domain_model import DomainModel
    from src.domain.models.tool_adapter import ToolAdapterProtocol


class SupportedTool(enum.Enum):
    """AI coding tools that alty can generate configurations for."""

    CLAUDE_CODE = "claude-code"
    CURSOR = "cursor"
    ROO_CODE = "roo-code"
    OPENCODE = "opencode"


@dataclass(frozen=True)
class ConfigSection:
    """A single file section within a tool configuration.

    Attributes:
        file_path: Relative path where this section should be written.
        content: The rendered file content.
        section_name: Human-readable label for this section.
    """

    file_path: str
    content: str
    section_name: str


class ToolConfig:
    """Aggregate root: generates and manages tool-native configuration sections.

    Attributes:
        config_id: Unique identifier for this configuration.
        tool: The target AI coding tool.
    """

    def __init__(self, tool: SupportedTool) -> None:
        self.config_id: str = str(uuid.uuid4())
        self.tool: SupportedTool = tool
        self._sections: list[ConfigSection] = []
        self._events: list[ConfigsGenerated] = []
        self._approved: bool = False

    # -- Properties -----------------------------------------------------------

    @property
    def sections(self) -> tuple[ConfigSection, ...]:
        """All generated config sections (defensive copy)."""
        return tuple(self._sections)

    @property
    def events(self) -> tuple[ConfigsGenerated, ...]:
        """Domain events produced by this aggregate (defensive copy)."""
        return tuple(self._events)

    # -- Commands -------------------------------------------------------------

    def build_sections(
        self,
        model: DomainModel,
        adapter: ToolAdapterProtocol,
    ) -> None:
        """Generate config sections from a DomainModel using the given adapter.

        Args:
            model: A finalized DomainModel with classified bounded contexts.
            adapter: A tool-specific adapter that translates domain model to sections.

        Raises:
            InvariantViolationError: If this config has already been approved.
        """
        if self._approved:
            msg = "Cannot regenerate sections on an approved config"
            raise InvariantViolationError(msg)

        self._sections.clear()
        self._sections.extend(adapter.translate(model))

    def preview(self) -> str:
        """Return a human-readable preview of generated sections.

        Raises:
            InvariantViolationError: If no sections have been generated.
        """
        if not self._sections:
            msg = "No sections generated yet — call build_sections() first"
            raise InvariantViolationError(msg)

        lines: list[str] = [
            f"Tool: {self.tool.value}",
            f"Total sections: {len(self._sections)}",
            "",
        ]

        lines.extend(
            f"  {section.section_name}: {section.file_path}" for section in self._sections
        )

        return "\n".join(lines)

    def approve(self) -> None:
        """Approve the config, emitting ConfigsGenerated.

        Raises:
            InvariantViolationError: If config has no sections or is already approved.
        """
        if self._approved:
            msg = "Config already approved"
            raise InvariantViolationError(msg)

        if not self._sections:
            msg = "Cannot approve config with no sections"
            raise InvariantViolationError(msg)

        self._approved = True

        from src.domain.events.config_events import ConfigsGenerated

        self._events.append(
            ConfigsGenerated(
                tool_names=(self.tool.value,),
                output_paths=tuple(s.file_path for s in self._sections),
            )
        )
