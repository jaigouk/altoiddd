"""GapAnalysis aggregate root and supporting value objects.

Manages the lifecycle of a rescue operation (alty init --existing) through
a strict state machine: SCANNING -> ANALYZED -> PLANNED -> EXECUTING ->
COMPLETED (or FAILED from EXECUTING).

The core invariant is **scan-before-analyze**: no analysis without a scan,
no plan without analysis, no execution without a plan.
"""

from __future__ import annotations

import enum
import uuid
from dataclasses import dataclass
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.events.rescue_events import GapAnalysisCompleted

from src.domain.models.errors import InvariantViolationError as InvariantViolationError


class GapType(enum.Enum):
    """Classification of a structural gap found during project analysis.

    MISSING_DOC:       A required documentation file is absent.
    MISSING_CONFIG:    A required configuration file is absent.
    MISSING_STRUCTURE: A required directory structure is absent.
    MISSING_TOOLING:   A required tooling config is absent.
    MISSING_KNOWLEDGE: The knowledge base directory is absent.
    CONFLICT:          An existing file conflicts with the expected structure.
    """

    MISSING_DOC = "missing_doc"
    MISSING_CONFIG = "missing_config"
    MISSING_STRUCTURE = "missing_structure"
    MISSING_TOOLING = "missing_tooling"
    MISSING_KNOWLEDGE = "missing_knowledge"
    CONFLICT = "conflict"


@dataclass(frozen=True)
class ProjectScan:
    """Immutable snapshot of an existing project's current state.

    Attributes:
        project_dir: Root directory of the scanned project.
        existing_docs: Tuple of documentation file paths found.
        existing_configs: Tuple of configuration file paths found.
        existing_structure: Tuple of directory structure paths found.
        has_knowledge_dir: Whether .alty/knowledge/ exists.
        has_agents_md: Whether AGENTS.md exists at the project root.
        has_git: Whether the project is a git repository.
        has_tests: Whether a tests/ directory exists.
    """

    project_dir: Path
    existing_docs: tuple[str, ...] = ()
    existing_configs: tuple[str, ...] = ()
    existing_structure: tuple[str, ...] = ()
    has_knowledge_dir: bool = False
    has_agents_md: bool = False
    has_git: bool = False
    has_tests: bool = False


@dataclass(frozen=True)
class Gap:
    """A single structural gap found during project analysis.

    Attributes:
        gap_id: Unique identifier for this gap.
        gap_type: Classification of the gap.
        path: Relative path where the gap was detected.
        description: Human-readable explanation of the gap.
        severity: How critical the gap is: "required", "recommended", or "optional".
    """

    gap_id: str
    gap_type: GapType
    path: str
    description: str
    severity: str  # "required" | "recommended" | "optional"


@dataclass(frozen=True)
class MigrationPlan:
    """Immutable plan for resolving gaps in an existing project.

    Attributes:
        plan_id: Unique identifier for this plan.
        gaps: Tuple of gaps to resolve.
        branch_name: Git branch to create for the migration.
        skip_agents_md: Whether to skip generating AGENTS.md (already exists).
    """

    plan_id: str
    gaps: tuple[Gap, ...]
    branch_name: str = "alty/init"
    skip_agents_md: bool = False


class AnalysisStatus(enum.Enum):
    """States in the gap analysis lifecycle.

    SCANNING:   Initial state, project is being scanned.
    ANALYZED:   Gaps have been identified.
    PLANNED:    A migration plan has been created.
    EXECUTING:  The plan is being executed.
    COMPLETED:  Execution finished successfully.
    FAILED:     Execution encountered an error.
    """

    SCANNING = "scanning"
    ANALYZED = "analyzed"
    PLANNED = "planned"
    EXECUTING = "executing"
    COMPLETED = "completed"
    FAILED = "failed"


class GapAnalysis:
    """Aggregate root for the rescue flow.

    Enforces the scan-before-analyze invariant and produces domain events
    on completion.

    Attributes:
        analysis_id: Unique identifier for this analysis.
        project_dir: The directory being analyzed.
    """

    def __init__(self, project_dir: Path) -> None:
        self.analysis_id: str = str(uuid.uuid4())
        self.project_dir: Path = project_dir
        self._status: AnalysisStatus = AnalysisStatus.SCANNING
        self._scan: ProjectScan | None = None
        self._gaps: tuple[Gap, ...] = ()
        self._plan: MigrationPlan | None = None
        self._events: list[GapAnalysisCompleted] = []

    @property
    def status(self) -> AnalysisStatus:
        """Current analysis state."""
        return self._status

    @property
    def scan(self) -> ProjectScan | None:
        """The project scan result, or None if not yet scanned."""
        return self._scan

    @property
    def gaps(self) -> tuple[Gap, ...]:
        """Identified structural gaps."""
        return self._gaps

    @property
    def plan(self) -> MigrationPlan | None:
        """The migration plan, or None if not yet planned."""
        return self._plan

    @property
    def events(self) -> list[GapAnalysisCompleted]:
        """Domain events produced by this aggregate (defensive copy)."""
        return list(self._events)

    def set_scan(self, scan: ProjectScan) -> None:
        """Record scan results. Only from SCANNING state.

        Args:
            scan: The project scan result.

        Raises:
            InvariantViolationError: If not in SCANNING state.
        """
        if self._status != AnalysisStatus.SCANNING:
            msg = f"Cannot set scan in {self._status.value} state"
            raise InvariantViolationError(msg)
        self._scan = scan

    def analyze(self, gaps: tuple[Gap, ...]) -> None:
        """Set analysis results. Requires scan to exist.

        Args:
            gaps: The identified gaps.

        Raises:
            InvariantViolationError: If not in SCANNING state or scan is missing.
        """
        if self._status != AnalysisStatus.SCANNING:
            msg = f"Cannot analyze in {self._status.value} state"
            raise InvariantViolationError(msg)
        if self._scan is None:
            raise InvariantViolationError("Cannot analyze without scan")
        self._gaps = gaps
        self._status = AnalysisStatus.ANALYZED

    def create_plan(self, plan: MigrationPlan) -> None:
        """Create migration plan from gaps. Only from ANALYZED.

        Args:
            plan: The migration plan to apply.

        Raises:
            InvariantViolationError: If not in ANALYZED state.
        """
        if self._status != AnalysisStatus.ANALYZED:
            msg = f"Cannot create plan in {self._status.value} state"
            raise InvariantViolationError(msg)
        self._plan = plan
        self._status = AnalysisStatus.PLANNED

    def begin_execution(self) -> None:
        """Start executing the plan. Only from PLANNED.

        Raises:
            InvariantViolationError: If not in PLANNED state.
        """
        if self._status != AnalysisStatus.PLANNED:
            msg = f"Cannot begin execution in {self._status.value} state"
            raise InvariantViolationError(msg)
        self._status = AnalysisStatus.EXECUTING

    def complete(self) -> None:
        """Mark as completed and emit event.

        Raises:
            InvariantViolationError: If not in EXECUTING state.
        """
        from src.domain.events.rescue_events import GapAnalysisCompleted

        if self._status != AnalysisStatus.EXECUTING:
            msg = f"Cannot complete in {self._status.value} state"
            raise InvariantViolationError(msg)
        self._status = AnalysisStatus.COMPLETED
        self._events.append(
            GapAnalysisCompleted(
                analysis_id=self.analysis_id,
                project_dir=self.project_dir,
                gaps_found=len(self._gaps),
                gaps_resolved=len(self._gaps),
            )
        )

    def fail(self, reason: str) -> None:
        """Mark as failed. Only from EXECUTING.

        Args:
            reason: Human-readable failure reason.

        Raises:
            InvariantViolationError: If not in EXECUTING state.
        """
        if self._status != AnalysisStatus.EXECUTING:
            msg = f"Cannot fail in {self._status.value} state"
            raise InvariantViolationError(msg)
        self._status = AnalysisStatus.FAILED
