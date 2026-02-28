"""Tests for DetectionHandler application command.

Verifies the handler orchestrates tool detection via the port
and classifies results via the ToolScanner domain service.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.application.commands.detection_handler import DetectionHandler
from src.domain.models.detection_result import ConflictSeverity, DetectionResult

if TYPE_CHECKING:
    from pathlib import Path


# ── Fake adapter ─────────────────────────────────────────────


class FakeToolDetection:
    """In-memory test double implementing ToolDetectionPort."""

    def __init__(
        self,
        tools: list[str] | None = None,
        conflicts: list[str] | None = None,
    ) -> None:
        self._tools = tools or []
        self._conflicts = conflicts or []

    def detect(self, project_dir: Path) -> list[str]:
        return self._tools

    def scan_conflicts(self, project_dir: Path) -> list[str]:
        return self._conflicts


# ── Tests ────────────────────────────────────────────────────


class TestDetectionHandler:
    def test_detect_returns_detection_result(self, tmp_path):
        fake = FakeToolDetection(tools=["claude-code"])
        handler = DetectionHandler(tool_detection=fake)
        result = handler.detect(tmp_path)
        assert isinstance(result, DetectionResult)

    def test_detect_no_tools(self, tmp_path):
        fake = FakeToolDetection(tools=[], conflicts=[])
        handler = DetectionHandler(tool_detection=fake)
        result = handler.detect(tmp_path)
        assert result.detected_tools == ()
        assert result.conflicts == ()

    def test_detect_with_tools(self, tmp_path):
        fake = FakeToolDetection(tools=["claude-code", "cursor"])
        handler = DetectionHandler(tool_detection=fake)
        result = handler.detect(tmp_path)
        assert len(result.detected_tools) == 2
        names = [t.name for t in result.detected_tools]
        assert "claude-code" in names
        assert "cursor" in names

    def test_detect_with_conflicts(self, tmp_path):
        fake = FakeToolDetection(
            tools=["cursor"],
            conflicts=["cursor: SQLite-based config detected, cannot read"],
        )
        handler = DetectionHandler(tool_detection=fake)
        result = handler.detect(tmp_path)
        assert len(result.conflicts) == 1
        assert ConflictSeverity.WARNING in result.severity_map.values()

    def test_detect_passes_project_dir_to_port(self, tmp_path):
        """Ensure the handler passes the project_dir through to the port."""
        received_dirs: list[Path] = []

        class RecordingDetection:
            def detect(self, project_dir: Path) -> list[str]:
                received_dirs.append(project_dir)
                return []

            def scan_conflicts(self, project_dir: Path) -> list[str]:
                received_dirs.append(project_dir)
                return []

        handler = DetectionHandler(tool_detection=RecordingDetection())
        handler.detect(tmp_path)
        assert tmp_path in received_dirs

    def test_detect_with_multiple_severity_levels(self, tmp_path):
        fake = FakeToolDetection(
            tools=["claude-code", "cursor"],
            conflicts=[
                "cursor: SQLite-based config detected, cannot read",
                "claude-code: global setting 'model' contradicts local value",
            ],
        )
        handler = DetectionHandler(tool_detection=fake)
        result = handler.detect(tmp_path)
        severities = set(result.severity_map.values())
        assert ConflictSeverity.WARNING in severities
        assert ConflictSeverity.CONFLICT in severities
