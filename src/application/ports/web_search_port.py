"""Port for web search infrastructure.

Defines the interface for executing web search queries and returning
raw results. This is an internal port consumed by research adapters,
not exposed in AppContext.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from src.domain.models.research import WebSearchResult


@runtime_checkable
class WebSearchPort(Protocol):
    """Interface for web search queries.

    Adapters wrap specific search engines (e.g., DuckDuckGo via ddgs)
    and return domain-agnostic WebSearchResult VOs.
    """

    async def search(
        self,
        query: str,
        max_results: int = 10,
    ) -> tuple[WebSearchResult, ...]:
        """Execute a web search query.

        Args:
            query: The search query string.
            max_results: Maximum results to return.

        Returns:
            Tuple of WebSearchResult VOs.
        """
        ...
