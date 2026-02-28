"""Port for the Tool Translation bounded context (tool detection).

Defines the interface for detecting installed AI coding tools and
scanning for configuration conflicts in a project directory.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class ToolDetectionPort(Protocol):
    """Interface for detecting AI coding tools.

    Adapters implement this to scan a project directory for installed
    AI coding tools and identify configuration conflicts.
    """

    def detect(self, project_dir: Path) -> list[str]:
        """Detect installed AI coding tools in the project directory.

        Args:
            project_dir: The project directory to scan.

        Returns:
            List of detected tool identifiers (e.g., ["claude", "cursor"]).
        """
        ...

    def scan_conflicts(self, project_dir: Path) -> list[str]:
        """Scan for configuration conflicts between detected tools.

        Args:
            project_dir: The project directory to scan.

        Returns:
            List of conflict descriptions, empty if no conflicts found.
        """
        ...
