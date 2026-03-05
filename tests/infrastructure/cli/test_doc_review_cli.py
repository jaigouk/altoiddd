"""Tests for the alty doc-review CLI command (alty-2j7.8).

Verifies that the doc-review command:
- Shows reviewable docs when invoked without args
- Marks a specific doc as reviewed
- Marks all stale docs with --all
- Shows confirmation after writing
- Exits 1 when no registry found
- Exits 1 when doc not tracked
- Rejects doc_path + --all together
"""

from __future__ import annotations

from datetime import date
from typing import TYPE_CHECKING
from unittest.mock import patch

from typer.testing import CliRunner

from tests.helpers.doc_review_helpers import write_doc, write_registry

if TYPE_CHECKING:
    from pathlib import Path

from src.infrastructure.cli.main import app

runner = CliRunner()


# ---------------------------------------------------------------------------
# Tests: Display reviewable docs (no args)
# ---------------------------------------------------------------------------


class TestDocReviewDisplaysReviewable:
    """doc-review (no args) shows docs due for review."""

    def test_shows_stale_docs(self, tmp_path: Path) -> None:
        write_registry(
            tmp_path, [{"path": "docs/PRD.md", "review_interval_days": 7}]
        )
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")

        with patch("src.infrastructure.cli.main._get_project_dir", return_value=tmp_path):
            result = runner.invoke(app, ["doc-review"], catch_exceptions=False)

        assert result.exit_code == 0
        assert "docs/PRD.md" in result.output

    def test_shows_no_docs_when_all_fresh(self, tmp_path: Path) -> None:
        today = date.today().isoformat()
        write_registry(
            tmp_path, [{"path": "docs/PRD.md", "review_interval_days": 30}]
        )
        write_doc(tmp_path, "docs/PRD.md", last_reviewed=today)

        with patch("src.infrastructure.cli.main._get_project_dir", return_value=tmp_path):
            result = runner.invoke(app, ["doc-review"], catch_exceptions=False)

        assert result.exit_code == 0
        assert "all fresh" in result.output.lower() or "no docs" in result.output.lower()


# ---------------------------------------------------------------------------
# Tests: Mark specific doc
# ---------------------------------------------------------------------------


class TestDocReviewMarkSpecific:
    """doc-review <path> marks a specific doc as reviewed."""

    def test_marks_specific_doc(self, tmp_path: Path) -> None:
        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")

        with patch("src.infrastructure.cli.main._get_project_dir", return_value=tmp_path):
            result = runner.invoke(
                app, ["doc-review", "docs/PRD.md"], catch_exceptions=False
            )

        assert result.exit_code == 0
        assert "docs/PRD.md" in result.output
        assert "reviewed" in result.output.lower()

        content = (tmp_path / "docs/PRD.md").read_text()
        assert f"last_reviewed: {date.today().isoformat()}" in content

    def test_untracked_doc_exits_with_error(self, tmp_path: Path) -> None:
        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")

        with patch("src.infrastructure.cli.main._get_project_dir", return_value=tmp_path):
            result = runner.invoke(app, ["doc-review", "docs/random.md"])

        assert result.exit_code == 1
        assert "not" in result.output.lower()
        assert "track" in result.output.lower()


# ---------------------------------------------------------------------------
# Tests: --all flag
# ---------------------------------------------------------------------------


class TestDocReviewMarkAll:
    """doc-review --all marks all stale docs as reviewed."""

    def test_marks_all_stale_docs(self, tmp_path: Path) -> None:
        write_registry(
            tmp_path,
            [
                {"path": "docs/PRD.md", "review_interval_days": 7},
                {"path": "docs/DDD.md", "review_interval_days": 7},
            ],
        )
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")
        write_doc(tmp_path, "docs/DDD.md", last_reviewed="2020-01-01")

        with patch("src.infrastructure.cli.main._get_project_dir", return_value=tmp_path):
            result = runner.invoke(
                app, ["doc-review", "--all"], catch_exceptions=False
            )

        assert result.exit_code == 0
        assert "docs/PRD.md" in result.output
        assert "docs/DDD.md" in result.output

        for rel in ["docs/PRD.md", "docs/DDD.md"]:
            content = (tmp_path / rel).read_text()
            assert f"last_reviewed: {date.today().isoformat()}" in content

    def test_all_flag_with_nothing_stale(self, tmp_path: Path) -> None:
        today = date.today().isoformat()
        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed=today)

        with patch("src.infrastructure.cli.main._get_project_dir", return_value=tmp_path):
            result = runner.invoke(
                app, ["doc-review", "--all"], catch_exceptions=False
            )

        assert result.exit_code == 0
        assert "all fresh" in result.output.lower() or "no docs" in result.output.lower()


# ---------------------------------------------------------------------------
# Tests: No registry
# ---------------------------------------------------------------------------


class TestDocReviewNoRegistry:
    """doc-review with no registry exits with error."""

    def test_no_registry_shows_error(self, tmp_path: Path) -> None:
        with patch("src.infrastructure.cli.main._get_project_dir", return_value=tmp_path):
            result = runner.invoke(app, ["doc-review"])

        assert result.exit_code == 1
        assert "registry" in result.output.lower()


# ---------------------------------------------------------------------------
# Tests: doc_path + --all conflict (Fix #6)
# ---------------------------------------------------------------------------


class TestDocReviewConflictingArgs:
    """doc-review rejects doc_path and --all used together."""

    def test_doc_path_and_all_exits_with_error(self, tmp_path: Path) -> None:
        write_registry(tmp_path, [{"path": "docs/PRD.md"}])
        write_doc(tmp_path, "docs/PRD.md", last_reviewed="2020-01-01")

        with patch("src.infrastructure.cli.main._get_project_dir", return_value=tmp_path):
            result = runner.invoke(app, ["doc-review", "docs/PRD.md", "--all"])

        assert result.exit_code == 1
        assert "cannot" in result.output.lower()
