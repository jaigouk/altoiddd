"""Domain models for document health checking.

DocHealthStatus enumerates possible doc states.
DocRegistryEntry records a doc to track with its review interval.
DocStatus captures the health of a single document.
DocHealthReport aggregates statuses into a summary.
create_doc_status is a factory that auto-calculates status and days_since.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass
from datetime import date

from src.domain.models.errors import InvariantViolationError


class DocHealthStatus(enum.Enum):
    """Possible health states for a tracked document."""

    OK = "ok"
    STALE = "stale"
    MISSING = "missing"
    NO_FRONTMATTER = "no_frontmatter"


@dataclass(frozen=True)
class DocRegistryEntry:
    """A document registered for health tracking.

    Attributes:
        path: Relative path to the document from project root.
        owner: Optional owner or team responsible for the document.
        review_interval_days: Maximum days between reviews before staleness.
    """

    path: str
    owner: str | None = None
    review_interval_days: int = 30

    def __post_init__(self) -> None:
        if self.review_interval_days <= 0:
            msg = (
                f"review_interval_days must be positive, got {self.review_interval_days}"
            )
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class DocStatus:
    """Health status of a single document.

    Attributes:
        path: Relative path to the document.
        status: Current health status.
        last_reviewed: Date the document was last reviewed, if known.
        days_since: Days since last review, if known.
        review_interval_days: The review interval for this document.
        owner: Optional owner of this document.
    """

    path: str
    status: DocHealthStatus
    last_reviewed: date | None = None
    days_since: int | None = None
    review_interval_days: int = 30
    owner: str | None = None


@dataclass(frozen=True)
class DocHealthReport:
    """Aggregate report of document health statuses.

    Attributes:
        statuses: Tuple of individual document statuses.
    """

    statuses: tuple[DocStatus, ...]

    @property
    def issue_count(self) -> int:
        """Count of documents with non-OK status."""
        return sum(1 for s in self.statuses if s.status != DocHealthStatus.OK)

    @property
    def total_checked(self) -> int:
        """Total number of documents checked."""
        return len(self.statuses)

    @property
    def has_issues(self) -> bool:
        """Whether any document has a non-OK status."""
        return self.issue_count > 0


def create_doc_status(
    path: str,
    exists: bool,
    last_reviewed: date | None,
    review_interval_days: int = 30,
    owner: str | None = None,
    today: date | None = None,
) -> DocStatus:
    """Create DocStatus with auto-calculated status and days_since.

    Args:
        path: Relative path to the document.
        exists: Whether the file exists on disk.
        last_reviewed: Date of last review, or None if unknown.
        review_interval_days: Maximum days between reviews.
        owner: Optional owner of this document.

    Returns:
        A DocStatus with the appropriate status and computed days_since.
    """
    if not exists:
        return DocStatus(
            path=path,
            status=DocHealthStatus.MISSING,
            review_interval_days=review_interval_days,
            owner=owner,
        )
    if last_reviewed is None:
        return DocStatus(
            path=path,
            status=DocHealthStatus.NO_FRONTMATTER,
            review_interval_days=review_interval_days,
            owner=owner,
        )
    effective_today = today if today is not None else date.today()
    days = (effective_today - last_reviewed).days
    status = DocHealthStatus.OK if days <= review_interval_days else DocHealthStatus.STALE
    return DocStatus(
        path=path,
        status=status,
        last_reviewed=last_reviewed,
        days_since=days,
        review_interval_days=review_interval_days,
        owner=owner,
    )
