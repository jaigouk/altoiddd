"""Tests for RlmResearchAdapter — iterative search + LLM synthesis.

Verifies protocol compliance, finding generation from mocked WebSearchPort +
LLMClient, source attribution, degradation paths (LLM unavailable, search
failure), and iterative search (at least 2 search calls per area).
"""

from __future__ import annotations

from unittest.mock import AsyncMock

import pytest

from src.application.ports.domain_research_port import DomainResearchPort
from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import BoundedContext
from src.domain.models.errors import LLMUnavailableError
from src.domain.models.research import Confidence, TrustLevel, WebSearchResult
from src.infrastructure.external.llm_client import LLMResponse
from src.infrastructure.external.rlm_research_adapter import RlmResearchAdapter


def _make_model(*context_names: str) -> DomainModel:
    """Build a DomainModel with given bounded context names."""
    model = DomainModel()
    for name in context_names:
        model.add_bounded_context(
            BoundedContext(name=name, responsibility=f"{name} responsibility")
        )
    return model


def _make_search_results(n: int = 2) -> tuple[WebSearchResult, ...]:
    return tuple(
        WebSearchResult(
            url=f"https://example.com/{i}",
            title=f"Result {i}",
            snippet=f"Snippet {i}",
        )
        for i in range(n)
    )


class TestRlmResearchAdapterProtocol:
    def test_satisfies_domain_research_port(self) -> None:
        adapter = RlmResearchAdapter(
            llm_client=AsyncMock(),
            web_search=AsyncMock(),
        )
        assert isinstance(adapter, DomainResearchPort)


class TestRlmResearchAdapterHappyPath:
    @pytest.mark.asyncio
    async def test_returns_briefing_with_findings(self) -> None:
        web_search = AsyncMock()
        web_search.search.return_value = _make_search_results(2)

        llm = AsyncMock()
        llm.text_completion.return_value = LLMResponse(
            content="Refined query: Sales competition analysis\nSynthesis: Key patterns found.",
            model_used="test-model",
            usage_tokens=100,
        )

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales")

        briefing = await adapter.research(model)

        assert len(briefing.findings) > 0
        assert briefing.no_data_areas == ()

    @pytest.mark.asyncio
    async def test_summary_contains_llm_response(self) -> None:
        """Happy path: summary should come from LLM text_completion."""
        web_search = AsyncMock()
        web_search.search.return_value = _make_search_results(2)

        llm = AsyncMock()
        llm.text_completion.side_effect = [
            LLMResponse(
                content="Refined query: test\nSynthesis: done",
                model_used="test-model",
                usage_tokens=50,
            ),
            LLMResponse(
                content="Key patterns across Sales domain.",
                model_used="test-model",
                usage_tokens=30,
            ),
        ]

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales")

        briefing = await adapter.research(model)

        assert briefing.summary == "Key patterns across Sales domain."

    @pytest.mark.asyncio
    async def test_results_with_empty_url_or_snippet_filtered_out(self) -> None:
        """_build_findings must drop results with empty URL or snippet."""
        results_with_gaps = (
            WebSearchResult(url="https://example.com/1", title="Good", snippet="Content"),
            WebSearchResult(url="", title="No URL", snippet="Has snippet"),
            WebSearchResult(url="https://example.com/3", title="No Snippet", snippet=""),
        )
        web_search = AsyncMock()
        web_search.search.return_value = results_with_gaps

        llm = AsyncMock()
        llm.text_completion.return_value = LLMResponse(
            content="Refined query: test\nSynthesis: done",
            model_used="test-model",
            usage_tokens=50,
        )

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales")

        briefing = await adapter.research(model)

        # Only the first result should survive (has both url and snippet)
        # Results appear twice (round 1 + round 2 both return same mock)
        for finding in briefing.findings:
            assert finding.source.url != ""
            assert finding.content != ""

    @pytest.mark.asyncio
    async def test_every_finding_has_source_url(self) -> None:
        web_search = AsyncMock()
        web_search.search.return_value = _make_search_results(3)

        llm = AsyncMock()
        llm.text_completion.return_value = LLMResponse(
            content="Refined query: test\nSynthesis: patterns",
            model_used="test-model",
            usage_tokens=50,
        )

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales")

        briefing = await adapter.research(model)

        for finding in briefing.findings:
            assert finding.source.url.startswith("https://")

    @pytest.mark.asyncio
    async def test_iterative_search_at_least_two_calls(self) -> None:
        """RLM pattern requires at least 2 search rounds per area."""
        web_search = AsyncMock()
        web_search.search.return_value = _make_search_results(2)

        llm = AsyncMock()
        llm.text_completion.return_value = LLMResponse(
            content="Refined query: deeper Sales analysis\nSynthesis: done",
            model_used="test-model",
            usage_tokens=50,
        )

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales")

        await adapter.research(model)

        # At least 2 search calls (round 1 broad + round 2 refined)
        assert web_search.search.call_count >= 2


class TestRlmResearchAdapterDegradation:
    @pytest.mark.asyncio
    async def test_llm_unavailable_returns_raw_results_low_confidence(self) -> None:
        """When LLM is unavailable, degrade to raw results with LOW confidence."""
        web_search = AsyncMock()
        web_search.search.return_value = _make_search_results(2)

        llm = AsyncMock()
        llm.text_completion.side_effect = LLMUnavailableError("No API key")

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales")

        briefing = await adapter.research(model)

        assert len(briefing.findings) > 0
        for finding in briefing.findings:
            assert finding.source.confidence == Confidence.LOW
            assert finding.trust_level == TrustLevel.AI_INFERRED

    @pytest.mark.asyncio
    async def test_complete_search_failure_all_no_data(self) -> None:
        """When search returns nothing, all areas become no_data."""
        web_search = AsyncMock()
        web_search.search.return_value = ()

        llm = AsyncMock()

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales", "Billing")

        briefing = await adapter.research(model)

        assert len(briefing.findings) == 0
        assert set(briefing.no_data_areas) == {"Sales", "Billing"}

    @pytest.mark.asyncio
    async def test_unexpected_llm_error_degrades_to_no_data(self) -> None:
        """Non-LLM exceptions (ValueError, KeyError) must not crash the briefing."""
        web_search = AsyncMock()
        web_search.search.return_value = _make_search_results(2)

        llm = AsyncMock()
        # Simulate malformed LLM response causing a KeyError
        llm.text_completion.side_effect = KeyError("bad key")

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales")

        briefing = await adapter.research(model)

        # Should degrade gracefully, not propagate KeyError
        assert "Sales" in briefing.no_data_areas or len(briefing.findings) > 0

    @pytest.mark.asyncio
    async def test_unexpected_error_does_not_block_other_areas(self) -> None:
        """Error in one area must not prevent other areas from being researched."""
        call_count = 0

        async def _search_side_effect(query: str, **kwargs: object) -> tuple[WebSearchResult, ...]:
            nonlocal call_count
            call_count += 1
            return _make_search_results(2)

        web_search = AsyncMock()
        web_search.search.side_effect = _search_side_effect

        llm = AsyncMock()
        # First text_completion call (Sales synthesis) raises ValueError,
        # second call (Billing synthesis) succeeds, third (summary) succeeds
        llm.text_completion.side_effect = [
            ValueError("malformed response"),
            LLMResponse(
                content="Refined query: Billing test\nSynthesis: done",
                model_used="test-model",
                usage_tokens=50,
            ),
            LLMResponse(
                content="Summary of findings",
                model_used="test-model",
                usage_tokens=30,
            ),
        ]

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales", "Billing")

        briefing = await adapter.research(model)

        # Billing should still have findings despite Sales failing
        finding_areas = {f.domain_area for f in briefing.findings}
        assert "Billing" in finding_areas

    @pytest.mark.asyncio
    async def test_partial_failure_mixed_results(self) -> None:
        """One area succeeds, another fails → findings + no_data."""
        web_search = AsyncMock()
        # First area: results; second area: empty
        web_search.search.side_effect = [
            _make_search_results(2),  # Sales round 1
            _make_search_results(1),  # Sales round 2
            (),                        # Billing round 1
        ]

        llm = AsyncMock()
        llm.text_completion.return_value = LLMResponse(
            content="Refined query: test\nSynthesis: done",
            model_used="test-model",
            usage_tokens=50,
        )

        adapter = RlmResearchAdapter(llm_client=llm, web_search=web_search)
        model = _make_model("Sales", "Billing")

        briefing = await adapter.research(model)

        assert len(briefing.findings) > 0
        assert "Billing" in briefing.no_data_areas
        # Sales should have findings, not be in no_data
        finding_areas = {f.domain_area for f in briefing.findings}
        assert "Sales" in finding_areas
