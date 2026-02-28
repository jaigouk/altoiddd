"""ToolScanner domain service.

Stateless service that classifies detected tool data and configuration
conflicts. Receives data (not filesystem I/O) and produces a DetectionResult.
"""

from __future__ import annotations

from pathlib import Path
from typing import ClassVar

from src.domain.models.detected_tool import DetectedTool
from src.domain.models.detection_result import ConflictSeverity, DetectionResult


class ToolScanner:
    """Classifies tools and conflicts into a structured DetectionResult.

    This is a pure domain service with no external dependencies. All
    filesystem interaction is handled by infrastructure adapters.
    """

    # Known AI coding tools and their default global config paths
    # (relative to the user's home directory).
    _KNOWN_TOOLS: ClassVar[dict[str, Path]] = {
        "claude-code": Path(".claude"),
        "cursor": Path(".cursor"),
        "roo-code": Path(".roo"),
        "opencode": Path(".config/opencode"),
    }

    @property
    def known_tools(self) -> dict[str, Path]:
        """Return a copy of the known tool registry."""
        return dict(self._KNOWN_TOOLS)

    def build_result(
        self,
        tool_names: list[str],
        conflicts: list[str],
    ) -> DetectionResult:
        """Build a DetectionResult from raw detection data.

        Args:
            tool_names: List of detected tool identifiers.
            conflicts: List of human-readable conflict descriptions.

        Returns:
            A DetectionResult with classified tools and conflict severities.
        """
        detected_tools = tuple(self._build_tool(name) for name in tool_names)
        severity_map = {c: self._classify_conflict(c) for c in conflicts}
        return DetectionResult(
            detected_tools=detected_tools,
            conflicts=tuple(conflicts),
            severity_map=severity_map,
        )

    def _build_tool(self, name: str) -> DetectedTool:
        """Create a DetectedTool, mapping known tools to their relative config paths.

        Config paths are relative to the user's home directory. Resolution to
        absolute paths is the infrastructure adapter's responsibility.
        """
        config_path = self._KNOWN_TOOLS.get(name)
        return DetectedTool(name=name, config_path=config_path)

    def _classify_conflict(self, description: str) -> ConflictSeverity:
        """Classify a conflict description into a severity level.

        Classification rules:
        - "compatible" keyword -> COMPATIBLE
        - "contradict" keyword -> CONFLICT
        - "sqlite", "restriction", or anything else -> WARNING
        """
        lower = description.lower()
        if "compatible" in lower:
            return ConflictSeverity.COMPATIBLE
        if "contradict" in lower:
            return ConflictSeverity.CONFLICT
        return ConflictSeverity.WARNING
