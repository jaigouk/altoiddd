"""Edge case tests for doc-health CLI command (2j7.7 QA).

BICEP analysis uncovered:
- Boundary: project with doc-registry referencing nonexistent docs
- Boundary: doc with future last_reviewed date
- Error: doc-registry.toml with invalid TOML
- Cross-check: output format contains expected status keywords
- Boundary: project with only one doc present out of defaults
"""

from __future__ import annotations

from datetime import date, timedelta
from typing import TYPE_CHECKING

from typer.testing import CliRunner

if TYPE_CHECKING:
    from pathlib import Path

from src.infrastructure.cli.main import app

runner = CliRunner()


# ── Helpers ────────────────────────────────────────────────


def _create_doc_with_frontmatter(path: Path, days_ago: int) -> None:
    """Create a markdown doc with last_reviewed frontmatter."""
    path.parent.mkdir(parents=True, exist_ok=True)
    reviewed_date = date.today() - timedelta(days=days_ago)
    path.write_text(
        f"---\nlast_reviewed: {reviewed_date.isoformat()}\n---\n# Document\nContent here.\n"
    )


def _create_registry(project_dir: Path, doc_entries: list[dict[str, str | int]]) -> None:
    """Create a doc-registry.toml with [[docs]] entries."""
    registry_dir = project_dir / ".alty" / "maintenance"
    registry_dir.mkdir(parents=True, exist_ok=True)
    lines: list[str] = []
    for entry in doc_entries:
        lines.append("[[docs]]")
        for k, v in entry.items():
            if isinstance(v, int):
                lines.append(f"{k} = {v}")
            else:
                lines.append(f'{k} = "{v}"')
        lines.append("")
    (registry_dir / "doc-registry.toml").write_text("\n".join(lines))


# ── Tests ────────────────────────────────────────────────


class TestDocHealthRegistryReferencesNonexistent:
    """Registry points to docs that do not exist."""

    def test_registry_with_nonexistent_docs_shows_missing(self, tmp_path: Path) -> None:
        """When registry references docs that don't exist, they show MISSING."""
        _create_registry(
            tmp_path,
            [
                {"path": "docs/PRD.md"},
                {"path": "docs/NONEXISTENT.md"},
            ],
        )
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "MISSING" in result.output
        assert result.exit_code == 1


class TestDocHealthFutureDate:
    """Doc with last_reviewed in the future."""

    def test_future_reviewed_date_shows_ok(self, tmp_path: Path) -> None:
        """A doc reviewed 'in the future' should still show OK (0 days since)."""
        future_date = date.today() + timedelta(days=10)
        doc_path = tmp_path / "docs" / "PRD.md"
        doc_path.parent.mkdir(parents=True, exist_ok=True)
        doc_path.write_text(
            f"---\nlast_reviewed: {future_date.isoformat()}\n---\n# PRD\nContent.\n"
        )
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert result.exit_code == 0


class TestDocHealthInvalidRegistry:
    """Registry file with invalid/corrupt content."""

    def test_corrupted_registry_falls_back_to_defaults(self, tmp_path: Path) -> None:
        """When registry is invalid TOML, fall back to default docs."""
        registry_dir = tmp_path / ".alty" / "maintenance"
        registry_dir.mkdir(parents=True, exist_ok=True)
        (registry_dir / "doc-registry.toml").write_text("this is not valid toml {{{}}")

        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        # Should fall back to defaults since corrupt registry produces no entries
        assert result.exit_code == 0
        assert "docs/PRD.md" in result.output


class TestDocHealthOutputFormat:
    """Verify output format matches expected patterns."""

    def test_output_contains_separator_line(self, tmp_path: Path) -> None:
        """Output should have a separator line (horizontal rule)."""
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        # Should contain Unicode horizontal line separator
        assert "\u2500" in result.output

    def test_ok_docs_show_days_since(self, tmp_path: Path) -> None:
        """OK docs should include 'reviewed X days ago' in output."""
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "5 days ago" in result.output

    def test_stale_docs_show_interval(self, tmp_path: Path) -> None:
        """STALE docs should include the review interval in output."""
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=45)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "interval:" in result.output


class TestDocHealthCustomRegistryInterval:
    """Custom review intervals in registry."""

    def test_custom_short_interval_causes_stale(self, tmp_path: Path) -> None:
        """A doc reviewed 10 days ago with a 7-day interval is STALE."""
        _create_registry(
            tmp_path,
            [{"path": "docs/PRD.md", "review_interval_days": 7}],
        )
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=10)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "STALE" in result.output
        assert result.exit_code == 1

    def test_custom_long_interval_keeps_ok(self, tmp_path: Path) -> None:
        """A doc reviewed 45 days ago with a 90-day interval is OK."""
        _create_registry(
            tmp_path,
            [{"path": "docs/PRD.md", "review_interval_days": 90}],
        )
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=45)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "OK" in result.output
        assert result.exit_code == 0
