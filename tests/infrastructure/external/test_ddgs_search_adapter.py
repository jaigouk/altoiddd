"""Tests for DuckDuckGoSearchAdapter — ddgs web search wrapper.

Verifies protocol compliance, result mapping (href→url, body→snippet),
error handling (network errors → empty tuple), and max_results respect.
"""

from __future__ import annotations

from unittest.mock import MagicMock, patch

import pytest

from src.application.ports.web_search_port import WebSearchPort
from src.domain.models.research import WebSearchResult
from src.infrastructure.external.ddgs_search_adapter import DuckDuckGoSearchAdapter


class TestDuckDuckGoSearchAdapter:
    def test_satisfies_web_search_port(self) -> None:
        adapter = DuckDuckGoSearchAdapter()
        assert isinstance(adapter, WebSearchPort)

    @pytest.mark.asyncio
    async def test_maps_ddgs_dict_to_web_search_result(self) -> None:
        raw_results = [
            {
                "title": "Example Page",
                "href": "https://example.com/page",
                "body": "A snippet from the page",
            },
        ]
        mock_ddgs = MagicMock()
        mock_ddgs.return_value.text.return_value = raw_results

        with patch(
            "src.infrastructure.external.ddgs_search_adapter.DDGS",
            mock_ddgs,
        ):
            adapter = DuckDuckGoSearchAdapter()
            results = await adapter.search("test query", max_results=5)

        assert len(results) == 1
        assert isinstance(results[0], WebSearchResult)
        assert results[0].url == "https://example.com/page"
        assert results[0].title == "Example Page"
        assert results[0].snippet == "A snippet from the page"

    @pytest.mark.asyncio
    async def test_network_error_returns_empty_tuple(self) -> None:
        mock_ddgs = MagicMock()
        mock_ddgs.return_value.text.side_effect = Exception("Network error")

        with patch(
            "src.infrastructure.external.ddgs_search_adapter.DDGS",
            mock_ddgs,
        ):
            adapter = DuckDuckGoSearchAdapter()
            results = await adapter.search("failing query")

        assert results == ()

    @pytest.mark.asyncio
    async def test_respects_max_results(self) -> None:
        mock_ddgs = MagicMock()
        mock_ddgs.return_value.text.return_value = []

        with patch(
            "src.infrastructure.external.ddgs_search_adapter.DDGS",
            mock_ddgs,
        ):
            adapter = DuckDuckGoSearchAdapter()
            await adapter.search("test", max_results=3)

        mock_ddgs.return_value.text.assert_called_once_with("test", max_results=3)

    @pytest.mark.asyncio
    async def test_import_error_returns_empty_tuple(self) -> None:
        """When ddgs is not installed, search returns empty tuple."""
        with patch(
            "src.infrastructure.external.ddgs_search_adapter.DDGS",
            None,
        ):
            adapter = DuckDuckGoSearchAdapter()
            results = await adapter.search("test query")

        assert results == ()
