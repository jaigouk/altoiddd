"""GitOpsAdapter -- infrastructure adapter for GitOpsPort.

Uses subprocess to execute git commands. Branch names are sanitized to
prevent command injection (only alphanumeric, /, -, _ allowed).
"""

from __future__ import annotations

import re
import subprocess
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path


_BRANCH_NAME_PATTERN = re.compile(r"^[a-zA-Z0-9/_-]+$")

_CMD_IS_GIT_REPO = ["git", "rev-parse", "--is-inside-work-tree"]
_CMD_STATUS = ["git", "status", "--porcelain"]


class GitOpsAdapter:
    """Filesystem-based implementation of GitOpsPort.

    Delegates to the ``git`` command-line tool via subprocess.
    """

    def has_git(self, project_dir: Path) -> bool:
        """Check whether the directory is inside a git repository."""
        result = subprocess.run(  # noqa: S603
            _CMD_IS_GIT_REPO,
            cwd=project_dir,
            capture_output=True,
            text=True,
        )
        return result.returncode == 0

    def is_clean(self, project_dir: Path) -> bool:
        """Check whether the git working tree is clean."""
        result = subprocess.run(  # noqa: S603
            _CMD_STATUS,
            cwd=project_dir,
            capture_output=True,
            text=True,
        )
        return result.returncode == 0 and result.stdout.strip() == ""

    def branch_exists(self, project_dir: Path, branch_name: str) -> bool:
        """Check whether a git branch already exists locally."""
        self._validate_branch_name(branch_name)
        cmd = ["git", "rev-parse", "--verify", f"refs/heads/{branch_name}"]
        result = subprocess.run(  # noqa: S603
            cmd,
            cwd=project_dir,
            capture_output=True,
            text=True,
        )
        return result.returncode == 0

    def create_branch(self, project_dir: Path, branch_name: str) -> None:
        """Create and check out a new git branch."""
        self._validate_branch_name(branch_name)
        cmd = ["git", "checkout", "-b", branch_name]
        subprocess.run(  # noqa: S603
            cmd,
            cwd=project_dir,
            capture_output=True,
            text=True,
            check=True,
        )

    @staticmethod
    def _validate_branch_name(branch_name: str) -> None:
        """Validate that a branch name contains only safe characters.

        Args:
            branch_name: The branch name to validate.

        Raises:
            ValueError: If the branch name contains invalid characters.
        """
        if not _BRANCH_NAME_PATTERN.match(branch_name):
            msg = (
                f"Invalid branch name: {branch_name!r}. "
                "Only alphanumeric characters, '/', '-', and '_' are allowed."
            )
            raise ValueError(msg)
