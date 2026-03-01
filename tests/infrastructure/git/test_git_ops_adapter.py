"""Tests for GitOpsAdapter infrastructure adapter.

Uses unittest.mock.patch to mock subprocess.run calls.
"""

from __future__ import annotations

import subprocess
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from src.infrastructure.git.git_ops_adapter import GitOpsAdapter


class TestGitOpsAdapterHasGit:
    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_has_git_returns_true_for_git_repo(self, mock_run: MagicMock) -> None:
        mock_run.return_value = MagicMock(returncode=0, stdout="true\n")
        adapter = GitOpsAdapter()
        assert adapter.has_git(Path("/tmp/proj")) is True
        mock_run.assert_called_once()

    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_has_git_returns_false_for_non_git_repo(self, mock_run: MagicMock) -> None:
        mock_run.return_value = MagicMock(returncode=128, stdout="", stderr="fatal")
        adapter = GitOpsAdapter()
        assert adapter.has_git(Path("/tmp/proj")) is False


class TestGitOpsAdapterIsClean:
    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_is_clean_returns_true_for_clean_tree(self, mock_run: MagicMock) -> None:
        mock_run.return_value = MagicMock(returncode=0, stdout="")
        adapter = GitOpsAdapter()
        assert adapter.is_clean(Path("/tmp/proj")) is True

    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_is_clean_returns_false_for_dirty_tree(self, mock_run: MagicMock) -> None:
        mock_run.return_value = MagicMock(returncode=0, stdout="M src/main.py\n")
        adapter = GitOpsAdapter()
        assert adapter.is_clean(Path("/tmp/proj")) is False

    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_is_clean_returns_false_on_error(self, mock_run: MagicMock) -> None:
        mock_run.return_value = MagicMock(returncode=128, stdout="")
        adapter = GitOpsAdapter()
        assert adapter.is_clean(Path("/tmp/proj")) is False


class TestGitOpsAdapterBranchExists:
    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_branch_exists_returns_true(self, mock_run: MagicMock) -> None:
        mock_run.return_value = MagicMock(returncode=0)
        adapter = GitOpsAdapter()
        assert adapter.branch_exists(Path("/tmp/proj"), "alty/init") is True

    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_branch_exists_returns_false(self, mock_run: MagicMock) -> None:
        mock_run.return_value = MagicMock(returncode=128)
        adapter = GitOpsAdapter()
        assert adapter.branch_exists(Path("/tmp/proj"), "alty/init") is False


class TestGitOpsAdapterCreateBranch:
    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_create_branch_calls_git(self, mock_run: MagicMock) -> None:
        mock_run.return_value = MagicMock(returncode=0)
        adapter = GitOpsAdapter()
        adapter.create_branch(Path("/tmp/proj"), "alty/init")
        mock_run.assert_called_once_with(
            ["git", "checkout", "-b", "alty/init"],
            cwd=Path("/tmp/proj"),
            capture_output=True,
            text=True,
            check=True,
        )

    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_create_branch_raises_on_failure(self, mock_run: MagicMock) -> None:
        mock_run.side_effect = subprocess.CalledProcessError(128, "git")
        adapter = GitOpsAdapter()
        with pytest.raises(subprocess.CalledProcessError):
            adapter.create_branch(Path("/tmp/proj"), "alty/init")


class TestGitOpsAdapterBranchNameValidation:
    def test_valid_branch_name_passes(self) -> None:
        # Should not raise
        GitOpsAdapter._validate_branch_name("alty/init")
        GitOpsAdapter._validate_branch_name("feature/my-branch")
        GitOpsAdapter._validate_branch_name("fix_123")

    def test_invalid_branch_name_raises(self) -> None:
        with pytest.raises(ValueError, match="Invalid branch name"):
            GitOpsAdapter._validate_branch_name("alty init")

    def test_invalid_branch_name_with_special_chars_raises(self) -> None:
        with pytest.raises(ValueError, match="Invalid branch name"):
            GitOpsAdapter._validate_branch_name("branch;rm -rf /")

    def test_empty_branch_name_raises(self) -> None:
        with pytest.raises(ValueError, match="Invalid branch name"):
            GitOpsAdapter._validate_branch_name("")

    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_branch_exists_validates_name(self, mock_run: MagicMock) -> None:
        adapter = GitOpsAdapter()
        with pytest.raises(ValueError, match="Invalid branch name"):
            adapter.branch_exists(Path("/tmp/proj"), "bad name")

    @patch("src.infrastructure.git.git_ops_adapter.subprocess.run")
    def test_create_branch_validates_name(self, mock_run: MagicMock) -> None:
        adapter = GitOpsAdapter()
        with pytest.raises(ValueError, match="Invalid branch name"):
            adapter.create_branch(Path("/tmp/proj"), "bad;name")
