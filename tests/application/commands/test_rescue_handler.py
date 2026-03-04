"""Tests for RescueHandler application command.

Verifies the handler orchestrates project scanning, git validation,
gap analysis, plan creation, and plan execution correctly.
"""

from __future__ import annotations

from pathlib import Path

import pytest

from src.application.commands.rescue_handler import RescueHandler
from src.domain.models.errors import InvariantViolationError
from src.domain.models.gap_analysis import (
    AnalysisStatus,
    GapSeverity,
    GapType,
    ProjectScan,
)
from src.domain.models.stack_profile import PythonUvProfile
from tests.conftest import FakeGitOps, FakeScanner

_PROFILE = PythonUvProfile()

# -- Fake adapters ---------------------------------------------------------


class FakeFileWriter:
    """In-memory test double implementing FileWriterPort."""

    def __init__(self) -> None:
        self.written_files: dict[Path, str] = {}

    def write_file(self, path: Path, content: str) -> None:
        self.written_files[path] = content


# -- Validate Preconditions Tests ------------------------------------------


class TestValidatePreconditions:
    def test_validate_preconditions_raises_if_not_git_repo(self) -> None:
        git_ops = FakeGitOps(has_git=False)
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)

        with pytest.raises(InvariantViolationError, match="Not a git repository"):
            handler.validate_preconditions(Path("/tmp/proj"))

    def test_validate_preconditions_raises_on_dirty_tree(self) -> None:
        git_ops = FakeGitOps(is_clean=False)
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)

        with pytest.raises(InvariantViolationError, match="Working tree is dirty"):
            handler.validate_preconditions(Path("/tmp/proj"))

    def test_validate_preconditions_raises_if_branch_exists(self) -> None:
        git_ops = FakeGitOps(branch_exists=True)
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)

        with pytest.raises(InvariantViolationError, match="Branch alty/init already exists"):
            handler.validate_preconditions(Path("/tmp/proj"))

    def test_validate_preconditions_passes_for_clean_repo(self) -> None:
        git_ops = FakeGitOps()
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)

        # Should not raise
        handler.validate_preconditions(Path("/tmp/proj"))


# -- Git Precondition Tests ------------------------------------------------


class TestRescueHandlerGitPreconditions:
    def test_rescue_aborts_if_not_git_repo(self) -> None:
        git_ops = FakeGitOps(has_git=False)
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)

        with pytest.raises(InvariantViolationError, match="Not a git repository"):
            handler.rescue(Path("/tmp/proj"))

    def test_rescue_aborts_on_dirty_tree(self) -> None:
        git_ops = FakeGitOps(is_clean=False)
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)

        with pytest.raises(InvariantViolationError, match="Working tree is dirty"):
            handler.rescue(Path("/tmp/proj"))

    def test_rescue_aborts_if_branch_exists(self) -> None:
        git_ops = FakeGitOps(branch_exists=True)
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)

        with pytest.raises(InvariantViolationError, match="Branch alty/init already exists"):
            handler.rescue(Path("/tmp/proj"))


# -- Happy Path Tests ------------------------------------------------------


class TestRescueHandlerHappyPath:
    def test_rescue_happy_path_returns_gap_analysis(self) -> None:
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        # Should be in PLANNED state because there are gaps
        assert analysis.status == AnalysisStatus.PLANNED
        assert analysis.scan is not None
        assert analysis.plan is not None
        assert len(analysis.gaps) > 0

    def test_rescue_creates_branch_before_scanning(self) -> None:
        git_ops = FakeGitOps()
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)
        handler.rescue(Path("/tmp/proj"))

        assert "alty/init" in git_ops.created_branches

    def test_rescue_detects_missing_docs(self) -> None:
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        gap_paths = [g.path for g in analysis.gaps]
        assert "docs/PRD.md" in gap_paths
        assert "docs/DDD.md" in gap_paths
        assert "docs/ARCHITECTURE.md" in gap_paths

    def test_rescue_detects_missing_knowledge_dir(self) -> None:
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        knowledge_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_KNOWLEDGE]
        assert len(knowledge_gaps) == 1
        assert knowledge_gaps[0].path == ".alty/knowledge/"

    def test_rescue_skips_agents_md_if_exists(self) -> None:
        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            existing_docs=("AGENTS.md",),
            existing_configs=(),
            existing_structure=(),
            has_knowledge_dir=False,
            has_agents_md=True,
            has_git=True,
        )
        handler = RescueHandler(project_scan=FakeScanner(scan=scan), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        # AGENTS.md should not appear in gaps
        gap_paths = [g.path for g in analysis.gaps]
        assert "AGENTS.md" not in gap_paths

        # Plan should have skip_agents_md=True
        assert analysis.plan is not None
        assert analysis.plan.skip_agents_md is True

    def test_rescue_all_artifacts_present_returns_analyzed_with_no_gaps(self) -> None:
        """When everything exists, no gaps and no plan."""
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
        handler = RescueHandler(project_scan=FakeScanner(scan=scan), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        assert analysis.status == AnalysisStatus.ANALYZED
        assert analysis.gaps == ()
        assert analysis.plan is None

    def test_rescue_detects_missing_config(self) -> None:
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=_PROFILE)

        config_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_CONFIG]
        config_paths = [g.path for g in config_gaps]
        # alty-universal + Python manifest
        assert ".claude/CLAUDE.md" in config_paths
        assert "pyproject.toml" in config_paths

    def test_rescue_detects_missing_structure(self) -> None:
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=_PROFILE)

        structure_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_STRUCTURE]
        structure_paths = [g.path for g in structure_gaps]
        assert "src/domain/" in structure_paths
        assert "src/application/" in structure_paths
        assert "src/infrastructure/" in structure_paths

    def test_rescue_does_not_report_tests_gap(self) -> None:
        """alty does not prescribe where tests live — no tests/ gap."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        test_gaps = [
            g
            for g in analysis.gaps
            if g.gap_type == GapType.MISSING_STRUCTURE and g.path == "tests/"
        ]
        assert len(test_gaps) == 0

    def test_rescue_detects_missing_alty_config(self) -> None:
        """Gap when .alty/config.toml is absent."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        config_gaps = [
            g for g in analysis.gaps if g.path == ".alty/config.toml"
        ]
        assert len(config_gaps) == 1
        assert config_gaps[0].gap_type == GapType.MISSING_CONFIG

    def test_rescue_detects_missing_alty_maintenance(self) -> None:
        """Gap when .alty/maintenance/ is absent."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        maint_gaps = [
            g for g in analysis.gaps if g.path == ".alty/maintenance/"
        ]
        assert len(maint_gaps) == 1
        assert maint_gaps[0].gap_type == GapType.MISSING_STRUCTURE

    def test_rescue_no_alty_config_gap_when_present(self) -> None:
        """No gap when .alty/config.toml exists."""
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
        handler = RescueHandler(project_scan=FakeScanner(scan=scan), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        alty_gaps = [
            g for g in analysis.gaps
            if ".alty/" in g.path
        ]
        assert len(alty_gaps) == 0


# -- None Profile Fallback Tests -------------------------------------------


class TestRescueNoneProfileFallback:
    """When profile=None, rescue must not produce Python-specific gaps."""

    def test_none_profile_no_structure_gaps(self) -> None:
        """rescue(profile=None) must not report profile-specific structure gaps."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=None)

        # Profile-specific structure gaps (src/domain/, etc.) should not appear
        profile_structure_gaps = [
            g for g in analysis.gaps
            if g.gap_type == GapType.MISSING_STRUCTURE and not g.path.startswith(".alty/")
        ]
        assert profile_structure_gaps == []

    def test_none_profile_no_pyproject_gap(self) -> None:
        """rescue(profile=None) must not report pyproject.toml as missing config."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=None)

        config_paths = [
            g.path for g in analysis.gaps if g.gap_type == GapType.MISSING_CONFIG
        ]
        assert "pyproject.toml" not in config_paths


# -- Execute Plan Tests ----------------------------------------------------


class TestRescueHandlerExecutePlan:
    def test_execute_plan_completes_analysis(self) -> None:
        writer = FakeFileWriter()
        handler = RescueHandler(
            project_scan=FakeScanner(),
            git_ops=FakeGitOps(),
            file_writer=writer,
        )
        analysis = handler.rescue(Path("/tmp/proj"))

        handler.execute_plan(analysis)

        assert analysis.status == AnalysisStatus.COMPLETED
        assert len(analysis.events) == 1

    def test_execute_plan_writes_files(self) -> None:
        writer = FakeFileWriter()
        handler = RescueHandler(
            project_scan=FakeScanner(),
            git_ops=FakeGitOps(),
            file_writer=writer,
        )
        analysis = handler.rescue(Path("/tmp/proj"))

        handler.execute_plan(analysis)

        # Should have written files for the gaps
        assert len(writer.written_files) > 0

    def test_execute_plan_without_file_writer_raises(self) -> None:
        handler = RescueHandler(
            project_scan=FakeScanner(),
            git_ops=FakeGitOps(),
            file_writer=None,
        )
        analysis = handler.rescue(Path("/tmp/proj"))

        with pytest.raises(InvariantViolationError, match="No file writer configured"):
            handler.execute_plan(analysis)

    def test_execute_plan_wrong_state_raises(self) -> None:
        writer = FakeFileWriter()
        # Create analysis with no gaps (stays in ANALYZED, not PLANNED)
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
        scanner = FakeScanner(scan=scan)
        handler2 = RescueHandler(
            project_scan=scanner,
            git_ops=FakeGitOps(),
            file_writer=writer,
        )
        analysis = handler2.rescue(Path("/tmp/proj"))

        with pytest.raises(InvariantViolationError, match="Cannot execute plan in analyzed state"):
            handler2.execute_plan(analysis)

    def test_execute_plan_skips_agents_md_when_flagged(self) -> None:
        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            existing_docs=("AGENTS.md",),
            existing_configs=(),
            existing_structure=(),
            has_knowledge_dir=False,
            has_agents_md=True,
            has_git=True,
        )
        writer = FakeFileWriter()
        handler = RescueHandler(
            project_scan=FakeScanner(scan=scan),
            git_ops=FakeGitOps(),
            file_writer=writer,
        )
        analysis = handler.rescue(Path("/tmp/proj"))
        handler.execute_plan(analysis)

        # AGENTS.md should not be in written files
        written_paths = [str(p) for p in writer.written_files]
        assert not any("AGENTS.md" in p for p in written_paths)


# -- Gap Severity Tests ----------------------------------------------------


class TestGapSeverity:
    def test_required_docs_have_required_severity(self) -> None:
        """docs/PRD.md, DDD.md, ARCHITECTURE.md gaps are REQUIRED."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=_PROFILE)

        doc_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_DOC]
        for gap in doc_gaps:
            if gap.path in ("docs/PRD.md", "docs/DDD.md", "docs/ARCHITECTURE.md"):
                assert gap.severity == GapSeverity.REQUIRED

    def test_required_configs_have_required_severity(self) -> None:
        """.claude/CLAUDE.md and pyproject.toml gaps are REQUIRED."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=_PROFILE)

        config_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_CONFIG]
        for gap in config_gaps:
            if gap.path in (".claude/CLAUDE.md", "pyproject.toml"):
                assert gap.severity == GapSeverity.REQUIRED

    def test_structure_gaps_have_required_severity(self) -> None:
        """Profile structure gaps (src/domain/) are REQUIRED."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"), profile=_PROFILE)

        structure_gaps = [
            g for g in analysis.gaps
            if g.gap_type == GapType.MISSING_STRUCTURE
            and not g.path.startswith(".alty/")
        ]
        for gap in structure_gaps:
            assert gap.severity == GapSeverity.REQUIRED

    def test_alty_config_gap_has_recommended_severity(self) -> None:
        """.alty/config.toml gap is RECOMMENDED."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        alty_config_gap = next(
            g for g in analysis.gaps if g.path == ".alty/config.toml"
        )
        assert alty_config_gap.severity == GapSeverity.RECOMMENDED

    def test_alty_maintenance_gap_has_recommended_severity(self) -> None:
        """.alty/maintenance/ gap is RECOMMENDED."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        maint_gap = next(
            g for g in analysis.gaps if g.path == ".alty/maintenance/"
        )
        assert maint_gap.severity == GapSeverity.RECOMMENDED

    def test_knowledge_gap_has_recommended_severity(self) -> None:
        """.alty/knowledge/ gap is RECOMMENDED."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        knowledge_gap = next(
            g for g in analysis.gaps if g.path == ".alty/knowledge/"
        )
        assert knowledge_gap.severity == GapSeverity.RECOMMENDED

    def test_agents_md_gap_has_recommended_severity(self) -> None:
        """AGENTS.md gap is RECOMMENDED."""
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        agents_gap = next(
            g for g in analysis.gaps if g.path == "AGENTS.md"
        )
        assert agents_gap.severity == GapSeverity.RECOMMENDED


# -- Validated Parameter Tests ---------------------------------------------


class TestRescueValidatedParameter:
    def test_rescue_validated_skips_precondition_check(self) -> None:
        """rescue(validated=True) skips git precondition validation."""
        git_ops = FakeGitOps(has_git=False)
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)

        # Without validated=True, this would raise (not a git repo)
        # With validated=True, preconditions are skipped
        analysis = handler.rescue(Path("/tmp/proj"), validated=True)
        assert len(analysis.gaps) > 0

    def test_rescue_default_validates_preconditions(self) -> None:
        """rescue() without validated= still validates git preconditions."""
        git_ops = FakeGitOps(has_git=False)
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=git_ops)

        with pytest.raises(InvariantViolationError, match="Not a git repository"):
            handler.rescue(Path("/tmp/proj"))
