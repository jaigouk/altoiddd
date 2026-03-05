"""Command handler for marking documents as reviewed (alty-2j7.8).

DocReviewHandler orchestrates: load registry, identify stale docs,
update frontmatter last_reviewed dates on disk.
"""

from __future__ import annotations

import re
from datetime import date
from typing import TYPE_CHECKING

from src.domain.models.doc_health import (
    DocHealthStatus,
    DocReviewResult,
)
from src.domain.models.errors import InvariantViolationError

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.queries.doc_health_handler import DocScannerProtocol
    from src.domain.models.doc_health import DocStatus

_FRONTMATTER_RE = re.compile(r"^(---[ \t]*\n)(.*?)(\n---)", re.DOTALL)
_LAST_REVIEWED_RE = re.compile(r"^(last_reviewed:\s*)(.+)$", re.MULTILINE)


def _update_frontmatter(content: str, review_date: date) -> str:
    """Update or insert last_reviewed in markdown frontmatter.

    If frontmatter exists and has last_reviewed, replace the value.
    If frontmatter exists but lacks last_reviewed, add it.
    If no frontmatter exists, prepend a new frontmatter block.

    Handles both LF and CRLF line endings — preserves the original style.

    Args:
        content: Full text content of the markdown file.
        review_date: The date to set as last_reviewed.

    Returns:
        Updated file content with new last_reviewed date.
    """
    date_str = review_date.isoformat()

    # Detect line ending style
    eol = "\r\n" if "\r\n" in content else "\n"

    # Normalize to LF for regex processing
    normalized = content.replace("\r\n", "\n") if eol == "\r\n" else content

    fm_match = _FRONTMATTER_RE.search(normalized)

    if fm_match:
        frontmatter = fm_match.group(2)
        if _LAST_REVIEWED_RE.search(frontmatter):
            new_fm = _LAST_REVIEWED_RE.sub(
                rf"\g<1>{date_str}", frontmatter
            )
        else:
            new_fm = frontmatter + f"\nlast_reviewed: {date_str}"
        result = (
            normalized[: fm_match.start(2)]
            + new_fm
            + normalized[fm_match.end(2) :]
        )
    else:
        result = f"---\nlast_reviewed: {date_str}\n---\n{normalized}"

    # Restore original line ending style
    if eol == "\r\n":
        result = result.replace("\n", "\r\n")

    return result


def _validate_path_within_project(doc_path: str, project_dir: Path) -> None:
    """Ensure the resolved path does not escape project_dir.

    Raises:
        InvariantViolationError: If the path escapes the project directory.
    """
    resolved = (project_dir / doc_path).resolve()
    project_resolved = project_dir.resolve()
    if not str(resolved).startswith(str(project_resolved) + "/") and resolved != project_resolved:
        msg = f"Path escapes project directory: {doc_path}"
        raise InvariantViolationError(msg)


class DocReviewHandler:
    """Orchestrates document review operations.

    Loads the doc registry, identifies stale/reviewable docs,
    and writes updated last_reviewed dates to frontmatter.
    """

    def __init__(self, scanner: DocScannerProtocol) -> None:
        self._scanner = scanner

    def _load_registry_entries(
        self, project_dir: Path
    ) -> tuple[DocStatus, ...]:
        """Load registry and scan, raising if no registry found."""
        registry_path = project_dir / ".alty" / "maintenance" / "doc-registry.toml"
        entries = self._scanner.load_registry(registry_path)

        if not entries:
            msg = "No doc-registry found. Run alty init first."
            raise InvariantViolationError(msg)

        return self._scanner.scan_registered(entries, project_dir)

    def _tracked_paths(self, project_dir: Path) -> frozenset[str]:
        """Return the set of paths tracked in the registry."""
        registry_path = project_dir / ".alty" / "maintenance" / "doc-registry.toml"
        entries = self._scanner.load_registry(registry_path)
        return frozenset(e.path for e in entries)

    def reviewable_docs(self, project_dir: Path) -> tuple[DocStatus, ...]:
        """Return docs that are due for review (STALE or NO_FRONTMATTER).

        MISSING docs are excluded — they can't be reviewed if they don't exist.

        Args:
            project_dir: The project root directory.

        Returns:
            Tuple of DocStatus for reviewable documents.

        Raises:
            InvariantViolationError: If no doc-registry.toml is found.
        """
        statuses = self._load_registry_entries(project_dir)
        reviewable_states = {DocHealthStatus.STALE, DocHealthStatus.NO_FRONTMATTER}
        return tuple(s for s in statuses if s.status in reviewable_states)

    def mark_reviewed(
        self,
        doc_path: str,
        project_dir: Path,
        review_date: date | None = None,
    ) -> DocReviewResult:
        """Mark a single document as reviewed by updating its frontmatter.

        Args:
            doc_path: Relative path to the document (e.g. "docs/PRD.md").
            project_dir: The project root directory.
            review_date: Date to stamp; defaults to today.

        Returns:
            DocReviewResult with the path and new date.

        Raises:
            InvariantViolationError: If the doc is not tracked, the path
                escapes the project directory, or the file cannot be read.
        """
        tracked = self._tracked_paths(project_dir)
        if doc_path not in tracked:
            msg = f"Not tracked in doc-registry: {doc_path}. Add it first."
            raise InvariantViolationError(msg)

        _validate_path_within_project(doc_path, project_dir)

        effective_date = review_date if review_date is not None else date.today()
        file_path = project_dir / doc_path

        try:
            content = file_path.read_text()
        except OSError as exc:
            msg = f"Cannot read {doc_path}: {exc}"
            raise InvariantViolationError(msg) from exc

        updated = _update_frontmatter(content, effective_date)

        try:
            file_path.write_text(updated)
        except OSError as exc:
            msg = f"Cannot write {doc_path}: {exc}"
            raise InvariantViolationError(msg) from exc

        return DocReviewResult(path=doc_path, new_date=effective_date)

    def mark_all_reviewed(
        self,
        project_dir: Path,
        review_date: date | None = None,
    ) -> tuple[DocReviewResult, ...]:
        """Mark all stale/reviewable docs as reviewed.

        Args:
            project_dir: The project root directory.
            review_date: Date to stamp; defaults to today.

        Returns:
            Tuple of DocReviewResult for each updated doc.
        """
        reviewable = self.reviewable_docs(project_dir)
        effective_date = review_date if review_date is not None else date.today()
        results: list[DocReviewResult] = []
        for doc in reviewable:
            _validate_path_within_project(doc.path, project_dir)
            file_path = project_dir / doc.path
            try:
                content = file_path.read_text()
            except OSError as exc:
                msg = f"Cannot read {doc.path}: {exc}"
                raise InvariantViolationError(msg) from exc
            updated = _update_frontmatter(content, effective_date)
            try:
                file_path.write_text(updated)
            except OSError as exc:
                msg = f"Cannot write {doc.path}: {exc}"
                raise InvariantViolationError(msg) from exc
            results.append(DocReviewResult(path=doc.path, new_date=effective_date))
        return tuple(results)
