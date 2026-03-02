"""Tests for the alty generate CLI commands.

Verifies that `alty generate <subcommand>` wires to the correct handlers
via composition root, loads sessions from .alty/session.json, and
follows the preview-before-action pattern.
"""

from __future__ import annotations

import json
from typing import TYPE_CHECKING
from unittest.mock import MagicMock, patch

import pytest
from typer.testing import CliRunner

if TYPE_CHECKING:
    from pathlib import Path

from src.infrastructure.cli.main import app

runner = CliRunner()


# ── Helpers ───────────────────────────────────────────────────


def _completed_session_snapshot() -> dict[str, object]:
    """Minimal valid snapshot for a completed discovery session."""
    return {
        "session_id": "test-session-123",
        "readme_content": "Test project README",
        "status": "completed",
        "persona": "developer",
        "register": "technical",
        "answers": [
            {"question_id": "Q1", "response_text": "Users, Admin"},
            {"question_id": "Q2", "response_text": "Orders, Products"},
            {"question_id": "Q3", "response_text": "User creates order"},
            {"question_id": "Q4", "response_text": "Payment must succeed"},
            {"question_id": "Q5", "response_text": "Cancel order"},
            {"question_id": "Q6", "response_text": "OrderCreated"},
            {"question_id": "Q7", "response_text": "Send email on order"},
            {"question_id": "Q8", "response_text": "Order dashboard"},
            {"question_id": "Q9", "response_text": "Orders, Products"},
            {"question_id": "Q10", "response_text": "Orders is core, Products is supporting"},
        ],
        "skipped": [],
        "playback_confirmations": [],
        "answers_since_last_playback": 1,
    }


def _mock_artifact_preview() -> MagicMock:
    preview = MagicMock()
    preview.prd_content = "# PRD\nProduct requirements."
    preview.ddd_content = "# DDD\nDomain model."
    preview.architecture_content = "# Architecture\nTechnical design."
    return preview


def _mock_fitness_preview() -> MagicMock:
    preview = MagicMock()
    preview.summary = "2 import-linter contracts, 4 pytestarch rules"
    return preview


def _mock_ticket_preview() -> MagicMock:
    preview = MagicMock()
    preview.summary = "1 epic, 5 tickets generated"
    return preview


def _mock_config_preview() -> MagicMock:
    preview = MagicMock()
    preview.summary = "Config Generation Preview\n\nclaude-code: 3 sections"
    return preview


# ── Fixtures ──────────────────────────────────────────────────


@pytest.fixture
def session_dir(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> Path:
    """Create tmp dir with valid .alty/session.json and chdir to it."""
    alty_dir = tmp_path / ".alty"
    alty_dir.mkdir()
    (alty_dir / "session.json").write_text(json.dumps(_completed_session_snapshot()))
    monkeypatch.chdir(tmp_path)
    return tmp_path


@pytest.fixture
def no_session_dir(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> Path:
    """Create tmp dir without .alty/session.json and chdir to it."""
    monkeypatch.chdir(tmp_path)
    return tmp_path


# ── Session loading ───────────────────────────────────────────


class TestGenerateSessionLoading:
    """All generate subcommands fail gracefully when no session exists."""

    @pytest.mark.parametrize("subcommand", ["artifacts", "fitness", "tickets", "configs"])
    def test_missing_session_exits_with_error(self, subcommand: str, no_session_dir: Path) -> None:
        result = runner.invoke(app, ["generate", subcommand])
        assert result.exit_code == 1

    @pytest.mark.parametrize("subcommand", ["artifacts", "fitness", "tickets", "configs"])
    def test_missing_session_prints_guidance(self, subcommand: str, no_session_dir: Path) -> None:
        result = runner.invoke(app, ["generate", subcommand])
        assert "guide" in result.output.lower()

    def test_invalid_json_exits_with_error(
        self, tmp_path: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        alty_dir = tmp_path / ".alty"
        alty_dir.mkdir()
        (alty_dir / "session.json").write_text("{invalid json")
        monkeypatch.chdir(tmp_path)
        result = runner.invoke(app, ["generate", "artifacts"])
        assert result.exit_code == 1

    def test_corrupt_snapshot_exits_with_error(
        self, tmp_path: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        alty_dir = tmp_path / ".alty"
        alty_dir.mkdir()
        (alty_dir / "session.json").write_text(json.dumps({"status": "completed"}))
        monkeypatch.chdir(tmp_path)
        result = runner.invoke(app, ["generate", "artifacts"])
        assert result.exit_code == 1

    def test_incomplete_session_exits_with_error(
        self, tmp_path: Path, monkeypatch: pytest.MonkeyPatch
    ) -> None:
        """Session in ANSWERING state (not completed) should be rejected."""
        snapshot = _completed_session_snapshot()
        snapshot["status"] = "answering"
        alty_dir = tmp_path / ".alty"
        alty_dir.mkdir()
        (alty_dir / "session.json").write_text(json.dumps(snapshot))
        monkeypatch.chdir(tmp_path)
        result = runner.invoke(app, ["generate", "artifacts"])
        assert result.exit_code == 1


# ── artifacts ─────────────────────────────────────────────────


class TestGenerateArtifacts:
    """alty generate artifacts wires to ArtifactGenerationHandler."""

    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_calls_create_app(self, mock_create_app, mock_handler_cls, session_dir):
        mock_handler_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        runner.invoke(app, ["generate", "artifacts"], input="y\n")
        mock_create_app.assert_called_once()

    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_builds_handler_with_renderer_and_writer(
        self, mock_create_app, mock_handler_cls, session_dir
    ):
        ctx = mock_create_app.return_value
        mock_handler_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        runner.invoke(app, ["generate", "artifacts"], input="y\n")
        mock_handler_cls.assert_called_once_with(
            renderer=ctx.artifact_renderer,
            writer=ctx.file_writer,
        )

    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_calls_build_preview_with_event(
        self, mock_create_app, mock_handler_cls, session_dir
    ):
        mock_handler_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        runner.invoke(app, ["generate", "artifacts"], input="y\n")
        mock_handler_cls.return_value.build_preview.assert_called_once()
        # Verify the event has session data
        event = mock_handler_cls.return_value.build_preview.call_args[0][0]
        assert event.session_id == "test-session-123"

    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_approval_writes_artifacts(self, mock_create_app, mock_handler_cls, session_dir):
        mock_handler_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        result = runner.invoke(app, ["generate", "artifacts"], input="y\n")
        assert result.exit_code == 0
        mock_handler_cls.return_value.write_artifacts.assert_called_once()

    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_denial_writes_no_files(self, mock_create_app, mock_handler_cls, session_dir):
        mock_handler_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        result = runner.invoke(app, ["generate", "artifacts"], input="n\n")
        assert result.exit_code == 0
        mock_handler_cls.return_value.write_artifacts.assert_not_called()

    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_shows_preview_info(self, mock_create_app, mock_handler_cls, session_dir):
        mock_handler_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        result = runner.invoke(app, ["generate", "artifacts"], input="n\n")
        assert "PRD" in result.output


# ── fitness ───────────────────────────────────────────────────


class TestGenerateFitness:
    """alty generate fitness wires to FitnessGenerationHandler."""

    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_calls_create_app(
        self, mock_create_app, mock_artifact_cls, mock_fitness_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_fitness_cls.return_value.build_preview.return_value = _mock_fitness_preview()
        runner.invoke(app, ["generate", "fitness"], input="y\n")
        mock_create_app.assert_called_once()

    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_builds_handler_with_writer(
        self, mock_create_app, mock_artifact_cls, mock_fitness_cls, session_dir
    ):
        ctx = mock_create_app.return_value
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_fitness_cls.return_value.build_preview.return_value = _mock_fitness_preview()
        runner.invoke(app, ["generate", "fitness"], input="y\n")
        mock_fitness_cls.assert_called_once_with(writer=ctx.file_writer)

    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_approval_calls_approve_and_write(
        self, mock_create_app, mock_artifact_cls, mock_fitness_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_fitness_cls.return_value.build_preview.return_value = _mock_fitness_preview()
        result = runner.invoke(app, ["generate", "fitness"], input="y\n")
        assert result.exit_code == 0
        mock_fitness_cls.return_value.approve_and_write.assert_called_once()

    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_denial_writes_no_files(
        self, mock_create_app, mock_artifact_cls, mock_fitness_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_fitness_cls.return_value.build_preview.return_value = _mock_fitness_preview()
        result = runner.invoke(app, ["generate", "fitness"], input="n\n")
        assert result.exit_code == 0
        mock_fitness_cls.return_value.approve_and_write.assert_not_called()

    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_shows_preview_summary(
        self, mock_create_app, mock_artifact_cls, mock_fitness_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        fitness_preview = _mock_fitness_preview()
        mock_fitness_cls.return_value.build_preview.return_value = fitness_preview
        result = runner.invoke(app, ["generate", "fitness"], input="n\n")
        assert fitness_preview.summary in result.output


# ── tickets ───────────────────────────────────────────────────


class TestGenerateTickets:
    """alty generate tickets wires to TicketGenerationHandler."""

    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_calls_create_app(
        self, mock_create_app, mock_artifact_cls, mock_ticket_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_ticket_cls.return_value.build_preview.return_value = _mock_ticket_preview()
        runner.invoke(app, ["generate", "tickets"], input="y\n")
        mock_create_app.assert_called_once()

    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_builds_handler_with_writer(
        self, mock_create_app, mock_artifact_cls, mock_ticket_cls, session_dir
    ):
        ctx = mock_create_app.return_value
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_ticket_cls.return_value.build_preview.return_value = _mock_ticket_preview()
        runner.invoke(app, ["generate", "tickets"], input="y\n")
        mock_ticket_cls.assert_called_once_with(writer=ctx.file_writer)

    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_approval_calls_approve_and_write(
        self, mock_create_app, mock_artifact_cls, mock_ticket_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_ticket_cls.return_value.build_preview.return_value = _mock_ticket_preview()
        result = runner.invoke(app, ["generate", "tickets"], input="y\n")
        assert result.exit_code == 0
        mock_ticket_cls.return_value.approve_and_write.assert_called_once()

    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_denial_writes_no_files(
        self, mock_create_app, mock_artifact_cls, mock_ticket_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_ticket_cls.return_value.build_preview.return_value = _mock_ticket_preview()
        result = runner.invoke(app, ["generate", "tickets"], input="n\n")
        assert result.exit_code == 0
        mock_ticket_cls.return_value.approve_and_write.assert_not_called()

    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_shows_preview_summary(
        self, mock_create_app, mock_artifact_cls, mock_ticket_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        ticket_preview = _mock_ticket_preview()
        mock_ticket_cls.return_value.build_preview.return_value = ticket_preview
        result = runner.invoke(app, ["generate", "tickets"], input="n\n")
        assert ticket_preview.summary in result.output


# ── configs ───────────────────────────────────────────────────


class TestGenerateConfigs:
    """alty generate configs wires to ConfigGenerationHandler."""

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_calls_create_app(
        self, mock_create_app, mock_artifact_cls, mock_config_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_config_cls.return_value.build_preview.return_value = _mock_config_preview()
        runner.invoke(app, ["generate", "configs"], input="y\n")
        mock_create_app.assert_called_once()

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_builds_handler_with_writer(
        self, mock_create_app, mock_artifact_cls, mock_config_cls, session_dir
    ):
        ctx = mock_create_app.return_value
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_config_cls.return_value.build_preview.return_value = _mock_config_preview()
        runner.invoke(app, ["generate", "configs"], input="y\n")
        mock_config_cls.assert_called_once_with(writer=ctx.file_writer)

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_approval_calls_approve_and_write(
        self, mock_create_app, mock_artifact_cls, mock_config_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_config_cls.return_value.build_preview.return_value = _mock_config_preview()
        result = runner.invoke(app, ["generate", "configs"], input="y\n")
        assert result.exit_code == 0
        mock_config_cls.return_value.approve_and_write.assert_called_once()

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_denial_writes_no_files(
        self, mock_create_app, mock_artifact_cls, mock_config_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        mock_config_cls.return_value.build_preview.return_value = _mock_config_preview()
        result = runner.invoke(app, ["generate", "configs"], input="n\n")
        assert result.exit_code == 0
        mock_config_cls.return_value.approve_and_write.assert_not_called()

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_shows_preview_summary(
        self, mock_create_app, mock_artifact_cls, mock_config_cls, session_dir
    ):
        mock_artifact_cls.return_value.build_preview.return_value = _mock_artifact_preview()
        config_preview = _mock_config_preview()
        mock_config_cls.return_value.build_preview.return_value = config_preview
        result = runner.invoke(app, ["generate", "configs"], input="n\n")
        assert config_preview.summary in result.output
