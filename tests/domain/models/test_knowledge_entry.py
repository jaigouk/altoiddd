"""Tests for Knowledge Base domain models.

Covers KnowledgeCategory enum, KnowledgePath value object,
EntryMetadata value object, and KnowledgeEntry entity.
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import InvariantViolationError

# ---------------------------------------------------------------------------
# KnowledgeCategory enum
# ---------------------------------------------------------------------------


class TestKnowledgeCategory:
    def test_knowledge_category_enum_values(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgeCategory

        assert KnowledgeCategory.DDD.value == "ddd"
        assert KnowledgeCategory.TOOLS.value == "tools"
        assert KnowledgeCategory.CONVENTIONS.value == "conventions"
        assert KnowledgeCategory.CROSS_TOOL.value == "cross-tool"


# ---------------------------------------------------------------------------
# KnowledgePath value object
# ---------------------------------------------------------------------------


class TestKnowledgePath:
    def test_knowledge_path_valid_ddd(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="ddd/tactical-patterns")
        assert path.raw == "ddd/tactical-patterns"

    def test_knowledge_path_valid_tools(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="tools/claude-code/agent-format")
        assert path.raw == "tools/claude-code/agent-format"

    def test_knowledge_path_valid_conventions(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="conventions/tdd")
        assert path.raw == "conventions/tdd"

    def test_knowledge_path_valid_cross_tool(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="cross-tool/agents-md")
        assert path.raw == "cross-tool/agents-md"

    def test_knowledge_path_rejects_empty(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        with pytest.raises(InvariantViolationError, match="empty"):
            KnowledgePath(raw="")

    def test_knowledge_path_rejects_traversal(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        with pytest.raises(InvariantViolationError, match="traversal"):
            KnowledgePath(raw="ddd/../secrets")

    def test_knowledge_path_rejects_invalid_category(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        with pytest.raises(InvariantViolationError, match="category"):
            KnowledgePath(raw="unknown/topic")

    def test_knowledge_path_category_extraction(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgeCategory, KnowledgePath

        assert KnowledgePath(raw="ddd/tactical-patterns").category == KnowledgeCategory.DDD
        assert KnowledgePath(raw="tools/cursor/rules-format").category == KnowledgeCategory.TOOLS
        assert KnowledgePath(raw="conventions/solid").category == KnowledgeCategory.CONVENTIONS
        assert KnowledgePath(raw="cross-tool/agents-md").category == KnowledgeCategory.CROSS_TOOL

    def test_knowledge_path_tool_extraction(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="tools/claude-code/agent-format")
        assert path.tool == "claude-code"

    def test_knowledge_path_tool_is_none_for_non_tools(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="ddd/tactical-patterns")
        assert path.tool is None

    def test_knowledge_path_subtopic_extraction(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="tools/claude-code/agent-format")
        assert path.subtopic == "agent-format"

    def test_knowledge_path_subtopic_is_none_for_non_tools(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="ddd/tactical-patterns")
        assert path.subtopic is None

    def test_knowledge_path_topic_ddd(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="ddd/tactical-patterns")
        assert path.topic == "tactical-patterns"

    def test_knowledge_path_topic_tools(self) -> None:
        """For tools, topic is tool/subtopic combined."""
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="tools/claude-code/agent-format")
        assert path.topic == "claude-code/agent-format"

    def test_knowledge_path_is_frozen(self) -> None:
        from src.domain.models.knowledge_entry import KnowledgePath

        path = KnowledgePath(raw="ddd/tactical-patterns")
        with pytest.raises(AttributeError):
            path.raw = "ddd/other"  # type: ignore[misc]


# ---------------------------------------------------------------------------
# EntryMetadata value object
# ---------------------------------------------------------------------------


class TestEntryMetadata:
    def test_entry_metadata_defaults(self) -> None:
        from src.domain.models.knowledge_entry import EntryMetadata

        meta = EntryMetadata()
        assert meta.last_verified is None
        assert meta.verified_against is None
        assert meta.confidence == "high"
        assert meta.deprecated is False
        assert meta.next_review_date is None
        assert meta.schema_version is None
        assert meta.source_urls == ()

    def test_entry_metadata_all_fields(self) -> None:
        from src.domain.models.knowledge_entry import EntryMetadata

        meta = EntryMetadata(
            last_verified="2026-01-01",
            verified_against="v2.0",
            confidence="medium",
            deprecated=True,
            next_review_date="2026-06-01",
            schema_version="1",
            source_urls=("https://example.com",),
        )
        assert meta.last_verified == "2026-01-01"
        assert meta.verified_against == "v2.0"
        assert meta.confidence == "medium"
        assert meta.deprecated is True
        assert meta.next_review_date == "2026-06-01"
        assert meta.schema_version == "1"
        assert meta.source_urls == ("https://example.com",)

    def test_entry_metadata_is_frozen(self) -> None:
        from src.domain.models.knowledge_entry import EntryMetadata

        meta = EntryMetadata()
        with pytest.raises(AttributeError):
            meta.confidence = "low"  # type: ignore[misc]


# ---------------------------------------------------------------------------
# KnowledgeEntry entity
# ---------------------------------------------------------------------------


class TestKnowledgeEntry:
    def test_knowledge_entry_equality_by_path(self) -> None:
        from src.domain.models.knowledge_entry import (
            KnowledgeEntry,
            KnowledgePath,
        )

        entry_a = KnowledgeEntry(
            path=KnowledgePath(raw="ddd/tactical-patterns"),
            title="Tactical Patterns",
            content="# Patterns",
        )
        entry_b = KnowledgeEntry(
            path=KnowledgePath(raw="ddd/tactical-patterns"),
            title="Different Title",
            content="Different content",
        )
        assert entry_a == entry_b
        assert hash(entry_a) == hash(entry_b)

    def test_knowledge_entry_inequality_different_paths(self) -> None:
        from src.domain.models.knowledge_entry import (
            KnowledgeEntry,
            KnowledgePath,
        )

        entry_a = KnowledgeEntry(
            path=KnowledgePath(raw="ddd/tactical-patterns"),
            title="Tactical Patterns",
            content="# Patterns",
        )
        entry_b = KnowledgeEntry(
            path=KnowledgePath(raw="ddd/strategic-patterns"),
            title="Tactical Patterns",
            content="# Patterns",
        )
        assert entry_a != entry_b

    def test_knowledge_entry_default_format(self) -> None:
        from src.domain.models.knowledge_entry import (
            KnowledgeEntry,
            KnowledgePath,
        )

        entry = KnowledgeEntry(
            path=KnowledgePath(raw="ddd/tactical-patterns"),
            title="Tactical Patterns",
            content="# Patterns",
        )
        assert entry.format == "markdown"
        assert entry.metadata is None
