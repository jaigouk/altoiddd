"""Application command handler for tool config generation.

ConfigGenerationHandler builds ToolConfig aggregates from a DomainModel
for each requested tool, generates config sections via adapters, and
writes them via the FileWriterPort.

Supports a preview-before-write workflow: build_preview() renders content
for user review, approve_and_write() commits approved content to disk.
"""

from __future__ import annotations

from dataclasses import dataclass
from typing import TYPE_CHECKING

from src.domain.models.tool_adapter import (
    ClaudeCodeAdapter,
    CursorAdapter,
    OpenCodeAdapter,
    RooCodeAdapter,
    ToolAdapterProtocol,
)
from src.domain.models.tool_config import SupportedTool, ToolConfig

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.ports.file_writer_port import FileWriterPort
    from src.domain.models.domain_model import DomainModel

_ADAPTER_REGISTRY: dict[SupportedTool, type[ToolAdapterProtocol]] = {
    SupportedTool.CLAUDE_CODE: ClaudeCodeAdapter,
    SupportedTool.CURSOR: CursorAdapter,
    SupportedTool.ROO_CODE: RooCodeAdapter,
    SupportedTool.OPENCODE: OpenCodeAdapter,
}


@dataclass
class ConfigPreview:
    """Rendered tool configurations ready for user review before writing.

    Attributes:
        configs: The ToolConfig aggregates, one per requested tool.
        summary: Human-readable preview summary of all configs.
    """

    configs: tuple[ToolConfig, ...]
    summary: str


class ConfigGenerationHandler:
    """Orchestrates tool config generation from a DomainModel.

    Reads bounded contexts and ubiquitous language from a finalized
    DomainModel, builds ToolConfig aggregates for each requested tool,
    and writes config files via the FileWriterPort.
    """

    def __init__(self, writer: FileWriterPort) -> None:
        self._writer = writer

    def build_preview(
        self,
        model: DomainModel,
        tools: tuple[SupportedTool, ...],
    ) -> ConfigPreview:
        """Build tool configs and render for preview without writing.

        Args:
            model: A finalized DomainModel with classified bounded contexts.
            tools: Which tools to generate configs for.

        Returns:
            ConfigPreview with rendered configs and summary.

        Raises:
            ValueError: If no tools specified.
        """
        if not tools:
            msg = "No tools specified for config generation"
            raise ValueError(msg)

        configs: list[ToolConfig] = []
        summary_lines: list[str] = ["Config Generation Preview", ""]

        for tool in tools:
            adapter_cls = _ADAPTER_REGISTRY[tool]
            adapter = adapter_cls()
            config = ToolConfig(tool=tool)
            config.build_sections(model=model, adapter=adapter)
            configs.append(config)
            summary_lines.append(config.preview())
            summary_lines.append("")

        return ConfigPreview(
            configs=tuple(configs),
            summary="\n".join(summary_lines),
        )

    def approve_and_write(
        self,
        preview: ConfigPreview,
        output_dir: Path,
    ) -> None:
        """Approve all configs (emitting domain events) and write to disk.

        This is the only way to finalize configs -- enforcing the
        preview-before-action pattern per ARCHITECTURE.md Design Principle 3.

        Args:
            preview: The ConfigPreview from build_preview().
            output_dir: Project root directory.
        """
        for config in preview.configs:
            config.approve()
            for section in config.sections:
                self._writer.write_file(
                    output_dir / section.file_path,
                    section.content,
                )
