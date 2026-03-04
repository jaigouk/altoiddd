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
    GapType,
    ProjectScan,
)

# -- Fake adapters ---------------------------------------------------------


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
        )


class FakeGitOps:
    """In-memory test double implementing GitOpsPort."""

    def __init__(
        self,
        has_git: bool = True,
        is_clean: bool = True,
        branch_exists: bool = False,
    ) -> None:
        self._has_git = has_git
        self._is_clean = is_clean
        self._branch_exists = branch_exists
        self.created_branches: list[str] = []

    def has_git(self, project_dir: Path) -> bool:
        return self._has_git

    def is_clean(self, project_dir: Path) -> bool:
        return self._is_clean

    def branch_exists(self, project_dir: Path, branch_name: str) -> bool:
        return self._branch_exists

    def create_branch(self, project_dir: Path, branch_name: str) -> None:
        self.created_branches.append(branch_name)


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
        )
        handler = RescueHandler(project_scan=FakeScanner(scan=scan), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        assert analysis.status == AnalysisStatus.ANALYZED
        assert analysis.gaps == ()
        assert analysis.plan is None

    def test_rescue_detects_missing_config(self) -> None:
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

        config_gaps = [g for g in analysis.gaps if g.gap_type == GapType.MISSING_CONFIG]
        config_paths = [g.path for g in config_gaps]
        # alty-universal + default manifest (pyproject.toml)
        assert ".claude/CLAUDE.md" in config_paths
        assert "pyproject.toml" in config_paths

    def test_rescue_detects_missing_structure(self) -> None:
        handler = RescueHandler(project_scan=FakeScanner(), git_ops=FakeGitOps())
        analysis = handler.rescue(Path("/tmp/proj"))

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
