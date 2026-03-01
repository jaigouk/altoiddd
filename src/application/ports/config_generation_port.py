"""Port for the Tool Translation bounded context (config generation).

Defines the interface for generating tool-native configurations from
a domain model for AI coding tools (Claude Code, Cursor, etc.).
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.domain_model import DomainModel
    from src.domain.models.tool_config import SupportedTool


@runtime_checkable
class ConfigGenerationPort(Protocol):
    """Interface for generating tool-native configurations.

    Adapters implement this to translate a single domain model into
    native config formats for each supported AI coding tool.

    Handlers using this port implement the preview-before-action pattern:
    build_preview() renders content, approve_and_write() commits it.
    """

    def generate(
        self,
        model: DomainModel,
        tools: tuple[SupportedTool, ...],
        output_dir: Path,
    ) -> None:
        """Generate tool-native configurations for the specified tools.

        Args:
            model: DomainModel with classified bounded contexts.
            tools: Which AI coding tools to generate configs for.
            output_dir: Directory where generated config files will be written.
        """
        ...
