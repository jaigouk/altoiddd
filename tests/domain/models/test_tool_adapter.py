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
from src.domain.models.tool_adapter import (
    ClaudeCodeAdapter,
    CursorAdapter,
    OpenCodeAdapter,
    RooCodeAdapter,
    ToolAdapterProtocol,
)

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
        sections = adapter.translate(model)
        assert len(sections) == 1
        assert sections[0].file_path == ".claude/CLAUDE.md"

    def test_content_includes_ubiquitous_language(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model)
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
        sections = adapter.translate(model)
        assert "Orders" in sections[0].content
        assert "Notifications" in sections[0].content
        assert "Bounded Contexts" in sections[0].content

    def test_content_includes_ddd_layer_rules(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model)
        assert "DDD Layer Rules" in sections[0].content


# ---------------------------------------------------------------------------
# 3. CursorAdapter
# ---------------------------------------------------------------------------


class TestCursorAdapter:
    def test_produces_agents_md(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = CursorAdapter()
        sections = adapter.translate(model)
        paths = [s.file_path for s in sections]
        assert "AGENTS.md" in paths

    def test_produces_cursor_rules(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = CursorAdapter()
        sections = adapter.translate(model)
        paths = [s.file_path for s in sections]
        assert ".cursor/rules/project-conventions.mdc" in paths

    def test_produces_two_sections(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = CursorAdapter()
        sections = adapter.translate(model)
        assert len(sections) == 2

    def test_mdc_has_frontmatter(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = CursorAdapter()
        sections = adapter.translate(model)
        mdc = next(s for s in sections if s.file_path.endswith(".mdc"))
        assert mdc.content.startswith("---")


# ---------------------------------------------------------------------------
# 4. RooCodeAdapter
# ---------------------------------------------------------------------------


class TestRooCodeAdapter:
    def test_produces_agents_md(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model)
        paths = [s.file_path for s in sections]
        assert "AGENTS.md" in paths

    def test_produces_roomodes(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model)
        paths = [s.file_path for s in sections]
        assert ".roomodes" in paths

    def test_produces_roo_rules(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model)
        paths = [s.file_path for s in sections]
        assert ".roo/rules/project-conventions.md" in paths

    def test_produces_three_sections(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model)
        assert len(sections) == 3

    def test_roomodes_is_valid_json(self):
        import json

        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model)
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
        sections = adapter.translate(model)
        paths = [s.file_path for s in sections]
        assert "AGENTS.md" in paths

    def test_produces_opencode_json(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = OpenCodeAdapter()
        sections = adapter.translate(model)
        paths = [s.file_path for s in sections]
        assert "opencode.json" in paths

    def test_produces_opencode_rules(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = OpenCodeAdapter()
        sections = adapter.translate(model)
        paths = [s.file_path for s in sections]
        assert ".opencode/rules/project-conventions.md" in paths

    def test_produces_three_sections(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = OpenCodeAdapter()
        sections = adapter.translate(model)
        assert len(sections) == 3

    def test_opencode_json_is_valid(self):
        import json

        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = OpenCodeAdapter()
        sections = adapter.translate(model)
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
        sections = adapter.translate(model)
        agents = next(s for s in sections if s.file_path == "AGENTS.md")
        assert "Orders" in agents.content

    def test_agents_md_includes_classification(self):
        model = _make_model_with_contexts([("Orders", SubdomainClassification.CORE)])
        adapter = RooCodeAdapter()
        sections = adapter.translate(model)
        agents = next(s for s in sections if s.file_path == "AGENTS.md")
        assert "core" in agents.content
