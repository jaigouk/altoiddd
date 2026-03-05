"""Port for document review operations.

Defines the interface for marking documentation as reviewed and
querying which docs are due for review.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from datetime import date
    from pathlib import Path

    from src.domain.models.doc_health import DocReviewResult, DocStatus


@runtime_checkable
class DocReviewPort(Protocol):
    """Interface for document review operations.

    Adapters implement this to identify reviewable documents,
    mark them as reviewed, and batch-update all stale docs.
    """

    def reviewable_docs(self, project_dir: Path) -> tuple[DocStatus, ...]:
        """Return docs that are due for review.

        Args:
            project_dir: The project root directory.

        Returns:
            Tuple of DocStatus for reviewable documents.
        """
        ...

    def mark_reviewed(
        self,
        doc_path: str,
        project_dir: Path,
        review_date: date | None = None,
    ) -> DocReviewResult:
        """Mark a document as reviewed.

        Args:
            doc_path: Relative path to the document.
            project_dir: The project root directory.
            review_date: Date to stamp; defaults to today.

        Returns:
            DocReviewResult with the path and new date.
        """
        ...

    def mark_all_reviewed(
        self,
        project_dir: Path,
        review_date: date | None = None,
    ) -> tuple[DocReviewResult, ...]:
        """Mark all stale docs as reviewed.

        Args:
            project_dir: The project root directory.
            review_date: Date to stamp; defaults to today.

        Returns:
            Tuple of DocReviewResult for each updated doc.
        """
        ...
