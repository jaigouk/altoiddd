"""Tests for FilesystemDocScanner infrastructure adapter.

Covers TOML registry loading, registered doc scanning with frontmatter
parsing, unregistered doc discovery, and edge cases.
"""

from __future__ import annotations

from datetime import date, timedelta
from pathlib import Path  # noqa: TC003 — used at runtime for tmp_path operations

from src.domain.models.doc_health import (
    DocHealthStatus,
    DocRegistryEntry,
)
from src.infrastructure.persistence.filesystem_doc_scanner import FilesystemDocScanner

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _write_doc_with_frontmatter(path: Path, last_reviewed: str | None) -> None:
    """Write a markdown file with optional YAML frontmatter."""
    path.parent.mkdir(parents=True, exist_ok=True)
    if last_reviewed is not None:
        content = f"---\nlast_reviewed: {last_reviewed}\n---\n\n# Title\n\nContent here.\n"
    else:
        content = "# Title\n\nContent without frontmatter.\n"
    path.write_text(content)


def _write_registry_toml(path: Path, entries: list[dict[str, str | int]]) -> None:
    """Write a doc-registry.toml file with the given entries."""
    path.parent.mkdir(parents=True, exist_ok=True)
    lines = []
    for entry in entries:
        lines.append("[[docs]]")
        for key, value in entry.items():
            if isinstance(value, str):
                lines.append(f'{key} = "{value}"')
            else:
                lines.append(f"{key} = {value}")
        lines.append("")
    path.write_text("\n".join(lines))


# ---------------------------------------------------------------------------
# 1. Registry loading
# ---------------------------------------------------------------------------


class TestLoadRegistry:
    def test_scanner_loads_registry_from_toml(self, tmp_path: Path) -> None:
        registry_path = tmp_path / "doc-registry.toml"
        _write_registry_toml(
            registry_path,
            [
                {"path": "docs/PRD.md", "owner": "pm", "review_interval_days": 14},
                {"path": "docs/DDD.md"},
            ],
        )

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(registry_path)

        assert len(entries) == 2
        assert entries[0].path == "docs/PRD.md"
        assert entries[0].owner == "pm"
        assert entries[0].review_interval_days == 14
        assert entries[1].path == "docs/DDD.md"
        assert entries[1].owner is None
        assert entries[1].review_interval_days == 30

    def test_scanner_returns_empty_on_missing_registry(self, tmp_path: Path) -> None:
        missing_path = tmp_path / "nonexistent.toml"

        scanner = FilesystemDocScanner()
        entries = scanner.load_registry(missing_path)

        assert entries == ()


# ---------------------------------------------------------------------------
# 2. Registered doc scanning
# ---------------------------------------------------------------------------


class TestScanRegistered:
    def test_scanner_scans_registered_ok(self, tmp_path: Path) -> None:
        reviewed = date.today() - timedelta(days=5)
        doc_path = tmp_path / "docs" / "PRD.md"
        _write_doc_with_frontmatter(doc_path, reviewed.isoformat())

        scanner = FilesystemDocScanner()
        entries = (DocRegistryEntry(path="docs/PRD.md"),)
        statuses = scanner.scan_registered(entries, project_dir=tmp_path)

        assert len(statuses) == 1
        assert statuses[0].status == DocHealthStatus.OK
        assert statuses[0].days_since == 5

    def test_scanner_scans_registered_missing(self, tmp_path: Path) -> None:
        scanner = FilesystemDocScanner()
        entries = (DocRegistryEntry(path="docs/MISSING.md"),)
        statuses = scanner.scan_registered(entries, project_dir=tmp_path)

        assert len(statuses) == 1
        assert statuses[0].status == DocHealthStatus.MISSING

    def test_scanner_scans_registered_no_frontmatter(self, tmp_path: Path) -> None:
        doc_path = tmp_path / "docs" / "plain.md"
        _write_doc_with_frontmatter(doc_path, None)

        scanner = FilesystemDocScanner()
        entries = (DocRegistryEntry(path="docs/plain.md"),)
        statuses = scanner.scan_registered(entries, project_dir=tmp_path)

        assert len(statuses) == 1
        assert statuses[0].status == DocHealthStatus.NO_FRONTMATTER

    def test_scanner_scans_registered_stale(self, tmp_path: Path) -> None:
        reviewed = date.today() - timedelta(days=45)
        doc_path = tmp_path / "docs" / "old.md"
        _write_doc_with_frontmatter(doc_path, reviewed.isoformat())

        scanner = FilesystemDocScanner()
        entries = (DocRegistryEntry(path="docs/old.md", review_interval_days=30),)
        statuses = scanner.scan_registered(entries, project_dir=tmp_path)

        assert len(statuses) == 1
        assert statuses[0].status == DocHealthStatus.STALE
        assert statuses[0].days_since == 45

    def test_scanner_handles_placeholder_date(self, tmp_path: Path) -> None:
        """YYYY-MM-DD placeholder should be treated as no date."""
        doc_path = tmp_path / "docs" / "placeholder.md"
        _write_doc_with_frontmatter(doc_path, "YYYY-MM-DD")

        scanner = FilesystemDocScanner()
        entries = (DocRegistryEntry(path="docs/placeholder.md"),)
        statuses = scanner.scan_registered(entries, project_dir=tmp_path)

        assert len(statuses) == 1
        assert statuses[0].status == DocHealthStatus.NO_FRONTMATTER

    def test_scanner_handles_invalid_date(self, tmp_path: Path) -> None:
        """Invalid date string should be treated as no date."""
        doc_path = tmp_path / "docs" / "bad_date.md"
        _write_doc_with_frontmatter(doc_path, "not-a-date")

        scanner = FilesystemDocScanner()
        entries = (DocRegistryEntry(path="docs/bad_date.md"),)
        statuses = scanner.scan_registered(entries, project_dir=tmp_path)

        assert len(statuses) == 1
        assert statuses[0].status == DocHealthStatus.NO_FRONTMATTER

    def test_scanner_passes_owner_through(self, tmp_path: Path) -> None:
        reviewed = date.today()
        doc_path = tmp_path / "docs" / "owned.md"
        _write_doc_with_frontmatter(doc_path, reviewed.isoformat())

        scanner = FilesystemDocScanner()
        entries = (DocRegistryEntry(path="docs/owned.md", owner="team-lead"),)
        statuses = scanner.scan_registered(entries, project_dir=tmp_path)

        assert statuses[0].owner == "team-lead"


# ---------------------------------------------------------------------------
# 3. Unregistered doc scanning
# ---------------------------------------------------------------------------


class TestScanUnregistered:
    def test_scanner_scans_unregistered_excludes_dirs(self, tmp_path: Path) -> None:
        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()

        # Regular doc
        (docs_dir / "notes.md").write_text("# Notes\n\nSome notes.\n")

        # Excluded dirs
        templates_dir = docs_dir / "templates"
        templates_dir.mkdir()
        (templates_dir / "template.md").write_text("# Template\n")

        beads_dir = docs_dir / "beads_templates"
        beads_dir.mkdir()
        (beads_dir / "ticket.md").write_text("# Ticket\n")

        scanner = FilesystemDocScanner()
        statuses = scanner.scan_unregistered(
            docs_dir=docs_dir,
            registered_paths=frozenset(),
            exclude_dirs=("templates", "beads_templates"),
        )

        paths = {s.path for s in statuses}
        assert "docs/notes.md" in paths or "notes.md" in paths
        # Should NOT include templates or beads_templates
        for s in statuses:
            assert "templates/" not in s.path or "beads_templates/" not in s.path

    def test_scanner_unregistered_skips_registered_paths(self, tmp_path: Path) -> None:
        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        (docs_dir / "PRD.md").write_text("---\nlast_reviewed: 2026-01-01\n---\n# PRD\n")
        (docs_dir / "extra.md").write_text("# Extra\n")

        scanner = FilesystemDocScanner()
        statuses = scanner.scan_unregistered(
            docs_dir=docs_dir,
            registered_paths=frozenset({"docs/PRD.md"}),
        )

        paths = {s.path for s in statuses}
        # PRD.md should be excluded (it's registered)
        assert all("PRD.md" not in p for p in paths)
        # extra.md should be included
        assert any("extra.md" in p for p in paths)

    def test_scanner_unregistered_detects_no_frontmatter(self, tmp_path: Path) -> None:
        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()
        (docs_dir / "bare.md").write_text("# Just a title\n\nNo frontmatter.\n")

        scanner = FilesystemDocScanner()
        statuses = scanner.scan_unregistered(
            docs_dir=docs_dir,
            registered_paths=frozenset(),
        )

        assert len(statuses) == 1
        assert statuses[0].status == DocHealthStatus.NO_FRONTMATTER

    def test_scanner_unregistered_empty_dir(self, tmp_path: Path) -> None:
        docs_dir = tmp_path / "docs"
        docs_dir.mkdir()

        scanner = FilesystemDocScanner()
        statuses = scanner.scan_unregistered(
            docs_dir=docs_dir,
            registered_paths=frozenset(),
        )

        assert statuses == ()

    def test_scanner_unregistered_nonexistent_dir(self, tmp_path: Path) -> None:
        docs_dir = tmp_path / "docs"  # does not exist

        scanner = FilesystemDocScanner()
        statuses = scanner.scan_unregistered(
            docs_dir=docs_dir,
            registered_paths=frozenset(),
        )

        assert statuses == ()
