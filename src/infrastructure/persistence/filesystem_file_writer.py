"""Filesystem implementation of FileWriterPort.

Writes files to disk with automatic parent directory creation.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path


class FilesystemFileWriter:
    """Writes files to the local filesystem.

    Implements FileWriterPort by writing content to the given path,
    creating parent directories as needed.
    """

    def write_file(self, path: Path, content: str) -> None:
        """Write content to a file at the given path.

        Creates parent directories if they don't exist. Overwrites
        existing files. Lets OSError propagate on permission failures.

        Args:
            path: Target file path.
            content: File content to write.
        """
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")
