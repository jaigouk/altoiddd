"""Query handler for document health checking.

DocHealthHandler orchestrates the doc health check flow: loads a registry,
scans registered and unregistered docs, and returns a DocHealthReport.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

from src.domain.models.doc_health import (
    DocHealthReport,
    DocRegistryEntry,
)

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.doc_health import DocStatus


@runtime_checkable
class DocScannerProtocol(Protocol):
    """Interface for scanning documents on disk.

    Adapters implement this to check file existence and parse
    frontmatter for last_reviewed dates.
    """

    def load_registry(self, registry_path: Path) -> tuple[DocRegistryEntry, ...]:
        """Load document registry entries from a TOML file.

        Args:
            registry_path: Path to the doc-registry.toml file.

        Returns:
            Tuple of registry entries, empty if file is missing or broken.
        """
        ...

    def scan_registered(
        self,
        entries: tuple[DocRegistryEntry, ...],
        project_dir: Path,
    ) -> tuple[DocStatus, ...]:
        """Scan registered documents for health status.

        Args:
            entries: Registry entries to scan.
            project_dir: Project root directory.

        Returns:
            Tuple of DocStatus for each registered document.
        """
        ...

    def scan_unregistered(
        self,
        docs_dir: Path,
        registered_paths: frozenset[str],
        exclude_dirs: tuple[str, ...] = ("templates", "beads_templates"),
    ) -> tuple[DocStatus, ...]:
        """Scan for markdown files not in the registry.

        Args:
            docs_dir: Directory to scan for .md files.
            registered_paths: Paths already tracked in the registry.
            exclude_dirs: Directory names to skip.

        Returns:
            Tuple of DocStatus for unregistered documents.
        """
        ...


_DEFAULT_ENTRIES = (
    DocRegistryEntry(path="docs/PRD.md"),
    DocRegistryEntry(path="docs/DDD.md"),
    DocRegistryEntry(path="docs/ARCHITECTURE.md"),
)


class DocHealthHandler:
    """Orchestrates document health checking.

    Loads a doc registry (or uses defaults), scans registered docs,
    discovers unregistered docs, and combines into a DocHealthReport.
    """

    def __init__(self, scanner: DocScannerProtocol) -> None:
        self._scanner = scanner

    def check(self, project_dir: Path) -> DocHealthReport:
        """Check the health of project documentation.

        Args:
            project_dir: The project root directory.

        Returns:
            A DocHealthReport with statuses for all checked documents.
        """
        registry_path = project_dir / ".alty" / "maintenance" / "doc-registry.toml"
        entries = self._scanner.load_registry(registry_path)

        if not entries:
            entries = _DEFAULT_ENTRIES

        registered_statuses = self._scanner.scan_registered(entries, project_dir)

        registered_paths = frozenset(e.path for e in entries)
        docs_dir = project_dir / "docs"
        unregistered_statuses = self._scanner.scan_unregistered(
            docs_dir=docs_dir,
            registered_paths=registered_paths,
            exclude_dirs=("templates", "beads_templates"),
        )

        all_statuses = registered_statuses + unregistered_statuses
        return DocHealthReport(statuses=all_statuses)
