"""DuckDuckGo search adapter — wraps the ddgs library.

ddgs is a sync-only library, so we wrap calls in asyncio.to_thread().
It is an optional dependency; if not installed, search returns an empty tuple.
"""

from __future__ import annotations

import asyncio
import logging

from src.domain.models.research import WebSearchResult

logger = logging.getLogger(__name__)

try:
    from ddgs import DDGS  # type: ignore[import-not-found]
except ImportError:
    DDGS = None


class DuckDuckGoSearchAdapter:
    """WebSearchPort implementation using DuckDuckGo via ddgs."""

    async def search(
        self,
        query: str,
        max_results: int = 10,
    ) -> tuple[WebSearchResult, ...]:
        """Execute a DuckDuckGo text search.

        Maps ddgs dict keys (href→url, body→snippet) to WebSearchResult VOs.
        Returns empty tuple on any error or if ddgs is not installed.
        """
        if DDGS is None:
            logger.warning("ddgs not installed; returning empty results")
            return ()

        try:
            raw = await asyncio.to_thread(
                DDGS().text,
                query,
                max_results=max_results,
            )
            return tuple(
                WebSearchResult(
                    url=r.get("href", ""),
                    title=r.get("title", ""),
                    snippet=r.get("body", ""),
                )
                for r in raw
            )
        except Exception:
            logger.exception("DuckDuckGo search failed for query: %s", query)
            return ()
