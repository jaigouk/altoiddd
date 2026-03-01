"""Tests for FileKnowledgeReader infrastructure adapter.

Covers reading markdown and TOML entries from the filesystem,
parsing [_meta] sections, error handling for missing files, listing
topics, and resolving versioned tool paths.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

import pytest

from src.domain.models.errors import InvariantViolationError
from src.domain.models.knowledge_entry import KnowledgeCategory, KnowledgePath

if TYPE_CHECKING:
    from pathlib import Path

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _create_markdown_entry(base: Path, category: str, topic: str, content: str) -> Path:
    """Create a markdown knowledge entry on disk."""
    entry_dir = base / category
    entry_dir.mkdir(parents=True, exist_ok=True)
    file_path = entry_dir / f"{topic}.md"
    file_path.write_text(content)
    return file_path


def _create_toml_entry(
    base: Path,
    tool: str,
    version: str,
    topic: str,
    content: str,
) -> Path:
    """Create a TOML tool knowledge entry on disk."""
    entry_dir = base / "tools" / tool / version
    entry_dir.mkdir(parents=True, exist_ok=True)
    file_path = entry_dir / f"{topic}.toml"
    file_path.write_text(content)
    return file_path


def _create_cross_tool_entry(base: Path, topic: str, content: str) -> Path:
    """Create a cross-tool TOML entry on disk."""
    entry_dir = base / "cross-tool"
    entry_dir.mkdir(parents=True, exist_ok=True)
    file_path = entry_dir / f"{topic}.toml"
    file_path.write_text(content)
    return file_path


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


class TestFileKnowledgeReaderMarkdown:
    def test_reader_reads_markdown_entry(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        _create_markdown_entry(
            tmp_path, "ddd", "tactical-patterns", "# Tactical Patterns\n\nContent here."
        )
        reader = FileKnowledgeReader(knowledge_dir=tmp_path)
        path = KnowledgePath(raw="ddd/tactical-patterns")

        entry = reader.read_entry(path)

        assert entry.title == "tactical-patterns"
        assert "Tactical Patterns" in entry.content
        assert entry.format == "markdown"

    def test_reader_reads_conventions_markdown(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        _create_markdown_entry(tmp_path, "conventions", "tdd", "# TDD\n\nRed-Green-Refactor.")
        reader = FileKnowledgeReader(knowledge_dir=tmp_path)
        path = KnowledgePath(raw="conventions/tdd")

        entry = reader.read_entry(path)

        assert entry.title == "tdd"
        assert "Red-Green-Refactor" in entry.content


class TestFileKnowledgeReaderToml:
    def test_reader_reads_toml_entry(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        toml_content = """\
[format]
file_extension = ".md"
description = "Agent definition format"
"""
        _create_toml_entry(tmp_path, "claude-code", "current", "agent-format", toml_content)
        reader = FileKnowledgeReader(knowledge_dir=tmp_path)
        path = KnowledgePath(raw="tools/claude-code/agent-format")

        entry = reader.read_entry(path)

        assert entry.title == "agent-format"
        assert entry.format == "toml"
        assert "Agent definition format" in entry.content

    def test_reader_reads_toml_metadata(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        toml_content = """\
[_meta]
last_verified = "2026-01-15"
verified_against = "v2.0"
confidence = "medium"
deprecated = true
next_review_date = "2026-07-01"
schema_version = 1
source_urls = ["https://example.com/docs"]

[format]
description = "Test entry"
"""
        _create_toml_entry(tmp_path, "claude-code", "current", "config-structure", toml_content)
        reader = FileKnowledgeReader(knowledge_dir=tmp_path)
        path = KnowledgePath(raw="tools/claude-code/config-structure")

        entry = reader.read_entry(path)

        assert entry.metadata is not None
        assert entry.metadata.last_verified == "2026-01-15"
        assert entry.metadata.verified_against == "v2.0"
        assert entry.metadata.confidence == "medium"
        assert entry.metadata.deprecated is True
        assert entry.metadata.next_review_date == "2026-07-01"
        assert entry.metadata.schema_version == "1"
        assert entry.metadata.source_urls == ("https://example.com/docs",)

    def test_reader_reads_cross_tool_toml(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        toml_content = """\
[mapping]
description = "Cross-tool concept mapping"
"""
        _create_cross_tool_entry(tmp_path, "concept-mapping", toml_content)
        reader = FileKnowledgeReader(knowledge_dir=tmp_path)
        path = KnowledgePath(raw="cross-tool/concept-mapping")

        entry = reader.read_entry(path)

        assert entry.title == "concept-mapping"
        assert entry.format == "toml"

    def test_reader_resolves_tool_version_path(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        toml_content = "[format]\ndescription = 'v1 content'\n"
        _create_toml_entry(tmp_path, "cursor", "v1", "rules-format", toml_content)
        reader = FileKnowledgeReader(knowledge_dir=tmp_path)
        path = KnowledgePath(raw="tools/cursor/rules-format")

        entry = reader.read_entry(path, version="v1")

        assert entry.title == "rules-format"
        assert "v1 content" in entry.content


class TestFileKnowledgeReaderErrors:
    def test_reader_not_found_raises(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        reader = FileKnowledgeReader(knowledge_dir=tmp_path)
        path = KnowledgePath(raw="ddd/nonexistent-topic")

        with pytest.raises(InvariantViolationError, match="not found"):
            reader.read_entry(path)

    def test_reader_tool_not_found_raises(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        reader = FileKnowledgeReader(knowledge_dir=tmp_path)
        path = KnowledgePath(raw="tools/nonexistent-tool/some-topic")

        with pytest.raises(InvariantViolationError, match="not found"):
            reader.read_entry(path)


class TestFileKnowledgeReaderListTopics:
    def test_reader_list_topics(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        _create_markdown_entry(tmp_path, "ddd", "tactical-patterns", "# TP")
        _create_markdown_entry(tmp_path, "ddd", "strategic-patterns", "# SP")
        reader = FileKnowledgeReader(knowledge_dir=tmp_path)

        topics = reader.list_topics(KnowledgeCategory.DDD)

        assert topics == ("strategic-patterns", "tactical-patterns")

    def test_reader_list_tool_topics(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        _create_toml_entry(tmp_path, "claude-code", "current", "agent-format", "[f]\nx = 1\n")
        _create_toml_entry(tmp_path, "claude-code", "current", "config-structure", "[f]\nx = 1\n")
        reader = FileKnowledgeReader(knowledge_dir=tmp_path)

        topics = reader.list_topics(KnowledgeCategory.TOOLS, tool="claude-code")

        assert topics == ("agent-format", "config-structure")

    def test_reader_list_topics_empty_dir(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        reader = FileKnowledgeReader(knowledge_dir=tmp_path)

        topics = reader.list_topics(KnowledgeCategory.DDD)
        assert topics == ()

    def test_reader_list_cross_tool_topics(self, tmp_path: Path) -> None:
        from src.infrastructure.persistence.file_knowledge_reader import (
            FileKnowledgeReader,
        )

        _create_cross_tool_entry(tmp_path, "agents-md", "[f]\nx = 1\n")
        _create_cross_tool_entry(tmp_path, "concept-mapping", "[f]\nx = 1\n")
        reader = FileKnowledgeReader(knowledge_dir=tmp_path)

        topics = reader.list_topics(KnowledgeCategory.CROSS_TOOL)

        assert topics == ("agents-md", "concept-mapping")
