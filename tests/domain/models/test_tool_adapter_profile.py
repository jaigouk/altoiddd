"""Tests for tool adapter profile integration.

Verifies that adapters accept a StackProfile parameter and use
quality_gate_display and file_glob from the profile instead of
hardcoded values. GenericProfile produces output with no quality
gates sections.
"""

from __future__ import annotations

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.stack_profile import GenericProfile, PythonUvProfile
from src.domain.models.tool_adapter import (
    ClaudeCodeAdapter,
    CursorAdapter,
    OpenCodeAdapter,
    RooCodeAdapter,
)

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_model() -> DomainModel:
    """Build a minimal valid DomainModel."""
    model = DomainModel()
    model.add_domain_story(
        DomainStory(
            name="Test flow",
            actors=("User",),
            trigger="User starts",
            steps=("User manages Orders",),
        )
    )
    model.add_term(term="Orders", definition="Order domain", context_name="Orders")
    model.add_bounded_context(
        BoundedContext(
            name="Orders",
            responsibility="Manages Orders",
            classification=SubdomainClassification.CORE,
        )
    )
    model.design_aggregate(
        AggregateDesign(
            name="OrderRoot",
            context_name="Orders",
            root_entity="OrderRoot",
            invariants=("must be valid",),
        )
    )
    model.finalize()
    return model


# ---------------------------------------------------------------------------
# ClaudeCodeAdapter
# ---------------------------------------------------------------------------


class TestClaudeCodeAdapterProfile:
    def test_python_profile_includes_quality_gates(self) -> None:
        """ClaudeCodeAdapter with PythonUvProfile includes quality gates."""
        model = _make_model()
        profile = PythonUvProfile()
        adapter = ClaudeCodeAdapter()

        sections = adapter.translate(model, profile)

        content = sections[0].content
        assert "Quality Gates" in content
        assert "uv run ruff" in content

    def test_generic_profile_omits_quality_gates(self) -> None:
        """ClaudeCodeAdapter with GenericProfile has no quality gates section."""
        model = _make_model()
        profile = GenericProfile()
        adapter = ClaudeCodeAdapter()

        sections = adapter.translate(model, profile)

        content = sections[0].content
        assert "Quality Gates" not in content
        assert "uv run" not in content

    def test_python_profile_output_matches_current(self) -> None:
        """PythonUvProfile produces identical output to current hardcoded."""
        model = _make_model()
        profile = PythonUvProfile()
        adapter = ClaudeCodeAdapter()

        sections = adapter.translate(model, profile)

        content = sections[0].content
        # Verify code block format (not bullet format)
        assert "```bash" in content
        assert "uv run ruff check ." in content
        assert "uv run mypy ." in content
        assert "uv run pytest" in content


# ---------------------------------------------------------------------------
# CursorAdapter
# ---------------------------------------------------------------------------


class TestCursorAdapterProfile:
    def test_uses_profile_file_glob(self) -> None:
        """CursorAdapter uses profile.file_glob instead of hardcoded **/*.py."""
        model = _make_model()
        profile = PythonUvProfile()
        adapter = CursorAdapter()

        sections = adapter.translate(model, profile)

        mdc = next(s for s in sections if s.file_path.endswith(".mdc"))
        assert "globs: **/*.py" in mdc.content

    def test_generic_profile_uses_star_glob(self) -> None:
        """CursorAdapter with GenericProfile uses * glob."""
        model = _make_model()
        profile = GenericProfile()
        adapter = CursorAdapter()

        sections = adapter.translate(model, profile)

        mdc = next(s for s in sections if s.file_path.endswith(".mdc"))
        assert "globs: *" in mdc.content
        assert "globs: **/*.py" not in mdc.content

    def test_generic_profile_omits_quality_gates_in_mdc(self) -> None:
        """CursorAdapter mdc file has no quality gates for GenericProfile."""
        model = _make_model()
        profile = GenericProfile()
        adapter = CursorAdapter()

        sections = adapter.translate(model, profile)

        mdc = next(s for s in sections if s.file_path.endswith(".mdc"))
        assert "Quality Gates" not in mdc.content

    def test_generic_profile_omits_quality_gates_in_agents(self) -> None:
        """CursorAdapter AGENTS.md has no quality gates for GenericProfile."""
        model = _make_model()
        profile = GenericProfile()
        adapter = CursorAdapter()

        sections = adapter.translate(model, profile)

        agents = next(s for s in sections if s.file_path == "AGENTS.md")
        assert "Quality Gates" not in agents.content


# ---------------------------------------------------------------------------
# RooCodeAdapter
# ---------------------------------------------------------------------------


class TestRooCodeAdapterProfile:
    def test_generic_profile_omits_quality_gates_in_rules(self) -> None:
        """RooCodeAdapter rules file has no quality gates for GenericProfile."""
        model = _make_model()
        profile = GenericProfile()
        adapter = RooCodeAdapter()

        sections = adapter.translate(model, profile)

        rules = next(
            s for s in sections if s.file_path == ".roo/rules/project-conventions.md"
        )
        assert "Quality Gates" not in rules.content


# ---------------------------------------------------------------------------
# OpenCodeAdapter
# ---------------------------------------------------------------------------


class TestOpenCodeAdapterProfile:
    def test_generic_profile_omits_quality_gates_in_rules(self) -> None:
        """OpenCodeAdapter rules file has no quality gates for GenericProfile."""
        model = _make_model()
        profile = GenericProfile()
        adapter = OpenCodeAdapter()

        sections = adapter.translate(model, profile)

        rules = next(
            s for s in sections
            if s.file_path == ".opencode/rules/project-conventions.md"
        )
        assert "Quality Gates" not in rules.content


# ---------------------------------------------------------------------------
# ToolAdapterProtocol
# ---------------------------------------------------------------------------


class TestProtocolSignature:
    def test_translate_accepts_profile(self) -> None:
        """All adapters accept (model, profile) in translate()."""
        model = _make_model()
        profile = PythonUvProfile()

        adapters: list[ClaudeCodeAdapter | CursorAdapter | RooCodeAdapter | OpenCodeAdapter] = [
            ClaudeCodeAdapter(),
            CursorAdapter(),
            RooCodeAdapter(),
            OpenCodeAdapter(),
        ]
        for adapter in adapters:
            sections = adapter.translate(model, profile)
            assert len(sections) >= 1
