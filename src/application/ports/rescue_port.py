"""Ports for the Rescue bounded context.

Defines interfaces for project scanning, git operations, and the rescue
orchestration flow (alty init --existing).
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.models.gap_analysis import GapAnalysis, MigrationPlan, ProjectScan


@runtime_checkable
class ProjectScanPort(Protocol):
    """Interface for scanning an existing project's structure.

    Adapters implement this to inspect the filesystem and report
    what documentation, configs, and structure already exist.
    """

    def scan(self, project_dir: Path) -> ProjectScan:
        """Scan a project directory and return a frozen snapshot.

        Args:
            project_dir: The project directory to scan.

        Returns:
            A ProjectScan value object describing the current state.
        """
        ...


@runtime_checkable
class GitOpsPort(Protocol):
    """Interface for git operations needed by the rescue flow.

    Adapters implement this to interact with the local git repository
    (check status, create branches, etc.).
    """

    def has_git(self, project_dir: Path) -> bool:
        """Check whether the directory is inside a git repository.

        Args:
            project_dir: The directory to check.

        Returns:
            True if the directory is a git repository.
        """
        ...

    def is_clean(self, project_dir: Path) -> bool:
        """Check whether the git working tree is clean.

        Args:
            project_dir: The directory to check.

        Returns:
            True if there are no uncommitted changes.
        """
        ...

    def branch_exists(self, project_dir: Path, branch_name: str) -> bool:
        """Check whether a git branch already exists.

        Args:
            project_dir: The directory to check.
            branch_name: The branch name to look for.

        Returns:
            True if the branch exists locally.
        """
        ...

    def create_branch(self, project_dir: Path, branch_name: str) -> None:
        """Create and check out a new git branch.

        Args:
            project_dir: The directory to operate in.
            branch_name: The branch name to create.
        """
        ...


@runtime_checkable
class RescuePort(Protocol):
    """Interface for rescue mode operations.

    Adapters implement this to handle analyzing an existing project,
    planning migration steps, and executing the rescue flow.
    """

    def analyze(self, project_dir: Path) -> GapAnalysis:
        """Analyze an existing project for structural gaps.

        Args:
            project_dir: The existing project directory to analyze.

        Returns:
            A GapAnalysis comparing the project against a fully-seeded reference.
        """
        ...

    def plan(self, gap_analysis: GapAnalysis) -> MigrationPlan:
        """Create a migration plan from the gap analysis.

        Args:
            gap_analysis: The completed gap analysis to plan from.

        Returns:
            A MigrationPlan with ordered steps to fill the gaps.
        """
        ...

    def execute(self, plan: MigrationPlan) -> None:
        """Execute the migration plan.

        Args:
            plan: The migration plan to execute.
        """
        ...
