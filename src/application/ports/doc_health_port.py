"""Port for document health checking.

Defines the interface for checking the health and consistency of
project documentation and knowledge base entries.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class DocHealthPort(Protocol):
    """Interface for document health check operations.

    Adapters implement this to verify that project documentation is
    complete, consistent, and up to date.
    """

    def check(self, project_dir: Path) -> str:
        """Check the health of project documentation.

        Args:
            project_dir: The project directory containing docs to check.

        Returns:
            A health report for the project documentation.
        """
        ...

    def check_knowledge(self, knowledge_dir: Path) -> str:
        """Check the health of knowledge base entries.

        Args:
            knowledge_dir: The knowledge base directory to check.

        Returns:
            A health report for the knowledge base.
        """
        ...
