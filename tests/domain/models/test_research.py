"""Tests for research domain value objects.

Verifies TrustLevel IntEnum ordering, Confidence enum, SourceAttribution
validation, WebSearchResult capture, ResearchFinding trust+source linkage,
and ResearchBriefing frozen aggregation.
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import InvariantViolationError
from src.domain.models.research import (
    Confidence,
    ResearchBriefing,
    ResearchFinding,
    SourceAttribution,
    TrustLevel,
    WebSearchResult,
)

# ---------------------------------------------------------------------------
# TrustLevel
# ---------------------------------------------------------------------------


class TestTrustLevel:
    def test_has_four_members(self) -> None:
        assert len(TrustLevel) == 4

    def test_ordering_lower_value_is_higher_trust(self) -> None:
        assert TrustLevel.USER_STATED < TrustLevel.USER_CONFIRMED
        assert TrustLevel.USER_CONFIRMED < TrustLevel.AI_RESEARCHED
        assert TrustLevel.AI_RESEARCHED < TrustLevel.AI_INFERRED

    def test_is_int(self) -> None:
        assert isinstance(TrustLevel.USER_STATED, int)

    def test_values(self) -> None:
        assert int(TrustLevel.USER_STATED) == 1
        assert int(TrustLevel.USER_CONFIRMED) == 2
        assert int(TrustLevel.AI_RESEARCHED) == 3
        assert int(TrustLevel.AI_INFERRED) == 4


# ---------------------------------------------------------------------------
# Confidence
# ---------------------------------------------------------------------------


class TestConfidence:
    def test_has_three_members(self) -> None:
        assert len(Confidence) == 3

    def test_values(self) -> None:
        assert Confidence.HIGH.value == "high"
        assert Confidence.MEDIUM.value == "medium"
        assert Confidence.LOW.value == "low"


# ---------------------------------------------------------------------------
# SourceAttribution
# ---------------------------------------------------------------------------


class TestSourceAttribution:
    def test_frozen(self) -> None:
        sa = SourceAttribution(
            url="https://example.com",
            title="Example",
            retrieved_date="2026-03-06",
            confidence=Confidence.HIGH,
        )
        with pytest.raises(AttributeError):
            sa.url = "other"  # type: ignore[misc]

    def test_requires_url(self) -> None:
        with pytest.raises(InvariantViolationError, match="url"):
            SourceAttribution(
                url="",
                title="Example",
                retrieved_date="2026-03-06",
                confidence=Confidence.HIGH,
            )

    def test_requires_title(self) -> None:
        with pytest.raises(InvariantViolationError, match="title"):
            SourceAttribution(
                url="https://example.com",
                title="",
                retrieved_date="2026-03-06",
                confidence=Confidence.HIGH,
            )

    def test_whitespace_only_url_rejected(self) -> None:
        with pytest.raises(InvariantViolationError, match="url"):
            SourceAttribution(
                url="   ",
                title="Example",
                retrieved_date="2026-03-06",
                confidence=Confidence.HIGH,
            )

    def test_valid_construction(self) -> None:
        sa = SourceAttribution(
            url="https://example.com",
            title="Example",
            retrieved_date="2026-03-06",
            confidence=Confidence.MEDIUM,
        )
        assert sa.url == "https://example.com"
        assert sa.title == "Example"
        assert sa.retrieved_date == "2026-03-06"
        assert sa.confidence == Confidence.MEDIUM


# ---------------------------------------------------------------------------
# WebSearchResult
# ---------------------------------------------------------------------------


class TestWebSearchResult:
    def test_frozen(self) -> None:
        wsr = WebSearchResult(
            url="https://example.com",
            title="Example",
            snippet="A snippet",
        )
        with pytest.raises(AttributeError):
            wsr.url = "other"  # type: ignore[misc]

    def test_captures_all_fields(self) -> None:
        wsr = WebSearchResult(
            url="https://example.com",
            title="Search Result",
            snippet="Some snippet text",
        )
        assert wsr.url == "https://example.com"
        assert wsr.title == "Search Result"
        assert wsr.snippet == "Some snippet text"


# ---------------------------------------------------------------------------
# ResearchFinding
# ---------------------------------------------------------------------------


class TestResearchFinding:
    def _make_source(self) -> SourceAttribution:
        return SourceAttribution(
            url="https://example.com",
            title="Source",
            retrieved_date="2026-03-06",
            confidence=Confidence.MEDIUM,
        )

    def test_frozen(self) -> None:
        finding = ResearchFinding(
            content="Some finding",
            source=self._make_source(),
            trust_level=TrustLevel.AI_RESEARCHED,
            domain_area="Sales",
        )
        with pytest.raises(AttributeError):
            finding.content = "other"  # type: ignore[misc]

    def test_carries_trust_level_and_source(self) -> None:
        source = self._make_source()
        finding = ResearchFinding(
            content="Industry pattern",
            source=source,
            trust_level=TrustLevel.AI_RESEARCHED,
            domain_area="Sales",
        )
        assert finding.trust_level == TrustLevel.AI_RESEARCHED
        assert finding.source is source
        assert finding.domain_area == "Sales"

    def test_outdated_defaults_false(self) -> None:
        finding = ResearchFinding(
            content="Finding",
            source=self._make_source(),
            trust_level=TrustLevel.AI_INFERRED,
            domain_area="Marketing",
        )
        assert finding.outdated is False


# ---------------------------------------------------------------------------
# ResearchBriefing
# ---------------------------------------------------------------------------


class TestResearchBriefing:
    def _make_finding(self, area: str = "Sales") -> ResearchFinding:
        return ResearchFinding(
            content="Some finding",
            source=SourceAttribution(
                url="https://example.com",
                title="Source",
                retrieved_date="2026-03-06",
                confidence=Confidence.MEDIUM,
            ),
            trust_level=TrustLevel.AI_RESEARCHED,
            domain_area=area,
        )

    def test_frozen(self) -> None:
        briefing = ResearchBriefing(
            findings=(),
            no_data_areas=(),
            summary="",
        )
        with pytest.raises(AttributeError):
            briefing.summary = "other"  # type: ignore[misc]

    def test_separates_findings_from_no_data(self) -> None:
        finding = self._make_finding("Sales")
        briefing = ResearchBriefing(
            findings=(finding,),
            no_data_areas=("Marketing",),
            summary="Partial research",
        )
        assert len(briefing.findings) == 1
        assert briefing.findings[0].domain_area == "Sales"
        assert briefing.no_data_areas == ("Marketing",)

    def test_empty_briefing(self) -> None:
        briefing = ResearchBriefing(
            findings=(),
            no_data_areas=("Sales", "Marketing"),
            summary="",
        )
        assert len(briefing.findings) == 0
        assert len(briefing.no_data_areas) == 2
        assert briefing.summary == ""
