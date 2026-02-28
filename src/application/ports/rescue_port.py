"""Port for the Rescue bounded context.

Defines the interface for adopting an existing project via gap analysis
and structural migration (alty init --existing).
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class RescuePort(Protocol):
    """Interface for rescue mode operations.

    Adapters implement this to handle analyzing an existing project,
    planning migration steps, and executing the rescue flow.
    """

    def analyze(self, project_dir: Path) -> str:
        """Analyze an existing project for structural gaps.

        Args:
            project_dir: The existing project directory to analyze.

        Returns:
            A gap analysis report comparing the project against a fully-seeded reference.
        """
        ...

    def plan(self, session_id: str) -> str:
        """Create a migration plan from the gap analysis.

        Args:
            session_id: The identifier for the analysis session.

        Returns:
            A human-readable migration plan.
        """
        ...

    def execute(self, session_id: str) -> str:
        """Execute the migration plan.

        Args:
            session_id: The identifier for the planned session to execute.

        Returns:
            Summary of the executed rescue actions.
        """
        ...
