"""Port for the Bootstrap bounded context.

Defines the interface for bootstrapping a new project from a README idea
into a fully-seeded project with DDD artifacts, configs, and tickets.
"""

from __future__ import annotations

from typing import TYPE_CHECKING, Protocol, runtime_checkable

if TYPE_CHECKING:
    from pathlib import Path


@runtime_checkable
class BootstrapPort(Protocol):
    """Interface for project bootstrap operations.

    Adapters implement this to handle the preview-confirm-execute flow
    for creating a new project seed from a README idea.
    """

    def preview(self, project_dir: Path) -> str:
        """Preview what will be created during bootstrap.

        Args:
            project_dir: The target project directory containing a README.

        Returns:
            A human-readable preview of the planned bootstrap actions.
        """
        ...

    def confirm(self, session_id: str) -> str:
        """Confirm a previewed bootstrap session.

        Args:
            session_id: The identifier for the preview session to confirm.

        Returns:
            Confirmation status message.
        """
        ...

    def execute(self, session_id: str) -> str:
        """Execute a confirmed bootstrap session.

        Args:
            session_id: The identifier for the confirmed session to execute.

        Returns:
            Summary of the executed bootstrap actions.
        """
        ...
