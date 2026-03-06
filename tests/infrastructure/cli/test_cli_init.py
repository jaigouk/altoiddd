"""Tests for the alty init CLI command (new project flow — else branch).

Verifies that `alty init` (without --existing) wires to the composition root,
runs interactive discovery, and orchestrates the full bootstrap pipeline:
artifacts → fitness → tickets → configs.

The --existing branch is tested elsewhere; these tests only cover the else branch.
"""

from __future__ import annotations

import json
from unittest.mock import MagicMock, patch

from typer.testing import CliRunner

from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.discovery_values import Answer, Persona, Register
from src.infrastructure.cli.main import app

runner = CliRunner()


# ── Helpers ─────────────────────────────────────────────────────


def _make_session(
    *,
    status: DiscoveryStatus = DiscoveryStatus.CREATED,
    register: Register = Register.TECHNICAL,
    persona: Persona = Persona.DEVELOPER,
    answers: tuple[Answer, ...] = (),
) -> MagicMock:
    """Create a mock DiscoverySession with sensible defaults."""
    session = MagicMock(spec=DiscoverySession)
    session.session_id = "init-test-session"
    session.status = status
    session.register = register
    session.persona = persona
    session.readme_content = "# My Project\nA test project."
    session.answers = answers
    session.playback_confirmations = ()
    session.to_snapshot.return_value = {
        "session_id": "init-test-session",
        "readme_content": "# My Project\nA test project.",
        "status": status.value,
        "persona": persona.value,
        "register": register.value,
        "answers": [
            {"question_id": a.question_id, "response_text": a.response_text}
            for a in answers
        ],
        "skipped": [],
        "playback_confirmations": [],
        "answers_since_last_playback": 0,
    }
    return session


def _setup_discovery_mocks(mock_ctx: MagicMock) -> None:
    """Configure discovery port mocks for a full happy-path discovery run."""
    mock_ctx.discovery.start_session.return_value = _make_session()
    mock_ctx.discovery.set_tech_stack.return_value = _make_session()
    mock_ctx.discovery.get_session.return_value = _make_session()
    mock_ctx.discovery.detect_persona.return_value = _make_session(
        status=DiscoveryStatus.PERSONA_DETECTED
    )

    answering = _make_session(status=DiscoveryStatus.ANSWERING)
    playback = _make_session(status=DiscoveryStatus.PLAYBACK_PENDING)
    completed = _make_session(status=DiscoveryStatus.COMPLETED)

    mock_ctx.discovery.answer_question.side_effect = [
        answering, answering, playback,   # Q1-Q3 → playback
        answering, answering, playback,   # Q4-Q6 → playback
        answering, answering, playback,   # Q7-Q9 → playback
        answering,                        # Q10
    ]
    mock_ctx.discovery.confirm_playback.return_value = answering
    mock_ctx.discovery.complete.return_value = completed


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
    preview.summary = "Config Generation Preview\nclaude-code: 3 sections"
    return preview


def _discovery_input() -> str:
    """Build stdin for full discovery: tech stack + mode + persona + 10 answers + 3 playbacks."""
    return (
        "y\n"                   # tech stack: Python
        "1\n"                   # mode: Express
        "1\n"                   # persona
        "answer1\n"             # Q1
        "answer2\n"             # Q2
        "answer3\n"             # Q3
        "y\n"                   # playback 1
        "answer4\n"             # Q4
        "answer5\n"             # Q5
        "answer6\n"             # Q6
        "y\n"                   # playback 2
        "answer7\n"             # Q7
        "answer8\n"             # Q8
        "answer9\n"             # Q9
        "y\n"                   # playback 3
        "answer10\n"            # Q10
    )


def _full_pipeline_input() -> str:
    """Build stdin for full pipeline: discovery + 4 approvals."""
    return _discovery_input() + "y\ny\ny\ny\n"


def _setup_all_handler_mocks(
    mock_artifact_cls: MagicMock,
    mock_fitness_cls: MagicMock,
    mock_ticket_cls: MagicMock,
    mock_config_cls: MagicMock,
) -> tuple[MagicMock, MagicMock, MagicMock, MagicMock]:
    """Configure all four handler class mocks with preview returns."""
    artifact_preview = _mock_artifact_preview()
    mock_artifact_cls.return_value.build_preview.return_value = artifact_preview

    fitness_preview = _mock_fitness_preview()
    mock_fitness_cls.return_value.build_preview.return_value = fitness_preview

    ticket_preview = _mock_ticket_preview()
    mock_ticket_cls.return_value.build_preview.return_value = ticket_preview

    config_preview = _mock_config_preview()
    mock_config_cls.return_value.build_preview.return_value = config_preview

    return artifact_preview, fitness_preview, ticket_preview, config_preview


# ── Wiring ──────────────────────────────────────────────────────


class TestInitWiring:
    """alty init (new project) uses composition root."""

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_calls_create_app(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# My Project")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        mock_create_app.assert_called_once()

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_uses_discovery_port(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# My Project")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        mock_ctx.discovery.start_session.assert_called_once_with("# My Project")


# ── README Requirement ──────────────────────────────────────────


class TestInitRequiresReadme:
    """alty init requires README.md in the current directory."""

    @patch("src.infrastructure.composition.create_app")
    def test_missing_readme_exits_with_error(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        result = runner.invoke(app, ["init"])
        assert result.exit_code == 1

    @patch("src.infrastructure.composition.create_app")
    def test_missing_readme_shows_error_message(self, mock_create_app, tmp_path, monkeypatch):
        monkeypatch.chdir(tmp_path)
        result = runner.invoke(app, ["init"])
        assert "readme" in result.output.lower()


# ── Discovery Flow ──────────────────────────────────────────────


class TestInitDiscoveryFlow:
    """alty init runs the full interactive discovery flow."""

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_calls_detect_persona(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        mock_ctx.discovery.detect_persona.assert_called_once_with(
            "init-test-session", "1"
        )

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_answers_all_ten_questions(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        assert mock_ctx.discovery.answer_question.call_count == 10

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_confirms_three_playbacks(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        assert mock_ctx.discovery.confirm_playback.call_count == 3

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_calls_complete(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        mock_ctx.discovery.complete.assert_called_once_with("init-test-session")


# ── Artifact Generation ─────────────────────────────────────────


class TestInitArtifactGeneration:
    """alty init invokes all four generation handlers."""

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_builds_artifact_handler_with_ports(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        mock_artifact_cls.assert_called_once_with(
            renderer=mock_ctx.artifact_renderer,
            writer=mock_ctx.file_writer,
        )

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_builds_fitness_handler_with_writer(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        mock_fitness_cls.assert_called_once_with(writer=mock_ctx.file_writer)

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_builds_ticket_handler_with_writer(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        mock_ticket_cls.assert_called_once_with(writer=mock_ctx.file_writer)

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_builds_config_handler_with_writer(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        mock_config_cls.assert_called_once_with(writer=mock_ctx.file_writer)

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_approval_writes_all_artifacts(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        result = runner.invoke(app, ["init"], input=_full_pipeline_input())
        assert result.exit_code == 0
        mock_artifact_cls.return_value.write_artifacts.assert_called_once()
        mock_fitness_cls.return_value.approve_and_write.assert_called_once()
        mock_ticket_cls.return_value.approve_and_write.assert_called_once()
        mock_config_cls.return_value.approve_and_write.assert_called_once()


# ── Cancel Behavior ─────────────────────────────────────────────


class TestInitCancelBehavior:
    """Cancelling at any stage writes no files."""

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_cancel_at_artifacts_writes_nothing(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        # Deny at artifact stage
        cancel_input = _discovery_input() + "n\n"
        result = runner.invoke(app, ["init"], input=cancel_input)
        assert result.exit_code == 0
        mock_artifact_cls.return_value.write_artifacts.assert_not_called()
        mock_fitness_cls.return_value.approve_and_write.assert_not_called()
        mock_ticket_cls.return_value.approve_and_write.assert_not_called()
        mock_config_cls.return_value.approve_and_write.assert_not_called()

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_cancel_at_fitness_writes_nothing(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        # Approve artifacts, deny fitness
        cancel_input = _discovery_input() + "y\nn\n"
        result = runner.invoke(app, ["init"], input=cancel_input)
        assert result.exit_code == 0
        # Artifacts were approved but fitness wasn't — whole pipeline should abort
        mock_fitness_cls.return_value.approve_and_write.assert_not_called()
        mock_ticket_cls.return_value.approve_and_write.assert_not_called()
        mock_config_cls.return_value.approve_and_write.assert_not_called()

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_cancel_at_tickets_writes_nothing(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        # Approve artifacts + fitness, deny tickets
        cancel_input = _discovery_input() + "y\ny\nn\n"
        result = runner.invoke(app, ["init"], input=cancel_input)
        assert result.exit_code == 0
        mock_ticket_cls.return_value.approve_and_write.assert_not_called()
        mock_config_cls.return_value.approve_and_write.assert_not_called()

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_cancel_at_configs_writes_nothing(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        # Approve artifacts + fitness + tickets, deny configs
        cancel_input = _discovery_input() + "y\ny\ny\nn\n"
        result = runner.invoke(app, ["init"], input=cancel_input)
        assert result.exit_code == 0
        mock_config_cls.return_value.approve_and_write.assert_not_called()


# ── Session Save ────────────────────────────────────────────────


class TestInitSessionSave:
    """alty init saves session snapshot to .alty/session.json."""

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_saves_session_snapshot(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        result = runner.invoke(app, ["init"], input=_full_pipeline_input())
        assert result.exit_code == 0
        session_file = tmp_path / ".alty" / "session.json"
        assert session_file.exists()
        data = json.loads(session_file.read_text())
        assert data["session_id"] == "init-test-session"

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_creates_alty_dir_if_missing(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        assert not (tmp_path / ".alty").exists()
        runner.invoke(app, ["init"], input=_full_pipeline_input())
        assert (tmp_path / ".alty").is_dir()


# ── Full Pipeline ───────────────────────────────────────────────


class TestInitFullPipeline:
    """End-to-end happy path for alty init."""

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_happy_path_exit_code_zero(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        result = runner.invoke(app, ["init"], input=_full_pipeline_input())
        assert result.exit_code == 0

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_shows_completion_message(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        result = runner.invoke(app, ["init"], input=_full_pipeline_input())
        assert "complete" in result.output.lower() or "bootstrap" in result.output.lower()

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_shows_artifact_preview(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        _setup_discovery_mocks(mock_ctx)
        _setup_all_handler_mocks(
            mock_artifact_cls, mock_fitness_cls, mock_ticket_cls, mock_config_cls
        )
        result = runner.invoke(app, ["init"], input=_full_pipeline_input())
        assert "PRD" in result.output


# ── Error Handling ──────────────────────────────────────────────


class TestInitErrorHandling:
    """alty init handles errors gracefully."""

    @patch("src.application.commands.config_generation_handler.ConfigGenerationHandler")
    @patch("src.application.commands.ticket_generation_handler.TicketGenerationHandler")
    @patch("src.application.commands.fitness_generation_handler.FitnessGenerationHandler")
    @patch("src.application.commands.artifact_generation_handler.ArtifactGenerationHandler")
    @patch("src.infrastructure.composition.create_app")
    def test_discovery_error_exits_with_error(
        self,
        mock_create_app,
        mock_artifact_cls,
        mock_fitness_cls,
        mock_ticket_cls,
        mock_config_cls,
        tmp_path,
        monkeypatch,
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = MagicMock()
        mock_create_app.return_value = mock_ctx
        mock_ctx.discovery.start_session.return_value = _make_session()
        mock_ctx.discovery.set_tech_stack.return_value = _make_session()
        mock_ctx.discovery.detect_persona.side_effect = ValueError("Invalid choice '9'")
        result = runner.invoke(app, ["init"], input="y\n9\n")
        assert result.exit_code == 1
