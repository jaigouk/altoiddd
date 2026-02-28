"""Tests for ToolScanner domain service.

Verifies detection logic and conflict classification.
The ToolScanner is a stateless domain service that receives data
(not filesystem I/O) and classifies conflicts.
"""

from __future__ import annotations

from pathlib import Path

from src.domain.models.detection_result import ConflictSeverity, DetectionResult
from src.domain.services.tool_scanner import ToolScanner


class TestToolScannerBuildResult:
    def test_build_result_with_no_tools_no_conflicts(self):
        scanner = ToolScanner()
        result = scanner.build_result(tool_names=[], conflicts=[])
        assert isinstance(result, DetectionResult)
        assert result.detected_tools == ()
        assert result.conflicts == ()
        assert result.severity_map == {}

    def test_build_result_with_known_tools(self):
        scanner = ToolScanner()
        result = scanner.build_result(
            tool_names=["claude-code", "cursor"],
            conflicts=[],
        )
        assert len(result.detected_tools) == 2
        tools_by_name = {t.name: t for t in result.detected_tools}
        assert "claude-code" in tools_by_name
        assert "cursor" in tools_by_name

    def test_build_result_maps_known_config_paths(self):
        scanner = ToolScanner()
        result = scanner.build_result(tool_names=["claude-code"], conflicts=[])
        tool = result.detected_tools[0]
        assert tool.name == "claude-code"
        # config_path should be the known path for claude-code
        assert tool.config_path is not None
        assert ".claude" in str(tool.config_path)

    def test_build_result_with_unknown_tool(self):
        scanner = ToolScanner()
        result = scanner.build_result(tool_names=["unknown-tool"], conflicts=[])
        assert len(result.detected_tools) == 1
        assert result.detected_tools[0].name == "unknown-tool"
        assert result.detected_tools[0].config_path is None


class TestToolScannerConflictClassification:
    def test_classify_no_conflicts(self):
        scanner = ToolScanner()
        result = scanner.build_result(tool_names=["claude-code"], conflicts=[])
        assert result.severity_map == {}

    def test_classify_cursor_sqlite_as_warning(self):
        scanner = ToolScanner()
        conflict = "cursor: SQLite-based config detected, cannot read"
        result = scanner.build_result(tool_names=["cursor"], conflicts=[conflict])
        assert result.severity_map[conflict] == ConflictSeverity.WARNING

    def test_classify_contradiction_as_conflict(self):
        scanner = ToolScanner()
        conflict = "claude-code: global setting 'model' contradicts local value"
        result = scanner.build_result(
            tool_names=["claude-code"],
            conflicts=[conflict],
        )
        assert result.severity_map[conflict] == ConflictSeverity.CONFLICT

    def test_classify_compatible_setting(self):
        scanner = ToolScanner()
        conflict = "claude-code: global setting 'theme' is compatible with local"
        result = scanner.build_result(
            tool_names=["claude-code"],
            conflicts=[conflict],
        )
        assert result.severity_map[conflict] == ConflictSeverity.COMPATIBLE

    def test_classify_restriction_as_warning(self):
        scanner = ToolScanner()
        conflict = "opencode: global restriction on model access"
        result = scanner.build_result(
            tool_names=["opencode"],
            conflicts=[conflict],
        )
        assert result.severity_map[conflict] == ConflictSeverity.WARNING

    def test_classify_unknown_defaults_to_warning(self):
        scanner = ToolScanner()
        conflict = "some weird message that does not match keywords"
        result = scanner.build_result(
            tool_names=["claude-code"],
            conflicts=[conflict],
        )
        assert result.severity_map[conflict] == ConflictSeverity.WARNING


class TestToolScannerKnownToolRegistry:
    def test_known_tools_include_claude_code(self):
        scanner = ToolScanner()
        assert "claude-code" in scanner.known_tools

    def test_known_tools_include_cursor(self):
        scanner = ToolScanner()
        assert "cursor" in scanner.known_tools

    def test_known_tools_include_roo_code(self):
        scanner = ToolScanner()
        assert "roo-code" in scanner.known_tools

    def test_known_tools_include_opencode(self):
        scanner = ToolScanner()
        assert "opencode" in scanner.known_tools

    def test_known_tool_config_paths(self):
        scanner = ToolScanner()
        # Each known tool has a config path pattern
        for tool_name, config_path in scanner.known_tools.items():
            assert isinstance(config_path, Path), f"{tool_name} should have a Path config"


class TestToolScannerMultipleTools:
    def test_multiple_tools_with_mixed_conflicts(self):
        scanner = ToolScanner()
        conflicts = [
            "cursor: SQLite-based config detected, cannot read",
            "claude-code: global setting 'model' contradicts local value",
            "opencode: global setting 'theme' is compatible with local",
        ]
        result = scanner.build_result(
            tool_names=["claude-code", "cursor", "opencode"],
            conflicts=conflicts,
        )
        assert len(result.detected_tools) == 3
        assert len(result.conflicts) == 3
        assert result.severity_map[conflicts[0]] == ConflictSeverity.WARNING
        assert result.severity_map[conflicts[1]] == ConflictSeverity.CONFLICT
        assert result.severity_map[conflicts[2]] == ConflictSeverity.COMPATIBLE
