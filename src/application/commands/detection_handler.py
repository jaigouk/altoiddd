"""DetectionHandler -- application command for the detect flow.

Orchestrates tool detection via the ToolDetectionPort and classifies
results via the ToolScanner domain service.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.services.tool_scanner import ToolScanner

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.ports.tool_detection_port import ToolDetectionPort
    from src.domain.models.detection_result import DetectionResult


class DetectionHandler:
    """Orchestrates the detect flow: scan tools, classify conflicts.

    Attributes:
        _tool_detection: Port for detecting installed AI coding tools.
        _scanner: Domain service for classifying results.
    """

    def __init__(self, tool_detection: ToolDetectionPort) -> None:
        self._tool_detection = tool_detection
        self._scanner = ToolScanner()

    def detect(self, project_dir: Path) -> DetectionResult:
        """Detect installed AI coding tools and classify conflicts.

        Args:
            project_dir: The project directory to scan.

        Returns:
            A DetectionResult with detected tools and classified conflicts.
        """
        tool_names = self._tool_detection.detect(project_dir)
        conflicts = self._tool_detection.scan_conflicts(project_dir)
        return self._scanner.build_result(tool_names=tool_names, conflicts=conflicts)
