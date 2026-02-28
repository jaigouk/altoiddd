"""Port for document review operations.

Defines the interface for marking documentation as reviewed and
tracking review metadata.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class DocReviewPort(Protocol):
    """Interface for document review operations.

    Adapters implement this to mark documentation as reviewed and
    record review timestamps and reviewer information.
    """

    def mark_reviewed(self, doc_path: Path, reviewer: str) -> str:
        """Mark a document as reviewed.

        Args:
            doc_path: Path to the document being reviewed.
            reviewer: Identifier for the reviewer.

        Returns:
            Confirmation message with review details.
        """
        ...

    def review_status(self, project_dir: Path) -> str:
        """Get the review status of all project documentation.

        Args:
            project_dir: The project directory containing docs to check.

        Returns:
            A report of review status for each document.
        """
        ...
