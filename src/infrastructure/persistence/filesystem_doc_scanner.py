"""Filesystem adapter for document health scanning.

FilesystemDocScanner implements DocScannerProtocol by reading TOML registries,
checking file existence, and parsing YAML frontmatter for last_reviewed dates.
Uses simple regex parsing -- no pyyaml dependency required.
"""

from __future__ import annotations

import re
from datetime import date
from typing import TYPE_CHECKING

from src.domain.models.doc_health import (
    DocRegistryEntry,
    DocStatus,
    create_doc_status,
)

if TYPE_CHECKING:
    from pathlib import Path

_FRONTMATTER_RE = re.compile(r"^---\s*\n(.*?)\n---", re.DOTALL)
_LAST_REVIEWED_RE = re.compile(r"^last_reviewed:\s*(.+)$", re.MULTILINE)
_PLACEHOLDER_RE = re.compile(r"^[A-Z]{4}-[A-Z]{2}-[A-Z]{2}$")


def _parse_last_reviewed(content: str) -> date | None:
    """Extract last_reviewed date from markdown file content.

    Looks for YAML frontmatter delimited by --- and extracts the
    last_reviewed field. Returns None if not found, if the value is
    a placeholder like YYYY-MM-DD, or if the date is invalid.

    Args:
        content: Full text content of a markdown file.

    Returns:
        Parsed date, or None if unavailable.
    """
    fm_match = _FRONTMATTER_RE.search(content)
    if not fm_match:
        return None

    frontmatter = fm_match.group(1)
    lr_match = _LAST_REVIEWED_RE.search(frontmatter)
    if not lr_match:
        return None

    raw = lr_match.group(1).strip().strip('"').strip("'")

    if _PLACEHOLDER_RE.match(raw):
        return None

    try:
        return date.fromisoformat(raw)
    except ValueError:
        return None


def _parse_toml_value(value: str) -> str | int:
    """Parse a single TOML value (string or integer).

    Args:
        value: Stripped TOML value string.

    Returns:
        Unquoted string or parsed integer.
    """
    if (value.startswith('"') and value.endswith('"')) or (
        value.startswith("'") and value.endswith("'")
    ):
        return value[1:-1]
    try:
        return int(value)
    except ValueError:
        return value


def _parse_toml_docs(content: str) -> list[dict[str, str | int]]:
    """Parse [[docs]] entries from a simple TOML file.

    This is a minimal parser that handles the specific format used by
    doc-registry.toml. It does not handle all TOML features.

    Args:
        content: TOML file content.

    Returns:
        List of dicts with string/int values per entry.
    """
    entries: list[dict[str, str | int]] = []
    current: dict[str, str | int] | None = None

    for line in content.splitlines():
        stripped = line.strip()
        if stripped == "[[docs]]":
            if current is not None:
                entries.append(current)
            current = {}
            continue

        if current is None or not stripped or stripped.startswith("#") or "=" not in stripped:
            continue

        key, _, value = stripped.partition("=")
        current[key.strip()] = _parse_toml_value(value.strip())

    if current is not None:
        entries.append(current)

    return entries


class FilesystemDocScanner:
    """Scans the filesystem for document health status.

    Implements DocScannerProtocol by reading TOML registries,
    checking file existence, and parsing YAML frontmatter.
    """

    def load_registry(self, registry_path: Path) -> tuple[DocRegistryEntry, ...]:
        """Load document registry entries from a TOML file.

        Args:
            registry_path: Path to the doc-registry.toml file.

        Returns:
            Tuple of DocRegistryEntry, empty if file is missing or broken.
        """
        if not registry_path.is_file():
            return ()

        try:
            content = registry_path.read_text()
        except OSError:
            return ()

        raw_entries = _parse_toml_docs(content)
        entries: list[DocRegistryEntry] = []
        for raw in raw_entries:
            path = raw.get("path")
            if not isinstance(path, str):
                continue
            owner = raw.get("owner")
            owner_str = str(owner) if owner is not None else None
            interval = raw.get("review_interval_days", 30)
            interval_int = int(interval) if not isinstance(interval, int) else interval
            entries.append(
                DocRegistryEntry(
                    path=path,
                    owner=owner_str,
                    review_interval_days=interval_int,
                )
            )

        return tuple(entries)

    def scan_registered(
        self,
        entries: tuple[DocRegistryEntry, ...],
        project_dir: Path,
    ) -> tuple[DocStatus, ...]:
        """Scan registered documents for health status.

        For each entry, checks file existence and parses frontmatter
        for the last_reviewed date. Uses create_doc_status factory.

        Args:
            entries: Registry entries to scan.
            project_dir: Project root directory.

        Returns:
            Tuple of DocStatus for each registered document.
        """
        statuses: list[DocStatus] = []
        for entry in entries:
            file_path = project_dir / entry.path
            exists = file_path.is_file()

            last_reviewed: date | None = None
            if exists:
                try:
                    content = file_path.read_text()
                    last_reviewed = _parse_last_reviewed(content)
                except OSError:
                    pass

            statuses.append(
                create_doc_status(
                    path=entry.path,
                    exists=exists,
                    last_reviewed=last_reviewed,
                    review_interval_days=entry.review_interval_days,
                    owner=entry.owner,
                )
            )

        return tuple(statuses)

    def scan_unregistered(
        self,
        docs_dir: Path,
        registered_paths: frozenset[str],
        exclude_dirs: tuple[str, ...] = ("templates", "beads_templates"),
    ) -> tuple[DocStatus, ...]:
        """Scan for markdown files not in the registry.

        Globs for *.md files, skips registered paths and excluded dirs,
        and checks frontmatter presence.

        Args:
            docs_dir: Directory to scan for .md files.
            registered_paths: Paths already tracked in the registry.
            exclude_dirs: Directory names to skip.

        Returns:
            Tuple of DocStatus for unregistered documents.
        """
        if not docs_dir.is_dir():
            return ()

        statuses: list[DocStatus] = []
        for md_file in sorted(docs_dir.rglob("*.md")):
            # Build relative path from docs_dir's parent (project root)
            try:
                rel_path = str(md_file.relative_to(docs_dir.parent))
            except ValueError:
                rel_path = str(md_file.relative_to(docs_dir))

            # Skip registered paths
            if rel_path in registered_paths:
                continue

            # Skip excluded directories
            parts = md_file.relative_to(docs_dir).parts
            if any(part in exclude_dirs for part in parts):
                continue

            # Check frontmatter
            try:
                content = md_file.read_text()
                last_reviewed = _parse_last_reviewed(content)
            except OSError:
                last_reviewed = None

            statuses.append(
                create_doc_status(
                    path=rel_path,
                    exists=True,
                    last_reviewed=last_reviewed,
                )
            )

        return tuple(statuses)
