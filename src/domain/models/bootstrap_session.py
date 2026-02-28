"""BootstrapSession aggregate root.

Manages the lifecycle of a single bootstrap operation through a strict
state machine: CREATED -> PREVIEWED -> CONFIRMED -> EXECUTING -> COMPLETED
(or CANCELLED from PREVIEWED).

The core invariant is **preview-before-action**: no confirmation without a
preview, no execution without confirmation. This guarantees the user always
sees what alty will do before it does anything.
"""

from __future__ import annotations

import enum
import uuid
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from pathlib import Path

    from src.domain.events.bootstrap_events import BootstrapCompleted
    from src.domain.models.preview import Preview


class SessionStatus(enum.Enum):
    """States in the bootstrap session lifecycle."""

    CREATED = "created"
    PREVIEWED = "previewed"
    CONFIRMED = "confirmed"
    EXECUTING = "executing"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class InvariantViolationError(Exception):
    """Raised when a domain invariant is violated."""


class SessionNotFoundError(Exception):
    """Raised when a session_id does not match any active session."""


class BootstrapSession:
    """Aggregate root for the bootstrap flow.

    Enforces the preview-before-action invariant and produces domain events
    on completion.

    Attributes:
        session_id: Unique identifier for this session.
        project_dir: The directory being bootstrapped.
    """

    def __init__(self, project_dir: Path) -> None:
        self.session_id: str = str(uuid.uuid4())
        self.project_dir: Path = project_dir
        self._status: SessionStatus = SessionStatus.CREATED
        self._preview: Preview | None = None
        self._detected_tools: tuple[str, ...] = ()
        self._events: list[BootstrapCompleted] = []

    @property
    def status(self) -> SessionStatus:
        """Current session state."""
        return self._status

    @property
    def preview(self) -> Preview | None:
        """The current preview, or None if not yet set."""
        return self._preview

    @property
    def detected_tools(self) -> tuple[str, ...]:
        """AI coding tools detected in the project directory."""
        return self._detected_tools

    def set_detected_tools(self, tools: list[str]) -> None:
        """Record which AI coding tools were found.

        Args:
            tools: List of tool identifiers (e.g. ["claude", "cursor"]).
        """
        self._detected_tools = tuple(tools)

    @property
    def events(self) -> list[BootstrapCompleted]:
        """Domain events produced by this aggregate (defensive copy)."""
        return list(self._events)

    def set_preview(self, preview: Preview) -> None:
        """Set or replace the preview.

        Allowed from CREATED (first preview) or PREVIEWED (idempotent replace).

        Raises:
            InvariantViolationError: If the session is not in CREATED or PREVIEWED state.
        """
        if self._status not in (SessionStatus.CREATED, SessionStatus.PREVIEWED):
            msg = f"Cannot preview in {self._status.value} state"
            raise InvariantViolationError(msg)
        self._preview = preview
        self._status = SessionStatus.PREVIEWED

    def confirm(self) -> None:
        """Confirm the preview, allowing execution.

        Raises:
            InvariantViolationError: If the session has not been previewed.
        """
        if self._status != SessionStatus.PREVIEWED:
            raise InvariantViolationError("Cannot confirm without preview")
        self._status = SessionStatus.CONFIRMED

    def cancel(self) -> None:
        """Cancel the session after preview.

        Raises:
            InvariantViolationError: If the session is not in PREVIEWED state.
        """
        if self._status != SessionStatus.PREVIEWED:
            raise InvariantViolationError("Can only cancel from previewed state")
        self._status = SessionStatus.CANCELLED

    def begin_execution(self) -> None:
        """Transition to EXECUTING state.

        Raises:
            InvariantViolationError: If the session has not been confirmed.
        """
        if self._status != SessionStatus.CONFIRMED:
            raise InvariantViolationError("Cannot execute without confirmation")
        self._status = SessionStatus.EXECUTING

    def complete(self) -> None:
        """Mark execution as completed and emit BootstrapCompleted event.

        Raises:
            InvariantViolationError: If the session is not executing.
        """
        from src.domain.events.bootstrap_events import BootstrapCompleted

        if self._status != SessionStatus.EXECUTING:
            raise InvariantViolationError("Cannot complete unless executing")
        self._status = SessionStatus.COMPLETED
        self._events.append(
            BootstrapCompleted(
                session_id=self.session_id,
                project_dir=self.project_dir,
            )
        )
