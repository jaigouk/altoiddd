"""File-based knowledge reader adapter.

FileKnowledgeReader implements KnowledgeReaderProtocol by reading
knowledge entries from the filesystem.  TOML files are parsed for
``[_meta]`` sections to extract EntryMetadata.  Markdown files are
returned with their raw content.
"""

from __future__ import annotations

import tomllib
from typing import TYPE_CHECKING

from src.domain.models.errors import InvariantViolationError

if TYPE_CHECKING:
    from pathlib import Path
from src.domain.models.knowledge_entry import (
    EntryMetadata,
    KnowledgeCategory,
    KnowledgeEntry,
    KnowledgePath,
)


class FileKnowledgeReader:
    """Reads knowledge entries from a directory on the local filesystem.

    Directory layout::

        {knowledge_dir}/
        ├── ddd/{topic}.md
        ├── conventions/{topic}.md
        ├── tools/{tool}/{version}/{topic}.toml
        └── cross-tool/{topic}.toml
    """

    def __init__(self, knowledge_dir: Path) -> None:
        self._root = knowledge_dir

    def read_entry(
        self,
        path: KnowledgePath,
        version: str = "current",
    ) -> KnowledgeEntry:
        """Read a knowledge entry from the filesystem.

        Args:
            path: The RLM-addressable knowledge path.
            version: Version directory for tools entries (default "current").

        Returns:
            A KnowledgeEntry populated from the file.

        Raises:
            InvariantViolationError: If the file does not exist.
        """
        file_path = self._resolve_file_path(path, version)
        if not file_path.exists():
            msg = f"Knowledge entry not found: {path.raw} (looked at {file_path})"
            raise InvariantViolationError(msg)

        if file_path.suffix == ".toml":
            return self._read_toml_entry(path, file_path)
        return self._read_markdown_entry(path, file_path)

    def list_topics(
        self,
        category: KnowledgeCategory,
        tool: str | None = None,
    ) -> tuple[str, ...]:
        """List available topics within a category.

        Args:
            category: The knowledge category to scan.
            tool: For TOOLS category, the specific tool to list topics for.

        Returns:
            Sorted tuple of topic stem names.
        """
        if category == KnowledgeCategory.TOOLS:
            if tool is None:
                return ()
            scan_dir = self._root / "tools" / tool / "current"
            pattern = "*.toml"
        elif category == KnowledgeCategory.CROSS_TOOL:
            scan_dir = self._root / "cross-tool"
            pattern = "*.toml"
        elif category == KnowledgeCategory.DDD:
            scan_dir = self._root / "ddd"
            pattern = "*.md"
        else:  # CONVENTIONS
            scan_dir = self._root / "conventions"
            pattern = "*.md"

        if not scan_dir.exists():
            return ()

        return tuple(sorted(f.stem for f in scan_dir.glob(pattern)))

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    def _resolve_file_path(self, path: KnowledgePath, version: str) -> Path:
        """Map a KnowledgePath to a concrete filesystem path."""
        if path.category == KnowledgeCategory.TOOLS:
            tool = path.tool
            subtopic = path.subtopic
            return self._root / "tools" / (tool or "") / version / f"{subtopic}.toml"

        if path.category == KnowledgeCategory.CROSS_TOOL:
            topic = path.topic
            return self._root / "cross-tool" / f"{topic}.toml"

        # DDD and CONVENTIONS: {category}/{topic}.md
        category_dir: str = path.category.value
        return self._root / category_dir / f"{path.topic}.md"

    def _read_toml_entry(
        self,
        path: KnowledgePath,
        file_path: Path,
    ) -> KnowledgeEntry:
        """Parse a TOML file into a KnowledgeEntry."""
        raw_text = file_path.read_text()
        data = tomllib.loads(raw_text)
        metadata = self._extract_metadata(data)
        return KnowledgeEntry(
            path=path,
            title=path.subtopic or path.topic,
            content=raw_text,
            metadata=metadata,
            format="toml",
        )

    def _read_markdown_entry(
        self,
        path: KnowledgePath,
        file_path: Path,
    ) -> KnowledgeEntry:
        """Read a Markdown file into a KnowledgeEntry."""
        content = file_path.read_text()
        return KnowledgeEntry(
            path=path,
            title=path.topic,
            content=content,
            metadata=None,
            format="markdown",
        )

    @staticmethod
    def _extract_metadata(data: dict[str, object]) -> EntryMetadata | None:
        """Extract EntryMetadata from a parsed TOML ``[_meta]`` section."""
        meta_raw = data.get("_meta")
        if not isinstance(meta_raw, dict):
            return None

        source_urls_raw = meta_raw.get("source_urls", ())
        source_urls: tuple[str, ...] = ()
        if isinstance(source_urls_raw, list):
            source_urls = tuple(str(u) for u in source_urls_raw)

        schema_raw = meta_raw.get("schema_version")
        schema_version = str(schema_raw) if schema_raw is not None else None

        return EntryMetadata(
            last_verified=_str_or_none(meta_raw.get("last_verified")),
            verified_against=_str_or_none(meta_raw.get("verified_against")),
            confidence=str(meta_raw.get("confidence", "high")),
            deprecated=bool(meta_raw.get("deprecated", False)),
            next_review_date=_str_or_none(meta_raw.get("next_review_date")),
            schema_version=schema_version,
            source_urls=source_urls,
        )


def _str_or_none(val: object) -> str | None:
    """Convert a value to str or None."""
    if val is None:
        return None
    return str(val)
