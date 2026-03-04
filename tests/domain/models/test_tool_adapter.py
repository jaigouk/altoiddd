"""Tests for tool adapters (ClaudeCodeAdapter, CursorAdapter, RooCodeAdapter, OpenCodeAdapter).

Verifies each adapter translates a DomainModel into correct file paths and content.
"""

from __future__ import annotations

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.stack_profile import PythonUvProfile
from src.domain.models.tool_adapter import (
    ClaudeCodeAdapter,
    CursorAdapter,
    OpenCodeAdapter,
    RooCodeAdapter,
    ToolAdapterProtocol,
)

_PROFILE = PythonUvProfile()

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


# ---------------------------------------------------------------------------
# 1. ToolAdapterProtocol conformance
# ---------------------------------------------------------------------------


class TestProtocolConformance:
    def test_claude_code_adapter_is_protocol_instance(self):
        assert isinstance(ClaudeCodeAdapter(), ToolAdapterProtocol)

    def test_cursor_adapter_is_protocol_instance(self):
        assert isinstance(CursorAdapter(), ToolAdapterProtocol)

    def test_roo_code_adapter_is_protocol_instance(self):
        assert isinstance(RooCodeAdapter(), ToolAdapterProtocol)

    def test_opencode_adapter_is_protocol_instance(self):
        assert isinstance(OpenCodeAdapter(), ToolAdapterProtocol)


# ---------------------------------------------------------------------------
# 2. ClaudeCodeAdapter
# ---------------------------------------------------------------------------


class TestClaudeCodeAdapter:
    def test_produces_claude_md(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert len(sections) == 2
        assert sections[0].file_path == ".claude/CLAUDE.md"
        assert sections[1].file_path == ".claude/memory/MEMORY.md"

    def test_content_includes_ubiquitous_language(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert "Orders" in sections[0].content
        assert "Ubiquitous Language" in sections[0].content

    def test_content_includes_bounded_contexts(self):
        model = _make_model_with_contexts(
            [
                ("Orders", SubdomainClassification.CORE),
                ("Notifications", SubdomainClassification.SUPPORTING),
            ]
        )
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert "Orders" in sections[0].content
        assert "Notifications" in sections[0].content
        assert "Bounded Contexts" in sections[0].content

    def test_content_includes_ddd_layer_rules(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert "DDD Layer Rules" in sections[0].content

    def test_claude_md_includes_after_close_protocol(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert "After-Close Protocol" in sections[0].content

    def test_claude_md_after_close_has_ripple_step(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert "bd-ripple" in sections[0].content

    def test_claude_md_after_close_has_review_step(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert "review_needed" in sections[0].content

    def test_claude_md_after_close_has_followup_step(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        content = sections[0].content
        assert "Follow-up" in content or "follow-up" in content

    def test_claude_md_after_close_has_groom_step(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert "bd ready" in sections[0].content


# ---------------------------------------------------------------------------
# 3. CursorAdapter
# ---------------------------------------------------------------------------


class TestCursorAdapter:
    def test_produces_agents_md(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = CursorAdapter()
        sections = adapter.translate(model, _PROFILE)
        paths = [s.file_path for s in sections]
        assert "AGENTS.md" in paths

    def test_produces_cursor_rules(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = CursorAdapter()
        sections = adapter.translate(model, _PROFILE)
        paths = [s.file_path for s in sections]
        assert ".cursor/rules/project-conventions.mdc" in paths

    def test_produces_two_sections(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = CursorAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert len(sections) == 2

    def test_mdc_has_frontmatter(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = CursorAdapter()
        sections = adapter.translate(model, _PROFILE)
        mdc = next(s for s in sections if s.file_path.endswith(".mdc"))
        assert mdc.content.startswith("---")


# ---------------------------------------------------------------------------
# 4. RooCodeAdapter
# ---------------------------------------------------------------------------


class TestRooCodeAdapter:
    def test_produces_agents_md(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        paths = [s.file_path for s in sections]
        assert "AGENTS.md" in paths

    def test_produces_roomodes(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        paths = [s.file_path for s in sections]
        assert ".roomodes" in paths

    def test_produces_roo_rules(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        paths = [s.file_path for s in sections]
        assert ".roo/rules/project-conventions.md" in paths

    def test_produces_three_sections(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert len(sections) == 3

    def test_roomodes_is_valid_json(self):
        import json

        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        roomodes = next(s for s in sections if s.file_path == ".roomodes")
        parsed = json.loads(roomodes.content)
        assert "customModes" in parsed


# ---------------------------------------------------------------------------
# 5. OpenCodeAdapter
# ---------------------------------------------------------------------------


class TestOpenCodeAdapter:
    def test_produces_agents_md(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = OpenCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        paths = [s.file_path for s in sections]
        assert "AGENTS.md" in paths

    def test_produces_opencode_json(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = OpenCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        paths = [s.file_path for s in sections]
        assert "opencode.json" in paths

    def test_produces_opencode_rules(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = OpenCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        paths = [s.file_path for s in sections]
        assert ".opencode/rules/project-conventions.md" in paths

    def test_produces_three_sections(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = OpenCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        assert len(sections) == 3

    def test_opencode_json_is_valid(self):
        import json

        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = OpenCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        oc_json = next(s for s in sections if s.file_path == "opencode.json")
        parsed = json.loads(oc_json.content)
        assert "context" in parsed


# ---------------------------------------------------------------------------
# 6. Content includes domain model data
# ---------------------------------------------------------------------------


class TestAdapterContent:
    def test_agents_md_includes_terms(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = CursorAdapter()
        sections = adapter.translate(model, _PROFILE)
        agents = next(s for s in sections if s.file_path == "AGENTS.md")
        assert "Orders" in agents.content

    def test_agents_md_includes_classification(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        agents = next(s for s in sections if s.file_path == "AGENTS.md")
        assert "core" in agents.content

    def test_agents_md_includes_after_close_protocol(self):
        """All adapters using _build_agents_md include the after-close protocol."""
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        for adapter_cls in (CursorAdapter, RooCodeAdapter, OpenCodeAdapter):
            adapter = adapter_cls()
            sections = adapter.translate(model, _PROFILE)
            agents = next(s for s in sections if s.file_path == "AGENTS.md")
            assert "After-Close Protocol" in agents.content, (
                f"{adapter_cls.__name__} AGENTS.md missing After-Close Protocol"
            )

    def test_claude_md_and_memory_md_protocol_consistent(self):
        """After-close protocol in CLAUDE.md matches MEMORY.md protocol steps."""
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, _PROFILE)
        claude_md = sections[0].content
        memory_md = sections[1].content
        # Both must contain the same 4 protocol steps
        for keyword in ("bd-ripple", "review_needed", "Follow-up", "bd ready"):
            assert keyword in claude_md, f"CLAUDE.md missing {keyword}"
            assert keyword in memory_md, f"MEMORY.md missing {keyword}"
