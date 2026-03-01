"""Domain events for the Tool Translation bounded context."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class ConfigsGenerated:
    """Emitted when a ToolConfig aggregate is approved and ready for output.

    Attributes:
        tool_names: Names of the tools configs were generated for.
        output_paths: File paths of the generated config files.
    """

    tool_names: tuple[str, ...]
    output_paths: tuple[str, ...]
