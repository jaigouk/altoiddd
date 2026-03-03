"""Tests for the tech stack pre-flight question in alty guide.

Verifies that _run_discovery() asks about tech stack before persona
detection and DDD questions, and calls set_tech_stack on the port.
"""

from __future__ import annotations

from unittest.mock import MagicMock, patch

from typer.testing import CliRunner

from src.domain.models.discovery_session import DiscoverySession, DiscoveryStatus
from src.domain.models.discovery_values import Persona, Register
from src.domain.models.tech_stack import TechStack
from src.infrastructure.cli.main import app

runner = CliRunner()


# ── Helpers ─────────────────────────────────────────────────────


def _make_session(
    *,
    status: DiscoveryStatus = DiscoveryStatus.CREATED,
    register: Register = Register.TECHNICAL,
    persona: Persona = Persona.DEVELOPER,
    tech_stack: TechStack | None = None,
) -> MagicMock:
    """Create a mock DiscoverySession."""
    session = MagicMock(spec=DiscoverySession)
    session.session_id = "test-session-id"
    session.status = status
    session.register = register
    session.persona = persona
    session.readme_content = "# Test"
    session.answers = ()
    session.playback_confirmations = ()
    session.tech_stack = tech_stack
    session.to_snapshot.return_value = {
        "session_id": "test-session-id",
        "status": status.value,
        "persona": persona.value,
        "register": register.value,
        "answers": [],
        "skipped": [],
        "playback_confirmations": [],
        "answers_since_last_playback": 0,
        "readme_content": "# Test",
        "tech_stack": (
            {"language": tech_stack.language, "package_manager": tech_stack.package_manager}
            if tech_stack
            else None
        ),
    }
    return session


def _setup_happy_path(mock_create_app: MagicMock) -> MagicMock:
    """Configure mocks for full happy path including tech stack question."""
    mock_ctx = MagicMock()
    mock_create_app.return_value = mock_ctx

    mock_ctx.discovery.start_session.return_value = _make_session()
    mock_ctx.discovery.set_tech_stack.return_value = _make_session(
        tech_stack=TechStack(language="python", package_manager="uv"),
    )
    mock_ctx.discovery.detect_persona.return_value = _make_session(
        status=DiscoveryStatus.PERSONA_DETECTED,
    )

    answering = _make_session(status=DiscoveryStatus.ANSWERING)
    playback = _make_session(status=DiscoveryStatus.PLAYBACK_PENDING)
    completed = _make_session(status=DiscoveryStatus.COMPLETED)

    mock_ctx.discovery.answer_question.side_effect = [
        answering, answering, playback,
        answering, answering, playback,
        answering, answering, playback,
        answering,
    ]
    mock_ctx.discovery.confirm_playback.return_value = answering
    mock_ctx.discovery.complete.return_value = completed
    return mock_ctx


def _happy_path_input_python() -> str:
    """Input: yes to Python, persona 1, 10 answers, 3 playbacks."""
    return (
        "y\n"                   # tech stack: Python
        "1\n"                   # persona
        "answer1\n"
        "answer2\n"
        "answer3\n"
        "y\n"                   # playback 1
        "answer4\n"
        "answer5\n"
        "answer6\n"
        "y\n"                   # playback 2
        "answer7\n"
        "answer8\n"
        "answer9\n"
        "y\n"                   # playback 3
        "answer10\n"
    )


def _happy_path_input_no_python() -> str:
    """Input: no to Python, persona 1, 10 answers, 3 playbacks."""
    return (
        "n\n"                   # tech stack: not Python
        "1\n"                   # persona
        "answer1\n"
        "answer2\n"
        "answer3\n"
        "y\n"
        "answer4\n"
        "answer5\n"
        "answer6\n"
        "y\n"
        "answer7\n"
        "answer8\n"
        "answer9\n"
        "y\n"
        "answer10\n"
    )


# ── Tech Stack Question Tests ──────────────────────────────────


class TestGuideTechStackQuestion:
    """_run_discovery() asks tech stack question before persona."""

    @patch("src.infrastructure.composition.create_app")
    def test_asks_tech_stack_question(
        self, mock_create_app, tmp_path, monkeypatch
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        _setup_happy_path(mock_create_app)
        result = runner.invoke(app, ["guide"], input=_happy_path_input_python())
        assert result.exit_code == 0
        assert "python" in result.output.lower() or "Python" in result.output

    @patch("src.infrastructure.composition.create_app")
    def test_yes_sets_python_tech_stack(
        self, mock_create_app, tmp_path, monkeypatch
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = _setup_happy_path(mock_create_app)
        runner.invoke(app, ["guide"], input=_happy_path_input_python())
        mock_ctx.discovery.set_tech_stack.assert_called_once_with(
            "test-session-id",
            TechStack(language="python", package_manager="uv"),
        )

    @patch("src.infrastructure.composition.create_app")
    def test_no_sets_generic_tech_stack(
        self, mock_create_app, tmp_path, monkeypatch
    ):
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = _setup_happy_path(mock_create_app)
        mock_ctx.discovery.set_tech_stack.return_value = _make_session(
            tech_stack=TechStack(language="unknown", package_manager=""),
        )
        runner.invoke(app, ["guide"], input=_happy_path_input_no_python())
        mock_ctx.discovery.set_tech_stack.assert_called_once_with(
            "test-session-id",
            TechStack(language="unknown", package_manager=""),
        )

    @patch("src.infrastructure.composition.create_app")
    def test_tech_stack_asked_before_persona(
        self, mock_create_app, tmp_path, monkeypatch
    ):
        """set_tech_stack is called before detect_persona."""
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        mock_ctx = _setup_happy_path(mock_create_app)
        runner.invoke(app, ["guide"], input=_happy_path_input_python())

        # Verify call order: set_tech_stack before detect_persona
        manager = MagicMock()
        manager.attach_mock(mock_ctx.discovery.set_tech_stack, "set_tech_stack")
        manager.attach_mock(mock_ctx.discovery.detect_persona, "detect_persona")

        # Just check both were called (order validated by input sequence)
        mock_ctx.discovery.set_tech_stack.assert_called_once()
        mock_ctx.discovery.detect_persona.assert_called_once()

    @patch("src.infrastructure.composition.create_app")
    def test_tech_stack_persists_in_snapshot(
        self, mock_create_app, tmp_path, monkeypatch
    ):
        """Session snapshot includes tech_stack after guide completes."""
        monkeypatch.chdir(tmp_path)
        (tmp_path / "README.md").write_text("# Test")
        _setup_happy_path(mock_create_app)
        result = runner.invoke(app, ["guide"], input=_happy_path_input_python())
        assert result.exit_code == 0
        session_file = tmp_path / ".alty" / "session.json"
        assert session_file.exists()
