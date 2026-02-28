"""FilesystemToolScanner -- infrastructure adapter for ToolDetectionPort.

Scans the filesystem for installed AI coding tool configuration directories
and detects configuration conflicts. Handles permission errors and missing
directories gracefully.
"""

from __future__ import annotations

from pathlib import Path
from typing import ClassVar


class FilesystemToolScanner:
    """Filesystem-based implementation of ToolDetectionPort.

    Scans for known AI coding tool config directories under the user's
    home directory.

    Attributes:
        _home_dir: The home directory to scan for tool configs.
    """

    # Tool name -> config directory relative to home
    _TOOL_DIRS: ClassVar[dict[str, Path]] = {
        "claude-code": Path(".claude"),
        "cursor": Path(".cursor"),
        "roo-code": Path(".roo"),
        "opencode": Path(".config/opencode"),
    }

    def __init__(self, home_dir: Path | None = None) -> None:
        """Initialize with an optional home directory override.

        Args:
            home_dir: Home directory to scan. Defaults to Path.home().
        """
        self._home_dir = home_dir if home_dir is not None else Path.home()

    def detect(self, project_dir: Path) -> list[str]:
        """Detect installed AI coding tools by checking config directories.

        Args:
            project_dir: The project directory (unused for global detection,
                         kept for port interface compliance).

        Returns:
            List of detected tool identifiers.
        """
        if not self._home_dir.exists():
            return []

        detected: list[str] = []
        for tool_name, config_rel_path in self._TOOL_DIRS.items():
            config_path = self._home_dir / config_rel_path
            try:
                if config_path.exists():
                    detected.append(tool_name)
            except PermissionError:
                # Cannot access the directory; still report the tool as
                # detected since the directory is present
                detected.append(tool_name)
        return detected

    def scan_conflicts(self, project_dir: Path) -> list[str]:
        """Scan for configuration conflicts between global and local settings.

        Args:
            project_dir: The project directory to compare against global configs.

        Returns:
            List of conflict descriptions.
        """
        if not self._home_dir.exists():
            return []

        conflicts: list[str] = []

        # Cursor: SQLite-based config, we cannot read it
        cursor_dir = self._home_dir / self._TOOL_DIRS["cursor"]
        try:
            if cursor_dir.exists():
                conflicts.append("cursor: SQLite-based config detected, cannot read")
        except PermissionError:
            conflicts.append("cursor: config directory not readable (permission denied)")

        return conflicts
