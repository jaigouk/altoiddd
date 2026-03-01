"""Tests for GapAnalysis aggregate root and supporting value objects.

Verifies the state machine (SCANNING -> ANALYZED -> PLANNED -> EXECUTING ->
COMPLETED / FAILED) and the scan-before-analyze invariant.
"""

from __future__ import annotations

from pathlib import Path

import pytest

from src.domain.events.rescue_events import GapAnalysisCompleted
from src.domain.models.errors import InvariantViolationError
from src.domain.models.gap_analysis import (
    AnalysisStatus,
    Gap,
    GapAnalysis,
    GapType,
    MigrationPlan,
    ProjectScan,
)

# -- Helpers ---------------------------------------------------------------


def _make_scan(project_dir: Path | None = None) -> ProjectScan:
    """Build a minimal valid ProjectScan for testing."""
    return ProjectScan(
        project_dir=project_dir or Path("/tmp/proj"),
        existing_docs=("docs/PRD.md",),
        has_git=True,
    )


def _make_gap() -> Gap:
    """Build a minimal valid Gap for testing."""
    return Gap(
        gap_id="gap-1",
        gap_type=GapType.MISSING_DOC,
        path="docs/DDD.md",
        description="Missing documentation: docs/DDD.md",
        severity="required",
    )


def _make_plan(gaps: tuple[Gap, ...] | None = None) -> MigrationPlan:
    """Build a minimal valid MigrationPlan for testing."""
    return MigrationPlan(
        plan_id="plan-1",
        gaps=gaps or (_make_gap(),),
    )


# -- Creation --------------------------------------------------------------


class TestGapAnalysisCreation:
    def test_gap_analysis_initial_state_is_scanning(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        assert analysis.status == AnalysisStatus.SCANNING

    def test_gap_analysis_has_unique_id(self) -> None:
        a1 = GapAnalysis(project_dir=Path("/tmp/a"))
        a2 = GapAnalysis(project_dir=Path("/tmp/b"))
        assert a1.analysis_id != a2.analysis_id

    def test_gap_analysis_stores_project_dir(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        assert analysis.project_dir == Path("/tmp/proj")

    def test_gap_analysis_initial_gaps_empty(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        assert analysis.gaps == ()

    def test_gap_analysis_initial_scan_none(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        assert analysis.scan is None

    def test_gap_analysis_initial_plan_none(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        assert analysis.plan is None

    def test_gap_analysis_initial_events_empty(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        assert analysis.events == []


# -- Set Scan --------------------------------------------------------------


class TestGapAnalysisSetScan:
    def test_gap_analysis_set_scan_from_scanning(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        scan = _make_scan()
        analysis.set_scan(scan)
        assert analysis.scan == scan
        # Status remains SCANNING (analyze() transitions)
        assert analysis.status == AnalysisStatus.SCANNING

    def test_gap_analysis_set_scan_wrong_state_raises(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        analysis.set_scan(_make_scan())
        analysis.analyze((_make_gap(),))
        with pytest.raises(InvariantViolationError, match="Cannot set scan in analyzed state"):
            analysis.set_scan(_make_scan())


# -- Analyze ---------------------------------------------------------------


class TestGapAnalysisAnalyze:
    def test_gap_analysis_analyze_sets_gaps(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        analysis.set_scan(_make_scan())
        gap = _make_gap()
        analysis.analyze((gap,))
        assert analysis.gaps == (gap,)
        assert analysis.status == AnalysisStatus.ANALYZED

    def test_gap_analysis_analyze_without_scan_raises(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        with pytest.raises(InvariantViolationError, match="Cannot analyze without scan"):
            analysis.analyze((_make_gap(),))

    def test_gap_analysis_analyze_wrong_state_raises(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        analysis.set_scan(_make_scan())
        analysis.analyze((_make_gap(),))
        with pytest.raises(InvariantViolationError, match="Cannot analyze in analyzed state"):
            analysis.analyze((_make_gap(),))


# -- Create Plan -----------------------------------------------------------


class TestGapAnalysisCreatePlan:
    def test_gap_analysis_create_plan_from_analyzed(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        analysis.set_scan(_make_scan())
        analysis.analyze((_make_gap(),))
        plan = _make_plan()
        analysis.create_plan(plan)
        assert analysis.plan == plan
        assert analysis.status == AnalysisStatus.PLANNED

    def test_gap_analysis_create_plan_wrong_state_raises(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        with pytest.raises(
            InvariantViolationError, match="Cannot create plan in scanning state"
        ):
            analysis.create_plan(_make_plan())


# -- Begin Execution -------------------------------------------------------


class TestGapAnalysisBeginExecution:
    def test_gap_analysis_begin_execution_from_planned(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        analysis.set_scan(_make_scan())
        analysis.analyze((_make_gap(),))
        analysis.create_plan(_make_plan())
        analysis.begin_execution()
        assert analysis.status == AnalysisStatus.EXECUTING

    def test_gap_analysis_begin_execution_wrong_state_raises(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        with pytest.raises(
            InvariantViolationError, match="Cannot begin execution in scanning state"
        ):
            analysis.begin_execution()


# -- Complete --------------------------------------------------------------


class TestGapAnalysisComplete:
    def test_gap_analysis_complete_emits_event(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        analysis.set_scan(_make_scan())
        gap = _make_gap()
        analysis.analyze((gap,))
        analysis.create_plan(_make_plan((gap,)))
        analysis.begin_execution()
        analysis.complete()

        assert analysis.status == AnalysisStatus.COMPLETED
        assert len(analysis.events) == 1
        event = analysis.events[0]
        assert isinstance(event, GapAnalysisCompleted)
        assert event.analysis_id == analysis.analysis_id
        assert event.project_dir == Path("/tmp/proj")
        assert event.gaps_found == 1
        assert event.gaps_resolved == 1

    def test_gap_analysis_complete_wrong_state_raises(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        with pytest.raises(InvariantViolationError, match="Cannot complete in scanning state"):
            analysis.complete()

    def test_gap_analysis_events_returns_defensive_copy(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        analysis.set_scan(_make_scan())
        analysis.analyze((_make_gap(),))
        analysis.create_plan(_make_plan())
        analysis.begin_execution()
        analysis.complete()
        events = analysis.events
        events.clear()
        assert len(analysis.events) == 1


# -- Fail ------------------------------------------------------------------


class TestGapAnalysisFail:
    def test_gap_analysis_fail_from_executing(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        analysis.set_scan(_make_scan())
        analysis.analyze((_make_gap(),))
        analysis.create_plan(_make_plan())
        analysis.begin_execution()
        analysis.fail("Something went wrong")
        assert analysis.status == AnalysisStatus.FAILED

    def test_gap_analysis_fail_wrong_state_raises(self) -> None:
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        with pytest.raises(InvariantViolationError, match="Cannot fail in scanning state"):
            analysis.fail("reason")


# -- Value Objects (frozen) ------------------------------------------------


class TestValueObjects:
    def test_project_scan_is_frozen(self) -> None:
        scan = _make_scan()
        with pytest.raises(AttributeError):
            scan.has_git = False  # type: ignore[misc]

    def test_gap_is_frozen(self) -> None:
        gap = _make_gap()
        with pytest.raises(AttributeError):
            gap.severity = "optional"  # type: ignore[misc]

    def test_migration_plan_is_frozen(self) -> None:
        plan = _make_plan()
        with pytest.raises(AttributeError):
            plan.branch_name = "other"  # type: ignore[misc]


# -- Enums -----------------------------------------------------------------


class TestGapTypeValues:
    def test_gap_type_values(self) -> None:
        assert GapType.MISSING_DOC.value == "missing_doc"
        assert GapType.MISSING_CONFIG.value == "missing_config"
        assert GapType.MISSING_STRUCTURE.value == "missing_structure"
        assert GapType.MISSING_TOOLING.value == "missing_tooling"
        assert GapType.MISSING_KNOWLEDGE.value == "missing_knowledge"
        assert GapType.CONFLICT.value == "conflict"

    def test_analysis_status_values(self) -> None:
        assert AnalysisStatus.SCANNING.value == "scanning"
        assert AnalysisStatus.ANALYZED.value == "analyzed"
        assert AnalysisStatus.PLANNED.value == "planned"
        assert AnalysisStatus.EXECUTING.value == "executing"
        assert AnalysisStatus.COMPLETED.value == "completed"
        assert AnalysisStatus.FAILED.value == "failed"


# -- Edge Cases ------------------------------------------------------------


class TestGapAnalysisEdgeCases:
    def test_gap_analysis_empty_gaps(self) -> None:
        """Analyzing with zero gaps transitions to ANALYZED."""
        analysis = GapAnalysis(project_dir=Path("/tmp/proj"))
        analysis.set_scan(_make_scan())
        analysis.analyze(())
        assert analysis.status == AnalysisStatus.ANALYZED
        assert analysis.gaps == ()

    def test_gap_analysis_missing_knowledge_dir(self) -> None:
        """ProjectScan with has_knowledge_dir=False is a valid scan."""
        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            has_knowledge_dir=False,
            has_git=True,
        )
        assert scan.has_knowledge_dir is False

    def test_gap_analysis_existing_agents_md_not_a_gap(self) -> None:
        """ProjectScan with has_agents_md=True records the fact."""
        scan = ProjectScan(
            project_dir=Path("/tmp/proj"),
            has_agents_md=True,
            has_git=True,
        )
        assert scan.has_agents_md is True

    def test_migration_plan_skip_agents_md(self) -> None:
        """MigrationPlan can flag that AGENTS.md should be skipped."""
        plan = MigrationPlan(
            plan_id="plan-1",
            gaps=(_make_gap(),),
            skip_agents_md=True,
        )
        assert plan.skip_agents_md is True

    def test_gap_conflict_type(self) -> None:
        """Gap with CONFLICT type is valid."""
        gap = Gap(
            gap_id="gap-conflict",
            gap_type=GapType.CONFLICT,
            path=".claude/CLAUDE.md",
            description="Conflicting config",
            severity="required",
        )
        assert gap.gap_type == GapType.CONFLICT
