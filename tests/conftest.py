"""Shared test fixtures.

Add project-wide fixtures here. Domain tests should rarely need fixtures
since domain objects are pure and can be constructed directly.
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from src.domain.models.gap_analysis import ProjectScan

if TYPE_CHECKING:
    from pathlib import Path


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
