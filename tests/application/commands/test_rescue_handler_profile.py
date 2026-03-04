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
from tests.conftest import FakeGitOps, FakeScanner

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
        """GenericProfile (empty source_layout) -> no profile-specific structure gaps."""
        from src.application.commands.rescue_handler import RescueHandler

        profile = GenericProfile()
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        profile_structure_gaps = [
            g for g in analysis.gaps
            if g.gap_type == GapType.MISSING_STRUCTURE
            and not g.path.startswith(".alty/")
        ]
        assert profile_structure_gaps == []


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
        """GenericProfile (empty manifest) -> no manifest gap reported."""
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
# 4. Full project with profile -> no gaps
# ---------------------------------------------------------------------------


class TestFullProjectWithProfile:
    def test_python_profile_all_present_no_gaps(self) -> None:
        """All artifacts present for Python profile -> no gaps."""
        from src.application.commands.rescue_handler import RescueHandler

        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            existing_docs=("docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md", "AGENTS.md"),
            existing_configs=(".claude/CLAUDE.md", "pyproject.toml"),
            existing_structure=("src/domain/", "src/application/", "src/infrastructure/"),
            has_knowledge_dir=True,
            has_agents_md=True,
            has_git=True,
            has_alty_config=True,
            has_maintenance_dir=True,
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
            has_alty_config=True,
            has_maintenance_dir=True,
        )
        profile = GenericProfile()
        handler = RescueHandler(project_scan=FakeScanner(scan=scan), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"), profile=profile)

        assert analysis.gaps == ()


# ---------------------------------------------------------------------------
# 5. Backward compatibility (no profile = generic fallback)
# ---------------------------------------------------------------------------


class TestBackwardCompat:
    def test_no_profile_still_works(self) -> None:
        """Calling rescue() without profile still returns gap analysis."""
        from src.application.commands.rescue_handler import RescueHandler

        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"))

        # Should still find doc/config/knowledge gaps even without profile
        assert len(analysis.gaps) > 0

    def test_no_profile_produces_no_structure_gaps(self) -> None:
        """Without profile, no profile-specific structure gaps are reported."""
        from src.application.commands.rescue_handler import RescueHandler

        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())

        analysis = handler.rescue(Path("/tmp/proj"))

        profile_structure_gaps = [
            g for g in analysis.gaps
            if g.gap_type == GapType.MISSING_STRUCTURE
            and not g.path.startswith(".alty/")
        ]
        assert profile_structure_gaps == []
