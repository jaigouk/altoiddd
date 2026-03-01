"""Tests for DocHealthHandler query handler.

Covers orchestration of the doc health check flow: loading registry,
scanning registered and unregistered docs, combining results.
"""

from __future__ import annotations

from pathlib import Path

from src.domain.models.doc_health import (
    DocHealthReport,
    DocHealthStatus,
    DocRegistryEntry,
    DocStatus,
)

# ---------------------------------------------------------------------------
# Fake scanner for testing
# ---------------------------------------------------------------------------


class FakeDocScanner:
    """In-memory doc scanner for handler tests."""

    def __init__(
        self,
        registry_entries: tuple[DocRegistryEntry, ...] = (),
        registered_statuses: tuple[DocStatus, ...] = (),
        unregistered_statuses: tuple[DocStatus, ...] = (),
    ) -> None:
        self._registry_entries = registry_entries
        self._registered_statuses = registered_statuses
        self._unregistered_statuses = unregistered_statuses
        self.load_registry_calls: list[Path] = []
        self.scan_registered_calls: list[tuple[tuple[DocRegistryEntry, ...], Path]] = []
        self.scan_unregistered_calls: list[tuple[Path, frozenset[str], tuple[str, ...]]] = []

    def load_registry(self, registry_path: Path) -> tuple[DocRegistryEntry, ...]:
        self.load_registry_calls.append(registry_path)
        return self._registry_entries

    def scan_registered(
        self,
        entries: tuple[DocRegistryEntry, ...],
        project_dir: Path,
    ) -> tuple[DocStatus, ...]:
        self.scan_registered_calls.append((entries, project_dir))
        return self._registered_statuses

    def scan_unregistered(
        self,
        docs_dir: Path,
        registered_paths: frozenset[str],
        exclude_dirs: tuple[str, ...] = ("templates", "beads_templates"),
    ) -> tuple[DocStatus, ...]:
        self.scan_unregistered_calls.append((docs_dir, registered_paths, exclude_dirs))
        return self._unregistered_statuses


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


class TestDocHealthHandler:
    def test_handler_checks_registered_docs(self) -> None:
        from src.application.queries.doc_health_handler import DocHealthHandler

        registered_statuses = (DocStatus(path="docs/PRD.md", status=DocHealthStatus.OK),)
        scanner = FakeDocScanner(
            registry_entries=(DocRegistryEntry(path="docs/PRD.md"),),
            registered_statuses=registered_statuses,
        )
        handler = DocHealthHandler(scanner=scanner)

        report = handler.check(project_dir=Path("/project"))

        assert isinstance(report, DocHealthReport)
        assert len(scanner.scan_registered_calls) == 1

    def test_handler_uses_defaults_when_registry_missing(self) -> None:
        from src.application.queries.doc_health_handler import DocHealthHandler

        # Empty registry -> handler uses defaults
        default_statuses = (
            DocStatus(path="docs/PRD.md", status=DocHealthStatus.OK),
            DocStatus(path="docs/DDD.md", status=DocHealthStatus.MISSING),
            DocStatus(path="docs/ARCHITECTURE.md", status=DocHealthStatus.NO_FRONTMATTER),
        )
        scanner = FakeDocScanner(
            registry_entries=(),  # empty -> triggers defaults
            registered_statuses=default_statuses,
        )
        handler = DocHealthHandler(scanner=scanner)

        handler.check(project_dir=Path("/project"))

        # Handler should have used the 3 default entries
        assert len(scanner.scan_registered_calls) == 1
        entries_used = scanner.scan_registered_calls[0][0]
        assert len(entries_used) == 3
        paths = {e.path for e in entries_used}
        assert paths == {"docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md"}

    def test_handler_skips_template_dirs(self) -> None:
        from src.application.queries.doc_health_handler import DocHealthHandler

        scanner = FakeDocScanner(
            registry_entries=(DocRegistryEntry(path="docs/PRD.md"),),
            registered_statuses=(DocStatus(path="docs/PRD.md", status=DocHealthStatus.OK),),
        )
        handler = DocHealthHandler(scanner=scanner)

        handler.check(project_dir=Path("/project"))

        # scan_unregistered should be called with exclude_dirs
        assert len(scanner.scan_unregistered_calls) == 1
        _, _, exclude_dirs = scanner.scan_unregistered_calls[0]
        assert "templates" in exclude_dirs
        assert "beads_templates" in exclude_dirs

    def test_handler_combines_registered_and_unregistered(self) -> None:
        from src.application.queries.doc_health_handler import DocHealthHandler

        registered = (DocStatus(path="docs/PRD.md", status=DocHealthStatus.OK),)
        unregistered = (DocStatus(path="docs/notes.md", status=DocHealthStatus.NO_FRONTMATTER),)
        scanner = FakeDocScanner(
            registry_entries=(DocRegistryEntry(path="docs/PRD.md"),),
            registered_statuses=registered,
            unregistered_statuses=unregistered,
        )
        handler = DocHealthHandler(scanner=scanner)

        report = handler.check(project_dir=Path("/project"))

        assert report.total_checked == 2
        paths = {s.path for s in report.statuses}
        assert paths == {"docs/PRD.md", "docs/notes.md"}

    def test_handler_returns_complete_report(self) -> None:
        from src.application.queries.doc_health_handler import DocHealthHandler

        registered = (
            DocStatus(path="docs/PRD.md", status=DocHealthStatus.OK),
            DocStatus(path="docs/DDD.md", status=DocHealthStatus.STALE, days_since=45),
        )
        unregistered = (DocStatus(path="docs/extra.md", status=DocHealthStatus.NO_FRONTMATTER),)
        scanner = FakeDocScanner(
            registry_entries=(
                DocRegistryEntry(path="docs/PRD.md"),
                DocRegistryEntry(path="docs/DDD.md"),
            ),
            registered_statuses=registered,
            unregistered_statuses=unregistered,
        )
        handler = DocHealthHandler(scanner=scanner)

        report = handler.check(project_dir=Path("/project"))

        assert report.total_checked == 3
        assert report.issue_count == 2  # STALE + NO_FRONTMATTER
        assert report.has_issues is True

    def test_handler_passes_registered_paths_to_unregistered_scan(self) -> None:
        from src.application.queries.doc_health_handler import DocHealthHandler

        scanner = FakeDocScanner(
            registry_entries=(
                DocRegistryEntry(path="docs/PRD.md"),
                DocRegistryEntry(path="docs/DDD.md"),
            ),
            registered_statuses=(
                DocStatus(path="docs/PRD.md", status=DocHealthStatus.OK),
                DocStatus(path="docs/DDD.md", status=DocHealthStatus.OK),
            ),
        )
        handler = DocHealthHandler(scanner=scanner)

        handler.check(project_dir=Path("/project"))

        # The registered_paths passed to scan_unregistered should match registry entries
        assert len(scanner.scan_unregistered_calls) == 1
        _, registered_paths, _ = scanner.scan_unregistered_calls[0]
        assert registered_paths == frozenset({"docs/PRD.md", "docs/DDD.md"})
