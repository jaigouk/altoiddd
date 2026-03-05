"""Tests for DocReviewHandler (alty-2j7.8).

Tests the mark_reviewed, reviewable_docs, and _update_frontmatter logic.
"""

from __future__ import annotations

from datetime import date
from typing import TYPE_CHECKING

import pytest

from src.domain.models.doc_health import DocHealthStatus
from src.domain.models.errors import InvariantViolationError
from tests.helpers.doc_review_helpers import write_doc, write_registry

if TYPE_CHECKING:
    from pathlib import Path


# ---------------------------------------------------------------------------
# Tests: reviewable_docs (query which docs are due for review)
# ---------------------------------------------------------------------------


class TestReviewableDocs:
    """DocReviewHandler.reviewable_docs returns docs needing review."""

    def test_stale_doc_is_reviewable(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(tmp_path, [{"path": "docs/PRD.md", "review_interval_days": 7}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        docs = handler.reviewable_docs(project_dir=tmp_path)

        assert len(docs) == 1
        assert docs[0].path == "docs/PRD.md"
        assert docs[0].status == DocHealthStatus.STALE

    def test_fresh_doc_is_not_reviewable(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        today = date.today().isoformat()
        write_registry(tmp_path, [{"path": "docs/PRD.md", "review_interval_days": 30}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed=today)

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        docs = handler.reviewable_docs(project_dir=tmp_path)

        assert len(docs) == 0

    def test_no_frontmatter_is_reviewable(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(tmp_path, [{"path": "docs/DDD.md"}])
        write_doc(tmp_path, "docs/DDD.md")  # no frontmatter

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        docs = handler.reviewable_docs(project_dir=tmp_path)

        assert len(docs) == 1
        assert docs[0].status == DocHealthStatus.NO_FRONTMATTER

    def test_no_registry_raises(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        handler = DocReviewHandler(scanner=FilesystemDocScanner())

        with pytest.raises(InvariantViolationError, match=r"[Nn]o.*registry"):
            handler.reviewable_docs(project_dir=tmp_path)

    def test_missing_doc_excluded_from_reviewable(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(
            tmp_path,
            [
                {"path": "docs/PRD.md", "review_interval_days": 7},
                {"path": "docs/GONE.md", "review_interval_days": 7},
            ],
        )
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        docs = handler.reviewable_docs(project_dir=tmp_path)

        paths = [d.path for d in docs]
        assert "docs/PRD.md" in paths
        assert "docs/GONE.md" not in paths


# ---------------------------------------------------------------------------
# Tests: mark_reviewed (write command)
# ---------------------------------------------------------------------------


class TestMarkReviewed:
    """DocReviewHandler.mark_reviewed updates frontmatter last_reviewed."""

    def test_updates_last_reviewed_in_frontmatter(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        result = handler.mark_reviewed(
            doc_path="docs/PRD.md",
            project_dir=tmp_path,
            review_date=date(2026, 3, 5),
        )

        assert result.path == "docs/PRD.md"
        assert result.new_date == date(2026, 3, 5)

        content = (tmp_path / "docs/PRD.md").read_text()
        assert "last_reviewed: 2026-03-05" in content
        assert "2020-01-01" not in content

    def test_adds_frontmatter_when_missing(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(tmp_path, [{"path": "docs/DDD.md"}])
        write_doc(tmp_path, "docs/DDD.md")

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        result = handler.mark_reviewed(
            doc_path="docs/DDD.md",
            project_dir=tmp_path,
            review_date=date(2026, 3, 5),
        )

        assert result.new_date == date(2026, 3, 5)
        content = (tmp_path / "docs/DDD.md").read_text()
        assert "---" in content
        assert "last_reviewed: 2026-03-05" in content

    def test_untracked_doc_raises(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")

        handler = DocReviewHandler(scanner=FilesystemDocScanner())

        with pytest.raises(InvariantViolationError, match=r"[Nn]ot.*tracked"):
            handler.mark_reviewed(
                doc_path="docs/random.md",
                project_dir=tmp_path,
                review_date=date(2026, 3, 5),
            )

    def test_mark_all_updates_multiple_docs(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(
            tmp_path,
            [
                {"path": "docs/PRD.md", "review_interval_days": 7},
                {"path": "docs/DDD.md", "review_interval_days": 7},
            ],
        )
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")
        write_doc(tmp_path, "docs/DDD.md", last_reviewed="2020-01-01")

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        results = handler.mark_all_reviewed(
            project_dir=tmp_path,
            review_date=date(2026, 3, 5),
        )

        assert len(results) == 2
        paths = {r.path for r in results}
        assert paths == {"docs/PRD.md", "docs/DDD.md"}

        for rel in ["docs/PRD.md", "docs/DDD.md"]:
            content = (tmp_path / rel).read_text()
            assert "last_reviewed: 2026-03-05" in content

    def test_mark_reviewed_is_idempotent(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        today = date.today()
        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed=today.isoformat())

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        result = handler.mark_reviewed(
            doc_path="docs/PRD.md",
            project_dir=tmp_path,
            review_date=today,
        )

        assert result.new_date == today
        content = (tmp_path / "docs/PRD.md").read_text()
        assert f"last_reviewed: {today.isoformat()}" in content

    def test_defaults_to_today(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        result = handler.mark_reviewed(
            doc_path="docs/PRD.md",
            project_dir=tmp_path,
        )

        assert result.new_date == date.today()


# ---------------------------------------------------------------------------
# Tests: Path traversal (Fix #1)
# ---------------------------------------------------------------------------


class TestPathTraversalGuard:
    """mark_reviewed rejects paths that escape the project directory."""

    def test_path_traversal_rejected(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        # Register a malicious path in the TOML
        write_registry(tmp_path, [{"path": "../../etc/passwd"}])

        handler = DocReviewHandler(scanner=FilesystemDocScanner())

        with pytest.raises(InvariantViolationError, match=r"[Ee]scape"):
            handler.mark_reviewed(
                doc_path="../../etc/passwd",
                project_dir=tmp_path,
                review_date=date(2026, 3, 5),
            )

    def test_normal_relative_path_accepted(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(tmp_path, [{"path": "docs/nested/PRD.md"}])
        write_doc(tmp_path, "docs/nested/PRD.md", last_reviewed="2020-01-01")

        handler = DocReviewHandler(scanner=FilesystemDocScanner())
        result = handler.mark_reviewed(
            doc_path="docs/nested/PRD.md",
            project_dir=tmp_path,
            review_date=date(2026, 3, 5),
        )

        assert result.path == "docs/nested/PRD.md"


# ---------------------------------------------------------------------------
# Tests: OSError handling (Fix #3)
# ---------------------------------------------------------------------------


class TestOSErrorHandling:
    """mark_reviewed wraps file I/O errors in InvariantViolationError."""

    def test_unreadable_file_raises_descriptive_error(self, tmp_path: Path) -> None:
        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        # Create the file so it passes existence checks, then delete it
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")
        # Remove the file to trigger FileNotFoundError on read
        (tmp_path / "docs/PRD.md").unlink()

        handler = DocReviewHandler(scanner=FilesystemDocScanner())

        with pytest.raises(InvariantViolationError, match=r"[Cc]annot read"):
            handler.mark_reviewed(
                doc_path="docs/PRD.md",
                project_dir=tmp_path,
                review_date=date(2026, 3, 5),
            )

    def test_read_only_file_raises_descriptive_error(self, tmp_path: Path) -> None:

        from src.application.commands.doc_review_handler import DocReviewHandler
        from src.infrastructure.persistence.filesystem_doc_scanner import (
            FilesystemDocScanner,
        )

        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")
        doc_file = tmp_path / "docs/PRD.md"
        doc_file.chmod(0o444)

        handler = DocReviewHandler(scanner=FilesystemDocScanner())

        try:
            with pytest.raises(InvariantViolationError, match=r"[Cc]annot write"):
                handler.mark_reviewed(
                    doc_path="docs/PRD.md",
                    project_dir=tmp_path,
                    review_date=date(2026, 3, 5),
                )
        finally:
            doc_file.chmod(0o644)


# ---------------------------------------------------------------------------
# Tests: _update_frontmatter edge cases (Fix #7)
# ---------------------------------------------------------------------------


class TestUpdateFrontmatter:
    """Direct unit tests for _update_frontmatter."""

    def test_replaces_existing_last_reviewed(self) -> None:
        from src.application.commands.doc_review_handler import _update_frontmatter

        content = "---\nlast_reviewed: 2020-01-01\n---\n\n# Title\n"
        result = _update_frontmatter(content, date(2026, 3, 5))

        assert "last_reviewed: 2026-03-05" in result
        assert "2020-01-01" not in result

    def test_preserves_other_frontmatter_fields(self) -> None:
        from src.application.commands.doc_review_handler import _update_frontmatter

        content = "---\ntitle: My Doc\nlast_reviewed: 2020-01-01\nowner: pm\n---\n\n# Body\n"
        result = _update_frontmatter(content, date(2026, 3, 5))

        assert "title: My Doc" in result
        assert "owner: pm" in result
        assert "last_reviewed: 2026-03-05" in result
        assert "2020-01-01" not in result

    def test_adds_last_reviewed_to_existing_frontmatter(self) -> None:
        from src.application.commands.doc_review_handler import _update_frontmatter

        content = "---\ntitle: My Doc\nowner: pm\n---\n\n# Body\n"
        result = _update_frontmatter(content, date(2026, 3, 5))

        assert "last_reviewed: 2026-03-05" in result
        assert "title: My Doc" in result

    def test_creates_frontmatter_for_bare_content(self) -> None:
        from src.application.commands.doc_review_handler import _update_frontmatter

        content = "# Just a heading\n\nSome text.\n"
        result = _update_frontmatter(content, date(2026, 3, 5))

        assert result.startswith("---\n")
        assert "last_reviewed: 2026-03-05" in result
        assert "# Just a heading" in result

    def test_empty_file(self) -> None:
        from src.application.commands.doc_review_handler import _update_frontmatter

        result = _update_frontmatter("", date(2026, 3, 5))

        assert "last_reviewed: 2026-03-05" in result
        assert result.startswith("---\n")

    def test_hr_not_confused_with_frontmatter(self) -> None:
        from src.application.commands.doc_review_handler import _update_frontmatter

        content = "# Title\n\n---\n\nSome text after a horizontal rule.\n"
        result = _update_frontmatter(content, date(2026, 3, 5))

        # Should prepend new frontmatter, not mistake the HR for existing
        assert result.startswith("---\n")
        assert result.count("---") >= 3  # new fm open + new fm close + original HR

    def test_crlf_line_endings_preserved(self) -> None:
        from src.application.commands.doc_review_handler import _update_frontmatter

        content = "---\r\nlast_reviewed: 2020-01-01\r\n---\r\n\r\n# Title\r\n"
        result = _update_frontmatter(content, date(2026, 3, 5))

        assert "last_reviewed: 2026-03-05" in result
        assert "2020-01-01" not in result
        # Verify CRLF is preserved (no bare \n)
        assert "\r\n" in result
        lines = result.split("\r\n")
        # No line should contain a bare \n (after splitting on \r\n)
        assert all("\n" not in line for line in lines)

    def test_crlf_new_frontmatter_uses_crlf(self) -> None:
        from src.application.commands.doc_review_handler import _update_frontmatter

        content = "# Title\r\n\r\nBody text.\r\n"
        result = _update_frontmatter(content, date(2026, 3, 5))

        assert "last_reviewed: 2026-03-05" in result
        assert "\r\n" in result
