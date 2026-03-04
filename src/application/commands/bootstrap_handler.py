"""BootstrapHandler -- application command for the bootstrap flow.

Orchestrates tool detection, preview generation, confirmation, and execution.
Depends on ports (abstractions), never on infrastructure directly.
"""

from __future__ import annotations

from pathlib import Path
from typing import TYPE_CHECKING

from src.domain.models.bootstrap_session import BootstrapSession, SessionNotFoundError
from src.domain.models.preview import FileAction, FileActionType, Preview

if TYPE_CHECKING:
    from src.application.ports.tool_detection_port import ToolDetectionPort

# Files alty plans to create in a new project.
_PLANNED_FILES: tuple[Path, ...] = (
    Path("docs/PRD.md"),
    Path("docs/DDD.md"),
    Path("docs/ARCHITECTURE.md"),
    Path("AGENTS.md"),
    Path(".alty/config.toml"),
    Path(".alty/knowledge/_index.toml"),
    Path(".alty/maintenance/doc-registry.toml"),
)


class BootstrapHandler:
    """Orchestrates the preview -> confirm -> execute bootstrap flow.

    Attributes:
        _tool_detection: Port for detecting installed AI coding tools.
        _sessions: In-memory store of active sessions (keyed by session_id).
    """

    def __init__(self, tool_detection: ToolDetectionPort) -> None:
        self._tool_detection = tool_detection
        self._sessions: dict[str, BootstrapSession] = {}

    def preview(self, project_dir: Path) -> BootstrapSession:
        """Create a new session and generate a preview of planned actions.

        Raises:
            FileNotFoundError: If the project directory has no README.md.
        """
        readme = project_dir / "README.md"
        if not readme.exists():
            raise FileNotFoundError("Create a README.md with your project idea first")

        detected_tools = self._tool_detection.detect(project_dir)
        conflict_descriptions = self._tool_detection.scan_conflicts(project_dir)

        file_actions: list[FileAction] = []
        for planned in _PLANNED_FILES:
            full_path = project_dir / planned
            if full_path.exists():
                file_actions.append(
                    FileAction(
                        path=planned,
                        action_type=FileActionType.SKIP,
                        reason="already exists",
                    )
                )
            else:
                file_actions.append(FileAction(path=planned, action_type=FileActionType.CREATE))

        preview = Preview(
            file_actions=tuple(file_actions),
            conflict_descriptions=tuple(conflict_descriptions),
        )
        session = BootstrapSession(project_dir=project_dir)
        session.set_detected_tools(detected_tools)
        session.set_preview(preview)
        self._sessions[session.session_id] = session
        return session

    def _get_session(self, session_id: str) -> BootstrapSession:
        """Look up a session by ID.

        Raises:
            SessionNotFoundError: If no session matches the given ID.
        """
        try:
            return self._sessions[session_id]
        except KeyError:
            raise SessionNotFoundError(f"No active session with id '{session_id}'") from None

    def confirm(self, session_id: str) -> BootstrapSession:
        """Confirm a previewed session, enabling execution."""
        session = self._get_session(session_id)
        session.confirm()
        return session

    def cancel(self, session_id: str) -> BootstrapSession:
        """Cancel a previewed session."""
        session = self._get_session(session_id)
        session.cancel()
        return session

    def execute(self, session_id: str) -> BootstrapSession:
        """Execute a confirmed session.

        Transitions through EXECUTING to COMPLETED. Emits BootstrapCompleted.
        """
        session = self._get_session(session_id)
        session.begin_execution()
        # Future: call DiscoveryPort, ArtifactRendererPort, etc.
        session.complete()
        return session
