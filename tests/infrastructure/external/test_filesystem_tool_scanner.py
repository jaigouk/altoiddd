"""Tests for FilesystemToolScanner infrastructure adapter.

Verifies filesystem-based detection of AI coding tools and conflict scanning.
Uses tmp_path fixtures to avoid touching the real filesystem.
"""

from __future__ import annotations

from src.application.ports.tool_detection_port import ToolDetectionPort
from src.infrastructure.external.filesystem_tool_scanner import FilesystemToolScanner


class TestFilesystemToolScannerProtocol:
    def test_implements_tool_detection_port(self):
        scanner = FilesystemToolScanner()
        assert isinstance(scanner, ToolDetectionPort)


class TestFilesystemToolScannerDetect:
    def test_no_tools_installed(self, tmp_path):
        """No tool directories exist -> empty list."""
        scanner = FilesystemToolScanner(home_dir=tmp_path)
        result = scanner.detect(tmp_path)
        assert result == []

    def test_detect_claude_code(self, tmp_path):
        """~/.claude/ directory exists -> claude-code detected."""
        (tmp_path / ".claude").mkdir()
        scanner = FilesystemToolScanner(home_dir=tmp_path)
        result = scanner.detect(tmp_path)
        assert "claude-code" in result

    def test_detect_cursor(self, tmp_path):
        """Cursor SQLite DB presence -> cursor detected."""
        # Cursor uses a SQLite DB in the home directory
        cursor_dir = tmp_path / ".cursor"
        cursor_dir.mkdir()
        scanner = FilesystemToolScanner(home_dir=tmp_path)
        result = scanner.detect(tmp_path)
        assert "cursor" in result

    def test_detect_roo_code(self, tmp_path):
        """~/.roo/ directory exists -> roo-code detected."""
        (tmp_path / ".roo").mkdir()
        scanner = FilesystemToolScanner(home_dir=tmp_path)
        result = scanner.detect(tmp_path)
        assert "roo-code" in result

    def test_detect_opencode(self, tmp_path):
        """~/.config/opencode/ directory exists -> opencode detected."""
        (tmp_path / ".config" / "opencode").mkdir(parents=True)
        scanner = FilesystemToolScanner(home_dir=tmp_path)
        result = scanner.detect(tmp_path)
        assert "opencode" in result

    def test_detect_multiple_tools(self, tmp_path):
        """Multiple tool directories -> all detected."""
        (tmp_path / ".claude").mkdir()
        (tmp_path / ".roo").mkdir()
        scanner = FilesystemToolScanner(home_dir=tmp_path)
        result = scanner.detect(tmp_path)
        assert "claude-code" in result
        assert "roo-code" in result
        assert len(result) == 2

    def test_detect_tool_dir_exists_but_empty(self, tmp_path):
        """Tool directory exists but is empty -> still detected."""
        (tmp_path / ".claude").mkdir()
        scanner = FilesystemToolScanner(home_dir=tmp_path)
        result = scanner.detect(tmp_path)
        assert "claude-code" in result


class TestFilesystemToolScannerConflicts:
    def test_no_conflicts_when_no_tools(self, tmp_path):
        scanner = FilesystemToolScanner(home_dir=tmp_path)
        result = scanner.scan_conflicts(tmp_path)
        assert result == []

    def test_cursor_sqlite_produces_warning(self, tmp_path):
        """Cursor dir exists -> warning about SQLite-based config."""
        (tmp_path / ".cursor").mkdir()
        scanner = FilesystemToolScanner(home_dir=tmp_path)
        result = scanner.scan_conflicts(tmp_path)
        assert any("cursor" in c.lower() and "sqlite" in c.lower() for c in result)


class TestFilesystemToolScannerEdgeCases:
    def test_permission_denied_handled_gracefully(self, tmp_path):
        """Unreadable directory -> graceful handling, no crash."""
        unreadable = tmp_path / ".claude"
        unreadable.mkdir()
        unreadable.chmod(0o000)
        try:
            scanner = FilesystemToolScanner(home_dir=tmp_path)
            # Should not raise, tool may or may not appear depending on
            # whether we can stat the directory
            result = scanner.detect(tmp_path)
            assert isinstance(result, list)
        finally:
            # Restore permissions so pytest can clean up
            unreadable.chmod(0o755)

    def test_home_directory_not_set(self, tmp_path):
        """If home dir doesn't exist -> graceful empty result."""
        nonexistent = tmp_path / "does_not_exist"
        scanner = FilesystemToolScanner(home_dir=nonexistent)
        result = scanner.detect(tmp_path)
        assert result == []

    def test_default_home_dir_uses_real_home(self):
        """When no home_dir passed, should use the actual home directory."""
        scanner = FilesystemToolScanner()
        # Just verify it can be created without error; we don't
        # want to assert on the actual home directory contents
        assert scanner is not None
