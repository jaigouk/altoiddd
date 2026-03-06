"""RLM Research Adapter — iterative search + LLM synthesis.

Implements the Recursive Language Model (RLM) pattern:
  search → read → reason → search again → build findings.

Degradation paths:
  - LLM unavailable → raw search results with LOW confidence
  - Search failure → area marked as no_data
  - Both fail → empty briefing
"""

from __future__ import annotations

import logging
from datetime import date
from typing import TYPE_CHECKING

from src.domain.models.errors import LLMUnavailableError
from src.domain.models.research import (
    Confidence,
    ResearchBriefing,
    ResearchFinding,
    SourceAttribution,
    TrustLevel,
    WebSearchResult,
)

if TYPE_CHECKING:
    from src.application.ports.web_search_port import WebSearchPort
    from src.domain.models.domain_model import DomainModel
    from src.infrastructure.external.llm_client import LLMClient

logger = logging.getLogger(__name__)


class RlmResearchAdapter:
    """DomainResearchPort implementation using RLM iterative search."""

    def __init__(
        self,
        llm_client: LLMClient,
        web_search: WebSearchPort,
    ) -> None:
        self._llm = llm_client
        self._web = web_search

    async def research(
        self,
        model: DomainModel,
        max_areas: int = 5,
    ) -> ResearchBriefing:
        """Research domain areas using iterative web search + LLM synthesis."""
        areas = self._extract_areas(model, max_areas)
        findings: list[ResearchFinding] = []
        no_data: list[str] = []

        for area in areas:
            area_findings = await self._research_area(area)
            if area_findings:
                findings.extend(area_findings)
            else:
                no_data.append(area)

        summary = await self._build_summary(findings) if findings else ""
        return ResearchBriefing(
            findings=tuple(findings),
            no_data_areas=tuple(no_data),
            summary=summary,
        )

    def _extract_areas(self, model: DomainModel, max_areas: int) -> list[str]:
        """Extract research areas from bounded context names."""
        return [ctx.name for ctx in model.bounded_contexts[:max_areas]]

    async def _research_area(self, area: str) -> list[ResearchFinding]:
        """Run 2-round RLM search for a single domain area."""
        # Round 1: broad search
        results = await self._web.search(f"{area} domain patterns best practices")
        if not results:
            return []

        # Try LLM synthesis + refined search
        try:
            synthesis = await self._llm.text_completion(
                self._synthesis_prompt(area, results)
            )
            refined_query = self._extract_refined_query(synthesis.content, area)

            # Round 2: refined search
            results_2 = await self._web.search(refined_query)
            all_results = results + results_2
            return self._build_findings(area, all_results, Confidence.MEDIUM)

        except LLMUnavailableError:
            logger.info("LLM unavailable for area %s, using raw results", area)
            return self._build_findings(area, results, Confidence.LOW)
        except Exception:
            logger.warning(
                "Unexpected error researching area %s, using raw results",
                area,
                exc_info=True,
            )
            return self._build_findings(area, results, Confidence.LOW)

    def _synthesis_prompt(
        self, area: str, results: tuple[WebSearchResult, ...]
    ) -> str:
        """Build the LLM prompt for synthesizing search results."""
        snippets = "\n".join(
            f"- [{r.title}]({r.url}): {r.snippet}" for r in results[:5]
        )
        return (
            f"Analyze these search results about '{area}' domain patterns:\n\n"
            f"{snippets}\n\n"
            f"Provide:\n"
            f"1. A refined search query to find deeper information "
            f"(prefix with 'Refined query: ')\n"
            f"2. A brief synthesis of key patterns found (prefix with 'Synthesis: ')"
        )

    def _extract_refined_query(self, content: str, area: str) -> str:
        """Extract the refined query from LLM response."""
        for line in content.splitlines():
            if line.strip().lower().startswith("refined query:"):
                return line.split(":", 1)[1].strip()
        return f"{area} industry patterns competitive analysis"

    def _build_findings(
        self,
        area: str,
        results: tuple[WebSearchResult, ...],
        confidence: Confidence,
    ) -> list[ResearchFinding]:
        """Convert search results into ResearchFinding VOs."""
        today = date.today().isoformat()
        trust = (
            TrustLevel.AI_RESEARCHED
            if confidence != Confidence.LOW
            else TrustLevel.AI_INFERRED
        )
        return [
            ResearchFinding(
                content=r.snippet,
                source=SourceAttribution(
                    url=r.url,
                    title=r.title,
                    retrieved_date=today,
                    confidence=confidence,
                ),
                trust_level=trust,
                domain_area=area,
            )
            for r in results
            if r.url and r.snippet
        ]

    async def _build_summary(self, findings: list[ResearchFinding]) -> str:
        """Build a human-readable summary of all findings."""
        areas = {f.domain_area for f in findings}
        try:
            prompt = (
                f"Summarize these research findings across {len(areas)} domain area(s) "
                f"({', '.join(sorted(areas))}):\n\n"
                + "\n".join(f"- [{f.domain_area}] {f.content}" for f in findings[:10])
                + "\n\nProvide a 2-3 sentence summary."
            )
            response = await self._llm.text_completion(prompt)
            return response.content
        except Exception:
            return f"Research found {len(findings)} finding(s) across {len(areas)} area(s)."
