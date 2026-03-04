"""Tests for ClaudeCodeAdapter MEMORY.md generation.

Verifies that ClaudeCodeAdapter generates a MEMORY.md file encoding the DDD
agile work process for bootstrapped projects. The file must be under 200 lines.
Quality gates come from StackProfile — GenericProfile omits them.
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
from src.domain.models.tool_adapter import ClaudeCodeAdapter

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


def _make_model_multi_context() -> DomainModel:
    """Build a DomainModel with multiple bounded contexts."""
    model = DomainModel()
    model.add_domain_story(
        DomainStory(
            name="E-commerce flow",
            actors=("Customer",),
            trigger="Customer places order",
            steps=(
                "Customer manages Orders",
                "Customer manages Payments",
                "Customer manages Notifications",
            ),
        )
    )
    for name, cls, defn in [
        ("Orders", SubdomainClassification.CORE, "Order management"),
        ("Payments", SubdomainClassification.SUPPORTING, "Payment processing"),
        ("Notifications", SubdomainClassification.GENERIC, "Notification delivery"),
    ]:
        model.add_term(term=name, definition=defn, context_name=name)
        model.add_bounded_context(
            BoundedContext(name=name, responsibility=defn, classification=cls)
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
# 1. ClaudeCodeAdapter generates MEMORY.md
# ---------------------------------------------------------------------------


class TestClaudeCodeAdapterMemoryMd:
    """ClaudeCodeAdapter.translate produces a MEMORY.md ConfigSection."""

    def test_generates_memory_md(self) -> None:
        """translate() returns a ConfigSection for .claude/memory/MEMORY.md."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, PythonUvProfile())

        paths = [s.file_path for s in sections]
        assert ".claude/memory/MEMORY.md" in paths

    def test_still_generates_claude_md(self) -> None:
        """translate() still produces .claude/CLAUDE.md alongside MEMORY.md."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, PythonUvProfile())

        paths = [s.file_path for s in sections]
        assert ".claude/CLAUDE.md" in paths

    def test_produces_two_sections(self) -> None:
        """translate() returns exactly 2 ConfigSections (CLAUDE.md + MEMORY.md)."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, PythonUvProfile())

        assert len(sections) == 2


# ---------------------------------------------------------------------------
# 2. MEMORY.md content — required sections
# ---------------------------------------------------------------------------


class TestMemoryMdContent:
    """MEMORY.md must contain the DDD agile work process sections."""

    def _get_memory_content(
        self,
        model: DomainModel | None = None,
        profile: PythonUvProfile | GenericProfile | None = None,
    ) -> str:
        model = model or _make_model()
        profile = profile or PythonUvProfile()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, profile)
        memory = next(s for s in sections if s.file_path == ".claude/memory/MEMORY.md")
        return memory.content

    def test_has_after_close_protocol(self) -> None:
        """MEMORY.md includes the after-close protocol section."""
        content = self._get_memory_content()
        assert "After-Close Protocol" in content

    def test_has_ripple_review(self) -> None:
        """After-close protocol includes ripple review."""
        content = self._get_memory_content()
        assert "bin/bd-ripple" in content

    def test_has_grooming_checklist(self) -> None:
        """MEMORY.md includes the grooming checklist."""
        content = self._get_memory_content()
        assert "Grooming Checklist" in content

    def test_grooming_has_template_compliance(self) -> None:
        """Grooming checklist mentions template compliance."""
        content = self._get_memory_content()
        assert "Template compliance" in content or "template compliance" in content.lower()

    def test_grooming_has_prd_traceability(self) -> None:
        """Grooming checklist mentions PRD traceability."""
        content = self._get_memory_content()
        assert "PRD traceability" in content or "prd-traceability" in content

    def test_has_beads_workflow(self) -> None:
        """MEMORY.md includes beads workflow commands."""
        content = self._get_memory_content()
        assert "bd ready" in content
        assert "bd show" in content
        assert "bd close" in content

    def test_has_bounded_contexts(self) -> None:
        """MEMORY.md includes bounded contexts from the domain model."""
        content = self._get_memory_content()
        assert "Bounded Contexts" in content
        assert "Orders" in content

    def test_has_ubiquitous_language(self) -> None:
        """MEMORY.md includes ubiquitous language from the domain model."""
        content = self._get_memory_content()
        assert "Ubiquitous Language" in content
        assert "Orders" in content

    def test_multi_context_model_includes_all_contexts(self) -> None:
        """MEMORY.md with multiple bounded contexts lists them all."""
        model = _make_model_multi_context()
        content = self._get_memory_content(model=model)
        assert "Orders" in content
        assert "Payments" in content
        assert "Notifications" in content


# ---------------------------------------------------------------------------
# 3. GenericProfile — no Python quality gates
# ---------------------------------------------------------------------------


class TestMemoryMdGenericProfile:
    """GenericProfile MEMORY.md must not have Python-specific quality gates."""

    def _get_memory_content(self, profile: PythonUvProfile | GenericProfile) -> str:
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, profile)
        memory = next(s for s in sections if s.file_path == ".claude/memory/MEMORY.md")
        return memory.content

    def test_generic_profile_omits_quality_gates(self) -> None:
        """GenericProfile MEMORY.md has no Quality Gates section."""
        content = self._get_memory_content(GenericProfile())
        assert "Quality Gates" not in content

    def test_generic_profile_no_uv_run(self) -> None:
        """GenericProfile MEMORY.md has no uv run commands."""
        content = self._get_memory_content(GenericProfile())
        assert "uv run" not in content

    def test_generic_profile_no_pytest(self) -> None:
        """GenericProfile MEMORY.md has no pytest references."""
        content = self._get_memory_content(GenericProfile())
        assert "pytest" not in content


# ---------------------------------------------------------------------------
# 4. PythonUvProfile — includes Python quality gates
# ---------------------------------------------------------------------------


class TestMemoryMdPythonProfile:
    """PythonUvProfile MEMORY.md includes Python-specific quality gates."""

    def _get_memory_content(self) -> str:
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, PythonUvProfile())
        memory = next(s for s in sections if s.file_path == ".claude/memory/MEMORY.md")
        return memory.content

    def test_python_profile_includes_quality_gates(self) -> None:
        """PythonUvProfile MEMORY.md has Quality Gates section."""
        content = self._get_memory_content()
        assert "Quality Gates" in content

    def test_python_profile_has_ruff(self) -> None:
        """PythonUvProfile quality gates include ruff."""
        content = self._get_memory_content()
        assert "uv run ruff" in content

    def test_python_profile_has_mypy(self) -> None:
        """PythonUvProfile quality gates include mypy."""
        content = self._get_memory_content()
        assert "uv run mypy" in content

    def test_python_profile_has_pytest(self) -> None:
        """PythonUvProfile quality gates include pytest."""
        content = self._get_memory_content()
        assert "uv run pytest" in content


# ---------------------------------------------------------------------------
# 5. MEMORY.md must be under 200 lines
# ---------------------------------------------------------------------------


class TestMemoryMdLineLimit:
    """MEMORY.md must be under 200 lines (Claude Code's critical window)."""

    def test_under_200_lines_python_profile(self) -> None:
        """PythonUvProfile MEMORY.md is under 200 lines."""
        model = _make_model_multi_context()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, PythonUvProfile())
        memory = next(s for s in sections if s.file_path == ".claude/memory/MEMORY.md")
        line_count = len(memory.content.splitlines())
        assert line_count < 200, f"MEMORY.md has {line_count} lines, must be under 200"

    def test_under_200_lines_generic_profile(self) -> None:
        """GenericProfile MEMORY.md is under 200 lines."""
        model = _make_model_multi_context()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, GenericProfile())
        memory = next(s for s in sections if s.file_path == ".claude/memory/MEMORY.md")
        line_count = len(memory.content.splitlines())
        assert line_count < 200, f"MEMORY.md has {line_count} lines, must be under 200"
