"""Tests for KnowledgeLookupHandler.

Covers query-side orchestration: delegating to a KnowledgeReaderProtocol,
listing categories, and listing topics.
"""

from __future__ import annotations

import pytest

from src.domain.models.errors import InvariantViolationError

# ---------------------------------------------------------------------------
# Fake reader for testing
# ---------------------------------------------------------------------------


class FakeKnowledgeReader:
    """In-memory knowledge reader for handler tests."""

    def __init__(self) -> None:
        from src.domain.models.knowledge_entry import (
            EntryMetadata,
            KnowledgeEntry,
            KnowledgePath,
        )

        self._entries: dict[str, KnowledgeEntry] = {
            "ddd/tactical-patterns": KnowledgeEntry(
                path=KnowledgePath(raw="ddd/tactical-patterns"),
                title="Tactical Patterns",
                content="# Tactical Patterns\nAggregates, Entities, VOs.",
                metadata=EntryMetadata(confidence="high"),
            ),
            "tools/claude-code/agent-format": KnowledgeEntry(
                path=KnowledgePath(raw="tools/claude-code/agent-format"),
                title="Agent Format",
                content="# Agent Format",
                metadata=EntryMetadata(confidence="high"),
                format="toml",
            ),
        }
        self._topics: dict[str, tuple[str, ...]] = {
            "ddd": ("tactical-patterns", "strategic-patterns"),
            "tools:claude-code": ("agent-format", "config-structure"),
        }

    def read_entry(
        self,
        path: object,
        version: str = "current",
    ) -> object:
        from src.domain.models.knowledge_entry import KnowledgePath

        assert isinstance(path, KnowledgePath)
        if path.raw not in self._entries:
            msg = f"Entry not found: {path.raw}"
            raise InvariantViolationError(msg)
        return self._entries[path.raw]

    def list_topics(
        self,
        category: object,
        tool: str | None = None,
    ) -> tuple[str, ...]:
        from src.domain.models.knowledge_entry import KnowledgeCategory

        assert isinstance(category, KnowledgeCategory)
        key = category.value
        if tool:
            key = f"{key}:{tool}"
        return self._topics.get(key, ())


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


class TestKnowledgeLookupHandler:
    def test_handler_lookup_delegates_to_reader(self) -> None:
        from src.application.queries.knowledge_lookup_handler import (
            KnowledgeLookupHandler,
        )

        reader = FakeKnowledgeReader()
        handler = KnowledgeLookupHandler(reader=reader)  # type: ignore[arg-type]

        entry = handler.lookup("ddd/tactical-patterns")
        assert entry.title == "Tactical Patterns"
        assert "Aggregates" in entry.content

    def test_handler_lookup_tools_path(self) -> None:
        from src.application.queries.knowledge_lookup_handler import (
            KnowledgeLookupHandler,
        )

        reader = FakeKnowledgeReader()
        handler = KnowledgeLookupHandler(reader=reader)  # type: ignore[arg-type]

        entry = handler.lookup("tools/claude-code/agent-format")
        assert entry.title == "Agent Format"

    def test_handler_list_categories(self) -> None:
        from src.application.queries.knowledge_lookup_handler import (
            KnowledgeLookupHandler,
        )

        reader = FakeKnowledgeReader()
        handler = KnowledgeLookupHandler(reader=reader)  # type: ignore[arg-type]

        categories = handler.list_categories()
        assert "ddd" in categories
        assert "tools" in categories
        assert "conventions" in categories
        assert "cross-tool" in categories

    def test_handler_list_topics_delegates_to_reader(self) -> None:
        from src.application.queries.knowledge_lookup_handler import (
            KnowledgeLookupHandler,
        )

        reader = FakeKnowledgeReader()
        handler = KnowledgeLookupHandler(reader=reader)  # type: ignore[arg-type]

        topics = handler.list_topics("ddd")
        assert "tactical-patterns" in topics
        assert "strategic-patterns" in topics

    def test_handler_list_topics_with_tool(self) -> None:
        from src.application.queries.knowledge_lookup_handler import (
            KnowledgeLookupHandler,
        )

        reader = FakeKnowledgeReader()
        handler = KnowledgeLookupHandler(reader=reader)  # type: ignore[arg-type]

        topics = handler.list_topics("tools", tool="claude-code")
        assert "agent-format" in topics
        assert "config-structure" in topics

    def test_handler_lookup_invalid_path_raises(self) -> None:
        from src.application.queries.knowledge_lookup_handler import (
            KnowledgeLookupHandler,
        )

        reader = FakeKnowledgeReader()
        handler = KnowledgeLookupHandler(reader=reader)  # type: ignore[arg-type]

        with pytest.raises(InvariantViolationError):
            handler.lookup("")

    def test_handler_lookup_nonexistent_entry_raises(self) -> None:
        from src.application.queries.knowledge_lookup_handler import (
            KnowledgeLookupHandler,
        )

        reader = FakeKnowledgeReader()
        handler = KnowledgeLookupHandler(reader=reader)  # type: ignore[arg-type]

        with pytest.raises(InvariantViolationError, match="not found"):
            handler.lookup("ddd/nonexistent-topic")

    def test_handler_list_topics_invalid_category_raises(self) -> None:
        from src.application.queries.knowledge_lookup_handler import (
            KnowledgeLookupHandler,
        )

        reader = FakeKnowledgeReader()
        handler = KnowledgeLookupHandler(reader=reader)  # type: ignore[arg-type]

        with pytest.raises(InvariantViolationError, match="category"):
            handler.list_topics("invalid-category")
