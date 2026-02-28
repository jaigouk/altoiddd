"""Port for the Knowledge Base bounded context.

Defines the interface for looking up RLM-addressable knowledge including
DDD patterns, tool conventions, and coding standards.
"""

from __future__ import annotations

from typing import Protocol, runtime_checkable


@runtime_checkable
class KnowledgeLookupPort(Protocol):
    """Interface for knowledge base lookup operations.

    Adapters implement this to provide versioned, RLM-addressable access
    to DDD patterns, tool conventions, and coding standards with drift
    detection support.
    """

    def lookup(self, category: str, topic: str, version: str = "current") -> str:
        """Look up a specific knowledge entry.

        Args:
            category: The knowledge category (e.g., "ddd", "tool", "coding").
            topic: The specific topic within the category.
            version: The version to look up, defaults to "current".

        Returns:
            The knowledge content for the requested entry.
        """
        ...

    def list_tools(self) -> list[str]:
        """List all known AI coding tools.

        Returns:
            List of tool identifiers with available knowledge entries.
        """
        ...

    def list_versions(self, tool: str) -> list[str]:
        """List available versions for a specific tool.

        Args:
            tool: The tool identifier to list versions for.

        Returns:
            List of version identifiers.
        """
        ...

    def list_topics(self, category: str, tool: str | None = None) -> list[str]:
        """List available topics within a category.

        Args:
            category: The knowledge category to list topics for.
            tool: Optional tool filter to narrow results.

        Returns:
            List of topic names within the category.
        """
        ...
