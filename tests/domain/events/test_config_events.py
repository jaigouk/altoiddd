"""Tests for ConfigsGenerated domain event."""

from __future__ import annotations

import pytest

from src.domain.events.config_events import ConfigsGenerated


class TestConfigsGenerated:
    def test_create_event(self):
        event = ConfigsGenerated(
            tool_names=("claude-code", "cursor"),
            output_paths=(".claude/CLAUDE.md", "AGENTS.md"),
        )
        assert event.tool_names == ("claude-code", "cursor")
        assert event.output_paths == (".claude/CLAUDE.md", "AGENTS.md")

    def test_event_is_frozen(self):
        event = ConfigsGenerated(
            tool_names=("claude-code",),
            output_paths=(".claude/CLAUDE.md",),
        )
        with pytest.raises(AttributeError):
            event.tool_names = ("changed",)  # type: ignore[misc]

    def test_event_stores_tuple_fields(self):
        event = ConfigsGenerated(
            tool_names=("claude-code", "cursor", "roo-code"),
            output_paths=("a.md", "b.md", "c.md"),
        )
        assert isinstance(event.tool_names, tuple)
        assert isinstance(event.output_paths, tuple)
        assert len(event.tool_names) == 3
        assert len(event.output_paths) == 3

    def test_empty_tuples_allowed(self):
        event = ConfigsGenerated(
            tool_names=(),
            output_paths=(),
        )
        assert event.tool_names == ()
        assert event.output_paths == ()
