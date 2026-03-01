"""Domain models for the Knowledge Base bounded context.

KnowledgeCategory enumerates the four knowledge categories.
KnowledgePath is a frozen value object for RLM-addressable paths.
EntryMetadata holds verification and confidence metadata.
KnowledgeEntry is an entity identified by its KnowledgePath.
"""

from __future__ import annotations

import enum
from dataclasses import dataclass, field

from src.domain.models.errors import InvariantViolationError

_VALID_CATEGORIES = {"ddd", "tools", "conventions", "cross-tool"}


class KnowledgeCategory(enum.Enum):
    """Categories of knowledge entries in the Knowledge Base."""

    DDD = "ddd"
    TOOLS = "tools"
    CONVENTIONS = "conventions"
    CROSS_TOOL = "cross-tool"


@dataclass(frozen=True)
class KnowledgePath:
    """RLM-addressable path to a knowledge entry.

    Format: ``{category}/{topic}`` or ``{category}/{tool}/{subtopic}`` for tools.

    Attributes:
        raw: The original path string, e.g. "ddd/tactical-patterns".
    """

    raw: str

    def __post_init__(self) -> None:
        if not self.raw:
            msg = "Knowledge path must not be empty"
            raise InvariantViolationError(msg)
        if ".." in self.raw:
            msg = "Knowledge path must not contain path traversal (..)"
            raise InvariantViolationError(msg)
        segments = self.raw.split("/")
        if segments[0] not in _VALID_CATEGORIES:
            msg = (
                f"Knowledge path must start with a valid category "
                f"({', '.join(sorted(_VALID_CATEGORIES))}), got '{segments[0]}'"
            )
            raise InvariantViolationError(msg)

    @property
    def category(self) -> KnowledgeCategory:
        """Extract the category from the first path segment."""
        first = self.raw.split("/")[0]
        return KnowledgeCategory(first)

    @property
    def topic(self) -> str:
        """Extract the topic portion of the path.

        For tools paths (tools/tool-name/subtopic), returns "tool-name/subtopic".
        For other categories, returns the second segment.
        """
        segments = self.raw.split("/")
        if self.category == KnowledgeCategory.TOOLS:
            return "/".join(segments[1:])
        return segments[1]

    @property
    def tool(self) -> str | None:
        """Extract the tool name for TOOLS category paths, None otherwise."""
        if self.category != KnowledgeCategory.TOOLS:
            return None
        return self.raw.split("/")[1]

    @property
    def subtopic(self) -> str | None:
        """Extract the subtopic for TOOLS category paths, None otherwise."""
        if self.category != KnowledgeCategory.TOOLS:
            return None
        segments = self.raw.split("/")
        if len(segments) < 3:
            return None
        return segments[2]


@dataclass(frozen=True)
class EntryMetadata:
    """Verification and confidence metadata for a knowledge entry.

    Attributes:
        last_verified: ISO date when entry was last verified.
        verified_against: Tool/library version verified against.
        confidence: Confidence level (high, medium, low).
        deprecated: Whether this entry is deprecated.
        next_review_date: ISO date for next scheduled review.
        schema_version: Schema version of the entry format.
        source_urls: Authoritative source URLs for this entry.
    """

    last_verified: str | None = None
    verified_against: str | None = None
    confidence: str = "high"
    deprecated: bool = False
    next_review_date: str | None = None
    schema_version: str | None = None
    source_urls: tuple[str, ...] = ()


@dataclass
class KnowledgeEntry:
    """A knowledge entry in the Knowledge Base.

    Entity identity is based on the KnowledgePath -- two entries with the
    same path are considered equal regardless of content differences.

    Attributes:
        path: The RLM-addressable path for this entry.
        title: Human-readable title.
        content: The entry content (markdown or serialized TOML).
        metadata: Optional verification metadata.
        format: Content format, e.g. "markdown" or "toml".
    """

    path: KnowledgePath
    title: str
    content: str
    metadata: EntryMetadata | None = field(default=None)
    format: str = "markdown"

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, KnowledgeEntry):
            return NotImplemented
        return self.path.raw == other.path.raw

    def __hash__(self) -> int:
        return hash(self.path.raw)
