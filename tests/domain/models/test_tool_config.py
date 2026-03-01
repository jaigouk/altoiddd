"""Tests for ToolConfig aggregate root and related value objects.

Covers section generation, preview, approve, invariant enforcement,
and event emission per k7m.21 ticket.
"""

from __future__ import annotations

import pytest

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.tool_config import ConfigSection, SupportedTool, ToolConfig

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_model_with_contexts(
    contexts: list[tuple[str, SubdomainClassification]],
) -> DomainModel:
    """Build a minimal valid DomainModel with the given classified contexts."""
    model = DomainModel()
    all_names = [name for name, _ in contexts]
    model.add_domain_story(
        DomainStory(
            name="Test flow",
            actors=("User",),
            trigger="User starts",
            steps=tuple(f"User manages {name}" for name in all_names),
        )
    )
    for name, classification in contexts:
        model.add_term(
            term=name,
            definition=f"{name} domain",
            context_name=name,
        )
        model.add_bounded_context(
            BoundedContext(
                name=name,
                responsibility=f"Manages {name}",
                classification=classification,
            )
        )
    for name, classification in contexts:
        if classification == SubdomainClassification.CORE:
            model.design_aggregate(
                AggregateDesign(
                    name=f"{name}Root",
                    context_name=name,
                    root_entity=f"{name}Root",
                    invariants=("must be valid",),
                )
            )
    model.finalize()
    return model


class FakeAdapter:
    """Adapter that returns a fixed set of sections."""

    def __init__(self, sections: tuple[ConfigSection, ...] | None = None):
        self._sections = sections or (
            ConfigSection(
                file_path=".claude/CLAUDE.md",
                content="# Test",
                section_name="Test section",
            ),
        )

    def translate(self, model: DomainModel) -> tuple[ConfigSection, ...]:
        return self._sections


# ---------------------------------------------------------------------------
# 1. SupportedTool enum
# ---------------------------------------------------------------------------


class TestSupportedTool:
    def test_has_four_values(self):
        assert len(SupportedTool) == 4

    def test_claude_code_value(self):
        assert SupportedTool.CLAUDE_CODE.value == "claude-code"

    def test_cursor_value(self):
        assert SupportedTool.CURSOR.value == "cursor"

    def test_roo_code_value(self):
        assert SupportedTool.ROO_CODE.value == "roo-code"

    def test_opencode_value(self):
        assert SupportedTool.OPENCODE.value == "opencode"


# ---------------------------------------------------------------------------
# 2. ConfigSection frozen dataclass
# ---------------------------------------------------------------------------


class TestConfigSection:
    def test_is_frozen(self):
        section = ConfigSection(
            file_path=".claude/CLAUDE.md",
            content="# Test",
            section_name="Test",
        )
        with pytest.raises(AttributeError):
            section.file_path = "changed"  # type: ignore[misc]

    def test_stores_fields(self):
        section = ConfigSection(
            file_path="AGENTS.md",
            content="content here",
            section_name="agents file",
        )
        assert section.file_path == "AGENTS.md"
        assert section.content == "content here"
        assert section.section_name == "agents file"


# ---------------------------------------------------------------------------
# 3. ToolConfig creation
# ---------------------------------------------------------------------------


class TestToolConfigCreation:
    def test_new_config_has_empty_sections(self):
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        assert config.sections == ()
        assert config.events == ()

    def test_config_has_unique_id(self):
        a = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        b = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        assert a.config_id != b.config_id

    def test_config_stores_tool(self):
        config = ToolConfig(tool=SupportedTool.CURSOR)
        assert config.tool == SupportedTool.CURSOR


# ---------------------------------------------------------------------------
# 4. build_sections
# ---------------------------------------------------------------------------


class TestBuildSections:
    def test_builds_sections_from_adapter(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        adapter = FakeAdapter()
        config.build_sections(model=model, adapter=adapter)
        assert len(config.sections) == 1
        assert config.sections[0].file_path == ".claude/CLAUDE.md"

    def test_build_clears_previous_sections(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        adapter = FakeAdapter()
        config.build_sections(model=model, adapter=adapter)
        assert len(config.sections) == 1

        # Rebuild with different adapter
        new_sections = (
            ConfigSection(file_path="a.md", content="a", section_name="A"),
            ConfigSection(file_path="b.md", content="b", section_name="B"),
        )
        config.build_sections(model=model, adapter=FakeAdapter(new_sections))
        assert len(config.sections) == 2

    def test_cannot_build_after_approve(self):
        from src.domain.models.errors import InvariantViolationError

        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        adapter = FakeAdapter()
        config.build_sections(model=model, adapter=adapter)
        config.approve()

        with pytest.raises(InvariantViolationError, match="approved"):
            config.build_sections(model=model, adapter=adapter)


# ---------------------------------------------------------------------------
# 5. preview
# ---------------------------------------------------------------------------


class TestPreview:
    def test_preview_returns_string(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        config.build_sections(model=model, adapter=FakeAdapter())
        preview = config.preview()
        assert isinstance(preview, str)
        assert "claude-code" in preview

    def test_preview_without_sections_raises(self):
        from src.domain.models.errors import InvariantViolationError

        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        with pytest.raises(InvariantViolationError, match="No sections"):
            config.preview()


# ---------------------------------------------------------------------------
# 6. approve
# ---------------------------------------------------------------------------


class TestApprove:
    def test_approve_emits_configs_generated(self):
        from src.domain.events.config_events import ConfigsGenerated

        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        config.build_sections(model=model, adapter=FakeAdapter())
        config.approve()

        assert len(config.events) == 1
        assert isinstance(config.events[0], ConfigsGenerated)

    def test_event_contains_tool_name(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        config = ToolConfig(tool=SupportedTool.CURSOR)
        config.build_sections(model=model, adapter=FakeAdapter())
        config.approve()

        assert config.events[0].tool_names == ("cursor",)

    def test_event_contains_output_paths(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        config.build_sections(model=model, adapter=FakeAdapter())
        config.approve()

        assert ".claude/CLAUDE.md" in config.events[0].output_paths

    def test_cannot_approve_without_sections(self):
        from src.domain.models.errors import InvariantViolationError

        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        with pytest.raises(InvariantViolationError, match="no sections"):
            config.approve()

    def test_cannot_approve_twice(self):
        from src.domain.models.errors import InvariantViolationError

        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        config.build_sections(model=model, adapter=FakeAdapter())
        config.approve()

        with pytest.raises(InvariantViolationError, match="already approved"):
            config.approve()


# ---------------------------------------------------------------------------
# 7. Defensive copies
# ---------------------------------------------------------------------------


class TestDefensiveCopies:
    def test_sections_returns_tuple(self):
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        assert isinstance(config.sections, tuple)

    def test_events_returns_tuple(self):
        config = ToolConfig(tool=SupportedTool.CLAUDE_CODE)
        assert isinstance(config.events, tuple)
