"""Query handler for Knowledge Base lookups.

KnowledgeLookupHandler provides read-side operations: looking up a single
knowledge entry by path, listing categories, and listing topics within a
category.  It delegates file I/O to a KnowledgeReaderProtocol adapter.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

from src.domain.models.knowledge_entry import (
    KnowledgeCategory,
    KnowledgePath,
)

if TYPE_CHECKING:
    from src.domain.models.knowledge_entry import KnowledgeEntry


@runtime_checkable
class KnowledgeReaderProtocol(Protocol):
    """Port for reading knowledge entries from storage."""

    def read_entry(
        self,
        path: KnowledgePath,
        version: str = "current",
    ) -> KnowledgeEntry:
        """Read a single knowledge entry by path.

        Args:
            path: The RLM-addressable knowledge path.
            version: Version to look up (default "current").

        Returns:
            The resolved KnowledgeEntry.

        Raises:
            InvariantViolationError: If the entry is not found.
        """
        ...

    def list_topics(
        self,
        category: KnowledgeCategory,
        tool: str | None = None,
    ) -> tuple[str, ...]:
        """List available topics within a category.

        Args:
            category: The knowledge category.
            tool: Optional tool filter for the TOOLS category.

        Returns:
            Sorted tuple of topic names.
        """
        ...


class KnowledgeLookupHandler:
    """Orchestrates knowledge base lookup operations.

    Parses user-facing path strings into KnowledgePath value objects
    and delegates to a KnowledgeReaderProtocol for actual retrieval.
    """

    def __init__(self, reader: KnowledgeReaderProtocol) -> None:
        self._reader = reader

    def lookup(self, path_str: str, version: str = "current") -> KnowledgeEntry:
        """Look up a knowledge entry by its RLM path string.

        Args:
            path_str: Path string, e.g. "ddd/tactical-patterns".
            version: Version to retrieve (default "current").

        Returns:
            The resolved KnowledgeEntry.

        Raises:
            InvariantViolationError: If the path is invalid or entry not found.
        """
        path = KnowledgePath(raw=path_str)
        return self._reader.read_entry(path, version=version)

    def list_categories(self) -> tuple[str, ...]:
        """Return all available knowledge category values.

        Returns:
            Tuple of category value strings.
        """
        return tuple(c.value for c in KnowledgeCategory)

    def list_topics(
        self,
        category: str,
        tool: str | None = None,
    ) -> tuple[str, ...]:
        """List topics within a category.

        Args:
            category: Category string (must match a KnowledgeCategory value).
            tool: Optional tool filter for the TOOLS category.

        Returns:
            Sorted tuple of topic names from the reader.

        Raises:
            ValueError: If the category string is not a valid KnowledgeCategory.
        """
        try:
            cat = KnowledgeCategory(category)
        except ValueError:
            from src.domain.models.errors import InvariantViolationError

            msg = (
                f"Invalid category '{category}'. "
                f"Valid categories: {[c.value for c in KnowledgeCategory]}"
            )
            raise InvariantViolationError(msg) from None
        return self._reader.list_topics(cat, tool=tool)
