"""Tests for the alty doc-health CLI command wiring.

Verifies that the doc-health command:
- Displays a formatted report of document health statuses
- Shows status icons for OK, STALE, MISSING, and NO_FRONTMATTER
- Displays a summary with total checked and issues found
- Exits 0 when no issues, exits 1 when issues exist
- Falls back to defaults when no doc-registry.toml exists
"""

from __future__ import annotations

from datetime import date, timedelta
from typing import TYPE_CHECKING

from typer.testing import CliRunner

if TYPE_CHECKING:
    from pathlib import Path

from src.infrastructure.cli.main import app

runner = CliRunner()


class TestDocHealthDisplaysReport:
    """doc-health command produces output with doc statuses."""

    def test_doc_health_displays_report_header(self, tmp_path: Path) -> None:
        """Command output includes a report header."""
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "Doc Health Report" in result.output

    def test_doc_health_shows_summary(self, tmp_path: Path) -> None:
        """Command output includes a summary line with counts."""
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "checked" in result.output
        assert "issue" in result.output.lower()


class TestDocHealthOkStatus:
    """OK status displayed for fresh docs."""

    def test_doc_health_shows_ok_for_healthy_docs(self, tmp_path: Path) -> None:
        """Docs reviewed recently show OK status."""
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert result.exit_code == 0
        assert "OK" in result.output


class TestDocHealthStaleStatus:
    """STALE status displayed for old docs."""

    def test_doc_health_shows_stale_for_old_docs(self, tmp_path: Path) -> None:
        """Docs not reviewed within interval show STALE status."""
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=45)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=45)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=45)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "STALE" in result.output


class TestDocHealthMissingStatus:
    """MISSING status displayed for absent docs."""

    def test_doc_health_shows_missing_for_absent_docs(self, tmp_path: Path) -> None:
        """Registered docs that do not exist show MISSING status."""
        # Create an empty project dir -- defaults will reference PRD, DDD, ARCHITECTURE
        # which are all missing
        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "MISSING" in result.output


class TestDocHealthNoFrontmatterStatus:
    """NO_FRONTMATTER status displayed for docs without frontmatter."""

    def test_doc_health_shows_no_frontmatter(self, tmp_path: Path) -> None:
        """Docs existing but without frontmatter show NO_FRONTMATTER."""
        doc_path = tmp_path / "docs" / "PRD.md"
        doc_path.parent.mkdir(parents=True, exist_ok=True)
        doc_path.write_text("# PRD\nNo frontmatter here.")
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        # Create a registry pointing to this doc
        _create_registry(tmp_path, ["docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md"])

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert "NO_FRONTMATTER" in result.output


class TestDocHealthNoRegistry:
    """Graceful fallback without .alty/maintenance/doc-registry.toml."""

    def test_doc_health_no_registry_uses_defaults(self, tmp_path: Path) -> None:
        """Without a registry file, default docs (PRD, DDD, ARCHITECTURE) are checked."""
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert result.exit_code == 0
        # All three default docs should appear in the output
        assert "docs/PRD.md" in result.output
        assert "docs/DDD.md" in result.output
        assert "docs/ARCHITECTURE.md" in result.output


class TestDocHealthExitCodes:
    """Exit code reflects whether issues were found."""

    def test_doc_health_exit_code_zero_when_clean(self, tmp_path: Path) -> None:
        """Exit code 0 when all docs are healthy."""
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "DDD.md", days_ago=5)
        _create_doc_with_frontmatter(tmp_path / "docs" / "ARCHITECTURE.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert result.exit_code == 0

    def test_doc_health_exit_code_one_when_issues(self, tmp_path: Path) -> None:
        """Exit code 1 when issues are found."""
        # Only create one doc -- the other two defaults will be MISSING
        _create_doc_with_frontmatter(tmp_path / "docs" / "PRD.md", days_ago=5)

        result = runner.invoke(
            app, ["doc-health", str(tmp_path)], catch_exceptions=False
        )

        assert result.exit_code == 1


class TestDocHealthProjectDirHandling:
    """Handle project_dir argument edge cases."""

    def test_doc_health_defaults_to_cwd(self) -> None:
        """Running doc-health without a directory argument should not crash."""
        result = runner.invoke(app, ["doc-health"], catch_exceptions=False)

        # Should complete without unhandled exception -- exit code 0 or 1 both acceptable
        assert result.exit_code in (0, 1)

    def test_doc_health_nonexistent_dir(self, tmp_path: Path) -> None:
        """Running doc-health on non-existent directory should not crash."""
        bogus = tmp_path / "nonexistent"
        result = runner.invoke(
            app, ["doc-health", str(bogus)], catch_exceptions=False
        )

        # Should still produce output without crashing
        assert result.exit_code in (0, 1)


# ── Helpers ────────────────────────────────────────────────


def _create_doc_with_frontmatter(path: Path, days_ago: int) -> None:
    """Create a markdown doc with last_reviewed frontmatter."""
    path.parent.mkdir(parents=True, exist_ok=True)
    reviewed_date = date.today() - timedelta(days=days_ago)
    path.write_text(
        f"---\nlast_reviewed: {reviewed_date.isoformat()}\n---\n# Document\nContent here.\n"
    )


def _create_registry(project_dir: Path, doc_paths: list[str]) -> None:
    """Create a doc-registry.toml in the project directory."""
    registry_dir = project_dir / ".alty" / "maintenance"
    registry_dir.mkdir(parents=True, exist_ok=True)
    lines: list[str] = []
    for doc_path in doc_paths:
        lines.append("[[docs]]")
        lines.append(f'path = "{doc_path}"')
        lines.append("")
    (registry_dir / "doc-registry.toml").write_text("\n".join(lines))
