"""Port for writing files to the filesystem.

Defines the interface for file output so that the application layer
remains decoupled from concrete file I/O.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class FileWriterPort(Protocol):
    """Interface for writing files.

    Adapters implement this to write generated artifacts to disk,
    allowing the application layer to remain infrastructure-agnostic.
    """

    def write_file(self, path: Path, content: str) -> None:
        """Write content to a file at the given path.

        Args:
            path: Target file path.
            content: File content to write.
        """
        ...
