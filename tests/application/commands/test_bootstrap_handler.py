"""Tests for BootstrapHandler application command.

Verifies the handler orchestrates tool detection, preview generation,
confirmation, and execution flows correctly.
"""

from __future__ import annotations

from pathlib import Path

import pytest

from src.application.commands.bootstrap_handler import BootstrapHandler
from src.domain.models.bootstrap_session import SessionNotFoundError, SessionStatus
from src.domain.models.preview import FileActionType

# ── Fake adapter ─────────────────────────────────────────────


class FakeToolDetection:
    """In-memory test double implementing ToolDetectionPort."""

    def __init__(
        self,
        tools: list[str] | None = None,
        conflicts: list[str] | None = None,
    ) -> None:
        self._tools = tools or ["claude"]
        self._conflicts = conflicts or []

    def detect(self, project_dir: Path) -> list[str]:
        return self._tools

    def scan_conflicts(self, project_dir: Path) -> list[str]:
        return self._conflicts


# ── Tests ────────────────────────────────────────────────────


class TestBootstrapHandler:
    def test_handler_calls_tool_detection(self, tmp_path):
        (tmp_path / "README.md").write_text("My project idea")
        fake = FakeToolDetection(tools=["claude", "cursor"])
        handler = BootstrapHandler(tool_detection=fake)

        session = handler.preview(tmp_path)
        # Session was created and preview set
        assert session.status == SessionStatus.PREVIEWED

    def test_handler_creates_preview_with_file_actions(self, tmp_path):
        (tmp_path / "README.md").write_text("My project idea")
        handler = BootstrapHandler(tool_detection=FakeToolDetection())

        session = handler.preview(tmp_path)
        preview = session.preview
        assert preview is not None
        assert len(preview.file_actions) > 0
        # All actions should be CREATE or SKIP, never OVERWRITE
        for action in preview.file_actions:
            assert action.action_type in (
                FileActionType.CREATE,
                FileActionType.SKIP,
                FileActionType.CONFLICT_RENAME,
            )

    def test_handler_full_flow_preview_confirm_execute(self, tmp_path):
        (tmp_path / "README.md").write_text("My project idea")
        handler = BootstrapHandler(tool_detection=FakeToolDetection())

        session = handler.preview(tmp_path)
        session_id = session.session_id

        handler.confirm(session_id)
        assert session.status == SessionStatus.CONFIRMED

        handler.execute(session_id)
        assert session.status == SessionStatus.COMPLETED  # type: ignore[comparison-overlap]
        assert len(session.events) == 1

    def test_handler_cancel_flow(self, tmp_path):
        (tmp_path / "README.md").write_text("My project idea")
        handler = BootstrapHandler(tool_detection=FakeToolDetection())

        session = handler.preview(tmp_path)
        session_id = session.session_id

        handler.cancel(session_id)
        assert session.status == SessionStatus.CANCELLED

    def test_handler_raises_on_missing_readme(self, tmp_path):
        handler = BootstrapHandler(tool_detection=FakeToolDetection())

        with pytest.raises(FileNotFoundError, match=r"README\.md"):
            handler.preview(tmp_path)

    def test_handler_skips_existing_files(self, tmp_path):
        (tmp_path / "README.md").write_text("My project idea")
        (tmp_path / "docs").mkdir()
        (tmp_path / "docs" / "PRD.md").write_text("existing")
        handler = BootstrapHandler(tool_detection=FakeToolDetection())

        session = handler.preview(tmp_path)
        preview = session.preview
        assert preview is not None

        prd_action = next(a for a in preview.file_actions if a.path == Path("docs/PRD.md"))
        assert prd_action.action_type == FileActionType.SKIP
        assert prd_action.reason == "already exists"

    def test_handler_stores_detected_tools_on_session(self, tmp_path):
        (tmp_path / "README.md").write_text("My project idea")
        fake = FakeToolDetection(tools=["claude", "cursor"])
        handler = BootstrapHandler(tool_detection=fake)

        session = handler.preview(tmp_path)
        assert session.detected_tools == ("claude", "cursor")

    def test_handler_stores_conflict_descriptions_on_preview(self, tmp_path):
        (tmp_path / "README.md").write_text("My project idea")
        fake = FakeToolDetection(conflicts=["Global cursor setting overrides local"])
        handler = BootstrapHandler(tool_detection=fake)

        session = handler.preview(tmp_path)
        preview = session.preview
        assert preview is not None
        assert preview.conflict_descriptions == ("Global cursor setting overrides local",)


class TestBootstrapPlannedFiles:
    """_PLANNED_FILES includes all .alty/ paths."""

    def test_planned_files_includes_alty_config(self, tmp_path: Path) -> None:
        (tmp_path / "README.md").write_text("idea")
        handler = BootstrapHandler(tool_detection=FakeToolDetection())
        session = handler.preview(tmp_path)
        preview = session.preview
        assert preview is not None
        paths = [str(a.path) for a in preview.file_actions]
        assert ".alty/config.toml" in paths

    def test_planned_files_includes_alty_maintenance(self, tmp_path: Path) -> None:
        (tmp_path / "README.md").write_text("idea")
        handler = BootstrapHandler(tool_detection=FakeToolDetection())
        session = handler.preview(tmp_path)
        preview = session.preview
        assert preview is not None
        paths = [str(a.path) for a in preview.file_actions]
        assert ".alty/maintenance/doc-registry.toml" in paths


class TestBootstrapHandlerSessionNotFound:
    def test_confirm_invalid_session_raises(self):
        handler = BootstrapHandler(tool_detection=FakeToolDetection())
        with pytest.raises(SessionNotFoundError, match="no-such-id"):
            handler.confirm("no-such-id")

    def test_cancel_invalid_session_raises(self):
        handler = BootstrapHandler(tool_detection=FakeToolDetection())
        with pytest.raises(SessionNotFoundError, match="no-such-id"):
            handler.cancel("no-such-id")

    def test_execute_invalid_session_raises(self):
        handler = BootstrapHandler(tool_detection=FakeToolDetection())
        with pytest.raises(SessionNotFoundError, match="no-such-id"):
            handler.execute("no-such-id")
