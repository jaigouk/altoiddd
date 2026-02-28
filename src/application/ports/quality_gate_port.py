"""Port for quality gate operations.

Defines the interface for running quality gates (lint, type-check, tests)
on a project directory. Used across multiple bounded contexts.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class QualityGatePort(Protocol):
    """Interface for running quality gate checks.

    Adapters implement this to execute lint, type-check, and test
    commands against a project directory.
    """

    def run_all(self, project_dir: Path) -> str:
        """Run all quality gates (lint, types, tests).

        Args:
            project_dir: The project directory to check.

        Returns:
            Combined results from all quality gate checks.
        """
        ...

    def run_lint(self, project_dir: Path) -> str:
        """Run the lint quality gate.

        Args:
            project_dir: The project directory to lint.

        Returns:
            Lint check results.
        """
        ...

    def run_types(self, project_dir: Path) -> str:
        """Run the type-check quality gate.

        Args:
            project_dir: The project directory to type-check.

        Returns:
            Type-check results.
        """
        ...

    def run_tests(self, project_dir: Path) -> str:
        """Run the test quality gate.

        Args:
            project_dir: The project directory to test.

        Returns:
            Test results.
        """
        ...
