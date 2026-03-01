"""Tests for ConfigGenerationHandler.

Covers the application-layer orchestration: building ToolConfig aggregates
from a DomainModel, preview-before-write, and file output.
"""

from __future__ import annotations

from pathlib import Path

import pytest

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.tool_config import SupportedTool

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


class FakeFileWriter:
    """In-memory file writer for testing."""

    def __init__(self) -> None:
        self.written: dict[str, str] = {}

    def write_file(self, path: Path, content: str) -> None:
        self.written[str(path)] = content


# ---------------------------------------------------------------------------
# 1. Build preview
# ---------------------------------------------------------------------------


class TestBuildPreview:
    def test_build_preview_returns_config_preview(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
            ConfigPreview,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)

        preview = handler.build_preview(
            model, tools=(SupportedTool.CLAUDE_CODE,)
        )

        assert isinstance(preview, ConfigPreview)
        assert len(preview.configs) == 1
        assert preview.summary  # non-empty

    def test_build_preview_does_not_write(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)
        handler.build_preview(model, tools=(SupportedTool.CLAUDE_CODE,))

        assert writer.written == {}

    def test_preview_for_multiple_tools(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)

        preview = handler.build_preview(
            model,
            tools=(SupportedTool.CLAUDE_CODE, SupportedTool.CURSOR),
        )

        assert len(preview.configs) == 2
        assert "claude-code" in preview.summary
        assert "cursor" in preview.summary

    def test_empty_tools_raises_value_error(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)

        with pytest.raises(ValueError, match="No tools"):
            handler.build_preview(model, tools=())


# ---------------------------------------------------------------------------
# 2. Approve and write
# ---------------------------------------------------------------------------


class TestApproveAndWrite:
    def test_approve_and_write_writes_files(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)
        preview = handler.build_preview(
            model, tools=(SupportedTool.CLAUDE_CODE,)
        )

        handler.approve_and_write(preview, output_dir=Path("/project"))

        assert len(writer.written) >= 1
        paths = list(writer.written.keys())
        assert any(".claude/CLAUDE.md" in p for p in paths)

    def test_approve_and_write_emits_events(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )
        from src.domain.events.config_events import ConfigsGenerated

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)
        preview = handler.build_preview(
            model, tools=(SupportedTool.CLAUDE_CODE,)
        )

        handler.approve_and_write(preview, output_dir=Path("/project"))

        for config in preview.configs:
            assert len(config.events) == 1
            assert isinstance(config.events[0], ConfigsGenerated)

    def test_approve_and_write_twice_raises(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )
        from src.domain.models.errors import InvariantViolationError

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)
        preview = handler.build_preview(
            model, tools=(SupportedTool.CLAUDE_CODE,)
        )

        handler.approve_and_write(preview, output_dir=Path("/project"))

        with pytest.raises(InvariantViolationError, match="already approved"):
            handler.approve_and_write(preview, output_dir=Path("/project"))

    def test_multiple_tools_writes_all_files(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)
        preview = handler.build_preview(
            model,
            tools=(SupportedTool.CLAUDE_CODE, SupportedTool.CURSOR),
        )

        handler.approve_and_write(preview, output_dir=Path("/project"))

        paths = list(writer.written.keys())
        # Claude Code: 1 file, Cursor: 2 files = 3 total
        assert len(paths) >= 3
        assert any(".claude/CLAUDE.md" in p for p in paths)
        assert any("AGENTS.md" in p for p in paths)
        assert any(".cursor/rules" in p for p in paths)


# ---------------------------------------------------------------------------
# 3. No generate() convenience method
# ---------------------------------------------------------------------------


class TestNoGenerateMethod:
    def test_no_generate_convenience_method(self):
        """Enforce preview-before-action: no shortcut generate() method."""
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )

        assert not hasattr(ConfigGenerationHandler, "generate")


# ---------------------------------------------------------------------------
# 4. Port signature
# ---------------------------------------------------------------------------


class TestPortSignature:
    def test_port_has_generate_accepting_model(self):
        import inspect

        from src.application.ports.config_generation_port import ConfigGenerationPort

        sig = inspect.signature(ConfigGenerationPort.generate)
        params = list(sig.parameters.keys())
        assert "model" in params

    def test_port_has_generate_accepting_tools(self):
        import inspect

        from src.application.ports.config_generation_port import ConfigGenerationPort

        sig = inspect.signature(ConfigGenerationPort.generate)
        params = list(sig.parameters.keys())
        assert "tools" in params

    def test_port_has_generate_accepting_output_dir(self):
        import inspect

        from src.application.ports.config_generation_port import ConfigGenerationPort

        sig = inspect.signature(ConfigGenerationPort.generate)
        params = list(sig.parameters.keys())
        assert "output_dir" in params


# ---------------------------------------------------------------------------
# 5. All four tools end-to-end
# ---------------------------------------------------------------------------


class TestAllToolsEndToEnd:
    def test_all_four_tools(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)

        all_tools = (
            SupportedTool.CLAUDE_CODE,
            SupportedTool.CURSOR,
            SupportedTool.ROO_CODE,
            SupportedTool.OPENCODE,
        )
        preview = handler.build_preview(model, tools=all_tools)

        assert len(preview.configs) == 4

        handler.approve_and_write(preview, output_dir=Path("/project"))

        paths = list(writer.written.keys())
        # Claude: 1, Cursor: 2 (AGENTS.md shared), Roo: 3 (AGENTS.md shared),
        # OpenCode: 3 (AGENTS.md shared) = 7 unique paths
        assert len(paths) >= 7

    def test_written_content_not_empty(self):
        from src.application.commands.config_generation_handler import (
            ConfigGenerationHandler,
        )

        model = _make_model_with_contexts(
            [("Orders", SubdomainClassification.CORE)]
        )
        writer = FakeFileWriter()
        handler = ConfigGenerationHandler(writer=writer)
        preview = handler.build_preview(
            model, tools=(SupportedTool.CLAUDE_CODE,)
        )
        handler.approve_and_write(preview, output_dir=Path("/project"))

        for content in writer.written.values():
            assert content  # non-empty
