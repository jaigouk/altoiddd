"""Tests for the ``alty kb`` CLI command (2j7.6).

Covers: topic lookup, category listing, unknown topic error,
missing .alty/ directory error, and tool-specific topics.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

import pytest
from typer.testing import CliRunner

from src.infrastructure.cli.main import app

if TYPE_CHECKING:
    from pathlib import Path

runner = CliRunner()


@pytest.fixture
def knowledge_dir(tmp_path: Path) -> Path:
    """Create a minimal .alty/knowledge/ tree for CLI tests."""
    kb = tmp_path / ".alty" / "knowledge"
    # DDD category
    ddd = kb / "ddd"
    ddd.mkdir(parents=True)
    (ddd / "tactical-patterns.md").write_text(
        "---\ntitle: Tactical Patterns\n---\n\n# Tactical Patterns\n\nAggregates, Entities, VOs."
    )
    # Conventions category
    conv = kb / "conventions"
    conv.mkdir(parents=True)
    (conv / "tdd.md").write_text("---\ntitle: TDD\n---\n\n# TDD\n\nRed, Green, Refactor.")
    # Tools category
    tools = kb / "tools" / "claude-code" / "current"
    tools.mkdir(parents=True)
    (tools / "config-structure.toml").write_text(
        '[_meta]\nlast_verified = "2026-03-01"\nconfidence = "high"\n\n'
        '[config]\ndescription = "Claude Code config structure"\n'
    )
    # Cross-tool category
    cross = kb / "cross-tool"
    cross.mkdir(parents=True)
    (cross / "concept-mapping.toml").write_text(
        '[_meta]\nconfidence = "medium"\n\n[mapping]\ndescription = "Cross-tool concepts"\n'
    )
    return tmp_path


class TestKbLookup:
    """alty kb <topic> looks up and displays knowledge entries."""

    def test_lookup_ddd_topic(self, knowledge_dir: Path) -> None:
        """alty kb ddd/tactical-patterns returns the entry content."""
        result = runner.invoke(
            app,
            ["kb", "ddd/tactical-patterns"],
            env={"ALTY_PROJECT_DIR": str(knowledge_dir)},
        )
        assert result.exit_code == 0
        assert "Tactical Patterns" in result.output

    def test_lookup_conventions_topic(self, knowledge_dir: Path) -> None:
        """alty kb conventions/tdd returns the entry content."""
        result = runner.invoke(
            app,
            ["kb", "conventions/tdd"],
            env={"ALTY_PROJECT_DIR": str(knowledge_dir)},
        )
        assert result.exit_code == 0
        assert "TDD" in result.output

    def test_lookup_tools_topic(self, knowledge_dir: Path) -> None:
        """alty kb tools/claude-code/config-structure returns TOML content."""
        result = runner.invoke(
            app,
            ["kb", "tools/claude-code/config-structure"],
            env={"ALTY_PROJECT_DIR": str(knowledge_dir)},
        )
        assert result.exit_code == 0
        assert "config" in result.output.lower()

    def test_lookup_cross_tool_topic(self, knowledge_dir: Path) -> None:
        """alty kb cross-tool/concept-mapping returns TOML content."""
        result = runner.invoke(
            app,
            ["kb", "cross-tool/concept-mapping"],
            env={"ALTY_PROJECT_DIR": str(knowledge_dir)},
        )
        assert result.exit_code == 0
        assert "mapping" in result.output.lower()


class TestKbListCategories:
    """alty kb (no args) lists available categories and topics."""

    def test_no_args_lists_categories(self, knowledge_dir: Path) -> None:
        """alty kb with empty topic lists categories."""
        result = runner.invoke(
            app,
            ["kb"],
            env={"ALTY_PROJECT_DIR": str(knowledge_dir)},
        )
        assert result.exit_code == 0
        assert "ddd" in result.output
        assert "conventions" in result.output
        assert "tools" in result.output
        assert "cross-tool" in result.output


class TestKbErrors:
    """Error handling for unknown topics and missing directories."""

    def test_unknown_topic_shows_error(self, knowledge_dir: Path) -> None:
        """Unknown topic returns exit code 1 with helpful message."""
        result = runner.invoke(
            app,
            ["kb", "ddd/nonexistent"],
            env={"ALTY_PROJECT_DIR": str(knowledge_dir)},
        )
        assert result.exit_code == 1
        assert "not found" in result.output.lower()

    def test_unknown_category_shows_error(self, knowledge_dir: Path) -> None:
        """Invalid category returns exit code 1 with valid categories."""
        result = runner.invoke(
            app,
            ["kb", "foo/bar"],
            env={"ALTY_PROJECT_DIR": str(knowledge_dir)},
        )
        assert result.exit_code == 1

    def test_missing_alty_dir_shows_error(self, tmp_path: Path) -> None:
        """No .alty/ directory returns exit code 1 with init message."""
        result = runner.invoke(
            app,
            ["kb", "ddd/tactical-patterns"],
            env={"ALTY_PROJECT_DIR": str(tmp_path)},
        )
        assert result.exit_code == 1
        assert "no .alty" in result.output.lower() or "not found" in result.output.lower()
