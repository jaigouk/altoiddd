"""RescueHandler -- application command for the rescue flow.

Orchestrates project scanning, git validation, gap analysis, plan creation,
and plan execution. Depends on ports (abstractions), never on infrastructure
directly.
"""

from __future__ import annotations

import uuid
from typing import TYPE_CHECKING

from src.domain.models.errors import InvariantViolationError
from src.domain.models.gap_analysis import (
    AnalysisStatus,
    Gap,
    GapAnalysis,
    GapSeverity,
    GapType,
    MigrationPlan,
)

if TYPE_CHECKING:
    from pathlib import Path

    from src.application.ports.file_writer_port import FileWriterPort
    from src.application.ports.rescue_port import GitOpsPort, ProjectScanPort
    from src.domain.models.gap_analysis import ProjectScan
    from src.domain.models.stack_profile import StackProfile

# Reference structure: docs and configs an alty project should have.
_REQUIRED_DOCS: tuple[str, ...] = (
    "docs/PRD.md",
    "docs/DDD.md",
    "docs/ARCHITECTURE.md",
)

_REQUIRED_CONFIGS: tuple[str, ...] = (".claude/CLAUDE.md",)

_BRANCH_NAME = "alty/init"


class RescueHandler:
    """Orchestrates the rescue flow: scan -> validate -> analyze -> plan.

    Attributes:
        _project_scan: Port for scanning the project directory.
        _git_ops: Port for git operations.
        _file_writer: Optional port for writing generated artifacts.
    """

    def __init__(
        self,
        project_scan: ProjectScanPort,
        git_ops: GitOpsPort,
        file_writer: FileWriterPort | None = None,
    ) -> None:
        self._project_scan = project_scan
        self._git_ops = git_ops
        self._file_writer = file_writer

    def validate_preconditions(self, project_dir: Path) -> None:
        """Validate git preconditions before rescue.

        Call this early (before asking tech stack) so the user doesn't
        answer questions only to discover git state is invalid.

        Args:
            project_dir: The existing project directory.

        Raises:
            InvariantViolationError: If git preconditions are not met.
        """
        if not self._git_ops.has_git(project_dir):
            raise InvariantViolationError("Not a git repository")

        if not self._git_ops.is_clean(project_dir):
            raise InvariantViolationError("Working tree is dirty")

        if self._git_ops.branch_exists(project_dir, _BRANCH_NAME):
            raise InvariantViolationError(
                f"Branch {_BRANCH_NAME} already exists. "
                "Delete it first or use --force-branch to override."
            )

    def rescue(
        self,
        project_dir: Path,
        profile: StackProfile | None = None,
        *,
        validated: bool = False,
    ) -> GapAnalysis:
        """Analyze an existing project and produce a gap analysis with plan.

        Validates git preconditions (unless already validated), scans the
        project, identifies gaps, and creates a migration plan. The caller
        decides whether to proceed (preview-before-action pattern).

        Args:
            project_dir: The existing project directory.
            profile: Stack profile providing structure and manifest targets.
                When None, uses default Python targets for backward compat.
            validated: If True, skip precondition validation (caller already
                called validate_preconditions).

        Returns:
            A GapAnalysis aggregate in PLANNED state (or ANALYZED if no gaps).

        Raises:
            InvariantViolationError: If git preconditions are not met.
        """
        if not validated:
            self.validate_preconditions(project_dir)

        # Create branch before scanning
        self._git_ops.create_branch(project_dir, _BRANCH_NAME)

        # Scan project (profile tells scanner which structure/config targets to check)
        scan = self._project_scan.scan(project_dir, profile=profile)

        # Analyze gaps
        gaps = self._identify_gaps(scan, profile)

        # Build aggregate
        analysis = GapAnalysis(project_dir=project_dir)
        analysis.set_scan(scan)
        analysis.analyze(gaps)

        if gaps:
            plan = MigrationPlan(
                plan_id=str(uuid.uuid4()),
                gaps=gaps,
                branch_name=_BRANCH_NAME,
                skip_agents_md=scan.has_agents_md,
            )
            analysis.create_plan(plan)

        return analysis

    def execute_plan(self, analysis: GapAnalysis) -> None:
        """Execute a planned migration.

        Transitions the aggregate through EXECUTING to COMPLETED and writes
        missing artifacts via the file writer port.

        Args:
            analysis: A GapAnalysis aggregate in PLANNED state.

        Raises:
            InvariantViolationError: If the analysis is not in PLANNED state
                or no file writer is configured.
        """
        if analysis.status != AnalysisStatus.PLANNED:
            msg = f"Cannot execute plan in {analysis.status.value} state"
            raise InvariantViolationError(msg)

        if self._file_writer is None:
            raise InvariantViolationError("No file writer configured for plan execution")

        analysis.begin_execution()

        plan = analysis.plan
        if plan is None:  # pragma: no cover — guarded by PLANNED state
            analysis.fail("No plan available")
            return

        from pathlib import PurePosixPath

        for gap in plan.gaps:
            if gap.gap_type == GapType.CONFLICT:
                continue  # Conflicts are flagged, not auto-resolved

            if plan.skip_agents_md and gap.path == "AGENTS.md":
                continue

            target = analysis.project_dir / gap.path
            stem = PurePosixPath(gap.path).stem
            self._file_writer.write_file(
                target,
                f"# {stem}\n\n> TODO: Fill in content.\n",
            )

        analysis.complete()

    def _identify_gaps(
        self,
        scan: ProjectScan,
        profile: StackProfile | None = None,
    ) -> tuple[Gap, ...]:
        """Compare scan results against reference structure to find gaps.

        Args:
            scan: The project scan result.
            profile: Stack profile providing structure and manifest targets.

        Returns:
            Tuple of identified gaps.
        """
        # Check required docs
        gaps: list[Gap] = [
            Gap(
                gap_id=str(uuid.uuid4()),
                gap_type=GapType.MISSING_DOC,
                path=doc_path,
                description=f"Missing documentation: {doc_path}",
                severity=GapSeverity.REQUIRED,
            )
            for doc_path in _REQUIRED_DOCS
            if doc_path not in scan.existing_docs
        ]

        # Check required configs (alty-universal)
        gaps.extend(
            Gap(
                gap_id=str(uuid.uuid4()),
                gap_type=GapType.MISSING_CONFIG,
                path=config_path,
                description=f"Missing configuration: {config_path}",
                severity=GapSeverity.REQUIRED,
            )
            for config_path in _REQUIRED_CONFIGS
            if config_path not in scan.existing_configs
        )

        # Check project manifest from profile
        manifest = profile.project_manifest if profile is not None else ""
        if manifest and manifest not in scan.existing_configs:
            gaps.append(
                Gap(
                    gap_id=str(uuid.uuid4()),
                    gap_type=GapType.MISSING_CONFIG,
                    path=manifest,
                    description=f"Missing configuration: {manifest}",
                    severity=GapSeverity.REQUIRED,
                )
            )

        # Check structure from profile
        structure_targets = profile.source_layout if profile is not None else ()
        gaps.extend(
            Gap(
                gap_id=str(uuid.uuid4()),
                gap_type=GapType.MISSING_STRUCTURE,
                path=structure_path,
                description=f"Missing directory: {structure_path}",
                severity=GapSeverity.REQUIRED,
            )
            for structure_path in structure_targets
            if structure_path not in scan.existing_structure
        )

        # Check knowledge directory
        if not scan.has_knowledge_dir:
            gaps.append(
                Gap(
                    gap_id=str(uuid.uuid4()),
                    gap_type=GapType.MISSING_KNOWLEDGE,
                    path=".alty/knowledge/",
                    description="Missing knowledge base directory",
                    severity=GapSeverity.RECOMMENDED,
                )
            )

        # Check .alty/config.toml
        if not scan.has_alty_config:
            gaps.append(
                Gap(
                    gap_id=str(uuid.uuid4()),
                    gap_type=GapType.MISSING_CONFIG,
                    path=".alty/config.toml",
                    description="Missing alty project configuration",
                    severity=GapSeverity.RECOMMENDED,
                )
            )

        # Check .alty/maintenance/
        if not scan.has_maintenance_dir:
            gaps.append(
                Gap(
                    gap_id=str(uuid.uuid4()),
                    gap_type=GapType.MISSING_STRUCTURE,
                    path=".alty/maintenance/",
                    description="Missing doc maintenance directory",
                    severity=GapSeverity.RECOMMENDED,
                )
            )

        # Check AGENTS.md
        if not scan.has_agents_md:
            gaps.append(
                Gap(
                    gap_id=str(uuid.uuid4()),
                    gap_type=GapType.MISSING_DOC,
                    path="AGENTS.md",
                    description="Missing AGENTS.md",
                    severity=GapSeverity.RECOMMENDED,
                )
            )

        return tuple(gaps)
