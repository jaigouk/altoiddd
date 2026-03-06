"""Research domain value objects for the Domain Research bounded context.

TrustLevel classifies how much a piece of knowledge can be trusted:
USER_STATED (highest) → USER_CONFIRMED → AI_RESEARCHED → AI_INFERRED (lowest).

SourceAttribution, WebSearchResult, ResearchFinding, and ResearchBriefing
are frozen VOs that carry research results with provenance metadata.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass

from src.domain.models.errors import InvariantViolationError


class TrustLevel(enum.IntEnum):
    """How much a piece of knowledge can be trusted.

    Lower numeric value = higher trust. IntEnum enables comparison operators.
    """

    USER_STATED = 1
    USER_CONFIRMED = 2
    AI_RESEARCHED = 3
    AI_INFERRED = 4


class Confidence(enum.Enum):
    """Confidence level of a research finding or source."""

    HIGH = "high"
    MEDIUM = "medium"
    LOW = "low"


@dataclass(frozen=True)
class SourceAttribution:
    """Provenance metadata for a research finding.

    Attributes:
        url: Where the information was found (must not be empty).
        title: Human-readable source title (must not be empty).
        retrieved_date: ISO-format date when the source was accessed.
        confidence: How confident the system is in this source.
    """

    url: str
    title: str
    retrieved_date: str
    confidence: Confidence

    def __post_init__(self) -> None:
        if not self.url.strip():
            msg = "SourceAttribution url cannot be empty"
            raise InvariantViolationError(msg)
        if not self.title.strip():
            msg = "SourceAttribution title cannot be empty"
            raise InvariantViolationError(msg)


@dataclass(frozen=True)
class WebSearchResult:
    """Raw result from a web search query.

    Attributes:
        url: Result URL.
        title: Result title.
        snippet: Text snippet from the search engine.
    """

    url: str
    title: str
    snippet: str


@dataclass(frozen=True)
class ResearchFinding:
    """A single research insight with source attribution and trust level.

    Attributes:
        content: The finding text.
        source: Where this finding came from.
        trust_level: How much to trust this finding.
        domain_area: Which domain area this relates to.
        outdated: Whether this finding may be stale (MVP: always False).
    """

    content: str
    source: SourceAttribution
    trust_level: TrustLevel
    domain_area: str
    outdated: bool = False


@dataclass(frozen=True)
class ResearchBriefing:
    """Complete research output for presentation to the user.

    Attributes:
        findings: Research findings with provenance.
        no_data_areas: Domain areas where no useful data was found.
        summary: Human-readable summary of findings.
    """

    findings: tuple[ResearchFinding, ...]
    no_data_areas: tuple[str, ...]
    summary: str
