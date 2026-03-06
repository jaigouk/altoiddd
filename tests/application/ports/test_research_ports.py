"""Tests for DomainResearchPort and WebSearchPort protocols.

Verifies runtime_checkable behaviour and that concrete classes
satisfying the protocol are recognised by isinstance().
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.application.ports.domain_research_port import DomainResearchPort
from src.application.ports.web_search_port import WebSearchPort
from src.domain.models.research import ResearchBriefing, WebSearchResult

if TYPE_CHECKING:
    from src.domain.models.domain_model import DomainModel


class TestDomainResearchPort:
    def test_runtime_checkable(self) -> None:
        """DomainResearchPort must be decorated with @runtime_checkable."""

        class _Impl:
            async def research(
                self,
                model: DomainModel,
                max_areas: int = 5,
            ) -> ResearchBriefing:
                return ResearchBriefing(findings=(), no_data_areas=(), summary="")

        assert isinstance(_Impl(), DomainResearchPort)

    def test_non_conforming_rejected(self) -> None:
        class _Bad:
            pass

        assert not isinstance(_Bad(), DomainResearchPort)


class TestWebSearchPort:
    def test_runtime_checkable(self) -> None:
        """WebSearchPort must be decorated with @runtime_checkable."""

        class _Impl:
            async def search(
                self,
                query: str,
                max_results: int = 10,
            ) -> tuple[WebSearchResult, ...]:
                return ()

        assert isinstance(_Impl(), WebSearchPort)

    def test_non_conforming_rejected(self) -> None:
        class _Bad:
            pass

        assert not isinstance(_Bad(), WebSearchPort)
