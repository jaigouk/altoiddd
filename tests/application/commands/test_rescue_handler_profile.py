"""Tests for RescueHandler profile integration.

Verifies that the handler reads structure and config targets from
StackProfile instead of hardcoded constants, and that GenericProfile
(empty source_layout / project_manifest) gracefully skips checks.
"""

from __future__ import annotations

from pathlib import Path

from src.domain.models.gap_analysis import (
    GapType,
    ProjectScan,
)
from src.domain.models.stack_profile import GenericProfile, PythonUvProfile

# -- Fake adapters -----------------------------------------------------------


class FakeScanner:
    """In-memory test double implementing ProjectScanPort."""

    def __init__(self, scan: ProjectScan | None = None) -> None:
        self._scan = scan

    def scan(self, project_dir: Path, profile: object = None) -> ProjectScan:
        if self._scan is not None:
            return self._scan
        return ProjectScan(
            project_dir=project_dir,
            existing_docs=(),
            existing_configs=(),
            existing_structure=(),
            has_knowledge_dir=False,
            has_agents_md=False,
            has_git=True,
            has_tests=False,
        )


class FakeGitOps:
    """In-memory test double implementing GitOpsPort."""

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


# ---------------------------------------------------------------------------
# 1. RescueHandler uses profile for structure checks
# ---------------------------------------------------------------------------


class TestRescueUsesProfileStructure:
    def test_rescue_uses_profile_source_layout(self) -> None:
        """RescueHandler checks profile.source_layout, not hardcoded dirs."""
        from src.application.commands.rescue_handler import RescueHandler

        profile = PythonUvProfile()
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        structure_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_STRUCTURE]
        structure_paths = [g.path for g in structure_gaps]
        # Python profile has src/domain/, src/application/, src/infrastructure/
        assert "src/domain/" in structure_paths
        assert "src/application/" in structure_paths
        assert "src/infrastructure/" in structure_paths

    def test_rescue_generic_profile_skips_structure_check(self) -> None:
        """GenericProfile (empty source_layout) → no structure gaps reported."""
        from src.application.commands.rescue_handler import RescueHandler

        profile = GenericProfile()
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        structure_gaps = [
            g
            for g in analysis.gaps
            if g.gap_type == GapType.MISSING_STRUCTURE and g.path != "tests/"
        ]
        assert structure_gaps == []


# ---------------------------------------------------------------------------
# 2. RescueHandler uses profile for config (project manifest) checks
# ---------------------------------------------------------------------------


class TestRescueUsesProfileManifest:
    def test_rescue_uses_profile_project_manifest(self) -> None:
        """RescueHandler checks profile.project_manifest in config targets."""
        from src.application.commands.rescue_handler import RescueHandler

        profile = PythonUvProfile()
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        config_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_CONFIG]
        config_paths = [g.path for g in config_gaps]
        # Python profile has pyproject.toml as manifest
        assert "pyproject.toml" in config_paths

    def test_rescue_generic_profile_skips_manifest_check(self) -> None:
        """GenericProfile (empty manifest) → no manifest gap reported."""
        from src.application.commands.rescue_handler import RescueHandler

        profile = GenericProfile()
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        config_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_CONFIG]
        config_paths = [g.path for g in config_gaps]
        # alty-universal configs (.claude/CLAUDE.md) still checked
        assert ".claude/CLAUDE.md" in config_paths
        # But no pyproject.toml gap for generic
        assert "pyproject.toml" not in config_paths


# ---------------------------------------------------------------------------
# 3. Alty-universal items stay hardcoded
# ---------------------------------------------------------------------------


class TestAltyUniversalItemsUnchanged:
    def test_docs_always_checked(self) -> None:
        """docs/PRD.md, docs/DDD.md, docs/ARCHITECTURE.md always checked."""
        from src.application.commands.rescue_handler import RescueHandler

        profile = GenericProfile()
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        doc_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_DOC]
        doc_paths = [g.path for g in doc_gaps]
        assert "docs/PRD.md" in doc_paths
        assert "docs/DDD.md" in doc_paths
        assert "docs/ARCHITECTURE.md" in doc_paths

    def test_claude_md_always_checked(self) -> None:
        """.claude/CLAUDE.md always checked regardless of profile."""
        from src.application.commands.rescue_handler import RescueHandler

        profile = GenericProfile()
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        config_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_CONFIG]
        config_paths = [g.path for g in config_gaps]
        assert ".claude/CLAUDE.md" in config_paths


# ---------------------------------------------------------------------------
# 4. Full project with profile → no gaps
# ---------------------------------------------------------------------------


class TestFullProjectWithProfile:
    def test_python_profile_all_present_no_gaps(self) -> None:
        """All artifacts present for Python profile → no gaps."""
        from src.application.commands.rescue_handler import RescueHandler

        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            existing_docs=("docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"),
            existing_configs=(".claude/CLAUDE.md", "pyproject.toml"),
            existing_structure=("src/domain/", "src/application/", "src/infrastructure/"),
            has_knowledge_dir=True,
            has_agents_md=True,
            has_git=True,
            has_tests=True,
        )
        profile = PythonUvProfile()
        handler = RescueHandler(project_scan=FakeScanner(scan=scan), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        assert analysis.gaps == ()

    def test_generic_profile_minimal_project_no_gaps(self) -> None:
        """Generic profile only needs docs + alty configs, no structure/manifest."""
        from src.application.commands.rescue_handler import RescueHandler

        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            existing_docs=("docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"),
            existing_configs=(".claude/CLAUDE.md",),
            existing_structure=(),
            has_knowledge_dir=True,
            has_agents_md=True,
            has_git=True,
            has_tests=True,
        )
        profile = GenericProfile()
        handler = RescueHandler(project_scan=FakeScanner(scan=scan), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        assert analysis.gaps == ()


# ---------------------------------------------------------------------------
# 5. Backward compatibility (no profile = default behavior)
# ---------------------------------------------------------------------------


class TestBackwardCompat:
    def test_no_profile_still_works(self) -> None:
        """Calling rescue() without profile still returns gap analysis."""
        from src.application.commands.rescue_handler import RescueHandler

        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"))

        assert len(analysis.gaps) > 0

    def test_no_profile_checks_python_structure(self) -> None:
        """Without profile, default behavior checks Python structure."""
        from src.application.commands.rescue_handler import RescueHandler

        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"))

        structure_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_STRUCTURE]
        structure_paths = [g.path for g in structure_gaps]
        assert "src/domain/" in structure_paths
