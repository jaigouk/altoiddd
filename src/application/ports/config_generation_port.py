"""Port for the Tool Translation bounded context (config generation).

Defines the interface for generating tool-native configurations from
a domain model for AI coding tools (Claude Code, Cursor, etc.).
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class ConfigGenerationPort(Protocol):
    """Interface for generating tool-native configurations.

    Adapters implement this to translate a single domain model into
    native config formats for each supported AI coding tool.
    """

    def generate(self, tools: list[str], output_dir: Path) -> str:
        """Generate tool-native configurations for the specified tools.

        Args:
            tools: List of AI coding tool identifiers (e.g., ["claude", "cursor"]).
            output_dir: Directory where generated config files will be written.

        Returns:
            Summary of the generated configurations.
        """
        ...
