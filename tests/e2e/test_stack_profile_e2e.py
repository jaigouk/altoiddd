"""End-to-end validation tests for GenericProfile vs PythonUvProfile output.

These tests validate the full pipeline: given a StackProfile, the generated
ticket details, CLAUDE.md, MEMORY.md, and rescue gap analysis must be
correctly stack-aware. GenericProfile output must contain ZERO Python-specific
terms. PythonUvProfile output must contain Python toolchain commands.

Ticket: jb0.3
"""

from __future__ import annotations

from pathlib import Path

from src.domain.models.domain_model import DomainModel
from src.domain.models.domain_values import (
    AggregateDesign,
    BoundedContext,
    DomainStory,
    SubdomainClassification,
)
from src.domain.models.gap_analysis import GapType, ProjectScan
from src.domain.models.stack_profile import GenericProfile, PythonUvProfile
from src.domain.models.ticket_values import TicketDetailLevel
from src.domain.models.tool_adapter import ClaudeCodeAdapter
from src.domain.services.ticket_detail_renderer import TicketDetailRenderer

# ---------------------------------------------------------------------------
# Shared fixtures
# ---------------------------------------------------------------------------

# Python-specific terms that MUST NOT appear in GenericProfile output
_PYTHON_TERMS: tuple[str, ...] = (
    "pytest",
    "uv run",
    "mypy",
    "ruff",
    "pyproject.toml",
    "src/domain/",
    "src/application/",
    "src/infrastructure/",
)

# DDD principles that ARE universal and SHOULD appear regardless of profile
_DDD_TERMS: tuple[str, ...] = (
    "domain",
    "Bounded Context",
    "DDD",
)


def _make_model() -> DomainModel:
    """Build a finalized DomainModel for end-to-end testing."""
    model = DomainModel()
    model.add_domain_story(
        DomainStory(
            name="Order Management",
            actors=("Customer",),
            trigger="Customer places order",
            steps=(
                "Customer manages Orders",
                "Customer manages Payments",
            ),
        )
    )
    for name, cls, defn in [
        ("Orders", SubdomainClassification.CORE, "Order lifecycle management"),
        ("Payments", SubdomainClassification.SUPPORTING, "Payment processing"),
    ]:
        model.add_term(term=name, definition=defn, context_name=name)
        model.add_bounded_context(
            BoundedContext(name=name, responsibility=defn, classification=cls)
        )
    model.design_aggregate(
        AggregateDesign(
            name="OrderAggregate",
            context_name="Orders",
            root_entity="Order",
            contained_objects=("OrderLine", "ShippingAddress"),
            invariants=("total must be positive", "at least one line item"),
            commands=("PlaceOrder", "CancelOrder"),
            domain_events=("OrderPlaced", "OrderCancelled"),
        )
    )
    model.finalize()
    return model


def _make_aggregate() -> AggregateDesign:
    """Build a rich aggregate for ticket rendering."""
    return AggregateDesign(
        name="OrderAggregate",
        context_name="Orders",
        root_entity="Order",
        contained_objects=("OrderLine", "ShippingAddress"),
        invariants=("total must be positive", "at least one line item"),
        commands=("PlaceOrder", "CancelOrder"),
        domain_events=("OrderPlaced", "OrderCancelled"),
    )


# ===========================================================================
# 1. GenericProfile ticket generation → no Python terms
# ===========================================================================


class TestGenericProfileTicketGeneration:
    """GenericProfile ticket output must contain ZERO Python-specific terms."""

    def test_full_detail_no_python_terms(self) -> None:
        """FULL detail ticket with GenericProfile has no Python terms."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, GenericProfile())

        for term in _PYTHON_TERMS:
            assert term not in output, (
                f"GenericProfile FULL ticket contains Python term '{term}'"
            )

    def test_standard_detail_no_python_terms(self) -> None:
        """STANDARD detail ticket with GenericProfile has no Python terms."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.STANDARD, GenericProfile())

        for term in _PYTHON_TERMS:
            assert term not in output, (
                f"GenericProfile STANDARD ticket contains Python term '{term}'"
            )

    def test_full_detail_has_ddd_principles(self) -> None:
        """GenericProfile FULL ticket still contains universal DDD principles."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, GenericProfile())

        # DDD-specific sections must be present
        assert "## DDD Alignment" in output
        assert "Bounded Context" in output
        assert "## SOLID Mapping" in output

    def test_full_detail_has_tdd_workflow(self) -> None:
        """GenericProfile FULL ticket still has TDD Workflow with placeholder."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, GenericProfile())

        assert "## TDD Workflow" in output
        assert "RED" in output
        assert "GREEN" in output
        assert "REFACTOR" in output
        assert "<test-runner>" in output

    def test_full_detail_no_quality_gates_section(self) -> None:
        """GenericProfile FULL ticket omits Quality Gates section."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, GenericProfile())

        assert "## Quality Gates" not in output


# ===========================================================================
# 2. PythonUvProfile ticket generation → Python commands present
# ===========================================================================


class TestPythonUvProfileTicketGeneration:
    """PythonUvProfile ticket output must contain Python toolchain commands."""

    def test_full_detail_has_pytest(self) -> None:
        """FULL detail ticket with PythonUvProfile includes pytest."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, PythonUvProfile())

        assert "pytest" in output

    def test_full_detail_has_uv_run(self) -> None:
        """FULL detail ticket with PythonUvProfile includes uv run."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, PythonUvProfile())

        assert "uv run" in output

    def test_full_detail_has_quality_gates(self) -> None:
        """FULL detail ticket with PythonUvProfile has quality gates."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, PythonUvProfile())

        assert "## Quality Gates" in output
        assert "uv run ruff" in output
        assert "uv run mypy" in output
        assert "uv run pytest" in output

    def test_standard_detail_has_quality_gates(self) -> None:
        """STANDARD detail ticket with PythonUvProfile has quality gates."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.STANDARD, PythonUvProfile())

        assert "## Quality Gates" in output

    def test_full_detail_has_ddd_sections(self) -> None:
        """PythonUvProfile FULL ticket also has DDD/SOLID sections."""
        agg = _make_aggregate()
        output = TicketDetailRenderer.render(agg, TicketDetailLevel.FULL, PythonUvProfile())

        assert "## DDD Alignment" in output
        assert "## SOLID Mapping" in output
        assert "## TDD Workflow" in output


# ===========================================================================
# 3. GenericProfile CLAUDE.md generation → no Python quality gates
# ===========================================================================


class TestGenericProfileClaudeMdGeneration:
    """GenericProfile CLAUDE.md must not contain Python-specific quality gates."""

    def test_no_uv_run_pytest(self) -> None:
        """GenericProfile CLAUDE.md has no 'uv run pytest'."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, GenericProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "uv run pytest" not in claude_md.content

    def test_no_uv_run_ruff(self) -> None:
        """GenericProfile CLAUDE.md has no 'uv run ruff'."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, GenericProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "uv run ruff" not in claude_md.content

    def test_no_uv_run_mypy(self) -> None:
        """GenericProfile CLAUDE.md has no 'uv run mypy'."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, GenericProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "uv run mypy" not in claude_md.content

    def test_no_quality_gates_section(self) -> None:
        """GenericProfile CLAUDE.md has no Quality Gates section."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, GenericProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "Quality Gates" not in claude_md.content

    def test_has_ddd_layer_rules(self) -> None:
        """GenericProfile CLAUDE.md still has universal DDD layer rules."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, GenericProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "DDD Layer Rules" in claude_md.content

    def test_has_ubiquitous_language(self) -> None:
        """GenericProfile CLAUDE.md still has ubiquitous language glossary."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, GenericProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "Ubiquitous Language" in claude_md.content
        assert "Orders" in claude_md.content

    def test_has_bounded_contexts(self) -> None:
        """GenericProfile CLAUDE.md still has bounded contexts."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, GenericProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "Bounded Contexts" in claude_md.content


# ===========================================================================
# 4. PythonUvProfile CLAUDE.md generation → Python quality gates present
# ===========================================================================


class TestPythonUvProfileClaudeMdGeneration:
    """PythonUvProfile CLAUDE.md must contain Python quality gates."""

    def test_has_uv_run_pytest(self) -> None:
        """PythonUvProfile CLAUDE.md has 'uv run pytest'."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, PythonUvProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "uv run pytest" in claude_md.content

    def test_has_uv_run_ruff(self) -> None:
        """PythonUvProfile CLAUDE.md has 'uv run ruff'."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, PythonUvProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "uv run ruff" in claude_md.content

    def test_has_uv_run_mypy(self) -> None:
        """PythonUvProfile CLAUDE.md has 'uv run mypy'."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, PythonUvProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "uv run mypy" in claude_md.content

    def test_has_quality_gates_section(self) -> None:
        """PythonUvProfile CLAUDE.md has Quality Gates section."""
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, PythonUvProfile())
        claude_md = next(s for s in sections if s.file_path == ".claude/CLAUDE.md")

        assert "Quality Gates" in claude_md.content


# ===========================================================================
# 5. MEMORY.md generation validation
# ===========================================================================


class TestMemoryMdGenerationValidation:
    """Validate MEMORY.md content for both profiles."""

    def _get_memory_content(self, profile: GenericProfile | PythonUvProfile) -> str:
        model = _make_model()
        adapter = ClaudeCodeAdapter()
        sections = adapter.translate(model, profile)
        memory = next(s for s in sections if s.file_path == ".claude/memory/MEMORY.md")
        return memory.content

    # -- GenericProfile --

    def test_generic_no_python_quality_gates(self) -> None:
        """GenericProfile MEMORY.md has no Python quality gates."""
        content = self._get_memory_content(GenericProfile())
        assert "uv run pytest" not in content
        assert "uv run ruff" not in content
        assert "uv run mypy" not in content

    def test_generic_no_quality_gates_section(self) -> None:
        """GenericProfile MEMORY.md has no Quality Gates section."""
        content = self._get_memory_content(GenericProfile())
        assert "Quality Gates" not in content

    # -- PythonUvProfile --

    def test_python_has_quality_gates(self) -> None:
        """PythonUvProfile MEMORY.md has Quality Gates with Python commands."""
        content = self._get_memory_content(PythonUvProfile())
        assert "Quality Gates" in content
        assert "uv run pytest" in content
        assert "uv run ruff" in content
        assert "uv run mypy" in content

    # -- Both profiles: under 200 lines --

    def test_generic_under_200_lines(self) -> None:
        """GenericProfile MEMORY.md is under 200 lines."""
        content = self._get_memory_content(GenericProfile())
        line_count = len(content.splitlines())
        assert line_count < 200, f"MEMORY.md has {line_count} lines, must be under 200"

    def test_python_under_200_lines(self) -> None:
        """PythonUvProfile MEMORY.md is under 200 lines."""
        content = self._get_memory_content(PythonUvProfile())
        line_count = len(content.splitlines())
        assert line_count < 200, f"MEMORY.md has {line_count} lines, must be under 200"

    # -- Both profiles: required process sections --

    def test_generic_has_after_close_protocol(self) -> None:
        """GenericProfile MEMORY.md has After-Close Protocol."""
        content = self._get_memory_content(GenericProfile())
        assert "After-Close Protocol" in content

    def test_python_has_after_close_protocol(self) -> None:
        """PythonUvProfile MEMORY.md has After-Close Protocol."""
        content = self._get_memory_content(PythonUvProfile())
        assert "After-Close Protocol" in content

    def test_generic_has_grooming_checklist(self) -> None:
        """GenericProfile MEMORY.md has Grooming Checklist."""
        content = self._get_memory_content(GenericProfile())
        assert "Grooming Checklist" in content

    def test_python_has_grooming_checklist(self) -> None:
        """PythonUvProfile MEMORY.md has Grooming Checklist."""
        content = self._get_memory_content(PythonUvProfile())
        assert "Grooming Checklist" in content

    def test_generic_has_beads_workflow(self) -> None:
        """GenericProfile MEMORY.md has beads workflow."""
        content = self._get_memory_content(GenericProfile())
        assert "bd ready" in content
        assert "bd close" in content

    def test_python_has_beads_workflow(self) -> None:
        """PythonUvProfile MEMORY.md has beads workflow."""
        content = self._get_memory_content(PythonUvProfile())
        assert "bd ready" in content
        assert "bd close" in content

    def test_generic_has_bounded_contexts(self) -> None:
        """GenericProfile MEMORY.md has bounded contexts from domain model."""
        content = self._get_memory_content(GenericProfile())
        assert "Bounded Contexts" in content
        assert "Orders" in content

    def test_python_has_bounded_contexts(self) -> None:
        """PythonUvProfile MEMORY.md has bounded contexts from domain model."""
        content = self._get_memory_content(PythonUvProfile())
        assert "Bounded Contexts" in content
        assert "Orders" in content


# ===========================================================================
# 6. Rescue handler with GenericProfile → no structure gaps
# ===========================================================================


class TestRescueHandlerGenericProfile:
    """GenericProfile rescue produces no structure or tests gaps."""

    def test_no_structure_gaps_for_generic_profile(self) -> None:
        """GenericProfile rescue reports no MISSING_STRUCTURE gaps."""
        from src.application.commands.rescue_handler import RescueHandler

        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            existing_docs=("docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"),
            existing_configs=(".claude/CLAUDE.md",),
            existing_structure=(),
            has_knowledge_dir=True,
            has_agents_md=True,
            has_git=True,
        )

        class FakeScanner:
            def scan(self, project_dir: Path, profile: object = None) -> ProjectScan:
                return scan

        class FakeGitOps:
            def __init__(self) -> None:
                self.created_branches: list[str] = []

            def has_git(self, project_dir: Path) -> bool:
                return True

            def is_clean(self, project_dir: Path) -> bool:
                return True

            def branch_exists(self, project_dir: Path, branch_name: str) -> bool:
                return False

            def create_branch(self, project_dir: Path, branch_name: str) -> None:
                self.created_branches.append(branch_name)

        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=GenericProfile())

        structure_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_STRUCTURE]
        assert structure_gaps == [], (
            f"GenericProfile should have no structure gaps: {structure_gaps}"
        )

    def test_no_manifest_gap_for_generic_profile(self) -> None:
        """GenericProfile rescue reports no pyproject.toml gap."""
        from src.application.commands.rescue_handler import RescueHandler

        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            existing_docs=("docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"),
            existing_configs=(".claude/CLAUDE.md",),
            existing_structure=(),
            has_knowledge_dir=True,
            has_agents_md=True,
            has_git=True,
        )

        class FakeScanner:
            def scan(self, project_dir: Path, profile: object = None) -> ProjectScan:
                return scan

        class FakeGitOps:
            def __init__(self) -> None:
                self.created_branches: list[str] = []

            def has_git(self, project_dir: Path) -> bool:
                return True

            def is_clean(self, project_dir: Path) -> bool:
                return True

            def branch_exists(self, project_dir: Path, branch_name: str) -> bool:
                return False

            def create_branch(self, project_dir: Path, branch_name: str) -> None:
                self.created_branches.append(branch_name)

        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=GenericProfile())

        config_paths = [g.path for g in analysis.gaps if g.gap_type == GapType.MISSING_CONFIG]
        assert "pyproject.toml" not in config_paths

    def test_no_tests_dir_gap_for_generic_profile(self) -> None:
        """GenericProfile rescue does not report a tests/ gap."""
        from src.application.commands.rescue_handler import RescueHandler

        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            existing_docs=(),
            existing_configs=(),
            existing_structure=(),
            has_knowledge_dir=False,
            has_agents_md=False,
            has_git=True,
        )

        class FakeScanner:
            def scan(self, project_dir: Path, profile: object = None) -> ProjectScan:
                return scan

        class FakeGitOps:
            def __init__(self) -> None:
                self.created_branches: list[str] = []

            def has_git(self, project_dir: Path) -> bool:
                return True

            def is_clean(self, project_dir: Path) -> bool:
                return True

            def branch_exists(self, project_dir: Path, branch_name: str) -> bool:
                return False

            def create_branch(self, project_dir: Path, branch_name: str) -> None:
                self.created_branches.append(branch_name)

        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=GenericProfile())

        all_paths = [g.path for g in analysis.gaps]
        assert "tests/" not in all_paths, "GenericProfile must not report tests/ gap"


# ===========================================================================
# 7. Beads templates — no Python-specific terms
# ===========================================================================


class TestBeadsTemplatesStackAgnostic:
    """Beads templates must use generic placeholders, not Python commands."""

    def _read_template(self, name: str) -> str:
        template_path = Path(__file__).parent.parent.parent / "docs" / "beads_templates" / name
        return template_path.read_text()

    def test_ticket_template_no_pytest(self) -> None:
        """beads-ticket-template.md has no 'pytest' reference."""
        content = self._read_template("beads-ticket-template.md")
        assert "pytest" not in content

    def test_ticket_template_no_uv_run(self) -> None:
        """beads-ticket-template.md has no 'uv run' reference."""
        content = self._read_template("beads-ticket-template.md")
        assert "uv run" not in content

    def test_ticket_template_no_src_domain(self) -> None:
        """beads-ticket-template.md has no 'src/domain/' reference."""
        content = self._read_template("beads-ticket-template.md")
        assert "src/domain/" not in content

    def test_ticket_template_has_generic_placeholders(self) -> None:
        """beads-ticket-template.md uses generic placeholders."""
        content = self._read_template("beads-ticket-template.md")
        assert "<test-runner>" in content
        assert "<lint-command>" in content
        assert "<type-check-command>" in content

    def test_epic_template_no_pytest(self) -> None:
        """beads-epic-template.md has no 'pytest' reference."""
        content = self._read_template("beads-epic-template.md")
        assert "pytest" not in content

    def test_epic_template_no_uv_run(self) -> None:
        """beads-epic-template.md has no 'uv run' reference."""
        content = self._read_template("beads-epic-template.md")
        assert "uv run" not in content

    def test_epic_template_has_generic_placeholders(self) -> None:
        """beads-epic-template.md uses generic placeholders."""
        content = self._read_template("beads-epic-template.md")
        assert "<lint-command>" in content
        assert "<type-check-command>" in content
        assert "<test-runner>" in content
